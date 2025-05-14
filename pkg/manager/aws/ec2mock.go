package aws

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

const INSTANCEID = "i-0af01c0123456789a"
const NODENAME = "test-cp-1"

type Ec2Client interface {
	DescribeInstances(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error)
	RunInstances(ctx context.Context, params *ec2.RunInstancesInput, optFns ...func(*ec2.Options)) (*ec2.RunInstancesOutput, error)
	TerminateInstances(ctx context.Context, params *ec2.TerminateInstancesInput, optFns ...func(*ec2.Options)) (*ec2.TerminateInstancesOutput, error)
	DescribeSecurityGroups(ctx context.Context, params *ec2.DescribeSecurityGroupsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeSecurityGroupsOutput, error)
}

//var blah = ec2.Client.DescribeSecurityGroups

type MockEc2ClientOneRunningInst struct {
	*ec2.Client
}

//var blah = ec2.Client.DescribeInstances(ctx, blah)

func (MockEc2ClientOneRunningInst) DescribeInstances(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error) {
	return &ec2.DescribeInstancesOutput{
		Reservations: []types.Reservation{
			{
				Instances: []types.Instance{
					{
						State:      &types.InstanceState{Name: types.InstanceStateNameRunning},
						InstanceId: aws.String(INSTANCEID),
						Tags: []types.Tag{
							{
								Key:   aws.String("Name"),
								Value: aws.String(NODENAME),
							},
						},
					},
				},
			},
		},
	}, nil
}

type MockEc2ClientStoppedInst struct {
	*ec2.Client
}

//var blah = ec2.Client.DescribeInstances(ctx, blah)

func (MockEc2ClientStoppedInst) DescribeInstances(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error) {
	return &ec2.DescribeInstancesOutput{
		Reservations: []types.Reservation{
			{
				Instances: []types.Instance{
					{
						State:      &types.InstanceState{Name: types.InstanceStateNameStopped},
						InstanceId: aws.String(INSTANCEID),
						Tags: []types.Tag{
							{
								Key:   aws.String("Name"),
								Value: aws.String(NODENAME),
							},
						},
					},
				},
			},
		},
	}, nil
}
