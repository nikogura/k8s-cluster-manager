package talos

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

	"github.com/nikogura/k8s-cluster-manager/pkg/manager"
	"github.com/pkg/errors"
	"github.com/siderolabs/talos/pkg/machinery/client"
)

const (
	defaultUpgradeTimeout    = 15 * time.Minute
	defaultHealthCheckPeriod = 10 * time.Second
	defaultHealthCheckWait   = 5 * time.Minute
)

// UpgradeNode upgrades a single Talos node to the specified version.
func UpgradeNode(ctx context.Context, node manager.ClusterNode, installerImage string, preserve bool, stage bool, verbose bool) (err error) {
	manager.VerboseOutput(verbose, "Upgrading node %s to %s\n", node.Name(), installerImage)

	tlsConfig := &tls.Config{
		InsecureSkipVerify: false, //nolint:gosec // False positive - this is explicitly false
	}

	// Create Talos Client
	tClient, clientErr := client.New(ctx, client.WithTLSConfig(tlsConfig), client.WithEndpoints(node.IP()))
	if clientErr != nil {
		err = errors.Wrapf(clientErr, "failed creating talos client for %s", node.Name())
		return err
	}

	defer tClient.Close()

	manager.VerboseOutput(verbose, "Sending upgrade request to %s\n", node.Name())

	// Execute upgrade - Talos v1.11.3 signature: Upgrade(ctx, image, preserve, stage, ...grpc.CallOption)
	_, upgradeErr := tClient.Upgrade(ctx, installerImage, preserve, stage)
	if upgradeErr != nil {
		err = errors.Wrapf(upgradeErr, "failed upgrading node %s to %s", node.Name(), installerImage)
		return err
	}

	manager.VerboseOutput(verbose, "Upgrade request sent successfully to %s\n", node.Name())

	return err
}

// WaitForNodeHealthy waits for a node to become healthy after upgrade.
func WaitForNodeHealthy(ctx context.Context, node manager.ClusterNode, timeout time.Duration, verbose bool) (err error) {
	manager.VerboseOutput(verbose, "Waiting for node %s to become healthy (timeout: %s)\n", node.Name(), timeout)

	deadline, hasDeadline := ctx.Deadline()
	if hasDeadline {
		remaining := time.Until(deadline)
		if remaining < timeout {
			timeout = remaining
		}
	}

	healthCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(defaultHealthCheckPeriod)
	defer ticker.Stop()

	for range ticker.C {
		select {
		case <-healthCtx.Done():
			err = fmt.Errorf("timeout waiting for node %s to become healthy after %s", node.Name(), timeout)
			return err
		default:
			// Check if node is accessible
			healthy, checkErr := isNodeHealthy(healthCtx, node, verbose)
			if checkErr != nil {
				manager.VerboseOutput(verbose, "Health check error for %s: %v (will retry)\n", node.Name(), checkErr)
				continue
			}

			if healthy {
				manager.VerboseOutput(verbose, "Node %s is healthy\n", node.Name())
				return err
			}

			manager.VerboseOutput(verbose, "Node %s not yet healthy, waiting...\n", node.Name())
		}
	}

	return err
}

func isNodeHealthy(ctx context.Context, node manager.ClusterNode, verbose bool) (healthy bool, err error) {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: false, //nolint:gosec // False positive - this is explicitly false
	}

	// Create Talos Client
	tClient, clientErr := client.New(ctx, client.WithTLSConfig(tlsConfig), client.WithEndpoints(node.IP()))
	if clientErr != nil {
		err = errors.Wrapf(clientErr, "failed creating talos client")
		return healthy, err
	}

	defer tClient.Close()

	// Get node version to verify it's responsive
	_, versionErr := tClient.Version(ctx)
	if versionErr != nil {
		manager.VerboseOutput(verbose, "Node %s not responding to version check: %v\n", node.Name(), versionErr)
		return healthy, err
	}

	healthy = true
	return healthy, err
}

// VerifyNodeVersion checks that a node is running the expected version.
func VerifyNodeVersion(ctx context.Context, node manager.ClusterNode, expectedVersion string, verbose bool) (matches bool, actualVersion string, err error) {
	manager.VerboseOutput(verbose, "Verifying node %s is running version %s\n", node.Name(), expectedVersion)

	tlsConfig := &tls.Config{
		InsecureSkipVerify: false, //nolint:gosec // False positive - this is explicitly false
	}

	// Create Talos Client
	tClient, clientErr := client.New(ctx, client.WithTLSConfig(tlsConfig), client.WithEndpoints(node.IP()))
	if clientErr != nil {
		err = errors.Wrapf(clientErr, "failed creating talos client")
		return matches, actualVersion, err
	}

	defer tClient.Close()

	// Get node version
	versionResp, versionErr := tClient.Version(ctx)
	if versionErr != nil {
		err = errors.Wrapf(versionErr, "failed getting version from node %s", node.Name())
		return matches, actualVersion, err
	}

	if len(versionResp.GetMessages()) == 0 {
		err = fmt.Errorf("no version messages returned from node %s", node.Name())
		return matches, actualVersion, err
	}

	actualVersion = versionResp.GetMessages()[0].GetVersion().GetTag()
	matches = actualVersion == expectedVersion

	manager.VerboseOutput(verbose, "Node %s version: %s (expected: %s, matches: %t)\n", node.Name(), actualVersion, expectedVersion, matches)

	return matches, actualVersion, err
}
