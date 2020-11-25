package guardrails_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/tidepool-org/devices/api"
	"github.com/tidepool-org/platform/guardrails"

	"github.com/tidepool-org/platform/data/types/settings/pump"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/guardrails/test"
	"github.com/tidepool-org/platform/pointer"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

var _ = Describe("ValidateInsulinSensitivitySchedule", func() {
	var guardRail *api.InsulinSensitivityGuardRail
	var validator *structureValidator.Validator

	BeforeEach(func() {
		guardRail = test.NewInsulinSensitivityGuardRail()
		validator = structureValidator.New()
	})

	It("doesn't return error with a single valid value", func() {
		var schedule pump.InsulinSensitivityStartArray = []*pump.InsulinSensitivityStart{
			{Amount: pointer.FromFloat64(120.00)},
		}
		guardrails.ValidateInsulinSensitivitySchedule(schedule, guardRail, validator)
		Expect(validator.Error()).To(BeNil())
	})

	It("doesn't return error with multiple valid values", func() {
		var schedule pump.InsulinSensitivityStartArray = []*pump.InsulinSensitivityStart{
			{Amount: pointer.FromFloat64(120.00)},
			{Amount: pointer.FromFloat64(10.00)},
		}
		guardrails.ValidateInsulinSensitivitySchedule(schedule, guardRail, validator)
		Expect(validator.Error()).To(BeNil())
	})

	It("returns an error with an invalid value", func() {
		var schedule pump.InsulinSensitivityStartArray = []*pump.InsulinSensitivityStart{
			{Amount: pointer.FromFloat64(120.00)},
			{Amount: pointer.FromFloat64(120.5)},
			{Amount: pointer.FromFloat64(10.00)},
		}
		expected := errorsTest.WithPointerSource(structureValidator.ErrorValueNotValid(), "/1/amount")
		guardrails.ValidateInsulinSensitivitySchedule(schedule, guardRail, validator)
		errorsTest.ExpectEqual(validator.Error(), expected)
	})
})
