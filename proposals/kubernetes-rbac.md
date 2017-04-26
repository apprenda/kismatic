# Enable RBAC support in Kubernetes

Status: Proposal

# Motivation
As of Kubernetes v1.6.0, the role-based access control authorizer API was promoted
to beta. RBAC is quickly becoming the recommended authorization scheme for
Kubernetes clusters, as it is API driven and more powerful than ABAC. RBAC is also
the preferred authorization mechanism in large enterprises.

We should enable RBAC support out of the box when installing Kubernetes with Kismatic.

# Admin user
We will continue enabling both basic-auth and certificate-based authentication.
Ideally, we would only support cert-based auth, but accessing the dashboard would
require running `kubectl proxy` locally, which might not always be an option. 

Kismatic will bootstrap a single user, `admin`, that will be used by the cluster administrator.
This user will be accessible via the generated admin certificate, and via HTTP basic
auth when accessing the dashboard.

The generated `admin` user will be part of the `system:masters` group, which is
bound to the `cluster-admin` role by the kubernetes default bindings.

Kismatic will use the `admin` user to perform all current management operations.
Defining roles and permissions for Kismatic functionality is out of scope for this
proposal, but one can imagine building a set of roles that will be Kismatic specific.
For example, a specific role could be created for Kismatic that will allow it to
create persistent volumes on behalf of the user (when using the storage features).

# Upgrade considerations
Current clusters deployed with Kismatic are using ABAC for authorization. Removing
ABAC support in the same release that RBAC is enabled would most likely break 
existing clusters and applications, as they are relying on the permissive ABAC policy.

For this reason, we will continue to enable both ABAC and RBAC authorizers, and 
deprecate ABAC. Users will be responsible for creating the necessary RBAC policy
according to their needs. After doing this, users may manually remove the ABAC 
authorizer (by modifying the flag and restarting API servers), or they may wait
until the next KET release where ABAC will be removed.


