# Certificates in v1.3.4

## User action required: See "Admin Certificate"

The certificate generation process has been revamped to produce a more secure cluster.

Previously, all nodes on the cluster had full access to the API server using the 
admin account. Now that RBAC is enabled, we can take advantage of more granural
authorization policies, and thus further secure the cluster.

During the upgrade to v1.3.4, you will notice that new component-specific certificates
will be generated. These certificates have a tighter access model than the previous node-level
certificates used in the past.

More information about certificates used in the cluster can be found [here](../../CERTIFICATES.md)

## Admin Certificate
One side effect of this change is that existing admin certificates are considered invalid. This 
is because the admin user must belong to the `system:masters` group, which is achieved
by including `system:masters` as an organization in the certificate. For this reason,
before performing an upgrade, the existing admin certificate must be removed, allowing kismatic to
regenerate it.