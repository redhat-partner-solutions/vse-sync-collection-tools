// package config encapsulates loading the config from file, and delegating
// loading of specific test config to that test package.

// To use a custom config section first define a struct into which it will be
// unmarshalled, as you would normally. Then create a package-scoped instance of
// that struct, and choose a key under which data for that struct will be stored
// in the config file.
// Finally, *WITHIN* the `init()` of the test suite package add a call to
// `config.RegisterCustomConfig`, passing in the chosen key and a *pointer* to
// the package-scoped instance.
// Then the framework-level configuration file load will populate that struct
// with values that can be used within the test package.

// NOTE: The call to `config.RegisterCustomConfig` must be within the `init()`
// of the package in order to run AFTER the loglevel is set. If this is not done
// then the custom config keys will still be registered, but debug log messages
// will not be shown regardless of the globally configured loglevel.
package config
