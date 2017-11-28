# Cloud Provider

## Motivation
Cloud Provider support is an important feature of kubernetes that KET did not configure. This was mainly due to the lack of documentation around this functionality, however it is something that has been requested by the community to be supported in KET.

## Implementation
Two new options are added to the `kubelet`, `kube-apiserver` and `kube-controller-manager` spec files  	

`--cloud-provider` - options are aws, azure, cloudstack, fake, gce, mesos, openstack, ovirt, photon, rackspace, vsphere, or empty for e.g. bare metal setups.  
`--cloud-config`- used by aws, gce, mesos, openshift, ovirt and rackspace  

More detail [here](https://kubernetes.io/docs/getting-started-guides/scratch/#cloud-providers)

### Plan File Changes
These options will be exposed in the plan file as follows: 
```
cluster:
  kube_apiserver:
    option_overrides: {}
  cloud_provider:
    provider:
    config:
```

`provider` - a string, maps to `--cloud-provider`  
`config` - absolute path to the config file on the bastion node. This file will be copied to all machines to `/etc/kubernetes/cloud_config` 

Initially we will target support(with tests) for `aws` and `openstack`, however we should not prevent the user from using any other provider.
This can be accomplished with a warning at runtime or documentation.

### aws provider
`aws` does not require the `cloud-config` and utilizes IAM policies to interact with the API.

Provider integration enables 2 features:
* using `LoadBalancer` service type, this will create an AWS ELB and assign a public DNS to the service
* using a `StorageClass` with a the `provisioner: kubernetes.io/aws-ebs`
* getting the required credentials to pull `ecr` images

Sample `StorageClass`
```
kind: StorageClass
apiVersion: storage.k8s.io/v1
metadata:
  name: slow
provisioner: kubernetes.io/aws-ebs
parameters:
  type: io1
  zones: us-east-1a, us-east-1c
  iopsPerGB: "10"
```

Sample IAM poicies below:

Master:
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
Worker:
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

### openstack provider
The only [example](https://stackoverflow.com/questions/32226108/kubernetes-openstack-integration) I've able to find.
The [sourcecode](https://github.com/kubernetes/kubernetes/blob/release-1.7/pkg/cloudprovider/providers/openstack/openstack.go) can also be used for reference.
```
[Global]
auth-url = OS_AUTH_URL
user-id = OS_USERNAME
api-key = OS_PASSWORD
tenant-id = OS_TENANT_ID
tenant-name = OS_TENANT_NAME
[LoadBalancer]
subnet-id = 11111111-1111-1111-1111-111111111111
```
This will require testing and someone with openstack experience.

### Validation
* Confirm `--cloud-provider` is a valid option
* Confirm `--cloud-config` is present on the local machine with the required permissions to copy the file 
* Prevent `--cloud-provider` and `--cloud-config` from being set in `cluster.kube_apiserver.option_overrides: {}`

### Documentation
* Modify Plan File Reference
* Document how to use `aws` provider
* Document how to use `openstack` provider