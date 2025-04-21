# k8s-cluster-manager

Utility for managing bare-metal k8s clusters Nik style.

I've had unsatisfactory experience with both managed kubernetes clusters, and the "LoadBalancer Service" pattern within K8s.

I prefer to put a simple layer 4 load balancer over my k8s instances, and add / remove nodes when I want to scale up or down.

This utility will create VM's, create load balancers, and attach the instances to the load balancers and configure the VM's for Talos Linux.

# Node Creation

Node creation basically looks like this:

1. Create a VM with a Talos image
2. Send the Talos Machine Config to the instance.
3. Add the node to the apiserver LB if it's a control plane node
4. Add the node to the various ingress load balancers if it's a worker instance

# Node Deletion

Node deletion removes the VM's from the load balancers, and then deletes the VM.


# Needs

controlplane machine config yaml
  -> patch with node name

worker machine config yaml
  -> patch with node name

general node configuration
  -> patch with instance type
  -> patch with root volume size, type, iops, throughput

