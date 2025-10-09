package aws

import (
	"context"
	"log"
	"os"
	"testing"
)

//nolint:gochecknoglobals,unused // Test globals
var awsClusterManager *AWSClusterManager

//nolint:gochecknoglobals // Test globals
var ctx context.Context

//nolint:gochecknoglobals,unused // Test globals
var awsProfile string

//nolint:gochecknoglobals,unused // Test globals
var clusterName string

//nolint:gochecknoglobals // Test globals
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
	_, statErr := os.Stat(tmpDir)
	if statErr == nil {
		rmErr := os.RemoveAll(tmpDir)
		if rmErr != nil {
			log.Printf("couldn't remove temp dir: %s", rmErr)
		}
	} else {
		log.Printf("couldn't find temp dir: %s", statErr)
	}
}
