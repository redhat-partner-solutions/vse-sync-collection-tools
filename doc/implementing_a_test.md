# Implementing a New Test

As described in the [readme](../README.md), these tests must separate their configuration from the environment, and Collectors from Checks.

The Environment definition is intended to be shared for all tests, and Collectors are intended to be reusable.

## Environment

Add something to the core Environment configuration file only if absolutely necessary. If possible then create one or more Collectors to gather than information at runtime if possible. This is critical to ensure that tests are, and remain, usable across systems and environments.

## Configuration

The framework provides a single configuration file, you can register a new section to be loaded as configuration for your test suite as show here:

```go
import (
	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/config"
)

const ptpCustomConfigKey = "ptp_tests_config"

type PtpConfig struct {
	TtyTimeout    int    `yaml:"tty_timeout"`
}

var ptpConfig = PtpConfig{}

func init() {
	config.RegisterCustomConfig(ptpCustomConfigKey, &ptpConfig)
}
```

The breaks down to:
1. Import the `config` module
1. Create a `struct` to hold you custom configuration. fields must be exported (capitalise) to be loaded from file.
1. Initialise a variable into which that configuration will be loaded
1. **in the `init` method** register the configuration section name, and the variable into which it will be loaded.

Your package-scoped variable will then be automatically populated with configuration when the file is loaded.


## Create a Collector

**TODO:** define a standard interface/approach

If an existing collector exists for the thing you wish to observe, then reuse that.

For an example Collector, see `GetClusterVersion` in [pkg/retrievers/cluster.go](../pkg/retrievers/cluster.go)

Defining these is work in progress, they should be as independent as possible from assumptions about the underlying cluster.


## Checks

**# TODO** Clarify test configuration patterns

Checks are implemented as standard Ginkgo specs. Create a new suite, or extend an existing one, in [tests/](../tests)

Detailed instructions are pending, see the current tests in [tests/ptp/](../tests/ptp/tgm.go) for a work-in-progress example.
