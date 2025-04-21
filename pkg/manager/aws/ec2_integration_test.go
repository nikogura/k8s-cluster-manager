//go:build integration

package aws

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAWSClusterManager_GetNode(t *testing.T) {
	testCases := []struct {
		name string
	}{
		{
			"alpha-cp-1",
		},
		//{
		//	fmt.Sprintf("%s-cp-1", clusterName),
		//	fmt.Sprintf("%s-cp-1", clusterName),
		//},
	}

	// TODO re-enable once we're creating EC2 Instances
	//t.Skip()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			out, err := awsClusterManager.GetNode(tc.name)
			if err != nil {
				t.Errorf("failed getting aws node %s: %s", tc.name, err)
			}

			fmt.Printf("Node Name: %s\n", out.Name)

			assert.Truef(t, out.Name == tc.name, "retrieved the wrong node")

		})
	}
}

func TestAWSClusterManager_GetSecurityGroupsForCluster(t *testing.T) {
	testCases := []struct {
		name string
	}{
		{
			"one",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := awsClusterManager.GetSecurityGroupsForCluster()
			if err != nil {
				t.Errorf("Failed getting security groups for cluster %s: %s", tc.name, err)
			}

			//for _, g := range actual {
			//	fmt.Printf("Name: %s  ID: %s\n", *g.GroupName, *g.GroupId)
			//	for _, perm := range g.IpPermissions {
			//		if perm.ToPort != nil {
			//			fmt.Printf("  %d\n", *perm.ToPort)
			//
			//		} else {
			//			fmt.Printf("  All\n")
			//		}
			//	}
			//}

			assert.Truef(t, len(actual) > 0, "No security groups found")

		})
	}

}

func TestAWSClusterManager_GetNodeSecurityGroupsForCluster(t *testing.T) {
	testCases := []struct {
		name string
	}{
		{
			"one",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := awsClusterManager.GetNodeSecurityGroupsForCluster()
			if err != nil {
				t.Errorf("Failed getting security groups for cluster %s: %s", tc.name, err)
			}

			//for _, g := range actual {
			//	fmt.Printf("Name: %s  ID: %s\n", *g.GroupName, *g.GroupId)
			//	for _, perm := range g.IpPermissions {
			//		if perm.ToPort != nil {
			//			fmt.Printf("  %d\n", *perm.ToPort)
			//
			//		} else {
			//			fmt.Printf("  All\n")
			//		}
			//	}
			//}

			assert.Truef(t, len(actual) > 0, "No security groups found")

		})
	}

}
