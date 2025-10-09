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

func ApplyConfig(ctx context.Context, node manager.ClusterNode, machineConfigBytes []byte, machineConfigPatches []string, insecure bool, verbose bool) (err error) {
	manager.VerboseOutput(verbose, "Applying config to %s (%s)\n", node.Name(), node.IP())

	tlsConfig := &tls.Config{}

	// The cert on a newly created node won't be trusted, so the initial config apply will need this.
	if insecure {
		tlsConfig.InsecureSkipVerify = true
	}

	// Crude yaml patch to put the node name into the machine config.  Note the spaces (not tabs) cos it's yaml.
	nodeNamePatch := fmt.Sprintf(`machine:
  network:
    hostname: %s.%s
`, node.Name(), node.Domain())

	machineConfigPatches = append(machineConfigPatches, nodeNamePatch)

	// Load config patches
	patches, patchErr := configpatcher.LoadPatches(machineConfigPatches)
	if patchErr != nil {
		err = errors.Wrapf(patchErr, "failed loading config patches")
		return err
	}

	// patch the machine config with things like the node name and other specifics
	cfg, cfgErr := configpatcher.Apply(configpatcher.WithBytes(machineConfigBytes), patches)
	if cfgErr != nil {
		err = errors.Wrapf(cfgErr, "failed applying config patches to machine config ")
		return err
	}

	// Extract the patched config bytes
	cfgBytes, bytesErr := cfg.Bytes()
	if bytesErr != nil {
		err = errors.Wrapf(bytesErr, "failed extracting config bytes")
		return err
	}

	// Create Talos Client
	tClient, clientErr := client.New(ctx, client.WithTLSConfig(tlsConfig), client.WithEndpoints(node.IP()))
	if clientErr != nil {
		err = errors.Wrapf(clientErr, "failed creating new talos client")
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
