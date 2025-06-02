package aws

import (
	"github.com/nikogura/k8s-cluster-manager/pkg/manager"
	"github.com/stretchr/testify/assert"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func TestTargetGroupName(t *testing.T) {
	cases := []struct {
		name     string
		tls      bool
		expected string
	}{
		{
			"foo",
			false,
			"ingress-foo-clear",
		},
		{
			"bar",
			true,
			"ingress-bar-tls",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			actual := TargetGroupName(tc.name, tc.tls)
			assert.Equal(t, tc.expected, actual, "actual target group name fails to meet expectations")
		})
	}

}

func TestLoadBalancerName(t *testing.T) {
	cases := []struct {
		name     string
		lbType   string
		expected string
	}{
		{
			"foo",
			"apiserver",
			"apiserver-foo",
		},
		{
			"bar",
			"int",
			"ingress-bar",
		},
		{
			"baz",
			"ext",
			"ingress-baz-ext",
		},
	}

	for _, tc := range cases {
		actual, err := LoadBalancerName(tc.name, tc.lbType)
		if err != nil {
			t.Errorf("failed generating lb name: %s", err)
		}
		assert.Equal(t, tc.expected, actual, "actual load balancer name does not meet expectations")
	}
}

func TestAWSClusterManager_GetClusterLBs(t *testing.T) {
	testCases := []struct {
		name   string
		acm    AWSClusterManager
		expect []manager.LBInfo
	}{
		{
			// test case
			name: "ACM.GetClusterLBs() - One Cluster LB",
			acm: AWSClusterManager{
				Name:      TEST_ELB_CLUSTER_TAG_VALUE,
				ELBClient: MockELBClient{},
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
		{
			//TODO: debug shows the test case is "len 0", but real function returns nil
			name: "ACM.GetClusterLBs() - No Cluster LBs",
			acm: AWSClusterManager{
				Name:      TEST_ELB_CLUSTER_TAG_VALUE,
				ELBClient: MockELBClientNoLB{},
			},
			// expected results from test case
			expect: nil,
		},
	}

	for i, tc := range testCases {
		t.Run(strings.Join([]string{strconv.Itoa(i + 1), tc.name}, "."), func(t *testing.T) {
			// you're testing the GetClusterLBs() method of the AWSClusterManager object, but you're mocking
			// two methods of the elasticloadbalancingv2.Client object inside that one test
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
