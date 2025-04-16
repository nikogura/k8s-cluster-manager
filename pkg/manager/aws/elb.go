package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"github.com/nikogura/k8s-cluster-manager/pkg/manager"
	"github.com/pkg/errors"
	"sort"
)

const API_SERVER_PORT = 6443
const CLEARTEXT_INGRESS_PORT_INT = 30080
const TLS_INGRESS_PORT_INT = 30443
const CLEARTEXT_INGRESS_PORT_EXT = 31080
const TLS_INGRESS_PORT_EXT = 31443
const ELB_CLUSTER_TAG = "Cluster"

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

func (am *AWSClusterManager) GetLBs(clusterName string) (lbs []manager.LBInfo, err error) {
	/*
		There is no way to filter LoadBalancers by tag
		aws elbv2 describe-load-balancers | jq -r '.LoadBalancers[].LoadBalancerArn' | xargs -I {} aws elbv2 describe-tags --resource-arns {} --query "TagDescriptions[?Tags[?Key=='env' &&Value=='dev'] && Tags[?Key=='created_by' &&Value=='xyz']].ResourceArn" --output text

	*/

	// DescribeLoadBalancers gives all by default, or filters by name or arn
	input := &elasticloadbalancingv2.DescribeLoadBalancersInput{}

	// Get all the load balancers
	output, err := am.ELBClient.DescribeLoadBalancers(am.Context, input)
	if err != nil {
		err = errors.Wrapf(err, "failed getting lbs for cluster %s", clusterName)
		return lbs, err
	}

	// Iterate through the list of load balancers, cos you can't filter by tag
	for _, lb := range output.LoadBalancers {
		// Find the ARN
		arn := *lb.LoadBalancerArn

		// Look for the tags by ARN
		tagInput := &elasticloadbalancingv2.DescribeTagsInput{
			ResourceArns: []string{arn},
		}

		// Describe the tags
		tagOutput, err := am.ELBClient.DescribeTags(am.Context, tagInput)
		if err != nil {
			err = errors.Wrapf(err, "failed fetching ")
		}

		// Iterate through the cruft and find one that has a tag for ELB_CLUSTER_TAG and a value equal to our cluster name
		for _, td := range tagOutput.TagDescriptions {
			for _, tag := range td.Tags {
				if *tag.Key == ELB_CLUSTER_TAG && *tag.Value == clusterName {

					lbInfo := manager.LBInfo{
						Name:    *lb.LoadBalancerName,
						Targets: make([]manager.LBTargetInfo, 0),
					}

					// Get Target Groups - have to get 'em all once again, since there's no filter
					tgOutput, err := am.GetTargetGroupsForLB(arn)
					if err != nil {
						err = errors.Wrapf(err, "failed getting target groups")
						return lbs, err
					}

					// iterate over the target groups, also looking for the cluster name  (we use the cluster name in all lb's and tg's, which makes this possible)
					for _, tg := range tgOutput.TargetGroups {
						tgArn := *tg.TargetGroupArn

						tagInput = &elasticloadbalancingv2.DescribeTagsInput{
							ResourceArns: []string{tgArn},
						}

						// Describe the tags on the TG
						tagOutput, err = am.ELBClient.DescribeTags(am.Context, tagInput)
						if err != nil {
							err = errors.Wrapf(err, "failed fetching tags for %s", tgArn)
							return lbs, err
						}
						for _, tag := range td.Tags {
							if *tag.Key == ELB_CLUSTER_TAG && *tag.Value == clusterName {
								// get the targets
								targets, err := am.GetTargets(*tg.TargetGroupName)
								if err != nil {
									err = errors.Wrapf(err, "failed getting target %s", *tg.TargetGroupName)
									return lbs, err
								}

								lbInfo.Targets = targets

							}
						}
					}

					// add it to the pile
					lbs = append(lbs, lbInfo)
				}
			}
		}
	}

	return lbs, err
}

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

func (am *AWSClusterManager) DeRegisterNode(nodeName string) (err error) {
	fmt.Printf("TODO DeregisterNode() Deregistering node from load balancers %s\n", nodeName)

	//// Get node Info - need the ID
	//nodeInfo, err := am.GetNode(nodeName)
	//if err != nil {
	//	err = errors.Wrapf(err, "failed getting info for %s", nodeName)
	//}
	//
	//// get target groups
	//tgOutput, err := am.GetTargetGroups("")
	//if err != nil {
	//	err = errors.Wrapf(err, "failed getting target groups")
	//	return err
	//}
	//
	//// iterate through the list of target groups
	//for _, tg := range tgOutput.TargetGroups {
	//	// only concern ourselves with target groups that have our cluster name in it
	//	if am.ClusterNameRegex.MatchString(*tg.TargetGroupName) {
	//
	//		// Remove from apiserver TG if we're in it
	//		err = am.DeregisterTarget(*tg.TargetGroupArn, nodeInfo.ID, API_SERVER_PORT)
	//		if err != nil {
	//			err = errors.Wrapf(err, "failed deregistering node id %s from tg %s port %d\n", nodeInfo.ID, *tg.TargetGroupArn, CLEARTEXT_INGRESS_PORT_INT)
	//			return err
	//		}
	//
	//		// Remove from internal ingress Cleartext TG
	//		err = am.DeregisterTarget(*tg.TargetGroupArn, nodeInfo.ID, CLEARTEXT_INGRESS_PORT_INT)
	//		if err != nil {
	//			err = errors.Wrapf(err, "failed deregistering node id %s from tg %s port %d\n", nodeInfo.ID, *tg.TargetGroupArn, CLEARTEXT_INGRESS_PORT_INT)
	//			return err
	//		}
	//
	//		// Remove from internal ingress TLS TG
	//		err = am.DeregisterTarget(*tg.TargetGroupArn, nodeInfo.ID, TLS_INGRESS_PORT_INT)
	//		if err != nil {
	//			err = errors.Wrapf(err, "failed deregistering node id %s from tg %s port %d\n", nodeInfo.ID, *tg.TargetGroupArn, TLS_INGRESS_PORT_INT)
	//			return err
	//		}
	//
	//		// Remove from external ingress Cleartext TG
	//		err = am.DeregisterTarget(*tg.TargetGroupArn, nodeInfo.ID, CLEARTEXT_INGRESS_PORT_EXT)
	//		if err != nil {
	//			err = errors.Wrapf(err, "failed deregistering node id %s from tg %s port %d\n", nodeInfo.ID, *tg.TargetGroupArn, CLEARTEXT_INGRESS_PORT_EXT)
	//			return err
	//		}
	//
	//		// Remove from external ingress TLS TG
	//		err = am.DeregisterTarget(*tg.TargetGroupArn, nodeInfo.ID, TLS_INGRESS_PORT_EXT)
	//		if err != nil {
	//			err = errors.Wrapf(err, "failed deregistering node id %s from tg %s port %d\n", nodeInfo.ID, *tg.TargetGroupArn, TLS_INGRESS_PORT_EXT)
	//			return err
	//		}
	//
	//		// TODO Remove from Kafka TG's
	//		fmt.Printf("TODO: Remove from Kafka Target Groups\n")
	//
	//	}
	//}

	return err
}

