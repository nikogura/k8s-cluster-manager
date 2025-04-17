package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/nikogura/k8s-cluster-manager/pkg/manager"
	"github.com/nikogura/k8s-cluster-manager/pkg/manager/cloudflare"
	"github.com/nikogura/k8s-cluster-manager/pkg/manager/talos"
	"github.com/pkg/errors"
	"sort"
)

func (am *AWSClusterManager) CreateNode(nodeName string, nodeRole string, config AWSNodeConfig) (err error) {
	// Create Instance
	fmt.Printf("Creating Node %s with role %s in cluster %s\n", nodeName, nodeRole, am.ClusterName())

	node := AWSNode{
		NodeName:  nodeName,
		IPAddress: "", // we don't know the IP address yet.
		NodeRole:  nodeRole,
		Config:    &config,
	}

	// Set up the "Name" Tag.
	tags := []types.TagSpecification{
		{
			ResourceType: types.ResourceTypeInstance,
			Tags: []types.Tag{
				{
					Key:   aws.String("Name"),
					Value: aws.String(nodeName),
				},
			},
		},
	}

	blockDeviceMappings := []types.BlockDeviceMapping{
		types.BlockDeviceMapping{
			DeviceName: aws.String("/dev/xvda"),
			Ebs: &types.EbsBlockDevice{
				DeleteOnTermination: aws.Bool(true),
				//Encrypted:           nil,
				//Iops:                nil,
				//KmsKeyId:            nil,
				//OutpostArn:          nil,
				//SnapshotId:          nil,
				//Throughput:          nil,
				VolumeSize: aws.Int32(int32(config.BlockDeviceGb)),
				VolumeType: types.VolumeType(config.BlockDeviceType),
			},
			NoDevice:    nil,
			VirtualName: nil,
		},
	}

	input := &ec2.RunInstancesInput{
		MaxCount:          aws.Int32(1),
		MinCount:          aws.Int32(1),
		ImageId:           aws.String(config.ImageID),
		TagSpecifications: tags,
		SecurityGroupIds:  config.SecurityGroupIDs,
		SubnetId:          aws.String(config.SubnetID),
		Placement:         nil, // *types.Placement
		InstanceType:      types.InstanceType(config.InstanceType),

		BlockDeviceMappings: blockDeviceMappings,

		//SecurityGroups: nil, // []string  Names of security groups.  Probably not needed if we use ID's, and vice versa.
	}

	if config.PlacementGroupName != "" {
		input.Placement = &types.Placement{
			GroupName: aws.String(config.PlacementGroupName),
		}
	}

	output, err := am.Ec2Client.RunInstances(am.Context, input)
	if err != nil {
		err = errors.Wrapf(err, "failed running instance %s", nodeName)
		return err
	}

	//Set IP on node struct
	node.IPAddress = *output.Instances[0].PrivateIpAddress
	node.NodeID = *output.Instances[0].InstanceId

	err = talos.ApplyConfig(&node)
	if err != nil {
		err = errors.Wrapf(err, "failed applying machine config to %s", nodeName)
		return err
	}

	err = am.RegisterNode(node)
	if err != nil {
		err = errors.Wrapf(err, "failed registering %s", nodeName)
		return err
	}

	err = cloudflare.RegisterNode(node)
	if err != nil {
		err = errors.Wrapf(err, "failed registering dns for %s", nodeName)
		return err
	}

	return err
}

func (am *AWSClusterManager) DeleteNode(nodeName string) (err error) {
	// Get Node info
	fmt.Printf("Getting node info\n")
	nodeInfo, err := am.GetNode(nodeName)
	if err != nil {
		err = errors.Wrapf(err, "failed getting node %s", nodeName)
		return err
	}

	err = cloudflare.DeRegisterNode(nodeName)
	if err != nil {
		err = errors.Wrapf(err, "failed deregistering dns for %s", nodeName)
		return err
	}

	err = am.DeRegisterNode(nodeName, nodeInfo.ID)
	if err != nil {
		err = errors.Wrapf(err, "failed deregistering node %s", nodeName)
		return err
	}

	fmt.Printf("Removing node %s from EC2\n", nodeName)
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

	fmt.Printf("Node Terminated\n")

	return err
}

func (am *AWSClusterManager) GetNode(nodeName string) (nodeInfo manager.NodeInfo, err error) {

	// aws ec2 describe-instances --filters Name=tag:Name,Values=alpha*
	filter := types.Filter{
		Name:   aws.String("tag:Name"),
		Values: []string{nodeName},
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
	// Unsure what we'd be updating.
	// TODO Update LB
	// TODO Update Instance
	// TODO Update DNS
	return err
}

func (am *AWSClusterManager) DescribeNode(nodeName string) (info manager.NodeInfo, err error) {
	// May be unnecessary?

	return info, err
}
