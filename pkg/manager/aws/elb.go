package aws

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"github.com/sirupsen/logrus"
	//"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types".
	"context"
	"github.com/nikogura/k8s-cluster-manager/pkg/manager"
	"github.com/pkg/errors"
	"regexp"
	"sort"
)

const APIServerPort = 6443
const CleartextIngressPortInt = 30080
const TLSIngressPortInt = 30443
const CleartextIngressPortExt = 31080
const TLSIngressPortExt = 31443
const ELBClusterTag = "Cluster"

type ELBClient interface {
	DescribeLoadBalancers(ctx context.Context, params *elasticloadbalancingv2.DescribeLoadBalancersInput, optFns ...func(*elasticloadbalancingv2.Options)) (*elasticloadbalancingv2.DescribeLoadBalancersOutput, error)
	DescribeTags(ctx context.Context, params *elasticloadbalancingv2.DescribeTagsInput, optFns ...func(*elasticloadbalancingv2.Options)) (*elasticloadbalancingv2.DescribeTagsOutput, error)
	DescribeTargetGroups(ctx context.Context, params *elasticloadbalancingv2.DescribeTargetGroupsInput, optFns ...func(*elasticloadbalancingv2.Options)) (*elasticloadbalancingv2.DescribeTargetGroupsOutput, error)
	DescribeTargetHealth(ctx context.Context, params *elasticloadbalancingv2.DescribeTargetHealthInput, optFns ...func(*elasticloadbalancingv2.Options)) (*elasticloadbalancingv2.DescribeTargetHealthOutput, error)
	RegisterTargets(ctx context.Context, params *elasticloadbalancingv2.RegisterTargetsInput, optFns ...func(*elasticloadbalancingv2.Options)) (*elasticloadbalancingv2.RegisterTargetsOutput, error)
	DeregisterTargets(ctx context.Context, params *elasticloadbalancingv2.DeregisterTargetsInput, optFns ...func(*elasticloadbalancingv2.Options)) (*elasticloadbalancingv2.DeregisterTargetsOutput, error)
}

func (am *AWSClusterManager) GetLB(lbName string) (lbOutput *elasticloadbalancingv2.DescribeLoadBalancersOutput, err error) {

	// DescribeLoadBalancers gives all by default, or filters by name or arn
	input := &elasticloadbalancingv2.DescribeLoadBalancersInput{}

	// Won't accept an empty string though.
	if lbName != "" {
		input.Names = []string{lbName}
	}

	lbOutput, err = am.ELBClient.DescribeLoadBalancers(am.Context, input)
	if err != nil {
		err = errors.Wrapf(err, "failed getting lb %s", lbName)
		return lbOutput, err
	}

	return lbOutput, err
}

func (am *AWSClusterManager) GetClusterLBs() (lbs []manager.LBInfo, err error) {
	/*
		There is no way to filter LoadBalancers by tag
		aws elbv2 describe-load-balancers | jq -r '.LoadBalancers[].LoadBalancerArn' | xargs -I {} aws elbv2 describe-tags --resource-arns {} --query "TagDescriptions[?Tags[?Key=='env' &&Value=='dev'] && Tags[?Key=='created_by' &&Value=='xyz']].ResourceArn" --output text

	*/

	// DescribeLoadBalancers gives all by default, or filters by name or arn
	input := &elasticloadbalancingv2.DescribeLoadBalancersInput{}

	// Get all the load balancers
	output, descErr := am.ELBClient.DescribeLoadBalancers(am.Context, input)
	if descErr != nil {
		err = errors.Wrapf(descErr, "failed getting lbs for cluster %s", am.ClusterName())
		return lbs, err
	}

	// Iterate through the list of load balancers, cos you can't filter by tag
	for _, lb := range output.LoadBalancers {
		lbInfo, shouldInclude, checkErr := am.checkAndBuildLBInfo(lb)
		if checkErr != nil {
			err = checkErr
			return lbs, err
		}
		if shouldInclude {
			lbs = append(lbs, lbInfo)
		}
	}

	return lbs, err
}

