// package config encapsulates loading the config from file, and delegating
// loading of specific test config to that test package.

// To use a custom config section first define a struct into which it will be
// unmarshalled, as you would normally. Then create a package-scoped instance of
// that struct, and choose a key under which data for that struct will be stored
// in the config file.
// Finally, at the top-level of the test package or file add a call to
// `config.RegisterCustomConfig`, passing in the chosen key and a *pointer* to
// the package-scoped instance.
// Then the framework-level configuration file load will populate that struct
// with values that can be used within the test package.
package config
