[![GitHub license](https://img.shields.io/badge/license-Apache%20License%202.0-blue.svg)](LICENSE)
[![CircleCI](https://circleci.com/gh/apprenda/kismatic.svg?style=svg)](https://circleci.com/gh/apprenda/kismatic)
[![Slackin](http://slack.kismatic.com/badge.svg)](http://slack.kismatic.com/)

# Kismatic Enterprise Toolkit (KET): Design, Deployment and Operations System for Production Kubernetes Clusters

Join our mailing list for updates on new releases: https://groups.google.com/forum/#!forum/kismatic-users

<img src="https://github.com/apprenda/kismatic/raw/master/docs/logo/KET_logo.png" width="500" />

## Introduction

KET is a set of production-ready defaults and best practice tools for creating enterprise-tuned Kubernetes clusters. KET was built to make it simple for organizations who fully manage their own infrastructure to deploy and run secure, highly-available Kubernetes installations with built-in sane defaults for scalable cross-cluster networking, distributed tracing, circuit-breaking, request-level routing, cluster health-checking and much more!

KET operational tools include:

1. [`Kismatic CLI`](docs/install.md)
   * Command-line control plane and lifecycle tool for installing and configuring Kubernetes on provisioned infrastructure.
2. [`Kismatic Inspector`](cmd/kismatic-inspector/README.md)
   * Cluster health and validation utility for assuring that software and network configurations of cluster nodes are correct when installing Kubernetes.
3. [`Kuberang`](https://github.com/apprenda/kuberang)
   * Cluster build verification to ensure networking and scaling work as intended. This tool is used to smoke-test a newly built cluster.
4. [`Kismatic Provision`](https://github.com/apprenda/kismatic-provision)
   * Quickly provision infrastructure on public clouds such as AWS and Packet. Makes building demo and development clusters a 2-step process.

## Components
| Component | Version |
| --- | --- |
| Kubernetes | v1.8.4 |
| Docker | v1.12.6 |
| Etcd (for Kubernetes) | v3.1.10 |
| Etcd (for Calico & Contiv) | v3.1.10 |
| Calico | v2.6.2 |
| Weave | v2.0.5 |
| Contiv | v1.1.1 |


[Download KET here](https://github.com/apprenda/kismatic/releases)

## Supported Operating Systems
- RHEL 7
- CentOS 7
- Ubuntu 16.04

# Usage Documentation

[Full Documentation](docs/README.md) -- Index of all the documentation

[Installation Overview](docs/install.md) -- Instructions on using KET to install a Kubernetes cluster.

[Upgrading Your Cluster](docs/upgrade.md) -- Instructions on using KET to upgrade your Kubernetes cluster.

[Plan File Reference](docs/plan-file-reference.md) -- Reference documentaion for the KET plan file.

[Cluster Examples](docs/intent.md) -- Examples for various ways you can use KET in your organization.

[CNI Providers](docs/networking.md) -- Information about the supported CNI providers by KET.

[Add Ons](docs/add_ons.md) -- Information about the Add-Ons supported by KET.

[Kismatic CLI](https://github.com/apprenda/kismatic/tree/master/docs/kismatic-cli) -- Dynamically generated documentation for the Kismatic CLI.

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

### Using your cluster

KET automatically configures and deploys [Kubernetes Dashboard](http://kubernetes.io/docs/user-guide/ui/) in your new cluster. Open the link provided at the end of the installation in your browser to use it.

Simply use the `kismatic dashboard` command to open the dashboard

You may be prompted for credentials, use `admin` for the **User Name** and `%admin_password%` (from your `kismatic-cluster.yaml` file) for the **Password**.

The installer also generates a [kubeconfig file](http://kubernetes.io/docs/user-guide/kubeconfig-file/) required for [kubectl](http://kubernetes.io/docs/user-guide/kubectl-overview/). Instructions are provided at the end of the installation on how to use it.
