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

package config_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/config"
)

const (
	simpleSectionKey  = "tests.simple_config"
	complexSectionKey = "tests.complex_config"
)

type SimpleCustomTestConfig struct {
	TargetLatency   int `yaml:"target_latency,omitempty"`
	PermittedJitter int `yaml:"permitted_jitter,omitempty"`
}

// Check that a more complex struct will be handled correctly.
type ComplexCustomTestConfig struct {
	InterfaceName   string                 `yaml:"interface_name,omitempty"`
	TargetIpAddress string                 `yaml:"target_ip,omitempty"`
	LatencyConfig   SimpleCustomTestConfig `yaml:"latency,omitempty"`
}

var _ = Describe("Registry", Ordered, func() {
	simpleConfigInstance := SimpleCustomTestConfig{}
	complexConfigInstance := ComplexCustomTestConfig{}
	When("registering a section", func() {
		It("should register a new section without error", func() {
			err := config.RegisterCustomConfig(simpleSectionKey, &simpleConfigInstance)
			Expect(err).NotTo(HaveOccurred())

			err = config.RegisterCustomConfig(complexSectionKey, &complexConfigInstance)
			Expect(err).NotTo(HaveOccurred())
		})
		It("should error if a section has already been registered", func() {
			err := config.RegisterCustomConfig(simpleSectionKey, &simpleConfigInstance)
			Expect(err).To(HaveOccurred())
		})
	})
	When("loading a config file", func() {
		It("should correctly populate the registered struct instances when loading a config", func() {
			err := config.LoadConfigFromFile("test_files/test_config.yaml")
			Expect(err).NotTo(HaveOccurred())

			Expect(simpleConfigInstance.TargetLatency).To(Equal(120))
			Expect(simpleConfigInstance.PermittedJitter).To(Equal(3))

			Expect(complexConfigInstance.InterfaceName).To(Equal("en0"))
			Expect(complexConfigInstance.TargetIpAddress).To(Equal("8.8.8.8"))
			Expect(complexConfigInstance.LatencyConfig.TargetLatency).To(Equal(999))
			Expect(complexConfigInstance.LatencyConfig.PermittedJitter).To(Equal(1))
		})
		It("should continue silently if there an unregistered section key", func() {
			err := config.LoadConfigFromFile("test_files/test_extra_section.yaml")
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
