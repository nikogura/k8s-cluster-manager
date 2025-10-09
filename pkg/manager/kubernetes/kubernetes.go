package kubernetes

import (
	"context"
	"github.com/nikogura/k8s-cluster-manager/pkg/manager"
	k8s_utility_client "github.com/nikogura/k8s-utility-client/pkg/k8s-utility-client"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
