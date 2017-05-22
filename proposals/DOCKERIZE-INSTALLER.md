# Dockerized Installer

This document describes the design decisions and implementation of running the `kismatic` installer in a Docker container.

## Motivation
Currently the installer is distributed as a _tar_ file with specific builds for Linux and OSX. Dealing with the tar file is cumbersome as it requires moving around multiple files and folders. Also, with new releases users are required to copy their `./generated` directory and `kismatic-cluster.yaml` file into a new directory to manage their cluster.

Another drawback with the current distribution is that it does not support running `kismatic` on a windows machine. Ansible does not support Windows.

## Design
* Build docker images containing the `kismatic` binary and the `ansible` directory.
  * The image could have `ENTRYPOINT [./kismatic]` so that it can be run as `apprenda/kismatic install` without needing to repeat `kismatic`
* Provide a mechanism to maintain current UX of `kismatic ...`
  * Provide samples `#!/bin/sh` scripts for Unix to be placed in their `$PATH` (`/usr/local/bin/kismatic`)
```
#!/bin/sh
docker run -it -v "$(pwd)":/workdir -v "$(HOME)/.ssh/kismaticuser.key":/root/.ssh/kismaticuser.key -w /workdir apprenda/kismatic:$TAG "$@"
```
This is only a sample file it would ultimately be up to user to run the container how they want to. For example the `-v` mount for the ssh key might be some other directory or even a docker _name value_.
  * Provide sample `.bat` files for Windows  
TODO `.bat` file  
TODO what is a common directory to place this?

This will allow the user to still run `kismatic install apply` but instead execute the installation from a docker container.

## Tagging/Building
* Each `kismatic` release version will have a corresponding docker tag
* The `latest` tag will always point to the latest `kismatic` release
* Continue building binaries

## Considerations
* Currently a new plan file is generated with `ssh_key: kismaticuser.key` but the validation requires it to be an absolute path.
  * Detect if the binary is running in a container: use `/root/.ssh/kismaticuser.key`
  * Otherwise: use `$(HOME)/.ssh/kismaticuser.key`
* `kismatic dashboard` will no longer work because the container would not have access to open a browser on the host machine
  * Detect if running in a container: instead of trying to open just print the dashboard URL
* How do we handle `provision`?
* What to use in integration tests?
  * Keep current process
  * Build a dev image, will this require pushing to Dockerhub? In SnapCI each step in the pipeline can be a different environment and may not have the docker image locally.
