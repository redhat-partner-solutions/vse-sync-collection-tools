// SPDX-License-Identifier: GPL-2.0-or-later

package callbacks_test

import (
	"bytes"
	"errors"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/redhat-partner-solutions/vse-sync-collection-tools/pkg/callbacks"
)

func NewTestFile() *testFile {
	return &testFile{Buffer: *bytes.NewBuffer([]byte{}), open: true}
}

type testFile struct {
	bytes.Buffer
	open bool
}

func (t *testFile) Close() error {
	if t.open {
		t.open = false
		return nil
	}
	return errors.New("File is already closed") // TODO look up actual errors
}

type testOutputType struct {
	Msg string `json:"msg"`
	// responder func() (*callbacks.AnalyserFormatType, error)
}

func (t *testOutputType) GetAnalyserFormat() ([]*callbacks.AnalyserFormatType, error) {
	fomatted := callbacks.AnalyserFormatType{
		ID:   "testOutput",
		Data: []any{"Hello"},
	}
	return []*callbacks.AnalyserFormatType{&fomatted}, nil
}

var _ = Describe("Callbacks", func() {
	var mockedFile *testFile

	BeforeEach(func() {
		mockedFile = NewTestFile()
	})

	When("Raw FileCallback is called", func() {
		It("should write to the file", func() {
			callback := callbacks.NewFileCallback(mockedFile, callbacks.Raw)
			out := testOutputType{
				Msg: "This is a test line",
			}
			err := callback.Call(&out, "testOut")
			Expect(err).NotTo(HaveOccurred())
			Expect(mockedFile.ReadString('\n')).To(ContainSubstring("This is a test line"))
		})
	})
	When("JSON FileCallback is called", func() {
		It("should write to the file", func() {
			callback := callbacks.NewFileCallback(mockedFile, callbacks.AnalyserJSON)
			out := testOutputType{
				Msg: "This is a test line",
			}
			err := callback.Call(&out, "testOut")
			Expect(err).NotTo(HaveOccurred())
			Expect(mockedFile.ReadString('\n')).To(Equal("{\"data\":[\"Hello\"],\"id\":\"testOutput\"}\n"))
		})
	})
	When("A FileCallback is cleaned up", func() {
		It("should close the file", func() {
			callback := callbacks.NewFileCallback(mockedFile, callbacks.Raw)
			err := callback.CleanUp()
			Expect(err).NotTo(HaveOccurred())
			Expect(mockedFile.open).To(BeFalse())
		})
	})

})

func TestCommand(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Clients Suite")
}
