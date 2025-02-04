package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/nikogura/k8s-cluster-manager/pkg/manager"
	"github.com/pkg/errors"
	"regexp"
	"sort"
)

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

	// DescribeLoadBalancers gives all by default, or filters by name or arn
	input := &elasticloadbalancingv2.DescribeLoadBalancersInput{}

	output, err := am.ELBClient.DescribeLoadBalancers(am.Context, input)
	if err != nil {
		err = errors.Wrapf(err, "failed getting lbs for cluster %s", clusterName)
		return lbs, err
	}

	re, err := regexp.Compile(fmt.Sprintf(".*%s.*", clusterName))
	if err != nil {
		err = errors.Wrapf(err, "cluster name %s doesn't compile into a regex", clusterName)
		return lbs, err
	}

	for _, lb := range output.LoadBalancers {
		if re.MatchString(*lb.LoadBalancerName) {
			lbInfo := manager.LBInfo{
				Name:    *lb.LoadBalancerName,
				Targets: make([]manager.LBTargetInfo, 0),
			}

			// Get Target Groups - have to get 'em all, since there's no filter
			tgOutput, err := am.GetTargetGroups("")
			if err != nil {
				err = errors.Wrapf(err, "failed getting target groups")
				return lbs, err
			}

			for _, tg := range tgOutput.TargetGroups {
				if re.MatchString(*tg.TargetGroupName) {

					// get targets
					targets, err := am.GetTargets(*tg.TargetGroupName)
					if err != nil {
						err = errors.Wrapf(err, "failed getting target %s", *tg.TargetGroupName)
						return lbs, err
					}

					lbInfo.Targets = targets
				}
			}

			lbs = append(lbs, lbInfo)
		}
	}

	return lbs, err
}

//	CreateLB() (err error)
//	DeleteLB(lbName string) (err error)
//	UpdateLB() (err error)

func (am *AWSClusterManager) AddToLB(nodeName string, lbName string) (err error) {
	return err
}

func (am *AWSClusterManager) RemoveFromLB(nodeName string, lbName string) (err error) {

	return err
}

func (am *AWSClusterManager) RegisterTarget(instanceName string, lbName string) (err error) {

	re := regexp.MustCompile(`.*worker.*`)
	if !re.MatchString(instanceName) {
		// TODO Add to apiserver TG
	}

	// TODO find TG ARN

	// TODO Add to internal ingress TG
	// TODO Add to external ingress TG

	elbinput := &elasticloadbalancingv2.RegisterTargetsInput{
		TargetGroupArn: nil,
		Targets:        nil, // []TargetDescription
	}

	_, err = am.ELBClient.RegisterTargets(am.Context, elbinput)
	if err != nil {
		err = errors.Wrapf(err, "failed registering target %s to lb %s", instanceName, lbName)
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
