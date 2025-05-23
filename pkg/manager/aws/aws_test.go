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
		log.Fatalf("error creating temp dir: %s", err)
	}

	tmpDir = tdir

}

func tearDown() {
	if _, err := os.Stat(tmpDir); err == nil {
		if err := os.RemoveAll(tmpDir); err != nil {
			log.Printf("couldn't remove temp dir: %s", err)
		}
	} else {
		log.Printf("couldn't find temp dir: %s", err)
	}
}