func (am *AWSClusterManager) checkAndBuildLBInfo(lb types.LoadBalancer) (lbInfo manager.LBInfo, shouldInclude bool, err error) {
	// Find the ARN
	arn := *lb.LoadBalancerArn

	// Look for the tags by ARN
	tagInput := &elasticloadbalancingv2.DescribeTagsInput{
		ResourceArns: []string{arn},
	}

	// Describe the tags
	tagOutput, tagErr := am.ELBClient.DescribeTags(am.Context, tagInput)
	if tagErr != nil {
		err = errors.Wrapf(tagErr, "failed fetching tags")
		shouldInclude = false
		return lbInfo, shouldInclude, err
	}

	// Check if this LB belongs to our cluster
	belongsToCluster := am.checkLBBelongsToCluster(tagOutput)
	if !belongsToCluster {
		shouldInclude = false
		return lbInfo, shouldInclude, err
	}

	// Build LB info
	apiserverRegex := regexp.MustCompile(`.*apiserver.*`)
	lbInfo = manager.LBInfo{
		Name:         *lb.LoadBalancerName,
		Targets:      make([]manager.LBTargetInfo, 0),
		TargetGroups: make([]manager.LBTargetGroupInfo, 0),
		IsAPIServer:  apiserverRegex.MatchString(*lb.LoadBalancerName),
	}

	// Get target groups and targets
	populateErr := am.populateLBTargetGroups(&lbInfo, arn)
	if populateErr != nil {
		err = populateErr
		shouldInclude = false
		return lbInfo, shouldInclude, err
	}

	shouldInclude = true
	return lbInfo, shouldInclude, err
}

func (am *AWSClusterManager) checkLBBelongsToCluster(tagOutput *elasticloadbalancingv2.DescribeTagsOutput) (belongsToCluster bool) {
	for _, td := range tagOutput.TagDescriptions {
		for _, tag := range td.Tags {
			if *tag.Key == ELBClusterTag && *tag.Value == am.ClusterName() {
				belongsToCluster = true
				return belongsToCluster
			}
		}
	}
	belongsToCluster = false
	return belongsToCluster
}

func (am *AWSClusterManager) populateLBTargetGroups(lbInfo *manager.LBInfo, lbArn string) (err error) {
	// Get Target Groups - have to get 'em all once again, since there's no filter
	tgOutput, tgErr := am.GetTargetGroupsForLB(lbArn)
	if tgErr != nil {
		err = errors.Wrapf(tgErr, "failed getting target groups")
		return err
	}

	// iterate over the target groups, also looking for the cluster name  (we use the cluster name in all lb's and tg's, which makes this possible)
	for _, tg := range tgOutput.TargetGroups {
		tgArn := *tg.TargetGroupArn

		tagInput := &elasticloadbalancingv2.DescribeTagsInput{
			ResourceArns: []string{tgArn},
		}

		// Describe the tags on the TG
		_, tagErr := am.ELBClient.DescribeTags(am.Context, tagInput)
		if tagErr != nil {
			err = errors.Wrapf(tagErr, "failed fetching tags for %s", tgArn)
			return err
		}

		// Record the Target Group info
		tgInfo := manager.LBTargetGroupInfo{
			Name: *tg.TargetGroupName,
			Arn:  *tg.TargetGroupArn,
			Port: *tg.Port,
		}
		lbInfo.TargetGroups = append(lbInfo.TargetGroups, tgInfo)

		// get the targets
		targets, targErr := am.GetTargets(*tg.TargetGroupName)
		if targErr != nil {
			err = errors.Wrapf(targErr, "failed getting target %s", *tg.TargetGroupName)
			return err
		}

		lbInfo.Targets = targets
	}

	return err
}

/*
func (am *AWSClusterManager) CreateLB() (err error) {
	// TODO implement CreateLB()
	return err
}

func (am *AWSClusterManager) DeleteLB(lbName string) (err error) {
	// TODO implement DeleteLB()
	return err
}

func (am *AWSClusterManager) UpdateLB() (err error) {
	// TODO implement UpdataLB()
	return err
}
*/

