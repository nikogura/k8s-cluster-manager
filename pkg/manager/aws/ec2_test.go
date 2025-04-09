package aws

import (
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLoadNodeName(t *testing.T) {
	cases := []struct {
		name     string
		nodeType string
		index    int
		expected string
	}{
		{
			"foo",
			"cp",
			1,
			"foo-cp-1",
		},
		{
			"bar",
			"worker",
			1,
			"bar-worker-1",
		},
		{
			"baz",
			"worker",
			2,
			"baz-worker-2",
		},
	}

	for _, tc := range cases {
		actual, err := NodeName(tc.name, tc.nodeType, tc.index)
		if err != nil {
			t.Errorf("failed generating lb name: %s", err)
		}
		assert.Equal(t, tc.expected, actual, "actual load balancer name does not meet expectations")
	}
}

func TestAWSClusterManager_GetNode(t *testing.T) {
	testCases := []struct {
		name string
	}{
		{
			"foo-cp-1",
		},
		//{
		//	fmt.Sprintf("%s-cp-1", clusterName),
		//	fmt.Sprintf("%s-cp-1", clusterName),
		//},
	}

	// TODO re-enable once we're creating EC2 Instances
	t.Skip()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			out, err := awsClusterManager.GetNode(tc.name)
			if err != nil {
				t.Errorf("failed getting aws node %s: %s", tc.name, err)
			}

			fmt.Printf("Node Name: %s\n", out.Name)
			spew.Dump(out)

			assert.Truef(t, out.Name == tc.name, "retrieved the wrong node")

		})
	}
}
