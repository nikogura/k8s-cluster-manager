package aws

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"github.com/aws/smithy-go/middleware"
	"strings"
)

const TestClusterTag = "Cluster"
const TestClusterTagValue = "test-cluster"
const TestELBClusterTag = "Cluster"
const TestELBClusterTagValue = "test-cluster"
const TestLoadBalancerArn = "arn:aws:elasticloadbalancing:ap-northeast-1:1234567890:loadbalancer/app/test-load-balancer/50dc6c495c0c9188"
const TestLoadBalancerNameValue = "not api server"
const TestTargetGroupNameValue = "test-targets"
const TestTargetGroupArnValue = "arn:aws:elasticloadbalancing:ap-northeast-1:1234567890:targetgroup/test-targets/73e2d6bc24d8a067"
const TestTargetGroupPortValue = int32(3333)

type MockELBClient struct {
	// the elasticloadbalancingv2.Client implements the ELBClient interface
	*elasticloadbalancingv2.Client
}

// DescribeLoadBalancers overrides the method of the same name on the elasticloadbalancingv2.Client and returns a specified result.
func (MockELBClient) DescribeLoadBalancers(ctx context.Context, params *elasticloadbalancingv2.DescribeLoadBalancersInput, optFns ...func(*elasticloadbalancingv2.Options)) (output *elasticloadbalancingv2.DescribeLoadBalancersOutput, err error) {
	output = &elasticloadbalancingv2.DescribeLoadBalancersOutput{
		LoadBalancers: []types.LoadBalancer{
			{
				LoadBalancerName: aws.String(TestLoadBalancerNameValue),
				LoadBalancerArn:  aws.String(TestLoadBalancerArn),
			},
		},
		NextMarker:     nil,
		ResultMetadata: middleware.Metadata{},
	}
	return output, err
}

// DescribeTags overrides the method of the same name on the elasticloadbalancingv2.Client and returns a specified result.
func (MockELBClient) DescribeTags(ctx context.Context, params *elasticloadbalancingv2.DescribeTagsInput, optFns ...func(*elasticloadbalancingv2.Options)) (output *elasticloadbalancingv2.DescribeTagsOutput, err error) {
	output = &elasticloadbalancingv2.DescribeTagsOutput{
		TagDescriptions: []types.TagDescription{
			{
				ResourceArn: aws.String(TestLoadBalancerArn),
				Tags: []types.Tag{
					{
						Key:   aws.String(TestELBClusterTag),
						Value: aws.String(TestELBClusterTagValue),
					},
				},
			},
		},
		ResultMetadata: middleware.Metadata{},
	}
	return output, err
}

func (MockELBClient) DescribeTargetGroups(ctx context.Context, params *elasticloadbalancingv2.DescribeTargetGroupsInput, optFns ...func(*elasticloadbalancingv2.Options)) (output *elasticloadbalancingv2.DescribeTargetGroupsOutput, err error) {
	output = &elasticloadbalancingv2.DescribeTargetGroupsOutput{
		TargetGroups: []types.TargetGroup{
			{
				TargetGroupName: aws.String(TestTargetGroupNameValue),
				TargetGroupArn:  aws.String(TestTargetGroupArnValue),
				Port:            aws.Int32(TestTargetGroupPortValue),
			},
		},
	}
	return output, err
}

func (MockELBClient) DescribeTargetHealth(ctx context.Context, params *elasticloadbalancingv2.DescribeTargetHealthInput, optFns ...func(*elasticloadbalancingv2.Options)) (output *elasticloadbalancingv2.DescribeTargetHealthOutput, err error) {
	output = &elasticloadbalancingv2.DescribeTargetHealthOutput{}
	return output, err
}

type MockELBClientNoLB struct {
	// the elasticloadbalancingv2.Client implements the ELBClient interface
	*elasticloadbalancingv2.Client
}

// DescribeLoadBalancers identical to MockELBClient method.
func (MockELBClientNoLB) DescribeLoadBalancers(ctx context.Context, params *elasticloadbalancingv2.DescribeLoadBalancersInput, optFns ...func(*elasticloadbalancingv2.Options)) (output *elasticloadbalancingv2.DescribeLoadBalancersOutput, err error) {
	output = &elasticloadbalancingv2.DescribeLoadBalancersOutput{
		LoadBalancers: []types.LoadBalancer{
			{
				LoadBalancerName: aws.String(TestLoadBalancerNameValue),
				LoadBalancerArn:  aws.String(TestLoadBalancerArn),
			},
		},
		NextMarker:     nil,
		ResultMetadata: middleware.Metadata{},
	}
	return output, err
}

func (MockELBClientNoLB) DescribeTags(ctx context.Context, params *elasticloadbalancingv2.DescribeTagsInput, optFns ...func(*elasticloadbalancingv2.Options)) (output *elasticloadbalancingv2.DescribeTagsOutput, err error) {
	output = &elasticloadbalancingv2.DescribeTagsOutput{
		TagDescriptions: []types.TagDescription{
			{
				ResourceArn: aws.String(TestLoadBalancerArn),
				Tags: []types.Tag{
					{
						Key:   aws.String(TestELBClusterTag),
						Value: aws.String(strings.Join([]string{"not", TestELBClusterTagValue}, " ")),
					},
				},
			},
		},
		ResultMetadata: middleware.Metadata{},
	}
	return output, err
}
