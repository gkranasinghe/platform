package prescription_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/pointer"

	"github.com/tidepool-org/platform/structure"
	"github.com/tidepool-org/platform/structure/validator"

	"github.com/tidepool-org/platform/prescription"
	"github.com/tidepool-org/platform/prescription/test"
)

var _ = Describe("Initial Settings", func() {
	var settings *prescription.InitialSettings
	var validate structure.Validator

	BeforeEach(func() {
		settings = test.RandomInitialSettings()
		validate = validator.New()
		Expect(validate.Validate(settings)).ToNot(HaveOccurred())
	})

	Describe("Validate", func() {
		BeforeEach(func() {
			validate = validator.New()
		})

		It("fails when pre-meal correction range upper bound is higher than upper bound of the correction range schedule", func() {
			bounds := settings.BloodGlucoseTargetSchedule.GetBounds()
			settings.BloodGlucoseTargetPreprandial.High = pointer.FromFloat64(bounds.Upper + 1)
			settings.Validate(validate)
			Expect(validate.Error()).To(HaveOccurred())
		})

		It("fails when workout correction range lower bound is lower than upper bound of the correction range schedule", func() {
			bounds := settings.BloodGlucoseTargetSchedule.GetBounds()
			settings.BloodGlucoseTargetPhysicalActivity.High = pointer.FromFloat64(bounds.Upper - 1)
			settings.Validate(validate)
			Expect(validate.Error()).To(HaveOccurred())
		})
	})

	Describe("ValidateAllRequired", func() {
		BeforeEach(func() {
			validate = validator.New()
		})

		It("fails with empty basal rate schedule", func() {
			settings.BasalRateSchedule = nil
			settings.ValidateAllRequired(validate)
			Expect(validate.Error()).To(HaveOccurred())
		})

		It("fails with empty blood glucose target schedule", func() {
			settings.BloodGlucoseTargetSchedule = nil
			settings.ValidateAllRequired(validate)
			Expect(validate.Error()).To(HaveOccurred())
		})

		It("fails with empty glucose safety limit", func() {
			settings.GlucoseSafetyLimit = nil
			settings.ValidateAllRequired(validate)
			Expect(validate.Error()).To(HaveOccurred())
		})

		It("fails with empty carbohydrate ratio schedule", func() {
			settings.CarbohydrateRatioSchedule = nil
			settings.ValidateAllRequired(validate)
			Expect(validate.Error()).To(HaveOccurred())
		})

		It("fails with empty insulin sensitivity schedule", func() {
			settings.InsulinSensitivitySchedule = nil
			settings.ValidateAllRequired(validate)
			Expect(validate.Error()).To(HaveOccurred())
		})

		It("fails with empty insulin model", func() {
			settings.InsulinModel = nil
			settings.ValidateAllRequired(validate)
			Expect(validate.Error()).To(HaveOccurred())
		})

		It("fails with empty basal rate maximum", func() {
			settings.BasalRateMaximum = nil
			settings.ValidateAllRequired(validate)
			Expect(validate.Error()).To(HaveOccurred())
		})

		It("fails with empty bolus amount maximum", func() {
			settings.BolusAmountMaximum = nil
			settings.ValidateAllRequired(validate)
			Expect(validate.Error()).To(HaveOccurred())
		})

		It("fails with empty pump id", func() {
			settings.PumpID = ""
			settings.ValidateAllRequired(validate)
			Expect(validate.Error()).To(HaveOccurred())
		})

		It("fails with empty cgm type", func() {
			settings.CgmID = ""
			settings.ValidateAllRequired(validate)
			Expect(validate.Error()).To(HaveOccurred())
		})
	})
})
