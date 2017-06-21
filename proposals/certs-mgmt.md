# Certificate Management

Status: Proposal

This proposal extends the initial certificates CLI with day-2 operations, such 
as validation of deployed certificates and certificate redeployment.

## Use Cases
* As an operator, I want to check the validity of all certificates currently in use on the cluster
* As an operator, I want to redeploy certificates that are nearing their expiration date
* As an operator, I want to redeploy certificates with a new CA without having to reinstall my cluster

## Certificate Validity Check
A command that can be run against an existing cluster to validate the certificates that are deployed on the cluster. Certificates that are withing the warning window will be flagged. Certificates that have expired will be flagged.

```
kismatic certificates validate [options]
```

Output: TBD. Most likely the execution output from an ansible playbook run to keep orchestration in ansible, but potentially a listing of per-node certificates:
```
# kismatic certificates validate
Node: etcd
Certificate          Expires
Etcd server          09/08/2017 10:00 AM

Node: master01
Certificate          Expires         
API server           09/08/2017 10:00 AM
Scheduler client     EXPIRES SOON (06/08/2017 10:00 AM)
...

Node: worker01
Certificate          Expires
Kubelet client       EXPIRED (09/08/2016 10:00 AM)
```

Pre-conditions:
* SSH access to nodes

Options:
* `--warning-window`: Warn about certificates that will expire within this number of days. Defaults to 45 days if not set.

## Certificate Redeployment
A command that can be used to deploy certificates on existing machines. If the certificates have not been generated, the command will generate the certificates for the cluster described in the plan file (As if an installation was being performed), and deploy them to the nodes.

Services are restarted whenever required.

```
kismatic certificates deploy
```

Output: Execution of ansible play for deploying certs.

Pre-conditions:
* SSH access to nodes


## Other considerations
* Have to figure out the mechanics of certificate redeployment. I _think_ service accounts 
would have to be recreated if the CA changes, or if the service account signing cert changes.