//go:build integration

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
		log.Fatalf("error creating temp dir: %s", err)
	}

	tmpDir = tdir

	var awsProfile string
	if ap, ok := os.LookupEnv("AWS_PROFILE"); ok {
		awsProfile = ap
	} //else {
	//	log.Printf("%s not set\n", "AWS_PROFILE")
	//	awsProfile = "blah"
	//}

	var clusterName string
	if cn, ok := os.LookupEnv("CLUSTER_NAME"); ok {
		clusterName = cn
	} //else {
	//	log.Printf("%s not set\n", "CLUSTER_NAME")
	//	clusterName = "blee"
	//}

	dnsManager := manager.DNSManagerStruct{}

	cm, err := NewAWSClusterManager(ctx, clusterName, awsProfile, dnsManager, true)
	if err != nil {
		log.Fatalf("couldn't create aws cluster manager: %s", err)
	}

	awsClusterManager = cm

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
