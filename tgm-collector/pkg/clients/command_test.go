// SPDX-License-Identifier: GPL-2.0-or-later

package clients_test

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/redhat-partner-solutions/vse-sync-collection-tools/tgm-collector/pkg/clients"
)

var _ = Describe("Cmd", func() {
	When("NewCmd is called ", func() {
		It("should return a Cmd ", func() {
			cmd, err := clients.NewCmd("TestKey", "Hello This is a test")
			Expect(cmd).ToNot(BeNil())
			Expect(err).ToNot(HaveOccurred())
		})
	})
	When("Cmd is passed a vaild result", func() {
		It("should parse it into a dict", func() {
			expected := "I am the correct answer"
			key := "TestKey"
			cmd, _ := clients.NewCmd(key, "Hello This is a test")
			result, err := cmd.ExtractResult(fmt.Sprintf("<%s>\n%s\n</%s>\n", key, expected, key))
			Expect(result[key]).To(Equal(expected))
			Expect(err).ToNot(HaveOccurred())
		})
	})
	When("Cmd is passed an invaild result", func() {
		It("should return an error", func() {
			cmd, _ := clients.NewCmd("TestKey", "Hello This is a test")
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
			cmd, _ := clients.NewCmd(key, "Hello This is a test")
			cmd.SetOutputProcessor(func(p1 string) (string, error) { return p1 + part2, nil })
			result, err := cmd.ExtractResult(fmt.Sprintf("<%s>\n%s\n</%s>\n", key, part1, key))
			Expect(result[key]).To(Equal(part1 + part2))
			Expect(err).ToNot(HaveOccurred())
		})
	})
})

var _ = Describe("CmdGrp", func() {
	When("a command is added", func() {
		It("should be added to GetCommand's output", func() {
			cmd, err := clients.NewCmd("TestKey", "Hello This is a test")
			Expect(err).ToNot(HaveOccurred())
			cmdGrp := &clients.CmdGroup{}
			cmdGrp.AddCommand(cmd)
			cmdString := cmdGrp.GetCommand()
			Expect(cmdString).To(Equal(cmd.GetCommand()))

			cmd2, err := clients.NewCmd("TestKey2", "This is another test goodbye.")
			Expect(err).ToNot(HaveOccurred())
			cmdGrp.AddCommand(cmd2)
			cmdString2 := cmdGrp.GetCommand()
			Expect(cmdString2).To(Equal(cmd.GetCommand() + cmd2.GetCommand()))
		})
	})
	When("passed a valid result", func() {
		It("should set the values to the key", func() {
			key1 := "TestKey"
			key2 := "TestKey2"
			answer1 := "Result of key1"
			answer2 := "Result of key2"
			cmd, err := clients.NewCmd(key1, "Hello This is a test")
			Expect(err).ToNot(HaveOccurred())
			cmd2, err := clients.NewCmd(key2, "This is another test goodbye.")
			Expect(err).ToNot(HaveOccurred())
			cmdGrp := &clients.CmdGroup{}
			cmdGrp.AddCommand(cmd)
			cmdGrp.AddCommand(cmd2)
			result, err := cmdGrp.ExtractResult(fmt.Sprintf(
				"<%s>\n%s\n</%s>\n<%s>\n%s\n</%s>\n",
				key1, answer1, key1,
				key2, answer2, key2,
			))
			Expect(result[key1]).To(Equal(answer1))
			Expect(result[key2]).To(Equal(answer2))
			Expect(err).ToNot(HaveOccurred())
		})
	})
})
