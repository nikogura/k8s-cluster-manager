package kubernetes

import (
	"context"
	"github.com/nikogura/k8s-cluster-manager/pkg/manager"
	k8s_utility_client "github.com/nikogura/k8s-utility-client/pkg/k8s-utility-client"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

func DeleteNode(ctx context.Context, nodeName string, verbose bool) (err error) {
	manager.VerboseOutput(verbose, "Deleting node %s from k8s\n", nodeName)

	client, clientErr := k8s_utility_client.NewK8sClients()
	if clientErr != nil {
		err = errors.Wrapf(clientErr, "failed creating k8s clients")
		return err
	}

	err = client.ClientSet.CoreV1().Nodes().Delete(ctx, nodeName, metav1.DeleteOptions{})
	if err != nil {
		err = errors.Wrapf(err, "failed deleting node %s", nodeName)
		return err
	}

	return err
}

// WaitForNodeReady waits for a node to register in Kubernetes and become Ready.
func WaitForNodeReady(ctx context.Context, nodeName string, timeout time.Duration, verbose bool) (err error) {
	manager.VerboseOutput(verbose, "Waiting for node %s to register in Kubernetes (timeout: %v)\n", nodeName, timeout)

	client, clientErr := k8s_utility_client.NewK8sClients()
	if clientErr != nil {
		err = errors.Wrapf(clientErr, "failed creating k8s clients")
		return err
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-timeoutCtx.Done():
			err = errors.Errorf("timeout waiting for node %s to become ready after %v", nodeName, timeout)
			return err
		case <-ticker.C:
			node, getErr := client.ClientSet.CoreV1().Nodes().Get(ctx, nodeName, metav1.GetOptions{})
			if getErr != nil {
				manager.VerboseOutput(verbose, "Node %s not found yet, continuing to wait...\n", nodeName)
				continue
			}

			// Check if node is Ready
			ready := isNodeReady(node)
			if ready {
				manager.VerboseOutput(verbose, "Node %s is ready\n", nodeName)
				return err
			}

			manager.VerboseOutput(verbose, "Node %s found but not ready yet, continuing to wait...\n", nodeName)
		}
	}
}

func isNodeReady(node *corev1.Node) (ready bool) {
	for _, condition := range node.Status.Conditions {
		if condition.Type == corev1.NodeReady {
			ready = condition.Status == corev1.ConditionTrue
			return ready
		}
	}
	ready = false
	return ready
}

// ListNodes returns a list of all node names in Kubernetes.
func ListNodes(ctx context.Context, verbose bool) (nodeNames []string, err error) {
	manager.VerboseOutput(verbose, "Listing all nodes from Kubernetes\n")

	client, clientErr := k8s_utility_client.NewK8sClients()
	if clientErr != nil {
		err = errors.Wrapf(clientErr, "failed creating k8s clients")
		return nodeNames, err
	}

	nodes, listErr := client.ClientSet.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if listErr != nil {
		err = errors.Wrapf(listErr, "failed listing nodes from kubernetes")
		return nodeNames, err
	}

	nodeNames = make([]string, 0, len(nodes.Items))
	for _, node := range nodes.Items {
		nodeNames = append(nodeNames, node.Name)
	}

	return nodeNames, err
}

// ApplyPurposeLabelsAndTaints applies a purpose label and corresponding taint to a node.
func ApplyPurposeLabelsAndTaints(ctx context.Context, nodeName string, purposeValue string, verbose bool) (err error) {
	manager.VerboseOutput(verbose, "Applying purpose label and taint to node %s (purpose: %s)\n", nodeName, purposeValue)

	client, clientErr := k8s_utility_client.NewK8sClients()
	if clientErr != nil {
		err = errors.Wrapf(clientErr, "failed creating k8s clients")
		return err
	}

	// Get the node
	node, getErr := client.ClientSet.CoreV1().Nodes().Get(ctx, nodeName, metav1.GetOptions{})
	if getErr != nil {
		err = errors.Wrapf(getErr, "failed getting node %s", nodeName)
		return err
	}

	// Add label
	if node.Labels == nil {
		node.Labels = make(map[string]string)
	}
	node.Labels["purpose"] = purposeValue

	// Add taint
	taint := corev1.Taint{
		Key:    "purpose",
		Value:  purposeValue,
		Effect: corev1.TaintEffectNoSchedule,
	}

	// Check if taint already exists
	taintExists := false
	for _, existingTaint := range node.Spec.Taints {
		if existingTaint.Key == taint.Key && existingTaint.Value == taint.Value && existingTaint.Effect == taint.Effect {
			taintExists = true
			break
		}
	}

	if !taintExists {
		node.Spec.Taints = append(node.Spec.Taints, taint)
	}

	// Update the node
	_, updateErr := client.ClientSet.CoreV1().Nodes().Update(ctx, node, metav1.UpdateOptions{})
	if updateErr != nil {
		err = errors.Wrapf(updateErr, "failed updating node %s with purpose label and taint", nodeName)
		return err
	}

	manager.VerboseOutput(verbose, "Successfully applied purpose=%s label and taint to node %s\n", purposeValue, nodeName)

	return err
}
