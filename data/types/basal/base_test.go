package basal_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	fixtures "github.com/tidepool-org/platform/data/_fixtures"
	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/data/types/basal"
)

var _ = Describe("Basal", func() {

	var helper *types.TestingHelper

	var basalObj = fixtures.TestingDatumBase()
	basalObj["type"] = "basal"
	basalObj["deliveryType"] = "scheduled"
	basalObj["rate"] = 1.0
	basalObj["duration"] = 28800000

	BeforeEach(func() {
		helper = types.NewTestingHelper()
	})

	Context("type from obj", func() {

		It("returns a valid basal type", func() {
			Expect(helper.ValidDataType(basal.Build(basalObj, helper.ErrorProcessing))).To(BeNil())
		})

	})

	Context("validation", func() {

		Context("duration", func() {

			It("is not required", func() {
				delete(basalObj, "duration")
				Expect(helper.ValidDataType(basal.Build(basalObj, helper.ErrorProcessing))).To(BeNil())
			})

			It("fails if < 0", func() {
				basalObj["duration"] = -1

				Expect(
					helper.ErrorIsExpected(
						basal.Build(basalObj, helper.ErrorProcessing),
						types.ExpectedErrorDetails{
							Path:   "0/duration",
							Detail: "Must be  >= 0 and <= 432000000 given '-1'",
						}),
				).To(BeNil())

			})

			It("fails if > 432000000", func() {
				basalObj["duration"] = 432000001

				Expect(
					helper.ErrorIsExpected(
						basal.Build(basalObj, helper.ErrorProcessing),
						types.ExpectedErrorDetails{
							Path:   "0/duration",
							Detail: "Must be  >= 0 and <= 432000000 given '432000001'",
						}),
				).To(BeNil())

			})

			It("valid when greater when >= 0", func() {
				basalObj["duration"] = 0
				Expect(helper.ValidDataType(basal.Build(basalObj, helper.ErrorProcessing))).To(BeNil())
			})
			It("valid when greater when <= 432000000", func() {
				basalObj["duration"] = 432000000
				Expect(helper.ValidDataType(basal.Build(basalObj, helper.ErrorProcessing))).To(BeNil())
			})

		})

		Context("deliveryType", func() {

			It("is required", func() {
				delete(basalObj, "deliveryType")

				Expect(
					helper.ErrorIsExpected(
						basal.Build(basalObj, helper.ErrorProcessing),
						types.ExpectedErrorDetails{
							Path:   "0/deliveryType",
							Detail: "Must be one of scheduled, suspend, temp given '<nil>'",
						}),
				).To(BeNil())

			})

			It("invalid when no matching type", func() {
				basalObj["deliveryType"] = "superfly"
				Expect(
					helper.ErrorIsExpected(
						basal.Build(basalObj, helper.ErrorProcessing),
						types.ExpectedErrorDetails{
							Path:   "0/deliveryType",
							Detail: "Must be one of scheduled, suspend, temp given 'superfly'",
						}),
				).To(BeNil())

			})

			It("invalid if unsupported injected type", func() {
				basalObj["deliveryType"] = "injected"

				Expect(
					helper.ErrorIsExpected(
						basal.Build(basalObj, helper.ErrorProcessing),
						types.ExpectedErrorDetails{
							Path:   "0/deliveryType",
							Detail: "Must be one of scheduled, suspend, temp given 'injected'",
						}),
				).To(BeNil())
			})

			It("valid if scheduled type", func() {
				basalObj["deliveryType"] = "scheduled"
				Expect(helper.ValidDataType(basal.Build(basalObj, helper.ErrorProcessing))).To(BeNil())
			})

			It("valid if suspend type", func() {
				basalObj["deliveryType"] = "suspend"
				Expect(helper.ValidDataType(basal.Build(basalObj, helper.ErrorProcessing))).To(BeNil())
			})

			It("valid if temp type", func() {
				basalObj["deliveryType"] = "temp"
				Expect(helper.ValidDataType(basal.Build(basalObj, helper.ErrorProcessing))).To(BeNil())
			})

		})
	})
})
