# k8s-cluster-manager

Utility for managing bare-metal k8s clusters Nik style.

I've had unsatisfactory experience with both managed kubernetes clusters, and the "LoadBalancer Service" pattern within K8s.

I prefer to put a simple layer 4 load balancer over my k8s instances, and add / remove nodes when I want to scale up or down.

This utility will create VM's, create load balancers, and attach the instances to the load balancers and configure the VM's for Talos Linux.

# Node Creation

Node creation basically looks like this:

1. Create a VM with a Talos image
2. Send the Talos Machine Config to the instance.  This requires 3 sources of data which can be provided in files, or loaded from Hashicorp Vault:
   1. Talos Machine Config for the node type (worker or control-plane)
   2. Talos Machine Config Patch
   3. Cloud Provider Node Information
3. Add the node to the apiserver LB if it's a control plane node
4. Add the node to the various ingress load balancers if it's a worker instance
5. Add the node name/IP to Cloudflare

Load Balancers, Security Groups, etc are discoverd based on the tags with the `Cluster` key.  Value is expected to be the name of the cluster.

# Node Deletion

Node deletion removes the VM's from the load balancers, kubernetes, cloudflare, and then deletes the VM.

# Hashicorp Vault Integration

The `--secretmount` or `-m` flag denotes a Hashicorp Vault KVv2 mount.  If provided, and if you have a current Vault token in the expected place (`~/.vault-token`), the `k8s-cluster-manager` will attempt to fetch data from secrets with the following name patterns:

* *Talos Machine Configuration* `<MOUNT>/cluster-<CLUSTER_NAME>-machine-<ROLE_NAME>` e.g. `dev/cluster-fargle-machine-worker`.
* *Talos Machine Configuration Patch* `<MOUNT>/cluster-<CLUSTER_NAME>-patch-<ROLE_NAME>` e.g. `dev/cluster-fargle-patch-worker`.
* *Cloud Provider Node Config Information* `<MOUNT>/cluster-<CLUSTER_NAME>-node-<ROLE_NAME>` e.g `dev/cluster-fargle-node-worker`.


## Talos Machine Configuration
This is the `controlplane.yaml` or `worker.yaml` produced from `talosctl`, rendered into JSON, since that's how Vault stores secrets.

    {
      "cluster": {
        "ca": {
          "crt": "...",
          "key": ""
        },
        "controlPlane": {
          "endpoint": "https://your-apiserver.com:6443"
        },
        "discovery": {
          "enabled": true,
          "registries": {
            "kubernetes": {
              "disabled": true
            },
            "service": {}
          }
        },
        "id": "123455667asdf",
        "network": {
          "dnsDomain": "cluster.local",
          "podSubnets": [
            "10.1.0.0/16"
          ],
          "serviceSubnets": [
            "10.2.0.0/12"
          ]
        },
        "secret": "...",
        "token": "..."
      },
      "debug": false,
      "machine": {
        "ca": {
          "crt": "...",
          "key": ""
        },
        "certSANs": [],
        "features": {
          "apidCheckExtKeyUsage": true,
          "diskQuotaSupport": true,
          "kubePrism": {
            "enabled": true,
            "port": 7445
          },
          "rbac": true,
          "stableHostname": true
        },
        "install": {
          "disk": "/dev/sda",
          "image": "ghcr.io/siderolabs/installer:v1.6.0",
          "wipe": false
        },
        "kubelet": {
          "defaultRuntimeSeccompProfileEnabled": true,
          "disableManifestsDirectory": true,
          "image": "ghcr.io/siderolabs/kubelet:v1.29.0"
        },
        "network": {},
        "registries": {},
        "token": "...",
        "type": "worker"
      },
      "persist": true,
      "version": "v1alpha1"
    }

## Talos Machine Configuration Patch

Vault stores things in JSON, basically, so here's an example patch:

    {
      "machine": {
        "install": {
          "disk": "/dev/xvda"
        },
        "kubelet": {
          "extraArgs": {
            "rotate-server-certificates": true
          }
        },
        "sysctls": {
          "net.netfilter.nf_conntrack_max": "1048576"
        }
      }
    }


## Cloud Provider Node Config

The Cloud Provider Node Config is a JSON map containing just enough info to run a Talos VM.

### AWS Version

    {
      "block_device_gb": 100,
      "block_device_name": "/dev/xvda",
      "block_device_type": "gp3",
      "domain": "some.domain",
      "image_id": "ami-003f31a78653be190",
      "instance_type": "r5.4xlarge",
      "placement_group_name": "some-placement-group",
      "subnet_id": "subnet-0e123456789"
    }
