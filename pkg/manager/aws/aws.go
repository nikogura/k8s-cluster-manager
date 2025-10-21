package aws

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/nikogura/k8s-cluster-manager/pkg/manager"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"math/rand"
	"os"
	"regexp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"strconv"
)

//nolint:gochecknoinits // Package-level initialization required
func init() {
	logrus.SetFormatter(&logrus.JSONFormatter{})

	logrus.SetOutput(os.Stdout)

	logrus.SetLevel(logrus.DebugLevel)
}

//type AWSClusterManager struct {
//	Name                       string
//	cloudProviderName          string
//	k8sProviderName            string
//	scheduleWorkloadsOnCPNodes bool
//	Domain                     string
//	DnsManager                 manager.DNSManager
//	GetVerbose                    bool
//	Config                     aws.Config
//	//TODO: delete the literal ec2 client after mocks are complete
//	Ec2ClientLiteral   *ec2.Client
//	Ec2Client          Ec2Client
//	ELBClient          ELBClient
//	Context            context.Context
//	Profile            string
//	KubeClient         client.Client
//	FetchedNodesById   map[string]manager.NodeInfo
//	FetchedNodesByName map[string]manager.NodeInfo
//	ClusterNameRegex   *regexp.Regexp
//}

type AWSClusterManager struct {
	Name                       string
	CloudProviderName          string
	K8sProviderName            string
	ScheduleWorkloadsOnCPNodes bool
	Domain                     string
	DnsManager                 manager.DNSManager //nolint:staticcheck // Changing to DNSManager would break API
	Verbose                    bool
	Config                     aws.Config
	//TODO: delete the literal ec2 client after mocks are complete
	//Ec2ClientLiteral   *ec2.Client
	Ec2Client          Ec2Client
	ELBClient          ELBClient
	Context            context.Context
	Profile            string
	KubeClient         client.Client
	FetchedNodesById   map[string]manager.NodeInfo //nolint:staticcheck // Changing to FetchedNodesByID would break API
	FetchedNodesByName map[string]manager.NodeInfo
	ClusterNameRegex   *regexp.Regexp
	CostEstimator      manager.CostEstimator // Optional: if provided, enables cost estimation
}

func NewAWSClusterManager(ctx context.Context, clusterName string, profile string, role string, dnsManager manager.DNSManager, verbose bool) (am *AWSClusterManager, err error) {
	_ = log.FromContext(ctx)
	var cfg aws.Config

	// if we are supplied a profile, use it to set up the aws config
	if profile != "" {
		cfg, err = config.LoadDefaultConfig(ctx, config.WithSharedConfigProfile(profile))
		if err != nil {
			err = errors.Wrapf(err, "failed creating aws config with shared profile")
			return am, err
		}
	} else { // if not, use the defaults
		cfg, err = config.LoadDefaultConfig(ctx)
		if err != nil {
			err = errors.Wrapf(err, "failed creating aws config")
			return am, err
		}
	}

	// If we have a role to assume, assume it
	if role != "" {
		sourceAccount := sts.NewFromConfig(cfg)

		assumeRoleInput := &sts.AssumeRoleInput{
			RoleArn:         aws.String(role),
			RoleSessionName: aws.String("k8s-cluster-manager-" + strconv.Itoa(10000+rand.Intn(25000))),
		}

		// Assume the role
		stsResp, stsErr := sourceAccount.AssumeRole(ctx, assumeRoleInput)
		if stsErr != nil {
			err = errors.Wrapf(stsErr, "failed assuming role %s", role)
			return am, err
		}

		// pull the creds out of the role assumption response, and use that to make a new config
		cfgTemp, cfgErr := config.LoadDefaultConfig(ctx, config.WithRegion(os.Getenv("AWS_REGION")), config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(*stsResp.Credentials.AccessKeyId, *stsResp.Credentials.SecretAccessKey, *stsResp.Credentials.SessionToken)))
		if cfgErr != nil {
			err = errors.Wrapf(cfgErr, "failed assuming role %s", role)
			return am, err
		}
		cfg = cfgTemp
	}

	// Create AWS clients
	ec2Client := ec2.NewFromConfig(cfg)
	elbClient := elasticloadbalancingv2.NewFromConfig(cfg)

	// Create k8s clients
	kubeClient, kubeErr := client.New(ctrl.GetConfigOrDie(), client.Options{})
	if kubeErr != nil {
		err = errors.Wrapf(kubeErr, "failed creating k8s clients")
		return am, err
	}

	re, reErr := regexp.Compile(fmt.Sprintf(".*%s.*", clusterName))
	if reErr != nil {
		err = errors.Wrapf(reErr, "cluster name %s doesn't compile into a regex", clusterName)
		return am, err
	}

	am = &AWSClusterManager{
		Name:               clusterName,
		CloudProviderName:  "aws",
		K8sProviderName:    "talos",
		DnsManager:         dnsManager,
		Verbose:            verbose,
		Config:             cfg,
		Ec2Client:          ec2Client,
		ELBClient:          elbClient,
		Context:            ctx,
		Profile:            profile,
		KubeClient:         kubeClient,
		FetchedNodesById:   make(map[string]manager.NodeInfo, 0),
		FetchedNodesByName: make(map[string]manager.NodeInfo, 0),
		ClusterNameRegex:   re,
	}

	return am, err
}

