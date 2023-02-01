// Test Manifest records the data about a test; for example its name, identifier, preconditions, and other labels.
package test_manifest

type TestLabel struct {
	label string
	value string
}

type TestManifest struct {
	identifer 	string
	name 		string
	labels		[]TestLabel
}
