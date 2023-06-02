// SPDX-License-Identifier: GPL-2.0-or-later

package clients_test

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/clients"
)

var _ = Describe("Cmd", func() {
	When("NewCmd is called ", func() {
		It("should return a Cmd ", func() {
			cmd, err := clients.NewCmd("TestKey", []string{"Hello This is a test"})
			Expect(cmd).ToNot(BeNil())
			Expect(err).ToNot(HaveOccurred())
		})
	})
	When("Cmd is passed a vaild result", func() {
		It("should parse it into a dict", func() {
			expected := "I am the correct answer"
			key := "TestKey"
			cmd, _ := clients.NewCmd(key, []string{"Hello This is a test"})
			result, err := cmd.ExtractResult(fmt.Sprintf("<%s>\n%s\n</%s>\n", key, expected, key))
			Expect(result[key]).To(Equal(expected))
			Expect(err).ToNot(HaveOccurred())
		})
	})
	When("Cmd is passed an invaild result", func() {
		It("should return an error", func() {
			cmd, _ := clients.NewCmd("TestKey", []string{"Hello This is a test"})
			result, err := cmd.ExtractResult("<SomethingElse>\nI am not the correct answer\n</SomethingElse>\n")
			Expect(err).To(HaveOccurred())
			Expect(result).To(BeEmpty())
		})
	})

	When("Cmd when supplied a clean up function", func() {
		It("should use it", func() {
			part1 := "I am part"
			part2 := "of the correct answer"
			key := "TestKey"
			cmd, _ := clients.NewCmd(key, []string{"Hello This is a test"})
			cmd.SetCleanupFunc(func(p1 string) string { return p1 + part2 })
			result, err := cmd.ExtractResult(fmt.Sprintf("<%s>\n%s\n</%s>\n", key, part1, key))
			Expect(result[key]).To(Equal(part1 + part2))
			Expect(err).ToNot(HaveOccurred())
		})
	})
})
