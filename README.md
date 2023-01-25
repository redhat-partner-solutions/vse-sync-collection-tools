# Go Plugin Tests

A small test to understand what would happen if a Go test framework attempted to use Go plugins to dynamically include multiple different sets of test specs using Ginkgo. The goal here is to allow test specs to be separate, and to be managed alongside metadata such as identifiers, tags, and other such static data. Longer term this approach would also allow subsets of tests to be recompiled and redistributed in isolation, without the overhead of inter-process or inter-pod communication.

The execution flow here is that the main framework loads pre-compiled plugins (`*.so`) at runtime. Loading the plugin registers its tests with ginkgo, and the framework runs all registered tests. The framework can call method in each plugin to configure them.

The outcome of this experiment is that it is achievable but has a few caveats that mean it is probably not worth pursuing. The caveats are:

1. The plugin must not add top-level setup or cleanup (e.g. BeforeEach). Any such methods will be run for every spec in every plugin. This could have unintended side effects.
2. All test plugins must be loaded and configured before any tests can be run. This is the same for a standard monolithic Ginkgo suite where `RunSpecs` may only be called once, but reduces the advantage of having separate plugins.
3. Two-way sharing of data between the main framework and the plugins is necessarily more heavyweight than without the plugin architecture.

Overall there is little benefit to plugins over standard packages for this use case and the potential for the caveats above to cause major problems is high enough to avoid the approach.

## Setup

1. [Install Go 1.19](https://go.dev/doc/install).


## Running

1. Install Tools and Dependencies

```shell
go install github.com/onsi/ginkgo/v2
go install github.com/onsi/gomega
go mod tidy
```

2. Build The Plugins:

```shell
go build -buildmode=plugin -o tests/ptp.so tests/ptp/ptp.go
go build -buildmode=plugin -o tests/quq.so tests/quq/quq.go
go build -buildmode=plugin -o tests/rvr.so tests/rvr/rvr.go
```

3. Run the test framework (hard-coded to run all three plugins for this test)

```shell
ginkgo ./tests 
```
