[![GitHub license](https://img.shields.io/badge/license-Apache%20License%202.0-blue.svg)](LICENSE)
[![Build Status](https://snap-ci.com/ulSrRsof30gMr7eaXZ_eufLs7XQtmS6Lw4eYwkmATn4/build_image)](https://snap-ci.com/apprenda/kismatic/branch/master)

# Kismatic Enterprise Toolkit (KET): Design, Deployment and Operations System for Production Kubernetes Clusters

Join our mailing list for updates on new releases: https://groups.google.com/forum/#!forum/kismatic-users

Join Slack to chat in real-time with the maintainers and community users of KET: https://kismatic.slack.com/signup

![KET](logo.png?raw=true "KET Logo")

## Introduction

KET is a set of production-ready defaults and best practice tools for creating enterprise-tuned Kubernetes clusters. KET was built to make it simple for organizations who fully manage their own infrastructure to deploy and run secure, highly-available Kubernetes installations with built-in sane defaults for scalable cross-cluster networking, distributed tracing, circuit-breaking, request-level routing, cluster health-checking and much more!

KET operational tools include:

1. [`Kismatic CLI`](docs/INSTALL.md)
   * Command-line control plane and lifecycle tool for installing and configuring Kubernetes on provisioned infrastructure.
2. [`Kismatic Inspector`](cmd/kismatic-inspector/README.md)
   * Cluster health and validation utility for assuring that software and network configurations of cluster nodes are correct when installing Kubernetes.
3. [`Kuberang`](https://github.com/apprenda/kuberang)
   * Cluster build verification to ensure networking and scaling work as intended. This tool is used to smoke-test a newly built cluster.
4. [Kismatic RPM & DEB Packages](docs/PACKAGES.md)
   * Packages for installing Kubernetes and its dependencies, focused on specific roles in an HA cluster.
   * With these packages installed on a local repo, it is possible to use Kismatic to install Kubernetes on nodes that do not have access to the public internet.
5. [`Kismatic Provision`](https://github.com/apprenda/kismatic-provision)
   * Quickly provision infrastructure on public clouds such as AWS and Packet. Makes building demo and development clusters a 2-step process.

## Dependencies
| Dependency | Current version |
| --- | --- |
| Kubernetes | 1.5.1 |
| Docker | 1.11.2 |
| Calico | 1.6 |
| Etcd (for Kubernetes) | 3.0.15 |
| Etcd (for Calico) | 2.37 |

[Download latest install tarball (Mac)](https://kismatic-installer.s3-accelerate.amazonaws.com/latest-darwin/kismatic.tar.gz)

[Download latest install tarball (Linux)](https://kismatic-installer.s3-accelerate.amazonaws.com/latest/kismatic.tar.gz)

# Usage Documentation

[Installation Overview](docs/INTENT.md) -- Useful examples for various ways you can use Kismatic in your organization.

[Plan & Build a Kubernetes cluster](docs/INSTALL.md) -- Details instructions on using KET to install a Kubernetes cluster.

[Using KET with linkerd](docs/LINKERD.md) -- Instructions on how to use KET with linkerd in 1 command.

[Using KET with Calico](docs/NETWORKING.md) -- Instructions on how to use KET with the built-in SDN controller Project Calico.

[Cert Generation](docs/CERT_GENERATION.md) -- Information on how KET handles certificates.

[Kismatic CLI](https://github.com/apprenda/kismatic/tree/master/docs/kismatic-cli) -- Dynamically generated Cobra documentation for the Kismatic CLI.

[Roadmap](ROADMAP.md) -- Insight into the near-term features roadmap for the next few releases of KET.

# Development Documentation

[How to Build KET](BUILDING.md)

[How to Contribute to KET](CONTRIBUTING.md)

[Running Integration Tests](INTEGRATION_TESTING.md)

[KET Code of Conduct](code-of-conduct.md)

[KET Release Process](RELEASE.md)

# Basic Installation Instructions
Use the `kismatic install` command to work through installation of a cluster. The installer expects the underlying infrastructure to be accessible via SSH using Public Key Authentication.

The installation consists of three phases:

1. **Plan**: `kismatic install plan`
   1. The installer will ask basic questions about the intent of your cluster.
   2. The installer will produce a `kismatic-cluster.yaml` which you will edit to capture your intent.
2. **Provision**
   1. You provision your own machines
   2. You tweak your network
   3. Review the installation plan in `kismatic-cluster.yaml` and add information for each node.
3. **Install**: `kismatic install apply`
   1. The installer checks your provisioned infrastructure against your intent.
   2. If the installation plan is valid, Kismatic will build you a cluster.
   3. After installation, Kismatic performs a basic test of scaling and networking on the cluster

###Using your cluster

KET automatically configures and deploys [Kubernetes Dashboard](http://kubernetes.io/docs/user-guide/ui/) in your new cluster. Open the link provided at the end of the installation in your browser to use it.

Simply use the `kismatic dashboard` command to open the dashboard

You may be prompted for credentials, use `admin` for the **User Name** and `%admin_password%` (from your `kismatic-cluster.yaml` file) for the **Password**.

The installer also generates a [kubeconfig file](http://kubernetes.io/docs/user-guide/kubeconfig-file/) required for [kubectl](http://kubernetes.io/docs/user-guide/kubectl-overview/), just follow the instructions provided at the end of the installation to use it.
