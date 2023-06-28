// SPDX-License-Identifier: GPL-2.0-or-later

package fetcher //nolint:testpackage // testing internal functions

import (
	"reflect"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type testStruct struct {
	TestString string `fetcherKey:"str"`
	TestSlice  []int  `fetcherKey:"slice"`
}

var _ = Describe("unmarshal", func() {
	When("calling setValue with the correct type", func() {
		It("should populate the field", func() {
			target := &testStruct{}
			testStrValue := "I am a test string"
			strField := reflect.TypeOf(target).Elem().Field(0)
			strFieldVal := reflect.ValueOf(target).Elem().FieldByIndex(strField.Index)
			err := setValue(&strField, strFieldVal, testStrValue)
			Expect(err).NotTo(HaveOccurred())
			Expect(target.TestString).To(Equal(testStrValue))

			testSliceValue := []int{12, 15, 18}
			sliceField := reflect.TypeOf(target).Elem().Field(1)
			sliceFieldVal := reflect.ValueOf(target).Elem().FieldByIndex(sliceField.Index)
			err = setValue(&sliceField, sliceFieldVal, testSliceValue)
			Expect(err).NotTo(HaveOccurred())
			Expect(target.TestSlice).To(Equal(testSliceValue))

		})
	})
	When("calling setValue with the incorrect type", func() {
		It("should return an error", func() {
			target := &testStruct{}
			testNonStrValue := 12
			strField := reflect.TypeOf(target).Elem().Field(0)
			strFieldVal := reflect.ValueOf(target).Elem().FieldByIndex(strField.Index)
			err := setValue(&strField, strFieldVal, testNonStrValue)
			Expect(err).To(HaveOccurred())

			testBadSliceValue := []string{"12", "15", "18"}
			sliceField := reflect.TypeOf(target).Elem().Field(1)
			sliceFieldVal := reflect.ValueOf(target).Elem().FieldByIndex(sliceField.Index)
			err = setValue(&sliceField, sliceFieldVal, testBadSliceValue)
			Expect(err).To(HaveOccurred())
		})
	})
	When("calling unmarshal with the correct types", func() {
		It("should populate the target", func() {
			target := &testStruct{}
			values := make(map[string]any)
			values["str"] = "I am a test string"
			values["slice"] = []int{1, 2, 3, 4}
			err := unmarshal(values, target)
			Expect(err).NotTo(HaveOccurred())
			Expect(target.TestString).To(Equal(values["str"]))
			Expect(target.TestSlice).To(Equal(values["slice"]))
		})
	})
	When("calling unmarshal with the incorrect str type", func() {
		It("should return an err", func() {
			target := &testStruct{}
			values := make(map[string]any)
			values["str"] = 12
			values["slice"] = []int{1, 2, 3, 4}
			err := unmarshal(values, target)
			Expect(err).To(HaveOccurred())
		})
	})
	When("calling unmarshal with the incorrect slice type", func() {
		It("should return an err", func() {
			target := &testStruct{}
			values := make(map[string]any)
			values["str"] = "I am a test string"
			values["slice"] = []string{"1", "2", "3", "4"}
			err := unmarshal(values, target)
			Expect(err).To(HaveOccurred())
		})
	})
})

func TestFetcher(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Fetcher Suite")
}
