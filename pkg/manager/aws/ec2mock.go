package aws

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"regexp"
	"strings"
)

const TEST_INSTANCEID = "i-0af01c0123456789a"
const TEST_NODENAME = "test-cp-1"
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
						State:        &types.InstanceState{Name: types.InstanceStateNameRunning},
						InstanceId:   aws.String(TEST_INSTANCEID),
						InstanceType: types.InstanceTypeT3Medium,
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
						State:        &types.InstanceState{Name: types.InstanceStateNameStopped},
						InstanceId:   aws.String(TEST_INSTANCEID),
						InstanceType: types.InstanceTypeT3Medium,
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

type MockEc2ClientGetNodeNoInst struct {
	*ec2.Client
}

// TODO: validate response content when using filters and no matching instances exist
func (MockEc2ClientGetNodeNoInst) DescribeInstances(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error) {
	return &ec2.DescribeInstancesOutput{
		Reservations: []types.Reservation{},
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
						InstanceId:   &params.InstanceIds[0],
						InstanceType: types.InstanceTypeT3Medium,
						Tags: []types.Tag{
							{
								Key:   aws.String("Name"),
								Value: aws.String(strings.Join([]string{"name of", params.InstanceIds[0]}, " ")),
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

type MockEc2ClientGetNodes struct {
	*ec2.Client
}

func (MockEc2ClientGetNodes) DescribeInstances(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error) {
	output := &ec2.DescribeInstancesOutput{
		Reservations: []types.Reservation{},
	}

	//mimic the regex that is happening in the real API call and GetNodes method
	//this mock is a "guard" on the regex pattern in the GetNodes method, and unit tests using it will fail if the regex is changed
	//TODO: a better fixture?
	nodeNames := []string{
		fmt.Sprintf("%s-b-node-name", TEST_CLUSTER_TAG_VALUE),
		fmt.Sprintf("%s-z-node-name", TEST_CLUSTER_TAG_VALUE),
		fmt.Sprintf("%s-a-node-name", TEST_CLUSTER_TAG_VALUE),
		fmt.Sprintf("not-%s-a-node-name", TEST_CLUSTER_TAG_VALUE),
	}

	nodeRegex := fmt.Sprintf("%s%s", "^", params.Filters[0].Values[0])

	for _, nodeName := range nodeNames {
		if match, _ := regexp.MatchString(nodeRegex, nodeName); match {
			output.Reservations = append(
				output.Reservations,
				types.Reservation{
					Instances: []types.Instance{
						{
							InstanceId:   aws.String(fmt.Sprintf("i-%s", nodeName)),
							InstanceType: types.InstanceTypeT3Medium,
							Tags: []types.Tag{
								{
									Key:   aws.String("Name"),
									Value: aws.String(nodeName),
								},
							},
						},
					},
				},
			)
		}
	}

	return output, nil
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
