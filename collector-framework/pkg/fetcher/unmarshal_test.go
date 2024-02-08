// SPDX-License-Identifier: GPL-2.0-or-later

package fetcher //nolint:testpackage // testing internal functions

import (
	"reflect"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type testStruct struct { //nolint:govet // changing the order would break the tests as it relies on the ordering
	TestString   string        `fetcherKey:"str"`
	TestSlice    []int         `fetcherKey:"slice"`
	TestDuration time.Duration `fetcherKey:"duration"`
	TestStuct    nestedStuct   `fetcherKey:"struct"`
}

type nestedStuct struct {
	Value string
}

var _ = Describe("setValueOnField", func() {
	When("calling setValueOnField with the correct type", func() {
		It("should populate the field", func() {
			target := &testStruct{}
			testStrValue := "I am a test string"
			strField := reflect.TypeOf(target).Elem().Field(0)
			strFieldVal := reflect.ValueOf(target).Elem().FieldByIndex(strField.Index)
			err := setValueOnField(strFieldVal, testStrValue)
			Expect(err).NotTo(HaveOccurred())
			Expect(target.TestString).To(Equal(testStrValue))

			testSliceValue := []int{12, 15, 18}
			sliceField := reflect.TypeOf(target).Elem().Field(1)
			sliceFieldVal := reflect.ValueOf(target).Elem().FieldByIndex(sliceField.Index)
			err = setValueOnField(sliceFieldVal, testSliceValue)
			Expect(err).NotTo(HaveOccurred())
			Expect(target.TestSlice).To(Equal(testSliceValue))

		})
	})
	When("calling setValueOnField with a non trivial type", func() {
		It("should populate the field", func() {
			target := &testStruct{}
			testDuration := 10 * time.Second
			sliceField := reflect.TypeOf(target).Elem().Field(2)
			sliceFieldVal := reflect.ValueOf(target).Elem().FieldByIndex(sliceField.Index)
			err := setValueOnField(sliceFieldVal, testDuration)
			Expect(err).NotTo(HaveOccurred())
			Expect(target.TestDuration).To(Equal(testDuration))

		})
	})
	When("calling setValueOnField with a struct type", func() {
		It("should populate the field", func() {
			target := &testStruct{}
			toNest := nestedStuct{Value: "I am nested"}
			sliceField := reflect.TypeOf(target).Elem().Field(3)
			sliceFieldVal := reflect.ValueOf(target).Elem().FieldByIndex(sliceField.Index)
			err := setValueOnField(sliceFieldVal, toNest)
			Expect(err).NotTo(HaveOccurred())
			Expect(target.TestStuct).To(Equal(toNest))
		})
	})
	When("calling setValueOnField with the incorrect type", func() {
		It("should return an error", func() {
			target := &testStruct{}
			testNonStrValue := 12
			strField := reflect.TypeOf(target).Elem().Field(0)
			strFieldVal := reflect.ValueOf(target).Elem().FieldByIndex(strField.Index)
			err := setValueOnField(strFieldVal, testNonStrValue)
			Expect(err).To(HaveOccurred())

			testBadSliceValue := []string{"12", "15", "18"}
			sliceField := reflect.TypeOf(target).Elem().Field(1)
			sliceFieldVal := reflect.ValueOf(target).Elem().FieldByIndex(sliceField.Index)
			err = setValueOnField(sliceFieldVal, testBadSliceValue)
			Expect(err).To(HaveOccurred())
		})
	})
})

var _ = Describe("unmarshal", func() {
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
