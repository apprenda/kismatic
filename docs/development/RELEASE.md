# Release

This document contains the release details for [Kismatic Enterprise Toolkit](https://github.com/apprenda/kismatic).

Kismatic leverages a CI/CD pipeline to build, test and release its software. The pipeline extensively tests the software before any release is published and a build must pass all unit and integration tests to be promoted by the pipeline.

Kismatic depends on [Kuberang](https://github.com/apprenda/kuberang) and the [Kismatic Distro Packages](https://github.com/apprenda/kismatic-distro-packages), along with all other components listed in the [README](https://github.com/apprenda/kismatic).
All components are developed and versioned independently, however a released version of the Kismatic Enterprise Toolkit will have specified versions of the underlying components that have been previously tested for compatibility.   

### Releasing
Once the master branch is ready for a release, it must be tagged and the tag pushed
to the remote repository.
```
git tag v1.13.0-playtika
git push origin v1.13.0-playtika
```

Pushing the tag will trigger a new build on the CI/CD system, which will have the
release job enabled. Once all tests are finished, the release job will publish a draft
release on GitHub and upload the binaries.

Once the build is complete, go to `https://github.com/apprenda/kismatic/releases`
to edit the release draft. 
* Include a section that highlights any actions required by the user, if any.
* Include a section that lists any deprecations that are being made in the release, if any.
* Include a section that lists the notable changes. Each entry in the list should 
include a link to the corresponding PR.
* Include a section that lists any plan file changes.
* Include a section that lists any component versions that have been changed in this new release
* Include any other information that is relevant to the release

### Useful Git commands
Get latest tag name:
```
git describe --abbrev=0 --tags
```

Get commit hash for a given tag
```
git rev-list -n 1 $TAG
```
