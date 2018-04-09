# Decouple Components Versions from KET Version

## Motivation

KET is used to install Kubernetes and many other components in the cluster. Historically the version of Kubernetes and all other components versions were tied to a version of KET.

For example KET `v1.7.0` can only install a cluster with Kubernetes `v1.9.0` and when the next version of Kubernetes is released KET would need to update its own version.

This has been working fine while KET was in active development, however as it became more stable and the release cycle slows down many KET releases will only contain the Kubernetes version change.

To alleviate the pressure on the KET team to release with every Kubernetes `Patch` version and to also allow users to stay with the latest bug fixes, the version of Kubernetes should be decoupled from the version of KET.  
Decoupling of Kubernetes will be done first and the decoupling of all other components that are installed can be discussed at a later time.

## Kubernetes Versions

Kubernetes follows the semantic versioning schema `X.Y.Z` - `Major.Minor.Patch`

The latest version as of this writing is `v1.9.0`

The Kubernetes release cadence has become quite stable, with a new `Minor` version about every 3 months and a `Patch` release about every 2 weeks.

---

`Minor` releases will contain API changes and may contain docker and etcd changes, in addition to other components changes, ie `dashboard`, `kube-dns`, `heapster` etc.

---

`Patch` releases are intended for critical bug fixes to the latest minor version, such as addressing security vulnerabilities, fixes to problems affecting a large number of users, severe problems with no workaround, and blockers for products based on Kubernetes.

They should not contain miscellaneous feature additions or improvements, and especially no incompatibilities should be introduced between patch versions of the same minor version.

## KET Versions

KET follows the semantic versioning schema `X.Y.Z` - `Major.Minor.Patch`

The latest version as of this writing is `v1.7.0`

KET had a release cycle of about every 8 weeks for a `Minor` version and 2-3 weeks for a `Patch` version.

---

`Minor` releases contained larger changes, but more importantly a Kubernetes `Minor` change could only be upgraded on a KET `Minor` release.

---
`Patch` releases are smaller bug fixes and improvements.

---

KET has always been up to date with the latest Kubernetes `Minor` version within 2 weeks of the release, however the Kubernetes `Patch` version tracking has been a "best effort" approach and not guaranteed to be at the latest. KET has also never retroactively released a `Patch` version of previous `Minor`.

## Implementation

A particular `Minor` release of KET will support **any** `Patch` version of Kubernetes. (Note multiple consecutive `Minor` versions of KET can support the same `Minor` version of Kubernetes, until KET is updated to support a new version of Kubernetes).

As discussed above, Kubernetes `Minor` changes will contain component and API changes and testing is required before KET can support and certify a new Kubernetes `Minor` release.

By contrast, Kubernetes `Patch` versions will only contain bug fixes and smaller patches and it should be safe to use any `Patch` version, and is actually recommended by the Kubernetes team to always use the latest patch version.

The tested and default Kubernetes `Patch` versions will still be upgraded whenever there are other KET bug fixes, however a new `Patch` version will NOT be released for the sole purpose of upgrading a Kubernetes `Patch` version.

### Example Versions

| KET Version                  | 1.8.0         | 1.8.1         | 1.8.2       | 1.9.0       | 1.10.0        |
|------------------------------|---------------|---------------|-------------|-------------|---------------|
| Tested Kubernetes Version    | 1.9.1         | 1.9.2         | 1.9.2       | 1.9.4       | 1.10.0        |
| Supported Kubernetes Version | 1.9.x         | 1.9.x         | 1.9.x       | 1.9.x       | 1.10.x        |

### Plan File Changes

``` yaml
# Set component versions to install.
cluster:
  version: "v1.9.0"
...
```

A new optional field will be added that will contain the Kubernetes `version` under the existing `cluster` tag.

### Install Changes

When running `kismatic plan` the version will be set to the version the the KET binary was built and tested with.

The `kubernetes` version string will also be validated to be the supported `Minor` version and less than or equal to the latest stable version of that `Minor` from:
``` bash
https://storage.googleapis.com/kubernetes-release/release/stable-1.9.txt
``` 
If the latest version cannot be fetched KET will only validate that the provided version the supported `Minor` version, it will be up to the user to select an existing valid version. This is required to support disconnected installs and any when errors retrieving latest stable version.

The `kubernetes` version will be propagated to Ansible:
``` yaml
  kube_proxy:
    name: gcr.io/google-containers/kube-proxy-amd64
    version: "v1.9.0"
  kube_controller_manager:
    name: gcr.io/google-containers/kube-controller-manager-amd64
    version: "v1.9.0"
  kube_scheduler:
    name: gcr.io/google-containers/kube-scheduler-amd64
    version: "v1.9.0"
  kube_apiserver:
    name: gcr.io/google-containers/kube-apiserver-amd64
    version: "v1.9.0"
```

**NOTE:** The `kubectl` version in the released tar will always be the default version.

### Upgrade Changes

With current version of KET prior tp upgrade the installer reads `/etc/kismatic-version` on all of the nodes to determine if the upgrade should proceed on the node. This will no longer be enough as the user may want to upgrade to a newer Kubernetes `Patch` version using the same KET version.

A new file will be placed in `/etc/installed-components.yaml`:

``` yaml
versions:
  kubernetes: "v1.9.0"
```

This new file will then also be read to determine if the node needs to be upgraded.

## Considerations

If there is a concern with installing a cluster with an untested Kubernetes version, the user can always leave the version field empty (or set it the version specified in the release notes). This would guarantee that the Kubernetes version used in the installation has been tested during the KET release process.

For the initial implementation of components decoupling only the Kubernetes version will be supported and this feature will be expanded to other components in the cluster at a later time.