package manager

import (
	"github.com/pkg/errors"
	"log"
	"os"
	"testing"
)

var tmpDir string

func TestMain(m *testing.M) {
	err := setUp()
	if err != nil {
		log.Fatalf("Setup Failed: %s", err)
	}

	code := m.Run()

	tearDown()

	os.Exit(code)
}

func setUp() (err error) {
	dir, err := os.MkdirTemp("", "k8s-cluster-manager")
	if err != nil {
		err = errors.Wrapf(err, "Error creating temp dir %q", tmpDir)
		return err
	}
	tmpDir = dir

	return err
}

func tearDown() {

}
