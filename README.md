# Kismatic Platform TEST

[![Build Status](https://snap-ci.com/On8xdVQV0xY5VXICf0Fx0Vq7fVMDUAfU6JFc8Wtt94A/build_image)](https://snap-ci.com/apprenda/kismatic-platform/branch/master)

# Kismatic CLI

`kismatic` is the command-line utility for managing your Kismatic platform.

## Installing Kismatic
Use the `install` command to setup your platform. The installer expects the underlying infrastructure to be accessible via SSH using Public Key Authentication.

The installation consists of three phases:
* Plan: On initial execution, the installer will ask questions about the platform. It will then produce a `kismatic-cluster.yaml` file that
describes the installation plan. Review the installation plan, and fill out the information for each node. Once done, run the `install` command again.
* Validate: The installer will validate the installation plan, and make sure that all fields are valid.
* Install: If the installation plan is valid, the installer will execute the installation.

### The install plan file
The install plan contains information about the desired cluster. Most options have sensible defaults, but it's definitely not a "one size fits all".
Depending on your infrastructure and needs, you will have to make changes to the installation plan.

The node has an optional internal IP attribute. This is used for public cloud scenarios where a machine has a public IP and a private IP.
If you are installing to private infrastructure, you can usually leave the internal IP attribute blank.


# Developer notes
### Pre-requisites
- Go installed
- Docker (required for building)

### Build using make
We use `make` to clean, build, and produce our distribution package. Take a look at the Makefile for more details.

In order to build the Go binaries (e.g. Kismatic CLI):
```
make build
```

In order to clean:
```
make clean
```

In order to produce the distribution package:
```
make dist
```
This will produce an `./out` directory, which contains the bits, and a tarball.

You may pass build options as necessary:
```
GOOS=linux make build 
```