func (am *AWSClusterManager) ClusterName() (name string) {
	name = am.Name
	return name
}

//	func (am *AWSClusterManager) CloudProviderName() string {
//		return am.CloudProviderName
//	}
//
//	func (am *AWSClusterManager) K8sProviderName() string {
//		return am.K8sProviderName
//	}
//
//nolint:staticcheck // Comment format is intentional for consistency
func (am *AWSClusterManager) GetScheduleWorkloadsOnCPNodes() (result bool) {
	result = am.ScheduleWorkloadsOnCPNodes
	return result
}
func (am *AWSClusterManager) GetVerbose() (result bool) {
	result = am.Verbose
	return result
}

// SetCostEstimator sets the cost estimator for the cluster manager.
func (am *AWSClusterManager) SetCostEstimator(estimator manager.CostEstimator) {
	am.CostEstimator = estimator
}

//
//func (am *AWSClusterManager) DNSManager() manager.DNSManager {
//	return am.DnsManager
//}

func (am *AWSClusterManager) DescribeCluster(clusterName string) (info manager.ClusterInfo, err error) {

	// construct the ClusterInfo struct
	info.Provider = "aws"
	info.Name = clusterName

	// Get the nodes for the cluster
	nodes, nodesErr := am.GetNodes(clusterName)
	if nodesErr != nil {
		err = errors.Wrapf(nodesErr, "failed getting cluster nodes")
		return info, err
	}

	info.Nodes = nodes

	// Get the load balancers for the cluster
	lbs, lbsErr := am.GetClusterLBs()
	if lbsErr != nil {
		err = errors.Wrapf(lbsErr, "failed getting loadbalancers for cluster %s", clusterName)
		return info, err
	}

	info.LoadBalancers = lbs

	// Calculate cost if estimator is available
	if am.CostEstimator != nil {
		totalCost, costErr := CalculateClusterDailyCost(info.Nodes, am.CostEstimator)
		if costErr != nil {
			// Log warning but don't fail the entire describe operation
			manager.VerboseOutput(am.Verbose, "Warning: failed to calculate cluster cost: %v", costErr)
		} else {
			info.EstimatedDailyCost = &totalCost
		}
	}

	return info, err
}

func NodeName(clusterName string, nodeType string, index int) (nodeName string, err error) {
	switch nodeType {
	case "cp":
		nodeName = fmt.Sprintf("%s-cp-%d", clusterName, index)
	case "worker":
		nodeName = fmt.Sprintf("%s-worker-%d", clusterName, index)
	default:
		err = errors.New(fmt.Sprintf("unkonwn node type %s", nodeType))
		return nodeName, err

	}

	return nodeName, err
}

func TargetGroupName(clusterName string, tls bool) (tgName string) {
	if tls {
		tgName = fmt.Sprintf("ingress-%s-tls", clusterName)
	} else {
		tgName = fmt.Sprintf("ingress-%s-clear", clusterName)
	}

	return tgName
}

func LoadBalancerName(clusterName string, lbType string) (lbName string, err error) {
	switch lbType {
	case "apiserver":
		lbName = fmt.Sprintf("apiserver-%s", clusterName)
	case "int":
		lbName = fmt.Sprintf("ingress-%s", clusterName)
	case "ext":
		lbName = fmt.Sprintf("ingress-%s-ext", clusterName)
	default:
		err = errors.New(fmt.Sprintf("unkonwn lb type %s", lbType))
		return lbName, err
	}

	return lbName, err
}
