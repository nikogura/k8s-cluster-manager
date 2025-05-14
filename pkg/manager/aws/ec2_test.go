package aws

import (
	"github.com/nikogura/k8s-cluster-manager/pkg/manager"
	"github.com/stretchr/testify/assert"
	"reflect"
	"strconv"
	"strings"
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

func TestGetNode(t *testing.T) {
	type expect struct {
		ni  manager.NodeInfo
		acm AWSClusterManager
	}

	testCases := []struct {
		name   string
		acm    AWSClusterManager
		expect expect
	}{
		{
			name: "ACM.GetNode() - One Running Instance",
			acm: AWSClusterManager{
				Ec2Client:          MockEc2ClientOneRunningInst{},
				FetchedNodesByName: make(map[string]manager.NodeInfo),
				FetchedNodesById:   make(map[string]manager.NodeInfo),
			},
			expect: expect{
				manager.NodeInfo{
					Name: NODENAME,
					ID:   INSTANCEID,
				},
				AWSClusterManager{
					Ec2Client: MockEc2ClientOneRunningInst{},
					FetchedNodesByName: map[string]manager.NodeInfo{
						NODENAME: {
							Name: NODENAME,
							ID:   INSTANCEID,
						},
					},
					FetchedNodesById: map[string]manager.NodeInfo{
						INSTANCEID: {
							Name: NODENAME,
							ID:   INSTANCEID,
						},
					},
				},
			},
		},
		{
			name: "ACM.GetNode() - Stopped Instance",
			acm: AWSClusterManager{
				Ec2Client:          MockEc2ClientStoppedInst{},
				FetchedNodesByName: make(map[string]manager.NodeInfo),
				FetchedNodesById:   make(map[string]manager.NodeInfo),
			},
			expect: expect{
				manager.NodeInfo{},
				AWSClusterManager{
					Ec2Client:          MockEc2ClientStoppedInst{},
					FetchedNodesByName: make(map[string]manager.NodeInfo),
					FetchedNodesById:   make(map[string]manager.NodeInfo),
				},
			},
		},
	}

	for i, tc := range testCases {
		t.Run(strings.Join([]string{strconv.Itoa(i + 1), tc.name}, "."), func(t *testing.T) {
			got, err := tc.acm.GetNode(NODENAME)

			if err != nil {
				t.Fatalf("no error expected with mocks, got %+v", err)
			}
			if e, g := tc.expect.ni, got; reflect.DeepEqual(e, g) != true {
				eStr := prettyPrint(e)
				gStr := prettyPrint(g)
				t.Errorf("\n expect:\n%s\n got:\n%s\n", eStr, gStr)
			}
			if e, g := tc.expect.acm, tc.acm; reflect.DeepEqual(e, g) != true {
				eStr := prettyPrint(e)
				gStr := prettyPrint(g)
				t.Logf("\n expect:\n%s\n got:\n%s\n", eStr, gStr)
				t.Errorf("\n expect (maps):\n%s\n got:\n%s\n", prettyPrintMap(tc.expect.acm.FetchedNodesByName), prettyPrintMap(tc.acm.FetchedNodesByName))
			}
		})
	}
}
