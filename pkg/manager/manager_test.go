package manager

import (
	"github.com/pkg/errors"
	"log"
	"os"
	"testing"
)

//nolint:gochecknoglobals // Test globals
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
	dir, mkdirErr := os.MkdirTemp("", "k8s-cluster-manager")
	if mkdirErr != nil {
		err = errors.Wrapf(mkdirErr, "Error creating temp dir %q", tmpDir)
		return err
	}
	tmpDir = dir

	return err
}

func tearDown() {

}
