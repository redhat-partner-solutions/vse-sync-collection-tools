// Copyright 2023 Red Hat, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// run_tests provides the entrypoint into the tests
package run_tests //nolint:revive,stylecheck,testpackage //entrypoint to framework, we will allow two words

import (
	"flag"
	"testing"

	. "github.com/onsi/ginkgo/v2" //nolint:stylecheck // ginkgo and gomega are dot imports by convention.
	. "github.com/onsi/gomega"    //nolint:stylecheck // ginkgo and gomega are dot imports by convention.

	log "github.com/sirupsen/logrus"

	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/clients"
	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/config"
	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/env"

	_ "github.com/redhat-partner-solutions/vse-sync-testsuite/tests/ptp"
)

var (
	configPath string
	envPath    string
)

// This is a workaround to set the LogLevel to Debug before flags are parsed or the config is loaded. It is necessary
// because flags can only be parsed once globally which must be after ginkgo registers its flags, and because logging
// needs to happen when config sections are registered by Suite packages.
func setLogLevel() bool {
	log.SetLevel(log.DebugLevel)
	return true
}

var _ = setLogLevel()

// The workaround could be made unnecessary if the framework had control over when the test Suite packages are loaded.
// One possibility for this would be dynamically loading Suite packages as Go plugins but that method has different
// compromises and restrictions. See https://github.com/redhat-partner-solutions/vse-sync-testsuite/tree/plugin-test/tests

func init() {
	flag.StringVar(&configPath, "config", "./config.yaml", "path to the config file")
	flag.StringVar(&envPath, "env", "./env.yaml", "path to the environment file")
}

// TestRun is the entrypoint for the framework trigger execution of all
// discovered Ginkgo specs
func TestRun(t *testing.T) {
	flag.Parse() // Ginkgo will trigger the parse, and a second call is ignored, so this does nothing in most cases.
	log.Info("Starting Tests")

	if err := config.LoadConfigFromFile(configPath); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	clusterInfo, err := env.LoadClusterDefFromFile(envPath)
	if err != nil {
		log.Fatal(err)
	}
	clients.GetClientset(clusterInfo.KubeconfigPath)

	RegisterFailHandler(Fail)
	RunSpecs(t, "Tests Suite")

	log.Info("Finished Tests")
}
