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

func init() {
	logrus.SetFormatter(&logrus.JSONFormatter{})

	logrus.SetOutput(os.Stdout)

	logrus.SetLevel(logrus.DebugLevel)
}

type AWSClusterManager struct {
	clusterName                string
	cloudProviderName          string
	k8sProviderName            string
	scheduleWorkloadsOnCPNodes bool
	domain                     string
	dnsManager                 manager.DNSManager
	verbose                    bool
	Config                     aws.Config
	Ec2Client                  *ec2.Client
	ELBClient                  *elasticloadbalancingv2.Client
	Context                    context.Context
	Profile                    string
	KubeClient                 client.Client
	FetchedNodesById           map[string]manager.NodeInfo
	FetchedNodesByName         map[string]manager.NodeInfo
	ClusterNameRegex           *regexp.Regexp
}

func NewAWSClusterManager(ctx context.Context, clusterName string, profile string, role string, dnsManager manager.DNSManager, verbose bool) (am *AWSClusterManager, err error) {
	_ = log.FromContext(ctx)
	var cfg aws.Config

	// if we are supplied a profile, use it to set up the aws config
	if profile != "" {
		cfg, err = config.LoadDefaultConfig(ctx, config.WithSharedConfigProfile(profile))
		if err != nil {
			err = errors.Wrapf(err, "failed creating aws config")
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
		stsResp, err := sourceAccount.AssumeRole(ctx, assumeRoleInput)
		if err != nil {
			err = errors.Wrapf(err, "failed assuming role %s", role)
			return am, err
		}

		// pull the creds out of the role assumption response, and use that to make a new config
		cfg, err = config.LoadDefaultConfig(ctx, config.WithRegion(os.Getenv("AWS_REGION")), config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(*stsResp.Credentials.AccessKeyId, *stsResp.Credentials.SecretAccessKey, *stsResp.Credentials.SessionToken)))
		if err != nil {
			err = errors.Wrapf(err, "failed assuming role %s", role)
			return am, err
		}
	}

	// Create AWS clients
	ec2Client := ec2.NewFromConfig(cfg)
	elbClient := elasticloadbalancingv2.NewFromConfig(cfg)

	// Create k8s clients
	kubeClient, err := client.New(ctrl.GetConfigOrDie(), client.Options{})
	if err != nil {
		err = errors.Wrapf(err, "failed creating k8s clients")
		return am, err
	}

	re, err := regexp.Compile(fmt.Sprintf(".*%s.*", clusterName))
	if err != nil {
		err = errors.Wrapf(err, "cluster name %s doesn't compile into a regex", clusterName)
		return am, err
	}

	am = &AWSClusterManager{
		clusterName:        clusterName,
		cloudProviderName:  "aws",
		k8sProviderName:    "talos",
		dnsManager:         dnsManager,
		verbose:            verbose,
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

func (am *AWSClusterManager) ClusterName() string {
	return am.clusterName
}

func (am *AWSClusterManager) CloudProviderName() string {
	return am.cloudProviderName
}

func (am *AWSClusterManager) K8sProviderName() string {
	return am.k8sProviderName
}

func (am *AWSClusterManager) ScheduleWorkloadsOnCPNodes() bool {
	return am.scheduleWorkloadsOnCPNodes
}

func (am *AWSClusterManager) Verbose() bool {
	return am.verbose
}

func (am *AWSClusterManager) DNSManager() manager.DNSManager {
	return am.dnsManager
}

func (am *AWSClusterManager) DescribeCluster(clusterName string) (info manager.ClusterInfo, err error) {

	// construct the ClusterInfo struct
	info.Provider = "aws"
	info.Name = clusterName

	// Get the nodes for the cluster
	nodes, err := am.GetNodes(clusterName)
	if err != nil {
		err = errors.Wrapf(err, "failed getting cluster nodes")
		return info, err
	}

	info.Nodes = nodes

	// Get the load balancers for the cluster
	lbs, err := am.GetClusterLBs()
	if err != nil {
		err = errors.Wrapf(err, "failed getting loadbalancers for cluster %s", clusterName)
		return info, err
	}

	info.LoadBalancers = lbs

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
