package aws

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAWSClusterManager_GetLB(t *testing.T) {
	testCases := []struct {
		name   string
		lbname string
	}{
		{
			"all",
			"",
		},
		{
			fmt.Sprintf("%s-apiserver", clusterName),
			fmt.Sprintf("%s-apiserver", clusterName),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			out, err := awsClusterManager.GetLB(tc.lbname)
			if err != nil {
				t.Errorf("failed getting aws load balancer %s: %s", tc.name, err)
			}

			assert.Truef(t, len(out.LoadBalancers) >= 1, "load balancer out put fails to meet expectations.")

		})
	}
}

func TestAWSClusterManager_GetTargetGroups(t *testing.T) {
	testCases := []struct {
		name   string
		tgname string
	}{
		{
			"all",
			"",
		},
		{
			fmt.Sprintf("%s-apiserver-6443", clusterName),
			fmt.Sprintf("%s-apiserver-6443", clusterName),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			out, err := awsClusterManager.GetTargetGroups(tc.tgname)
			if err != nil {
				t.Errorf("failed getting aws target group %s: %s", tc.name, err)
			}

			assert.Truef(t, len(out.TargetGroups) >= 1, "load balancer out put fails to meet expectations.")

		})
	}
}

func TestAWSClusterManager_GetTargets(t *testing.T) {
	testCases := []struct {
		name   string
		tgname string
	}{
		// all target will fail if any target group in AWS account has no targets
		//{
		//	"all",
		//	"",
		//},
		{
			fmt.Sprintf("%s-apiserver-6443", clusterName),
			fmt.Sprintf("%s-apiserver-6443", clusterName),
		},
	}

	// TODO re-enable when we have targets being spun up/down and registered.
	t.Skip()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			targets, err := awsClusterManager.GetTargets(tc.tgname)
			if err != nil {
				t.Errorf("failed getting aws load balancer %s: %s", tc.name, err)
			}

			fmt.Printf("Targets(%s):\n", tc.name)
			for _, t := range targets {
				fmt.Printf("ID: %s Port: %d State: %s\n", t.ID, t.Port, t.State)
			}

			assert.Truef(t, len(targets) >= 1, "No targets found for load balancer %s.", tc.tgname)

		})
	}
}

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
