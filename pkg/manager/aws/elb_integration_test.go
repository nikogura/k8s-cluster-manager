//go:build integration

package aws

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAWSClusterManager_GetClusterLBs(t *testing.T) {
	testCases := []struct {
		name string
	}{
		{
			"test",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			lbs, err := awsClusterManager.GetClusterLBs()
			if err != nil {
				t.Errorf("failed getting load balancers: %s", err)
			}

			assert.Nil(t, err, "Errors when retrieving load balancers.")

			assert.True(t, len(lbs) > 0, "No load balancers returned")

			for _, lb := range lbs {
				fmt.Printf("LB: %s API: %v\n", lb.Name, lb.IsApiServer)
			}

		})
	}
}

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
