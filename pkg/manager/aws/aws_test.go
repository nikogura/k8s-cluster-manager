package aws

import (
	"context"
	"log"
	"os"
	"testing"
)

var awsClusterManager *AWSClusterManager
var ctx context.Context
var awsProfile string
var clusterName string

func TestMain(m *testing.M) {
	setUp()

	code := m.Run()

	tearDown()

	os.Exit(code)
}

func setUp() {
	ctx = context.Background()

	awsProfile = os.Getenv("AWS_PROFILE")
	clusterName = os.Getenv("CLUSTER_NAME")

	cm, err := NewAWSClusterManager(ctx, clusterName, awsProfile)
	if err != nil {
		log.Fatalf("Couldn't create aws cluster manager: %s", err)
	}

	awsClusterManager = cm

}

func tearDown() {
}