func (am *AWSClusterManager) DeRegisterNode(nodeName string, nodeID string) (err error) {
	manager.VerboseOutput(am.GetVerbose(), "Deregistering node %s from load balancers in cluster %s \n", nodeName, am.ClusterName())

	lbs, lbsErr := am.GetClusterLBs()
	if lbsErr != nil {
		err = errors.Wrapf(lbsErr, "failed getting cluster LB's")
		return err
	}

	// Remove Node from all LB's.  It doesn't appear to generate an error if you try to remove a node from a target group it's not in.
	for _, lb := range lbs {
		for _, tg := range lb.TargetGroups {
			// Register the node in the TargetGroup
			manager.VerboseOutput(am.GetVerbose(), "Degistering Node %s with Target Group %s on Port %d\n", nodeName, tg.Arn, tg.Port)
			deregErr := am.DeregisterTarget(tg.Arn, nodeID, tg.Port)
			if deregErr != nil {
				err = errors.Wrapf(deregErr, "failed deregistering %s on tg %s", nodeName, tg.Arn)
				return err
			}
		}
	}

	return err
}

func (am *AWSClusterManager) DeregisterTarget(tgARN string, nodeID string, port int32) (err error) {
	input := &elasticloadbalancingv2.DeregisterTargetsInput{
		TargetGroupArn: aws.String(tgARN),
		Targets: []types.TargetDescription{
			{
				Id:   aws.String(nodeID),
				Port: &port,
			},
		},
	}

	_, deregErr := am.ELBClient.DeregisterTargets(am.Context, input)
	if deregErr != nil {
		err = errors.Wrapf(deregErr, "failed deregistering target ID %s", nodeID)
		return err
	}

	return err
}

func (am *AWSClusterManager) RegisterNode(node manager.ClusterNode) (err error) {
	manager.VerboseOutput(am.GetVerbose(), "Registering Node %s with role %s\n", node.Name(), node.Role())

	lbs, lbsErr := am.GetClusterLBs()
	if lbsErr != nil {
		err = errors.Wrapf(lbsErr, "failed getting cluster LB's")
		return err
	}

	// Add to all TG's and ports for all LB's for the cluster for workers
	for _, lb := range lbs {
		// If we're looking at the apiserver LB, and this is not a CP node, move on.
		if node.Role() != manager.NodeRoleCp && lb.IsAPIServer {
			continue // skip registration
		}

		// If it's not an apiserver LB, and this is a CP node, and we don't schedule workloads here, move on
		if !lb.IsAPIServer && node.Role() == manager.NodeRoleCp && !am.GetScheduleWorkloadsOnCPNodes() {
			continue // skip registration
		}

		// Register the node.
		for _, tg := range lb.TargetGroups {
			// Register the node in the TargetGroup
			manager.VerboseOutput(am.GetVerbose(), "Registering Node %s with Target Group %s on Port %d\n", node.ID(), tg.Arn, tg.Port)
			regErr := am.RegisterTarget(tg.Arn, node.ID(), tg.Port)
			if regErr != nil {
				err = errors.Wrapf(regErr, "failed registering %s on tg %s", node.ID(), tg.Arn)
				return err
			}
		}
	}

	return err
}

func (am *AWSClusterManager) RegisterTarget(tgARN string, nodeID string, port int32) (err error) {
	input := &elasticloadbalancingv2.RegisterTargetsInput{
		TargetGroupArn: aws.String(tgARN),
		Targets: []types.TargetDescription{
			{
				Id:   aws.String(nodeID),
				Port: &port,
			},
		},
	}

	_, regErr := am.ELBClient.RegisterTargets(am.Context, input)
	if regErr != nil {
		err = errors.Wrapf(regErr, "failed registering target ID %s", nodeID)
		return err
	}
	return err
}

