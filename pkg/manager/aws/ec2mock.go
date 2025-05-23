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

//to quickly find the signature of a mocked method, create a variable as below, use autocomplete, and Ctrl-Click right to the original method
//var blah = ec2.Client.DescribeInstances(ctx, blah)

type MockEc2ClientGetNodeOneRunningInst struct {
	*ec2.Client
}

func (MockEc2ClientGetNodeOneRunningInst) DescribeInstances(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error) {
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
								Value: &params.Filters[0].Values[0],
							},
						},
					},
				},
			},
		},
	}, nil
}

type MockEc2ClientGetNodeStoppedInst struct {
	*ec2.Client
}

func (MockEc2ClientGetNodeStoppedInst) DescribeInstances(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error) {
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
								Value: &params.Filters[0].Values[0],
							},
						},
					},
				},
			},
		},
	}, nil
}

type MockEc2ClientGetNodeByIdInstExists struct {
	*ec2.Client
}

func (MockEc2ClientGetNodeByIdInstExists) DescribeInstances(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error) {
	return &ec2.DescribeInstancesOutput{
		Reservations: []types.Reservation{
			{
				Instances: []types.Instance{
					{
						InstanceId: &params.InstanceIds[0],
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

type MockEc2ClientGetNodeByIdNoInst struct {
	*ec2.Client
}

func (MockEc2ClientGetNodeByIdNoInst) DescribeInstances(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error) {
	return &ec2.DescribeInstancesOutput{
		Reservations: []types.Reservation{},
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
