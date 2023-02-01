# synchronization-testsuites

The purpose of this project is to make a single repository for all the existent used synchronization testing suites for the OpenShift platform, typically testing suites that target  different levels of testing (integration, operator testing, performance testing)

## Setup

1. [Install Go](https://go.dev/doc/install)
1. Install required binaries: `make install-tools`. Ensure your `$GOBIN` is on your `$PATH`
1. Install dependencies with `go mod tidy`
1. Run tests with `ginkgo tests`

## Contributing to the repo

Send a pull request.

### Current Synchronization Testing Suites

Common testing suites:

* [ptp operator conformance testing](https://github.com/openshift/ptp-operator.git) - ptp operator testing
* [cnf ptp tests](https://github.com/openshift-kni/cnf-features-deploy.git) - cnf ptp tests
