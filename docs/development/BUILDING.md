# Kismatic

## Pre-requisites
- make
- Docker

## Build using make

### Checking out the code

If you're looking to contribute to Kismatic, the easiest way to get started is by using git.

Assuming you have `git` and `go` installed on your machine, run

```bash
ORG=your_github_account_name_here mkdir -p $(go env GOPATH)/src/github.com/$ORG && cd $(go env GOPATH)/src/github.com/$ORG && git clone https://github.com/$ORG/kismatic.git
```

### Building the distribution

Next, assuming you have `docker` and `make` installed, run

```bash
make dist
```

This will build a distribution for the host running the command. I.E. if you run this from a mac, you'll receive an `out-darwin` package in the working directory. Likewise, building from a linux machine will produce an `out-linux` package in the working directory.

If you do not wish to use `docker`, you can append `-host` to the command, and `make` will attempt to build locally. This is strongly discouraged, unless you really know what you're doing.

I.E. 

```bash
make dist-host
```

### Unit testing

We appreciate any contributions to be made to include unit tests for changes being made.

To run golang unit tests, simply run

```bash
make test
```

Again, you can append `-host` if you know what you're doing, and do not wish to run tests using docker.

### Integration testing

Kismatic uses [Ginkgo](https://github.com/onsi/ginkgo) for BDD integration testing.
We provide a guide for getting started with integration test: [here](docs/development/INTEGRATION_TESTING.md).

Similar to the unit tests, we also provide `make` recipes for integration testing.
To run the full suite of integration tests, run

```bash
make integration-test
```

The only caveat here is that since the integration tests are running from within a docker container, you need to have a linux distribution built.
If you are on a mac, this requires you to `GOOS=linux make dist`. If you already know you plan on running integration tests, you can also use `make all` to build a darwin and linux distribution.

Once again, you can append `-host` to run the tests outside of the container. Likewise, this requires you to have a distribution built for your host, I.E. a `make dist`

### Cutting down on iteration time

Assuming you already have a `dist` built. You can run a vanilla

```bash
make
```

to update the most commonly changed components of the build.

### Cleaning

If you want to run a clean build

```bash
make clean
```

will remove all build artifacts and distributions.

If you simply want to remove the distribution, without cleaning any of the vendored tools, use

```bash
make shallow-clean
```

### Advanced use cases

If you're very familiar with unix systems, and want a more detailed layout of the build process the makefile is roughly sorted by how user-facing each recipe is.

Be warned: the recipes near the bottom are intended to only be run on CI, and we do not guarantee any of them will work on your local.