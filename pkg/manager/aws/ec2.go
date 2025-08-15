package aws

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/nikogura/k8s-cluster-manager/pkg/manager"
	"github.com/nikogura/k8s-cluster-manager/pkg/manager/kubernetes"
	"github.com/nikogura/k8s-cluster-manager/pkg/manager/talos"
	"github.com/pkg/errors"
	"sort"
	"strconv"
	"time"
)

const TALOS_CONTROL_PORT = 50000

func (am *AWSClusterManager) CreateNode(nodeName string, nodeRole string, config AWSNodeConfig, machineConfigBytes []byte, machineConfigPatches []string) (err error) {
	// Create Instance
	fmt.Printf("Creating Node %s with role %s in cluster %s\n", nodeName, nodeRole, am.ClusterName())

	node := AWSNode{
		NodeName:   nodeName,
		IPAddress:  "", // we don't know the IP address yet.
		NodeRole:   nodeRole,
		Config:     &config,
		NodeDomain: config.Domain,
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

	blockSize, err := strconv.Atoi(config.BlockDeviceGb)
	if err != nil {
		err = errors.Wrapf(err, "failed converting %s to integer", config.BlockDeviceGb)
		return err
	}

	blockDeviceMappings := []types.BlockDeviceMapping{
		types.BlockDeviceMapping{
			DeviceName: aws.String("/dev/xvda"),
			Ebs: &types.EbsBlockDevice{
				DeleteOnTermination: aws.Bool(true),
				Encrypted:           aws.Bool(true),
				//Iops:                nil,
				//KmsKeyId:            nil,
				//OutpostArn:          nil,
				//SnapshotId:          nil,
				//Throughput:          nil,
				VolumeSize: aws.Int32(int32(blockSize)),
				VolumeType: types.VolumeType(config.BlockDeviceType),
			},
			NoDevice:    nil,
			VirtualName: nil,
		},
	}

	securityGroups, err := am.GetNodeSecurityGroupsForCluster()
	if err != nil {
		err = errors.Wrapf(err, "failed getting security groups")
		return err
	}

	sgIDs := make([]string, 0)

	for _, g := range securityGroups {
		sgIDs = append(sgIDs, *g.GroupId)
	}

	input := &ec2.RunInstancesInput{
		MaxCount:            aws.Int32(1),
		MinCount:            aws.Int32(1),
		ImageId:             aws.String(config.ImageID),
		TagSpecifications:   tags,
		SecurityGroupIds:    sgIDs,
		SubnetId:            aws.String(config.SubnetID),
		InstanceType:        types.InstanceType(config.InstanceType),
		BlockDeviceMappings: blockDeviceMappings,
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

	// Set IP on node struct
	node.IPAddress = *output.Instances[0].PrivateIpAddress
	node.NodeID = *output.Instances[0].InstanceId

	fullAddr := fmt.Sprintf("%s:50000", node.IP())

	manager.VerboseOutput(am.Verbose(), "Waiting for node %s to become ready (This will take several tries.)\n", fullAddr)

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second)
	defer cancel()

	conn, err := manager.DialWithRetry(ctx, "tcp", fullAddr, 150, 2*time.Second, am.Verbose())
	if err != nil {
		err = errors.Wrapf(err, "failed dialing %s", fullAddr)
		return err
	}
	conn.Close()

	// Apply Talos machine config
	err = talos.ApplyConfig(am.Context, &node, machineConfigBytes, machineConfigPatches, true, am.Verbose())
	if err != nil {
		err = errors.Wrapf(err, "failed applying machine config to %s", nodeName)
		return err
	}

	// Register Node with Load Balancers
	err = am.RegisterNode(node)
	if err != nil {
		err = errors.Wrapf(err, "failed registering %s", nodeName)
		return err
	}

	// Register Node with DNS
	err = am.DNSManager().RegisterNode(am.Context, node, am.Verbose())
	if err != nil {
		err = errors.Wrapf(err, "failed registering dns for %s", nodeName)
		return err
	}

	fmt.Printf("Node %s (%s) Successfully Created and Registered\n", node.Name(), node.NodeID)

	return err
}

func (am *AWSClusterManager) DeleteNode(nodeName string) (err error) {
	// Get Node info
	manager.VerboseOutput(am.Verbose(), "Getting node info\n")
	nodeInfo, err := am.GetNode(nodeName)
	if err != nil {
		err = errors.Wrapf(err, "failed getting node %s", nodeName)
		return err
	}

	err = am.DNSManager().DeregisterNode(am.Context, nodeName, am.Verbose())
	if err != nil {
		err = errors.Wrapf(err, "failed deregistering dns for %s", nodeName)
		return err
	}

	err = am.DeRegisterNode(nodeName, nodeInfo.ID)
	if err != nil {
		err = errors.Wrapf(err, "failed deregistering node %s", nodeName)
		return err
	}

	manager.VerboseOutput(am.Verbose(), "Removing node %s from EC2\n", nodeName)
	// Remove Instance
	input := &ec2.TerminateInstancesInput{
		InstanceIds: []string{nodeInfo.ID},
	}

	// Terminate Instances
	_, err = am.Ec2Client.TerminateInstances(am.Context, input)
	if err != nil {
		err = errors.Wrapf(err, "failed removing node %s (%s) from aws", nodeName, nodeInfo.ID)
		return err
	}

	err = kubernetes.DeleteNode(am.Context, nodeName, am.Verbose())
	if err != nil {
		err = errors.Wrapf(err, "failed deleting node %s from k8s", nodeName)
	}

	fmt.Printf("Node %s (%s) Terminated\n", nodeName, nodeInfo.ID)

	return err
}

func (am *AWSClusterManager) GetNode(nodeName string) (nodeInfo manager.NodeInfo, err error) {
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

	// There could be any number of instances out there with the same Name tag.  We're only interested in the one that's 'running'.
	for _, res := range output.Reservations {
		for _, inst := range res.Instances {
			//fmt.Printf("Found Reservation: %s ID: %s State: %v\n", *res.ReservationId, *inst.InstanceId, inst.State.Name)
			if inst.State.Name == types.InstanceStateNameRunning {
				//fmt.Printf("Selecting instance Reservation: %s ID: %s State: %v\n", *res.ReservationId, *inst.InstanceId, inst.State.Name)
				nodeInfo.Name = nodeName
				nodeInfo.ID = *inst.InstanceId
				nodeInfo.InstanceType = string(inst.InstanceType)

				am.FetchedNodesByName[nodeName] = nodeInfo
				am.FetchedNodesById[*inst.InstanceId] = nodeInfo

				return nodeInfo, err
			}
		}
	}

	//if len(output.Reservations) > 0 {
	//	if len(output.Reservations[0].Instances) > 0 {
	//		nodeInfo.Name = nodeName
	//		nodeInfo.ID = *output.Reservations[0].Instances[0].InstanceId
	//
	//		am.FetchedNodesByName[nodeName] = nodeInfo
	//		am.FetchedNodesById[*output.Reservations[0].Instances[0].InstanceId] = nodeInfo
	//
	//	} else {
	//		err = errors.New(fmt.Sprintf("instance for name %s not found", nodeName))
	//		return nodeInfo, err
	//
	//	}
	//} else {
	//	err = errors.New(fmt.Sprintf("reservation for name %s not found", nodeName))
	//	return nodeInfo, err
	//}

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
			nodeInfo.InstanceType = string(output.Reservations[0].Instances[0].InstanceType)

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
			Name:         name,
			ID:           *instance.InstanceId,
			InstanceType: string(instance.InstanceType),
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

func (am *AWSClusterManager) GetSecurityGroupsForCluster() (groups []types.SecurityGroup, err error) {
	input := &ec2.DescribeSecurityGroupsInput{
		//DryRun:     nil,
		Filters: []types.Filter{
			{
				Name: aws.String("tag:Cluster"),
				Values: []string{
					am.ClusterName(),
				},
			},
		},
		//GroupIds:   nil,
		//GroupNames: nil,
		//MaxResults: nil,
		//NextToken:  nil,
	}

	output, err := am.Ec2Client.DescribeSecurityGroups(am.Context, input)
	if err != nil {
		err = errors.Wrapf(err, "failed getting security groups for cluster %s", am.ClusterName())
		return groups, err
	}

	groups = output.SecurityGroups

	return groups, err
}

func (am *AWSClusterManager) GetNodeSecurityGroupsForCluster() (groups []types.SecurityGroup, err error) {
	groups = make([]types.SecurityGroup, 0)

	allGroups, err := am.GetSecurityGroupsForCluster()
	if err != nil {
		err = errors.Wrapf(err, "failed getting security groups for cluster")
		return groups, err
	}

	for _, g := range allGroups {
		for _, p := range g.IpPermissions {
			if p.ToPort != nil {
				if int(*p.ToPort) == TALOS_CONTROL_PORT {
					groups = append(groups, g)
				}
			}
		}
	}

	return groups, err
}

func (am *AWSClusterManager) UpdateNode(nodeName string) (err error) {
	// Unsure what we'd be updating beyond machine config.
	// TODO Update LB
	// TODO Update Instance
	// TODO Update DNS
	return err
}

func (am *AWSClusterManager) DescribeNode(nodeName string) (info manager.NodeInfo, err error) {
	// May be unnecessary?

	return info, err
}
