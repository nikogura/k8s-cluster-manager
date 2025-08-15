package aws

import (
	"fmt"
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
		name     string
		nodeName string
		acm      AWSClusterManager
		expect   expect
	}{
		{
			name:     "ACM.GetNode() - One Running Instance",
			nodeName: TEST_NODENAME,
			acm: AWSClusterManager{
				Ec2Client:          MockEc2ClientGetNodeOneRunningInst{},
				FetchedNodesByName: make(map[string]manager.NodeInfo),
				FetchedNodesById:   make(map[string]manager.NodeInfo),
			},
			expect: expect{
				manager.NodeInfo{
					Name:         TEST_NODENAME,
					ID:           TEST_INSTANCEID,
					InstanceType: "t3.medium",
				},
				AWSClusterManager{
					Ec2Client: MockEc2ClientGetNodeOneRunningInst{},
					FetchedNodesByName: map[string]manager.NodeInfo{
						TEST_NODENAME: {
							Name:         TEST_NODENAME,
							ID:           TEST_INSTANCEID,
							InstanceType: "t3.medium",
						},
					},
					FetchedNodesById: map[string]manager.NodeInfo{
						TEST_INSTANCEID: {
							Name:         TEST_NODENAME,
							ID:           TEST_INSTANCEID,
							InstanceType: "t3.medium",
						},
					},
				},
			},
		},
		{
			name:     "ACM.GetNode() - Stopped Instance",
			nodeName: TEST_NODENAME,
			acm: AWSClusterManager{
				Ec2Client:          MockEc2ClientGetNodeStoppedInst{},
				FetchedNodesByName: make(map[string]manager.NodeInfo),
				FetchedNodesById:   make(map[string]manager.NodeInfo),
			},
			expect: expect{
				manager.NodeInfo{},
				AWSClusterManager{
					Ec2Client:          MockEc2ClientGetNodeStoppedInst{},
					FetchedNodesByName: make(map[string]manager.NodeInfo),
					FetchedNodesById:   make(map[string]manager.NodeInfo),
				},
			},
		},
		{
			name:     "ACM.GetNode() - No Instance",
			nodeName: TEST_NODENAME,
			acm: AWSClusterManager{
				Ec2Client:          MockEc2ClientGetNodeNoInst{},
				FetchedNodesByName: make(map[string]manager.NodeInfo),
				FetchedNodesById:   make(map[string]manager.NodeInfo),
			},
			expect: expect{
				manager.NodeInfo{},
				AWSClusterManager{
					Ec2Client:          MockEc2ClientGetNodeNoInst{},
					FetchedNodesByName: make(map[string]manager.NodeInfo),
					FetchedNodesById:   make(map[string]manager.NodeInfo),
				},
			},
		},
	}

	for i, tc := range testCases {
		t.Run(strings.Join([]string{strconv.Itoa(i + 1), tc.name}, "."), func(t *testing.T) {
			got, err := tc.acm.GetNode(tc.nodeName)

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

	//this unit test cannot test whether the instance Name tag value is "correct",
	//as it is not part of the input. It *can* be tested in an integration test, however
	nodeName := fmt.Sprintf("name of %s", TEST_INSTANCEID)

	testCases := []struct {
		name   string
		id     string
		acm    AWSClusterManager
		expect expect
	}{
		{
			id:   TEST_INSTANCEID,
			name: "ACM.GetNodeById() - Instance Exists",
			acm: AWSClusterManager{
				Ec2Client:          MockEc2ClientGetNodeByIdInstExists{},
				FetchedNodesByName: make(map[string]manager.NodeInfo),
				FetchedNodesById:   make(map[string]manager.NodeInfo),
			},
			expect: expect{
				manager.NodeInfo{
					Name:         nodeName,
					ID:           TEST_INSTANCEID,
					InstanceType: "t3.medium",
				},
				AWSClusterManager{
					Ec2Client: MockEc2ClientGetNodeByIdInstExists{},
					FetchedNodesByName: map[string]manager.NodeInfo{
						nodeName: {
							Name:         nodeName,
							ID:           TEST_INSTANCEID,
							InstanceType: "t3.medium",
						},
					},
					FetchedNodesById: map[string]manager.NodeInfo{
						TEST_INSTANCEID: {
							Name:         nodeName,
							ID:           TEST_INSTANCEID,
							InstanceType: "t3.medium",
						},
					},
				},
			},
		},
		{
			name: "ACM.GetNodeById() - Instance Does Not Exist",
			acm: AWSClusterManager{
				Ec2Client:          MockEc2ClientGetNodeByIdNoInst{},
				FetchedNodesByName: make(map[string]manager.NodeInfo),
				FetchedNodesById:   make(map[string]manager.NodeInfo),
			},
			expect: expect{
				manager.NodeInfo{},
				AWSClusterManager{
					Ec2Client:          MockEc2ClientGetNodeByIdNoInst{},
					FetchedNodesByName: map[string]manager.NodeInfo{},
					FetchedNodesById:   map[string]manager.NodeInfo{},
				},
			},
		},
	}

	for i, tc := range testCases {
		t.Run(strings.Join([]string{strconv.Itoa(i + 1), tc.name}, "."), func(t *testing.T) {
			got, err := tc.acm.GetNodeById(tc.id)

			if err != nil {
				t.Fatalf("no error expected with mocks, got %+v", err)
			}
			if e, g := tc.expect.ni, got; reflect.DeepEqual(e, g) != true {
				eStr := prettyPrint(e)
				gStr := prettyPrint(g)
				t.Errorf("\nmanager.NodeInfo\n expect:\n%s\n got:\n%s\n", eStr, gStr)
			}
			if e, g := tc.expect.acm, tc.acm; reflect.DeepEqual(e, g) != true {
				eStr := prettyPrint(e)
				gStr := prettyPrint(g)
				t.Logf("\nAWSClusterManager:\n expect:\n%s\n got:\n%s\n", eStr, gStr)
				t.Errorf("\nexpansion of AWSClusterManager.FetchedNodesByName:\n expect:\n%s\n got:\n%s\n", prettyPrintMap(tc.expect.acm.FetchedNodesByName), prettyPrintMap(tc.acm.FetchedNodesByName))
			}
		})
	}
}

func TestGetNodes(t *testing.T) {
	type expect struct {
		ni  []manager.NodeInfo
		acm AWSClusterManager
	}

	testCases := []struct {
		name        string
		clusterName string
		acm         AWSClusterManager
		expect      expect
	}{
		{
			name:        "ACM.GetNodes()",
			clusterName: TEST_CLUSTER_TAG_VALUE,
			acm: AWSClusterManager{
				Ec2Client: MockEc2ClientGetNodes{},
			},
			//Note: there is a dependency between this expected result and the mocked method MockEc2ClientGetNodes
			//the output of the mock must match the below; a real API invocation would return an unknown number of instances,
			//so to test the sorting in the GetNodes method, we need to have "both sides" agree
			expect: expect{
				[]manager.NodeInfo{
					{
						Name:         fmt.Sprintf("%s-a-node-name", TEST_CLUSTER_TAG_VALUE),
						ID:           fmt.Sprintf("i-%s-a-node-name", TEST_CLUSTER_TAG_VALUE),
						InstanceType: "t3.medium",
					},
					{
						Name:         fmt.Sprintf("%s-b-node-name", TEST_CLUSTER_TAG_VALUE),
						ID:           fmt.Sprintf("i-%s-b-node-name", TEST_CLUSTER_TAG_VALUE),
						InstanceType: "t3.medium",
					},
					{
						Name:         fmt.Sprintf("%s-z-node-name", TEST_CLUSTER_TAG_VALUE),
						ID:           fmt.Sprintf("i-%s-z-node-name", TEST_CLUSTER_TAG_VALUE),
						InstanceType: "t3.medium",
					},
				},
				AWSClusterManager{
					Ec2Client: MockEc2ClientGetNodes{},
				},
			},
		},
	}

	for i, tc := range testCases {
		t.Run(strings.Join([]string{strconv.Itoa(i + 1), tc.name}, "."), func(t *testing.T) {
			got, err := tc.acm.GetNodes(tc.clusterName)
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
				Name:      fmt.Sprintf("not %s", TEST_EC2_SG_TAG_VALUE),
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
