package talos

import (
	"fmt"
	"github.com/nikogura/k8s-cluster-manager/pkg/manager"
)

// TODO  Talos Machine Config Apply
func ApplyConfig(config manager.ClusterNode) (err error) {
	fmt.Printf("TODO: talos.ApplyConfig() Applying config to %s\n", config.Name)
	return err
}
