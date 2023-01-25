package main

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	log "github.com/sirupsen/logrus"
)

var (
	kubeconfig string
	count int
)

var _ = Describe("Tests in a second Plugin", func() {
	var _ = BeforeEach(func() {
		count += 1
		log.Infof("Incremented count, new value: %d", count)
	})
	When("There is a test in a second Go plugin", func() {
		It("should be run isolated from other plugins", func() {
			// Test logic goes here
			log.Infof("Running first test case %d from a second plugin", count)
			Expect(kubeconfig).NotTo(BeEmpty())
			Expect(count).To(Equal(43))
		})
		It("should be run isolated from other plugins", func() {
			// Test logic goes here
			log.Infof("Running second test case %d from a second plugin", count)
			Expect(kubeconfig).NotTo(BeEmpty())
			Expect(count).To(Equal(44))
		})
	})
})

func Configure(kc string, c int) {
	log.Infof("Configure in plugin: %s, %d", kc, c)
	kubeconfig = kc
	count = c
}

// // WARNING: DO NOT use top-level setup and cleanup methods: they will
// // escape the scope of the plugin. and run for every spec in other plugins.
// var _ = BeforeEach(func() {
// 	count += 1
// 	log.Infof("Incremented count, new value: %d", count)
// })
