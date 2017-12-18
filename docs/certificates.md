# Certificates and Certificate Generation

## Overview
Certificates (and their corresponding private keys) are used to secure all communication endpoints
exposed by the components of the Kubernetes cluster. In this way, messages between
components are encrypted, and only the parties involved in the communication
are able to decrypt them.

Certificates are also used to authenticate the parties involved in the 
communication. The authentication function can be split into two buckets:

* Server authentication: As a client, I want to ensure that I am talking to who I think I am talking to, and not a rogue server.
* Client authentication: As a server, I want to ensure that the client initiating a connection is allowed access.

The Kubernetes cluster has a root Certificate Authority (CA) that is used to create trust between
all components in the cluster. This certificate is a root CA with respect to the cluster, but it doesn't have to be itself a root CA. It can be an intermediate certificate, a self-signed certificate, or any certificate chosen by the operator. 

Each component has it's own certificate that has been signed
by the root CA, but all components are configured to trust the root CA.
In this way, when a client issues a request to a server, the client can check that
the server's certificate has been signed by the CA. On the flip side, the server can
verify the client's certificate as well by making sure the cert has been signed by the root CA.

## Certificates in Kismatic
Kismatic will generate certificates for all the components in the cluster. Furthermore,
it will also generate an admin certificate that can be used to access the API server
as a cluster admin.

### Generated Certificates

| Certificate | Purpose | Filename |
|---|---|---|
| Self-Signed CA | Sign generated certificates |  ca.pem |
| Etcd Server Cert | Serving API over HTTPS, performing peer-authentication | $nodeName-etcd.pem | 
| API Server Cert | Serving API over HTTPS | $nodeName-apiserver.pem  |
| Controller Manager Client Cert  | Used by controller manager to talk to API Server  | kube-controller-manager.pem  |
| Scheduler Client Cert | Used by scheduler to talk to API Server | kube-scheduler.pem |
| Service Account Signing Cert | Used by controller mgr to sign service accounts | service-account.pem |
| Kubelet Client Cert | Used by Kubelet to talk to API Server | $nodeName-kubelet.pem |
| Kube Proxy Client Cert | Used by Kube-proxy to talk to API Server | kube-proxy.pem | 
| Etcd Client Cert | Used by calico to talk to etcd | etcd-client.pem |
| Admin Client Cert | Used by admin to authenticate with the cluster using kubectl | admin.pem | 

### Secured Interactions

The following is a list of the interactions that happen between all the components of the cluster:

* Etcd <-> Etcd peering
* API Server -> Etcd 
* Calico -> Etcd
* Controller Manager -> API Server
* Scheduler -> API Server
* Kubelet <-> API Server
* Kube-proxy -> API Server
* Calico -> API Server
* Workloads -> API Server (via Service Accounts)
* Kubectl -> API Server

### How are certs generated?
* Using cfssl (https://github.com/cloudflare/cfssl
  * Algorithm: RSA
  * Key Size: 2048
* Expiration: configurable, defaults to 17600h (2 years)

### Can I bring my own CA?
Yes. Kismatic allows you to provide your own Certificate Authority for generating certificates. Simply place the CA's private key (`ca-key.pem`) and certificate (`ca.pem`) in the `generated/keys` directory beside the `kismatic` binary.

### Certificate generation command
In Kubernetes, client certificates are used for authenticating with the Kubernetes API server. KET facilitates
the generation of certificates with the `certificates generate` subcommand. 

The `certificates generate` subcommand can be used to generate certificates using the 
same technique that is employed by KET. The main use case for this subcommand is to 
create client certificates when new team members join, or when you need to grant
access to the cluster to another system such as a CI/CD tool.

The Kubernetes API server derives username and group information from client certificates when they
are used for authentication purposes. More specifically, the username is derived
from the certificate's Common Name field, and the groups are derived from the certificate's
organization fields. The X509 Client Cert authentication strategy is documented 
[here](https://kubernetes.io/docs/admin/authentication/#x509-client-certs).

With that said, you can use the following command to generate a client certificate
for a new team member `alice` that belongs to the `dev` and `ops` groups:
```
./kismatic certificates generate alice --organizations dev,ops
```


Full documentation on the CLI command can be found [here](./kismatic-cli/kismatic_certificates.md)
