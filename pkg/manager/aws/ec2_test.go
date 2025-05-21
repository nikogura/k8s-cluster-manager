package aws

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
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

func TestGetNodeById(t *testing.T) {
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
			name: "ACM.GetNodeById()",
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
	}

	for i, tc := range testCases {
		t.Run(strings.Join([]string{strconv.Itoa(i + 1), tc.name}, "."), func(t *testing.T) {
			got, err := tc.acm.GetNodeById(INSTANCEID)

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

func TestGetSecurityGroupsForCluster(t *testing.T) {
	type expect struct {
		gp []types.SecurityGroup
	}

	testCases := []struct {
		name   string
		acm    AWSClusterManager
		expect expect
	}{
		{
			name: "ACM.GetSecurityGroupsForCluster() - One Security Group",
			acm: AWSClusterManager{
				Name:      TEST_EC2_SG_TAG_VALUE,
				Ec2Client: MockEc2ClientOneSecurityGroup{},
			},
			expect: expect{
				[]types.SecurityGroup{
					{
						Tags: []types.Tag{
							{
								Key:   aws.String(TEST_EC2_SG_TAG),
								Value: aws.String(TEST_EC2_SG_TAG_VALUE),
							},
						},
					},
				},
			},
		},
		{
			name: "ACM.GetSecurityGroupsForCluster() - No Security Groups",
			acm: AWSClusterManager{
				Name:      strings.Join([]string{"not", TEST_EC2_SG_TAG_VALUE}, " "),
				Ec2Client: MockEc2ClientNoSecurityGroups{},
			},
			expect: expect{
				nil,
			},
		},
	}

	for i, tc := range testCases {
		t.Run(strings.Join([]string{strconv.Itoa(i + 1), tc.name}, "."), func(t *testing.T) {
			got, err := tc.acm.GetSecurityGroupsForCluster()

			if err != nil {
				t.Fatalf("no error expected with mocks, got %+v", err)
			}
			if e, g := tc.expect.gp, got; reflect.DeepEqual(e, g) != true {
				eStr := prettyPrint(e)
				gStr := prettyPrint(g)
				t.Errorf("\n expect:\n%s\n got:\n%s\n", eStr, gStr)
			}
		})
	}
}

func TestGetNodeSecurityGroupsForCluster(t *testing.T) {
	type expect struct {
		gp []types.SecurityGroup
	}

	testCases := []struct {
		name   string
		acm    AWSClusterManager
		expect expect
	}{
		{
			name: "ACM.GetNodeSecurityGroupsForCluster() - One Security Group",
			acm: AWSClusterManager{
				Name:      TEST_EC2_SG_TAG_VALUE,
				Ec2Client: MockEc2ClientOneNodeSecurityGroup{},
			},
			expect: expect{
				[]types.SecurityGroup{
					{
						Tags: []types.Tag{
							{
								Key:   aws.String(TEST_EC2_SG_TAG),
								Value: aws.String(TEST_EC2_SG_TAG_VALUE),
							},
						},
						IpPermissions: []types.IpPermission{
							{
								ToPort: aws.Int32(TALOS_CONTROL_PORT),
							},
						},
					},
				},
			},
		},
		{
			name: "ACM.GetNodeSecurityGroupsForCluster() - Cluster Security Group but No Node Security Groups",
			acm: AWSClusterManager{
				Name:      TEST_EC2_SG_TAG_VALUE,
				Ec2Client: MockEc2ClientNoNodeSecurityGroup{},
			},
			expect: expect{
				nil,
			},
		},
	}

	for i, tc := range testCases {
		t.Run(strings.Join([]string{strconv.Itoa(i + 1), tc.name}, "."), func(t *testing.T) {
			got, err := tc.acm.GetNodeSecurityGroupsForCluster()

			if err != nil {
				t.Fatalf("no error expected with mocks, got %+v", err)
			}
			if e, g := tc.expect.gp, got; reflect.DeepEqual(e, g) != true {
				eStr := prettyPrint(e)
				gStr := prettyPrint(g)
				t.Errorf("\n expect:\n%s\n got:\n%s\n", eStr, gStr)
			}
		})
	}
}
