package test_manifest

type TestManifest struct {
	identifer string `yaml:"identifier"`
	name      string `yaml:"name"`
	labels    map[string]string `yaml:"labels"`
}
