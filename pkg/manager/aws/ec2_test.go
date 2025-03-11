package aws

import (
	"fmt"
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
			"alpha",
			"cp",
			1,
			"alpha-cp-1",
		},
		{
			"alpha",
			"worker",
			1,
			"alpha-worker-1",
		},
		{
			"charlie",
			"worker",
			2,
			"charlie-worker-2",
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
