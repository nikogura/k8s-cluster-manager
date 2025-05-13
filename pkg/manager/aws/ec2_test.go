package aws

import (
	"fmt"
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
			name: "Get Node - One Running Instance",
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
			name: "Get Node - Stopped Instance",
			acm: AWSClusterManager{
				Ec2Client:          MockEc2ClientOneRunningInst{},
				FetchedNodesByName: make(map[string]manager.NodeInfo),
				FetchedNodesById:   make(map[string]manager.NodeInfo),
			},
			expect: expect{
				manager.NodeInfo{},
				AWSClusterManager{
					Ec2Client:          MockEc2ClientOneRunningInst{},
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
				t.Errorf("\n expect:\n%s\n got:\n%s", eStr, gStr)
			}
			if e, g := tc.expect.acm, tc.acm; reflect.DeepEqual(e, g) != true {
				eStr := prettyPrint(e)
				gStr := prettyPrint(g)
				t.Logf("\n expect:\n%s\n got:\n%s", eStr, gStr)
				t.Errorf("\n maps:\n expect:\n%s\n got:\n%s", prettyPrintMap(tc.expect.acm.FetchedNodesByName), prettyPrintMap(tc.acm.FetchedNodesByName))
			}
		})
	}
}

func prettyPrint[T any](str T) string {
	s := reflect.ValueOf(&str).Elem()
	typeOf := s.Type()
	pretty := ""
	for i := 0; i < s.NumField(); i++ {
		f := s.Field(i)
		pretty = strings.Join([]string{pretty, fmt.Sprintf("%d: %s %s = %v\n", i,
			typeOf.Field(i).Name, f.Type(), f.Interface())}, "")
	}

	return pretty
}

func prettyPrintMap[T map[string]manager.NodeInfo](str T) string {
	pretty := ""
	i := 0
	for k, v := range str {
		pretty = strings.Join([]string{pretty, fmt.Sprintf("%d: %s =\n-\n|\n%s\n", i,
			k, prettyPrint(v))}, "")
		i++
	}

	return pretty
}
