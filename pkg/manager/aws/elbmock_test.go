package aws

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"github.com/aws/smithy-go/middleware"
)

const TEST_ELB_CLUSTER_TAG = "Cluster"
const TEST_ELB_CLUSTER_TAG_VALUE = "test-cluster"
const TEST_LOAD_BALANCER_ARN = "arn:aws:elasticloadbalancing:ap-northeast-1:1234567890:loadbalancer/app/test-load-balancer/50dc6c495c0c9188"
const TEST_LOAD_BALANCER_NAME = "not api server"
const TEST_TARGET_GROUP_NAME = "test-targets"
const TEST_TARGET_GROUP_ARN = "arn:aws:elasticloadbalancing:ap-northeast-1:1234567890:targetgroup/test-targets/73e2d6bc24d8a067"
const TEST_TARGET_GROUP_PORT = int32(3333)

type mockELBClient struct {
	// the elasticloadbalancingv2.Client implements the ELBClient interface
	*elasticloadbalancingv2.Client
}

// DescribeLoadBalancers overrides the method of the same name on the elasticloadbalancingv2.Client and returns a specified result
func (mockELBClient) DescribeLoadBalancers(ctx context.Context, params *elasticloadbalancingv2.DescribeLoadBalancersInput, optFns ...func(*elasticloadbalancingv2.Options)) (*elasticloadbalancingv2.DescribeLoadBalancersOutput, error) {
	loadBalancerName := TEST_LOAD_BALANCER_NAME
	loadBalancerArn := TEST_LOAD_BALANCER_ARN

	return &elasticloadbalancingv2.DescribeLoadBalancersOutput{
		LoadBalancers: []types.LoadBalancer{
			{
				LoadBalancerName: &loadBalancerName,
				LoadBalancerArn:  &loadBalancerArn,
			},
		},
		NextMarker:     nil,
		ResultMetadata: middleware.Metadata{},
	}, nil
}

// DescribeTags overrides the method of the same name on the elasticloadbalancingv2.Client and returns a specified result
func (mockELBClient) DescribeTags(ctx context.Context, params *elasticloadbalancingv2.DescribeTagsInput, optFns ...func(*elasticloadbalancingv2.Options)) (*elasticloadbalancingv2.DescribeTagsOutput, error) {
	loadBalancerArn := TEST_LOAD_BALANCER_ARN
	tagKey := TEST_ELB_CLUSTER_TAG
	tagValue := TEST_ELB_CLUSTER_TAG_VALUE

	return &elasticloadbalancingv2.DescribeTagsOutput{
		TagDescriptions: []types.TagDescription{
			{
				ResourceArn: &loadBalancerArn,
				Tags: []types.Tag{
					{
						Key:   &tagKey,
						Value: &tagValue,
					},
				},
			},
		},
		ResultMetadata: middleware.Metadata{},
	}, nil
}

func (mockELBClient) DescribeTargetGroups(ctx context.Context, params *elasticloadbalancingv2.DescribeTargetGroupsInput, optFns ...func(*elasticloadbalancingv2.Options)) (*elasticloadbalancingv2.DescribeTargetGroupsOutput, error) {
	targetGroupName := TEST_TARGET_GROUP_NAME
	targetGroupArn := TEST_TARGET_GROUP_ARN
	port := TEST_TARGET_GROUP_PORT

	return &elasticloadbalancingv2.DescribeTargetGroupsOutput{
		TargetGroups: []types.TargetGroup{
			{
				TargetGroupName: &targetGroupName,
				TargetGroupArn:  &targetGroupArn,
				Port:            &port,
			},
		},
	}, nil
}

func (mockELBClient) DescribeTargetHealth(ctx context.Context, params *elasticloadbalancingv2.DescribeTargetHealthInput, optFns ...func(*elasticloadbalancingv2.Options)) (*elasticloadbalancingv2.DescribeTargetHealthOutput, error) {
	return &elasticloadbalancingv2.DescribeTargetHealthOutput{}, nil
}
