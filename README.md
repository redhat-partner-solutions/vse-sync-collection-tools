# synchronization-testsuites

The purpose of this project is to make a single repository for all the existent used synchronization testing suites for the OpenShift platform, typically testing suites that target  different levels of testing (integration, operator testing, performance testing)

The core approach taken is to strongly encourage and enforce separation of concerns between:
1. Declarative description of the cluster(s) under test
1. Configuration of a test (e.g. number of repetitions, acceptable thresholds, etc.)
1. Collectors - methods of collecting indicative information about the cluster
1. Checks - performed on collected values

## Setup

1. [Install Go](https://go.dev/doc/install)
1. Install required binaries: `make install-tools`. Ensure your `$GOBIN` is on your `$PATH`
1. Install dependencies with `go mod tidy`

## Running Tests

Create test configuration and environment description files, then run the following command, referencing your created files as needed:

```shell
ginkgo run tests -- --config=../cormorant_test_config.yaml --env=../cormorant_env.yaml
```

For more information on `--config=../cormorant_test_config.yaml` see [Test Configuration](doc/test_configuration.md) 
For more information on `--env=../cormorant_env.yaml` see [Test Environment](doc/test_environment.md)

## Testing the Framework

TODO: implement tests for all packages

To test the framework components run `ginkgo pkg/<packagename>`, for example to run the unit tests for the `config` package use `ginkgo pkg/config`

## Contributing to the repo

See [Adding a Test](doc/implementing_a_test.md)

## To Do List

* unit tests for all of `pkg/`
* Specify an interface for a `collector`
* consider `collector`s as Go packages
* `collector` registry?
* reconsider dynamically loading tests suites from Go packages, would force people slightly further away from standard ginkgo but have significant potential benefits for modularity
