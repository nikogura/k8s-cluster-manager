package talos

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/nikogura/k8s-cluster-manager/pkg/manager"
	"github.com/pkg/errors"
	machineapi "github.com/siderolabs/talos/pkg/machinery/api/machine"
	"github.com/siderolabs/talos/pkg/machinery/client"
	"github.com/siderolabs/talos/pkg/machinery/config/configpatcher"
)

func ApplyConfig(ctx context.Context, node manager.ClusterNode, machineConfigBytes []byte, machineConfigPatchBytes []string, insecure bool) (err error) {
	fmt.Printf("Applying config to %s (%s)\n", node.Name(), node.IP())

	tlsConfig := &tls.Config{}

	// The cert on a newly created node won't be trusted, so the initial config apply will need this.
	if insecure {
		tlsConfig.InsecureSkipVerify = true
	}

	// Load config patches
	patches, err := configpatcher.LoadPatches(machineConfigPatchBytes)
	if err != nil {
		err = errors.Wrapf(err, "failed loading config patches")
		return err
	}

	// patch the machine config with things like the node name and other specifics
	cfg, err := configpatcher.Apply(configpatcher.WithBytes(machineConfigBytes), patches)
	if err != nil {
		err = errors.Wrapf(err, "failed applying config patches to machine config ")
		return err
	}

	// Extract the patched config bytes
	cfgBytes, err := cfg.Bytes()
	if err != nil {
		err = errors.Wrapf(err, "failed extracting config bytes")
		return err
	}

	// Create Talos Client
	tClient, err := client.New(ctx, client.WithTLSConfig(tlsConfig), client.WithEndpoints(node.IP()))
	if err != nil {
		err = errors.Wrapf(err, "failed creating new talos client")
		return err
	}

	defer tClient.Close()

	// Create apply config request
	req := &machineapi.ApplyConfigurationRequest{
		Data:   cfgBytes,
		Mode:   machineapi.ApplyConfigurationRequest_AUTO,
		DryRun: false,
	}

	// Actually apply the config.
	_, err = tClient.ApplyConfiguration(ctx, req)
	if err != nil {
		err = errors.Wrapf(err, "failed applying machine configuration to %s at %s", node.Name(), node.IP())
		return err
	}

	return err
}
