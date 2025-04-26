package aws

import (
	"github.com/google/go-cmp/cmp"
	"github.com/nikogura/k8s-cluster-manager/pkg/manager"
	"github.com/stretchr/testify/assert"
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

type AWSClusterManager_GetClusterLBsAPI interface {
	GetClusterLBs() (lbs []manager.LBInfo, err error)
}

func AWSClusterManager_GetClusterLBs(api AWSClusterManager_GetClusterLBsAPI) ([]manager.LBInfo, error) {
	object, err := api.GetClusterLBs()
	if err != nil {
		return nil, err
	}

	return object, err
}

type mockAWSClusterManager_GetClusterLBsAPI func() (lbs []manager.LBInfo, err error)

func (m mockAWSClusterManager_GetClusterLBsAPI) GetClusterLBs() (lbs []manager.LBInfo, err error) {
	return m()
}

func TestAWSClusterManager_GetClusterLBs(t *testing.T) {
	testCases := []struct {
		name   string
		client func(t *testing.T) AWSClusterManager_GetClusterLBsAPI
		expect []manager.LBInfo
	}{
		{
			name: "test",
			client: func(t *testing.T) AWSClusterManager_GetClusterLBsAPI {
				return mockAWSClusterManager_GetClusterLBsAPI(
					func() (lbs []manager.LBInfo, err error) {
						t.Helper()
						return []manager.LBInfo{
							{
								Name:        "TestLB",
								IsApiServer: false,
								Targets:     []manager.LBTargetInfo{},
							},
						}, err
					},
				)
			},
			expect: []manager.LBInfo{
				{
					Name:        "TestLB",
					IsApiServer: false,
					Targets:     []manager.LBTargetInfo{},
				},
			},
		},
	}

	for i, tc := range testCases {
		t.Run(strings.Join([]string{strconv.Itoa(i), tc.name}, ". "), func(t *testing.T) {
			got, err := AWSClusterManager_GetClusterLBs(tc.client(t))
			if err != nil {
				t.Fatalf("no error expected with mock, got %v", err)
			}
			if e, g := tc.expect, got; cmp.Equal(e, g) != true {
				t.Errorf("expect %v, got %v", e, g)
			}
		})
	}
}
