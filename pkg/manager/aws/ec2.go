package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/nikogura/k8s-cluster-manager/pkg/manager"
	"github.com/pkg/errors"
	"sort"
)

func (am *AWSClusterManager) CreateNode(instanceName string, lbName string) (err error) {
	// Create Instance
	input := &ec2.RunInstancesInput{}

	_, err = am.Ec2Client.RunInstances(am.Context, input)
	if err != nil {
		err = errors.Wrapf(err, "failed running instance %s", instanceName)
		return err
	}

	// TODO Add to Target Group

	// TODO Register DNS

	return err
}

func (am *AWSClusterManager) DeleteNode(nodeName string) (err error) {
	// Get Node info
	nodeInfo, err := am.GetNode(nodeName)
	if err != nil {
		err = errors.Wrapf(err, "failed getting node %s", nodeName)
		return err
	}

	// TODO Remove from DNS

	// TODO Remove from LB Target Group

	// Remove Instance
	input := &ec2.TerminateInstancesInput{
		InstanceIds: []string{nodeInfo.ID},
	}

	// Terminate Instances
	_, err = am.Ec2Client.TerminateInstances(am.Context, input)
	if err != nil {
		err = errors.Wrapf(err, "failed removing node %s from aws", nodeName)
		return err
	}

	return err
}

func (am *AWSClusterManager) GetNode(nodeName string) (nodeInfo manager.NodeInfo, err error) {

	// aws ec2 describe-instances --filters Name=tag:Name,Values=alpha*
	filter := types.Filter{
		Name:   aws.String("tag:Name"),
		Values: []string{fmt.Sprintf("%s-", nodeName)},
	}

	input := &ec2.DescribeInstancesInput{
		DryRun:      nil,
		Filters:     []types.Filter{filter},
		InstanceIds: nil,
		MaxResults:  nil,
		NextToken:   nil,
	}

	output, err := am.Ec2Client.DescribeInstances(am.Context, input)
	if err != nil {
		err = errors.Wrapf(err, "failed getting node %s", nodeName)
	}

	if len(output.Reservations) > 0 {
		if len(output.Reservations[0].Instances) > 0 {
			nodeInfo.Name = nodeName
			nodeInfo.ID = *output.Reservations[0].Instances[0].InstanceId

			am.FetchedNodesByName[nodeName] = nodeInfo
			am.FetchedNodesById[*output.Reservations[0].Instances[0].InstanceId] = nodeInfo

		} else {
			err = errors.New(fmt.Sprintf("instance for name %s not found", nodeName))
			return nodeInfo, err

		}
	} else {
		err = errors.New(fmt.Sprintf("reservation for name %s not found", nodeName))
		return nodeInfo, err
	}

	// TODO get DNS Status?

	return nodeInfo, err
}

func (am *AWSClusterManager) GetNodeById(id string) (nodeInfo manager.NodeInfo, err error) {

	input := &ec2.DescribeInstancesInput{
		InstanceIds: []string{id},
	}

	output, err := am.Ec2Client.DescribeInstances(am.Context, input)
	if err != nil {
		err = errors.Wrapf(err, "failed getting node by id %s", id)
	}

	var nodeName string

	if len(output.Reservations) > 0 {
		if len(output.Reservations[0].Instances) > 0 {
			// assuming there's only 1 reservation and 1 instance makes me nervous, but that is how I create 'em
			for _, tag := range output.Reservations[0].Instances[0].Tags {
				if *tag.Key == "Name" {
					nodeName = *tag.Value
				}
			}

			nodeInfo.ID = id
			nodeInfo.Name = nodeName

			am.FetchedNodesByName[nodeName] = nodeInfo
			am.FetchedNodesById[*output.Reservations[0].Instances[0].InstanceId] = nodeInfo

		} else {
			err = errors.New(fmt.Sprintf("instance for id %s not found", id))
			return nodeInfo, err
		}
	} else {
		err = errors.New(fmt.Sprintf("reservation for id %s not found", id))
		return nodeInfo, err
	}

	// TODO get DNS Status?

	return nodeInfo, err
}

func (am *AWSClusterManager) GetNodes(clusterName string) (nodeInfo []manager.NodeInfo, err error) {
	nodeInfo = make([]manager.NodeInfo, 0)

	// aws ec2 describe-instances --filters Name=tag:Name,Values=alpha*
	filter := types.Filter{
		Name:   aws.String("tag:Name"),
		Values: []string{fmt.Sprintf("%s-*", clusterName)},
	}

	input := &ec2.DescribeInstancesInput{
		DryRun:      nil,
		Filters:     []types.Filter{filter},
		InstanceIds: nil,
		MaxResults:  nil,
		NextToken:   nil,
	}

	output, err := am.Ec2Client.DescribeInstances(am.Context, input)
	if err != nil {
		err = errors.Wrapf(err, "failed getting nodes for cluster %s", clusterName)
		return nodeInfo, err
	}

	for _, reservation := range output.Reservations {

		instance := reservation.Instances[0] // I don't know why we'd ever have more than 1 instance per reservation

		var name string
		tags := instance.Tags
		for _, t := range tags {
			if *t.Key == "Name" {
				name = *t.Value
			}
		}

		info := manager.NodeInfo{
			Name: name,
		}

		nodeInfo = append(nodeInfo, info)

	}

	// TODO get DNS Status?

	// Sort the output alphabetically
	sort.Slice(nodeInfo, func(i, j int) bool {
		return nodeInfo[i].Name < nodeInfo[j].Name
	})

	return nodeInfo, err
}

func (am *AWSClusterManager) UpdateNode(nodeName string) (err error) {
	// TODO Update LB
	// TODO Update Instance
	// TODO Update DNS
	return err
}

func (am *AWSClusterManager) GlassNode(nodeName string) (err error) {
	// TODO Delete
	// TODO Recreate
	return err
}

func (am *AWSClusterManager) DescribeNode(nodeName string) (info manager.NodeInfo, err error) {
	return info, err
}