func (am *AWSClusterManager) GetTargetGroups(tgName string) (tgOutput *elasticloadbalancingv2.DescribeTargetGroupsOutput, err error) {
	// DescribeTargetGroups gives you all by default unless you give it a name or ARN

	input := &elasticloadbalancingv2.DescribeTargetGroupsInput{}

	// Won't accept an empty string though.
	if tgName != "" {
		input.Names = []string{tgName}
	}

	var descErr error
	tgOutput, descErr = am.ELBClient.DescribeTargetGroups(am.Context, input)
	if descErr != nil {
		err = errors.Wrapf(descErr, "failed getting targetGroup %s", tgName)
		return tgOutput, err
	}

	return tgOutput, err

}

func (am *AWSClusterManager) GetTargetGroupsForLB(lbArn string) (tgOutput *elasticloadbalancingv2.DescribeTargetGroupsOutput, err error) {
	// DescribeTargetGroups gives you all by default unless you give it a name or ARN

	input := &elasticloadbalancingv2.DescribeTargetGroupsInput{
		LoadBalancerArn: aws.String(lbArn),
	}

	var descErr error
	tgOutput, descErr = am.ELBClient.DescribeTargetGroups(am.Context, input)
	if descErr != nil {
		err = errors.Wrapf(descErr, "failed getting targetGroup for lb %s", lbArn)
		return tgOutput, err
	}

	return tgOutput, err

}

func (am *AWSClusterManager) GetTargets(tgName string) (targets []manager.LBTargetInfo, err error) {
	targets = make([]manager.LBTargetInfo, 0)
	// Need the ARN of the TG
	groups, groupsErr := am.GetTargetGroups(tgName)
	if groupsErr != nil {
		err = errors.Wrapf(groupsErr, "failed looking up target group %s", tgName)
		return targets, err
	}

	// TODO is mindlessly returning the first group found going to be safe?
	// return the first found.
	tg := groups.TargetGroups[0]

	//DescribeTargetHealth
	//aws elbv2 describe-target-health --target-group-arn ${TG}  --query 'TargetHealthDescriptions[*].Target.Id'

	input := &elasticloadbalancingv2.DescribeTargetHealthInput{
		TargetGroupArn: aws.String(*tg.TargetGroupArn),
		Include:        nil,
		Targets:        nil,
	}

	output, descErr := am.ELBClient.DescribeTargetHealth(am.Context, input)
	if descErr != nil {
		err = errors.Wrapf(descErr, "failed getting target group health for %s", tgName)
		return targets, err
	}

	for _, t := range output.TargetHealthDescriptions {
		var nodeInfo manager.NodeInfo

		// Try to pull the node info from cache, else we'll be looking up the same nodes over and over
		cachedNode, ok := am.FetchedNodesById[*t.Target.Id]
		if ok {
			nodeInfo = cachedNode
		} else {
			fetchedNode, nodeErr := am.GetNodeById(*t.Target.Id)
			if nodeErr != nil {
				err = errors.Wrapf(nodeErr, "failed getting node by ID %s", *t.Target.Id)
				return targets, err
			} else if len(fetchedNode.ID) == 0 {
				logrus.Warnf("id %s exists but is not owned by this account", *t.Target.Id)
				return targets, err
			}

			nodeInfo = fetchedNode
		}

		info := manager.LBTargetInfo{
			ID:    *t.Target.Id,
			Name:  nodeInfo.Name,
			Port:  int32(int(*t.Target.Port)),
			State: string(t.TargetHealth.State),
		}

		targets = append(targets, info)
	}

	// Sort the output alphabetically
	targets = sortLBTargetsByName(targets)

	return targets, err
}

func sortLBTargetsByName(targets []manager.LBTargetInfo) (sorted []manager.LBTargetInfo) {
	sorted = make([]manager.LBTargetInfo, len(targets))
	copy(sorted, targets)
	// Use sort.Slice with type assertion to avoid closure with named returns
	sort.Slice(sorted, lbTargetComparator{targets: sorted}.Less)
	return sorted
}

type lbTargetComparator struct {
	targets []manager.LBTargetInfo
}

func (c lbTargetComparator) Less(i, j int) (less bool) {
	less = c.targets[i].Name < c.targets[j].Name
	return less
}
