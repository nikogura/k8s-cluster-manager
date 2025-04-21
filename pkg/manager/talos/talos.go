package talos

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/nikogura/k8s-cluster-manager/pkg/manager"
	"github.com/pkg/errors"
	machineapi "github.com/siderolabs/talos/pkg/machinery/api/machine"
	"github.com/siderolabs/talos/pkg/machinery/client"
	"google.golang.org/protobuf/types/known/durationpb"
	"time"
)

func ApplyConfig(ctx context.Context, node manager.ClusterNode, machineConfigBytes []byte, insecure bool) (err error) {
	fmt.Printf("Applying config to %s (%s)\n", node.Name(), node.IP())

	tlsConfig := &tls.Config{}

	// The cert on the newly created node is not trusted.
	if insecure {
		tlsConfig.InsecureSkipVerify = true
	}

	tClient, err := client.New(ctx, client.WithTLSConfig(tlsConfig), client.WithEndpoints(node.IP()))
	if err != nil {
		err = errors.Wrapf(err, "failed creating new talos client")
		return err
	}

	defer tClient.Close()

	timeout := durationpb.New(300 * time.Second)

	req := &machineapi.ApplyConfigurationRequest{
		Data:           machineConfigBytes,
		Mode:           machineapi.ApplyConfigurationRequest_AUTO,
		DryRun:         false,
		TryModeTimeout: timeout,
	}

	_, err = tClient.ApplyConfiguration(ctx, req)
	if err != nil {
		err = errors.Wrapf(err, "failed applying configuration to %s at %s", node.Name(), node.IP())
		return err
	}

	return err
}
