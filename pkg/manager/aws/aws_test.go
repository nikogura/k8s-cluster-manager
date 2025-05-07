package aws

import (
	"context"
	"github.com/nikogura/k8s-cluster-manager/pkg/manager"
	"log"
	"os"
	"testing"
)

var awsClusterManager *AWSClusterManager
var ctx context.Context
var awsProfile string
var clusterName string
var tmpDir string

func TestMain(m *testing.M) {
	setUp()

	code := m.Run()

	tearDown()

	os.Exit(code)
}

func setUp() {
	ctx = context.Background()

	tdir, err := os.MkdirTemp("", "k8s-cluster-manager")
	if err != nil {
		log.Fatalf("Error creating temp dir: %s", err)
	}

	tmpDir = tdir

	awsProfile = os.Getenv("AWS_PROFILE")
	clusterName = os.Getenv("CLUSTER_NAME")
	dm := manager.DNSManagerStruct{}

	cm, err := NewAWSClusterManager(ctx, clusterName, awsProfile, dm, true)
	if err != nil {
		log.Fatalf("Couldn't create aws cluster manager: %s", err)
	}

	awsClusterManager = cm

}

func tearDown() {
	if _, err := os.Stat(tmpDir); err == nil {
		os.RemoveAll(tmpDir)
	}
}
