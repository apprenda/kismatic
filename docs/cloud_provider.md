# Cloud Provider Integration

KET can configure your Kubernetes cluster to integrate with the underlying cloud
provider. When enabled, Kubernetes will be able to create load balancers, storage
and other resources on the fly.

To enable the cloud provider integration, the cloud provider must be specified in
the [plan file](./plan-file-reference.md). Furthermore - depending on the provider - you might have to provide
a provider-specific configuration file that must be distributed to all nodes in the 
cluster.

## AWS

If you are setting up your cluster on AWS, you can enable the cloud provider
integration by setting the [cluster.cloud_provider.provider](./plan-file-reference.md#clustercloud_providerprovider) field of the plan 
file to `aws`. The [cluster.cloud_provider.config](./plan-file-reference.md#clustercloud_providerconfig)
field can be left blank.

### Prerequisites
The following requirements must be completed before installing a Kubernetes cluster
with the AWS cloud provider integration enabled:

* The hostnames set in the plan file must match the Private DNS name that is 
assigned to the machine on AWS.
* All machines that are part of the Kubernetes cluster must be tagged with a 
unique cluster name. More specifically, all machines in the cluster must have
the `kubernetes.io/cluster/<cluster id>` tag set, where `<cluster id>` is the same 
for all nodes that belong to the same cluster. More information about this tag 
can be found [here](https://github.com/kubernetes/kubernetes/commit/0b5ae5391ef9eeba337f1dfa62b6f3125e28c86c).
* To allow the Kubernetes components access to the AWS API, the following IAM policy
must be applied to the nodes:

### Master
```
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "ec2:*"
            ],
            "Resource": [
                "*"
            ]
        },
        {
            "Effect": "Allow",
            "Action": [
                "elasticloadbalancing:*"
            ],
            "Resource": [
                "*"
            ]
        }
    ]
}
```

### Worker, Ingress, Storage
```
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": "ec2:Describe*",
            "Resource": "*"
        },
        {
            "Effect": "Allow",
            "Action": "ec2:AttachVolume",
            "Resource": "*"
        },
        {
            "Effect": "Allow",
            "Action": "ec2:DetachVolume",
            "Resource": "*"
        },
        {
            "Effect": "Allow",
            "Action": [
                "ecr:GetAuthorizationToken",
                "ecr:BatchCheckLayerAvailability",
                "ecr:GetDownloadUrlForLayer",
                "ecr:GetRepositoryPolicy",
                "ecr:DescribeRepositories",
                "ecr:ListImages",
                "ecr:BatchGetImage"
            ],
            "Resource": "*"
        }
    ]
}
```