package aws

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

const INSTANCEID = "i-0af01c0123456789a"
const NODENAME = "test-cp-1"
const TEST_EC2_SG_TAG = "Cluster"
const TEST_EC2_SG_TAG_VALUE = "test-cluster"

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

type MockEc2ClientOneSecurityGroup struct {
	*ec2.Client
}

func (MockEc2ClientOneSecurityGroup) DescribeSecurityGroups(ctx context.Context, params *ec2.DescribeSecurityGroupsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeSecurityGroupsOutput, error) {
	return &ec2.DescribeSecurityGroupsOutput{
		SecurityGroups: []types.SecurityGroup{
			{
				Tags: []types.Tag{
					{
						Key:   aws.String(TEST_EC2_SG_TAG),
						Value: aws.String(TEST_EC2_SG_TAG_VALUE),
					},
				},
			},
		},
	}, nil
}

type MockEc2ClientNoSecurityGroups struct {
	*ec2.Client
}

func (MockEc2ClientNoSecurityGroups) DescribeSecurityGroups(ctx context.Context, params *ec2.DescribeSecurityGroupsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeSecurityGroupsOutput, error) {
	return &ec2.DescribeSecurityGroupsOutput{
		SecurityGroups: nil,
	}, nil
}

type MockEc2ClientOneNodeSecurityGroup struct {
	*ec2.Client
}

func (MockEc2ClientOneNodeSecurityGroup) DescribeSecurityGroups(ctx context.Context, params *ec2.DescribeSecurityGroupsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeSecurityGroupsOutput, error) {
	return &ec2.DescribeSecurityGroupsOutput{
		SecurityGroups: []types.SecurityGroup{
			{
				Tags: []types.Tag{
					{
						Key:   aws.String(TEST_EC2_SG_TAG),
						Value: aws.String(TEST_EC2_SG_TAG_VALUE),
					},
				},
				IpPermissions: []types.IpPermission{
					{
						ToPort: aws.Int32(TALOS_CONTROL_PORT),
					},
				},
			},
		},
	}, nil
}

type MockEc2ClientNoNodeSecurityGroup struct {
	*ec2.Client
}

func (MockEc2ClientNoNodeSecurityGroup) DescribeSecurityGroups(ctx context.Context, params *ec2.DescribeSecurityGroupsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeSecurityGroupsOutput, error) {
	return &ec2.DescribeSecurityGroupsOutput{
		SecurityGroups: []types.SecurityGroup{
			{
				Tags: []types.Tag{
					{
						Key:   aws.String(TEST_EC2_SG_TAG),
						Value: aws.String(TEST_EC2_SG_TAG_VALUE),
					},
				},
				IpPermissions: []types.IpPermission{
					{
						ToPort: aws.Int32(TALOS_CONTROL_PORT + 1),
					},
				},
			},
		},
	}, nil
}
