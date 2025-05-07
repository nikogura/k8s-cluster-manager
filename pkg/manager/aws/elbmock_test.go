package aws

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"github.com/aws/smithy-go/middleware"
	"github.com/nikogura/k8s-cluster-manager/pkg/manager"
	//"github.com/nikogura/k8s-cluster-manager/pkg/manager/aws"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

const TEST_ELB_CLUSTER_TAG = "Cluster"
const TEST_ELB_CLUSTER_TAG_VALUE = "test-cluster"
const TEST_LOAD_BALANCER_ARN = "arn:aws:elb:ap-northeast-1:1234567890:test-load-balancer"
const TEST_LOAD_BALANCER_NAME = "not api server"
const TEST_TARGET_GROUP_NAME = "test-target-group"

// TODO: verify format of target group arn
const TEST_TARGET_GROUP_ARN = "arn:aws:elb:ap-northeast-1:1234567890:test-target-group"
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

func TestAWSClusterManager_GetClusterLBs(t *testing.T) {
	testCases := []struct {
		name   string
		acm    AWSClusterManager
		expect []manager.LBInfo
	}{
		{
			// test case
			name: "Get Cluster LBs",
			acm: AWSClusterManager{
				Name:      TEST_ELB_CLUSTER_TAG_VALUE,
				ELBClient: mockELBClient{},
			},
			// expected results from test case
			expect: []manager.LBInfo{
				{
					Name:        TEST_LOAD_BALANCER_NAME,
					IsApiServer: false,
					//TODO: add mock target info to test case when enabling acm.GetTargets()
					Targets: []manager.LBTargetInfo{
						//	{
						//		ID:   TEST_TARGET_GROUP_NAME,
						//		Port: TEST_TARGET_GROUP_PORT,
						//	},
					},
					TargetGroups: []manager.LBTargetGroupInfo{
						{
							Name: TEST_TARGET_GROUP_NAME,
							Arn:  TEST_TARGET_GROUP_ARN,
							Port: TEST_TARGET_GROUP_PORT,
						},
					},
				},
			},
		},
	}

	for i, tc := range testCases {
		t.Run(strings.Join([]string{strconv.Itoa(i + 1), tc.name}, "."), func(t *testing.T) {
			// you're testing the GetClusterLBs() method of the AWSClusterManager object, but you're mocking
			// two methods of the elasticloadbalancingv2.Client object inside that one test; the blog post
			// only shows one mocked method of the external API per tested method of the top level object
			got, err := tc.acm.GetClusterLBs()
			if err != nil {
				t.Fatalf("no error expected with mocks, got %v", err)
			}
			if e, g := tc.expect, got; reflect.DeepEqual(e, g) != true {
				t.Errorf("expect %v, got %v", e, g)
			}
		})
	}
}
