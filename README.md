# synchronization-testsuites

The main purpose of this repo is build the necessary tooling to collect necessary synchronization-related data logs from a running OpenShift cluster. This data will then be leveraged by different analysis tools to determine if the cluster is running within acceptable bounds synchronization-wise.```

The core approach taken is to strongly encourage and enforce separation of concerns between:
1. Declarative description of the cluster(s) under test
1. Configuration of a test (e.g. number of repetitions, acceptable thresholds, etc.)
1. Collectors - methods of collecting indicative information about the cluster
1. Checks - performed on collected values

## Setup

1. [Install Go](https://go.dev/doc/install)
1. Install dependencies with `go mod tidy`

### Development Extras
1. Install dev binaries: `make install-tools`. Ensure your `$GOBIN` is on your `$PATH`
1. yamllint
    1. [Install yamllint](https://yamllint.readthedocs.io/en/stable/) with `sudo yum install yamllint`
    1. run with `yamllint ./`
1. golangci-lint
    1. [Install golangci-lint](https://golangci-lint.run/usage/install/#local-installation)
    1. run with `make lint`
1. license-eye
    1. [Install license-eye](https://github.com/apache/skywalking-eyes) with `go install github.com/apache/skywalking-eyes/cmd/license-eye@latest`
    1. run with `license-eye header check` or `license-eye header fix`
1. pre-commit
    1. on RHEL, `pre-commit` requires recompiling python to include optional sqlite modules:
        1. `sudo yum install sqlite-devel`
        1. See instructions [here](https://tecadmin.net/how-to-install-python-3-10-on-centos-rhel-8-fedora/)
    1. install pre-commit with `pip3.10 install pre-commit`
    1. configure your repository to run pre-commit hooks with `pre-commit install`
    1. manually run against all files with `pre-commit run --all-files` or against staged files with `pre-commit run`.
    1. Otherwise pre-commit will now run automatically when you make a new commit.

## Usage
### Building binary
Run the following command to build the binary used in the following commands:
```shell
go build
```
###### NOTE: use the `--race` flag when developing collectors.

### Checking Enviroment
Run the following command  (check help string for more details):

```shell
./vse-sync-collection-tools env verify --interface="<ptp interface>" --kubeconfig="${KUBECONFIG}"
```

### Running Collectors

Run the following command  (check help string for more details):

```shell
./vse-sync-collection-tools collect --interface="<ptp interface>" --kubeconfig="${KUBECONFIG}"
```

### Fetching logs
The log subcommand has been removed. Instead we have implimented at collector which is enabled by default.
If possible you should use a log aggregator. You can control the collectors running using the `--collector` flag.

## Running tests

TODO: implement tests for all packages

To test the framework components run `ginkgo pkg/<packagename>`, for example to run the unit tests for the `config` package use `ginkgo pkg/config`

## Contributing to the repo

See [Adding a collector](doc/implementing_a_collector.md)

## To Do List

* unit tests for all of `pkg/`
* add more collectors
* better data persistance options