func (am *AWSClusterManager) DeregisterTarget(tgARN string, nodeID string, port int) (err error) {
	p := int32(port)

	input := &elasticloadbalancingv2.DeregisterTargetsInput{
		TargetGroupArn: aws.String(tgARN),
		Targets: []types.TargetDescription{
			{
				Id:   aws.String(nodeID),
				Port: &p,
			},
		},
	}

	_, err = am.ELBClient.DeregisterTargets(am.Context, input)
	if err != nil {
		err = errors.Wrapf(err, "failed deregistering target ID %s", nodeID)
		return err
	}

	return err
}

func (am *AWSClusterManager) RegisterNode(config manager.ClusterNode) (err error) {
	fmt.Printf("Registering Node %s\n", config.Name())

	// TODO Differentiate between CP and Worker Nodes

	// TODO Add to APIserver if node is a CP Node

	// TODO Add to TG"s for all CP Nodes if workloads are scheduled on the CP nodes.

	// TODO Add to all TG's and ports for all LB's for the cluster for workers

	return err

}

func (am *AWSClusterManager) RegisterTarget(tgARN string, nodeID string, port int) (err error) {
	p := int32(port)

	input := &elasticloadbalancingv2.RegisterTargetsInput{
		TargetGroupArn: aws.String(tgARN),
		Targets: []types.TargetDescription{
			{
				Id:   aws.String(nodeID),
				Port: &p,
			},
		},
	}

	_, err = am.ELBClient.RegisterTargets(am.Context, input)
	if err != nil {
		err = errors.Wrapf(err, "failed registering target ID %s", nodeID)
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

	tgOutput, err = am.ELBClient.DescribeTargetGroups(am.Context, input)
	if err != nil {
		err = errors.Wrapf(err, "failed getting targetGroup %s", tgName)
		return tgOutput, err
	}

	return tgOutput, err

}

func (am *AWSClusterManager) GetTargetGroupsForLB(lbArn string) (tgOutput *elasticloadbalancingv2.DescribeTargetGroupsOutput, err error) {
	// DescribeTargetGroups gives you all by default unless you give it a name or ARN

	input := &elasticloadbalancingv2.DescribeTargetGroupsInput{
		LoadBalancerArn: aws.String(lbArn),
	}

	tgOutput, err = am.ELBClient.DescribeTargetGroups(am.Context, input)
	if err != nil {
		err = errors.Wrapf(err, "failed getting targetGroup for lb %s", lbArn)
		return tgOutput, err
	}

	return tgOutput, err

}

func (am *AWSClusterManager) GetTargets(tgName string) (targets []manager.LBTargetInfo, err error) {
	targets = make([]manager.LBTargetInfo, 0)
	// Need the ARN of the TG
	groups, err := am.GetTargetGroups(tgName)
	if err != nil {
		err = errors.Wrapf(err, "failed looking up target group %s", tgName)
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

	output, err := am.ELBClient.DescribeTargetHealth(am.Context, input)
	if err != nil {
		err = errors.Wrapf(err, "failed getting target group health for %s", tgName)
		return targets, err
	}

	for _, t := range output.TargetHealthDescriptions {
		var nodeInfo manager.NodeInfo

		// Try to pull the node info from cache, else we'll be looking up the same nodes over and over
		node, ok := am.FetchedNodesById[*t.Target.Id]
		if ok {
			nodeInfo = node
		} else {
			node, err := am.GetNodeById(*t.Target.Id)
			if err != nil {
				err = errors.Wrapf(err, "failed getting node by ID %s", *t.Target.Id)
				return targets, err
			}

			nodeInfo = node
		}

		info := manager.LBTargetInfo{
			ID:    *t.Target.Id,
			Name:  nodeInfo.Name,
			Port:  int(*t.Target.Port),
			State: string(t.TargetHealth.State),
		}

		targets = append(targets, info)
	}

	// Sort the output alphabetically
	sort.Slice(targets, func(i, j int) bool {
		return targets[i].Name < targets[j].Name
	})

	return targets, err
}
