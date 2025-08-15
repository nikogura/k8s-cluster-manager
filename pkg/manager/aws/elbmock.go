package aws

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"github.com/aws/smithy-go/middleware"
	"strings"
)

const TEST_CLUSTER_TAG = "Cluster"
const TEST_CLUSTER_TAG_VALUE = "test-cluster"
const TEST_ELB_CLUSTER_TAG = "Cluster"
const TEST_ELB_CLUSTER_TAG_VALUE = "test-cluster"
const TEST_LOAD_BALANCER_ARN = "arn:aws:elasticloadbalancing:ap-northeast-1:1234567890:loadbalancer/app/test-load-balancer/50dc6c495c0c9188"
const TEST_LOAD_BALANCER_NAME = "not api server"
const TEST_TARGET_GROUP_NAME = "test-targets"
const TEST_TARGET_GROUP_ARN = "arn:aws:elasticloadbalancing:ap-northeast-1:1234567890:targetgroup/test-targets/73e2d6bc24d8a067"
const TEST_TARGET_GROUP_PORT = int32(3333)

type MockELBClient struct {
	// the elasticloadbalancingv2.Client implements the ELBClient interface
	*elasticloadbalancingv2.Client
}

// DescribeLoadBalancers overrides the method of the same name on the elasticloadbalancingv2.Client and returns a specified result
func (MockELBClient) DescribeLoadBalancers(ctx context.Context, params *elasticloadbalancingv2.DescribeLoadBalancersInput, optFns ...func(*elasticloadbalancingv2.Options)) (*elasticloadbalancingv2.DescribeLoadBalancersOutput, error) {
	return &elasticloadbalancingv2.DescribeLoadBalancersOutput{
		LoadBalancers: []types.LoadBalancer{
			{
				LoadBalancerName: aws.String(TEST_LOAD_BALANCER_NAME),
				LoadBalancerArn:  aws.String(TEST_LOAD_BALANCER_ARN),
			},
		},
		NextMarker:     nil,
		ResultMetadata: middleware.Metadata{},
	}, nil
}

// DescribeTags overrides the method of the same name on the elasticloadbalancingv2.Client and returns a specified result
func (MockELBClient) DescribeTags(ctx context.Context, params *elasticloadbalancingv2.DescribeTagsInput, optFns ...func(*elasticloadbalancingv2.Options)) (*elasticloadbalancingv2.DescribeTagsOutput, error) {  
	return &elasticloadbalancingv2.DescribeTagsOutput{
		TagDescriptions: []types.TagDescription{
			{
				ResourceArn: aws.String(TEST_LOAD_BALANCER_ARN),
				Tags: []types.Tag{
					{
						Key:   aws.String(TEST_ELB_CLUSTER_TAG),
						Value: aws.String(TEST_ELB_CLUSTER_TAG_VALUE),
					},
				},
			},
		},
		ResultMetadata: middleware.Metadata{},
	}, nil
}

func (MockELBClient) DescribeTargetGroups(ctx context.Context, params *elasticloadbalancingv2.DescribeTargetGroupsInput, optFns ...func(*elasticloadbalancingv2.Options)) (*elasticloadbalancingv2.DescribeTargetGroupsOutput, error) {
	return &elasticloadbalancingv2.DescribeTargetGroupsOutput{
		TargetGroups: []types.TargetGroup{
			{
				TargetGroupName: aws.String(TEST_TARGET_GROUP_NAME),
				TargetGroupArn:  aws.String(TEST_TARGET_GROUP_ARN),
				Port:            aws.Int32(TEST_TARGET_GROUP_PORT),
			},
		},
	}, nil
}

func (MockELBClient) DescribeTargetHealth(ctx context.Context, params *elasticloadbalancingv2.DescribeTargetHealthInput, optFns ...func(*elasticloadbalancingv2.Options)) (*elasticloadbalancingv2.DescribeTargetHealthOutput, error) {
	return &elasticloadbalancingv2.DescribeTargetHealthOutput{}, nil
}

type MockELBClientNoLB struct {
	// the elasticloadbalancingv2.Client implements the ELBClient interface
	*elasticloadbalancingv2.Client
}

// DescribeLoadBalancers identical to MockELBClient method
func (MockELBClientNoLB) DescribeLoadBalancers(ctx context.Context, params *elasticloadbalancingv2.DescribeLoadBalancersInput, optFns ...func(*elasticloadbalancingv2.Options)) (*elasticloadbalancingv2.DescribeLoadBalancersOutput, error) {
	return &elasticloadbalancingv2.DescribeLoadBalancersOutput{
		LoadBalancers: []types.LoadBalancer{
			{
				LoadBalancerName: aws.String(TEST_LOAD_BALANCER_NAME),
				LoadBalancerArn:  aws.String(TEST_LOAD_BALANCER_ARN),
			},
		},
		NextMarker:     nil,
		ResultMetadata: middleware.Metadata{},
	}, nil
}

func (MockELBClientNoLB) DescribeTags(ctx context.Context, params *elasticloadbalancingv2.DescribeTagsInput, optFns ...func(*elasticloadbalancingv2.Options)) (*elasticloadbalancingv2.DescribeTagsOutput, error) {
	return &elasticloadbalancingv2.DescribeTagsOutput{
		TagDescriptions: []types.TagDescription{
			{
				ResourceArn: aws.String(TEST_LOAD_BALANCER_ARN),
				Tags: []types.Tag{
					{
						Key:   aws.String(TEST_ELB_CLUSTER_TAG),
						Value: aws.String(strings.Join([]string{"not", TEST_ELB_CLUSTER_TAG_VALUE}, " ")),
					},
				},
			},
		},
		ResultMetadata: middleware.Metadata{},
	}, nil
}
