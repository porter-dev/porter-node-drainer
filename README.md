## Porter Node Drainer

This repository contains source code for performing a node drain for self-managed EKS nodes.

**How does it work?**

The script is meant to run as a Lambda function which gets triggered on an ASG instance termination event. The script performs the following actions:

1. Performs a read on the relevant EC2 instance being terminated to discover the EC2 internal hostname.
2. Lists nodes on the EKS cluster and matches the above hostname on the `kubernetes.io/hostname` label.
3. Cordons the node to prevent new workloads from being scheduled on that node.
4. Evicts any non-daemonset pods running on the node.
5. Calls the ASG Complete Lifecycle action to allow ASG termination to proceed.

**Required environment variables:**

```
EKS_AWS_ACCESS_KEY_ID
EKS_AWS_SECRET_ACCESS_KEY
EKS_AWS_REGION
EKS_AWS_CLUSTER_ID
EKS_CLUSTER_SERVER
EKS_CA_DATA
```

### Project Goals

While this project currently contains a very simple drain script, the goal is eventually to be able to customize the drain behavior on the node and to notify the Porter server that a node is being cycled out.

### Similar Projects

- https://github.com/aws-samples/amazon-k8s-node-drainer
- https://github.com/ryan-a-baker/eks-node-drainer
