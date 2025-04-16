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
