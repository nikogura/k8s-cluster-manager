package cloudflare

import (
	"fmt"
	"github.com/nikogura/k8s-cluster-manager/pkg/manager"
)

// TODO CRUD of DNS A record
func RegisterNode(node manager.ClusterNode) (err error) {
	fmt.Printf("TODO: cloudflare.RegisterNode() Registering DNS for node\n")
	return err
}

func DeRegisterNode(nodeName string) (err error) {
	fmt.Printf("TODO: cloudflare.DeRegisterNode() Registering DNS for node\n")
	return err
}
