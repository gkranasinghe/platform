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

var _ = Describe("ValidateBasalRateSchedule", func() {
	var guardRail *api.BasalRatesGuardRail
	var validator *structureValidator.Validator

	BeforeEach(func() {
		guardRail = test.NewBasalRatesGuardRail()
		validator = structureValidator.New()
	})

	It("doesn't return error with a single valid value", func() {
		var schedule pump.BasalRateStartArray = []*pump.BasalRateStart{
			{Rate: pointer.FromFloat64(0.55)},
		}
		guardrails.ValidateBasalRateSchedule(schedule, guardRail, validator)
		Expect(validator.Error()).To(BeNil())
	})

	It("doesn't return error with multiple valid values", func() {
		var schedule pump.BasalRateStartArray = []*pump.BasalRateStart{
			{Rate: pointer.FromFloat64(0.55)},
			{Rate: pointer.FromFloat64(15.55)},
		}
		guardrails.ValidateBasalRateSchedule(schedule, guardRail, validator)
		Expect(validator.Error()).To(BeNil())
	})

	It("returns an error with an invalid value", func() {
		var schedule pump.BasalRateStartArray = []*pump.BasalRateStart{
			{Rate: pointer.FromFloat64(0.55)},
			{Rate: pointer.FromFloat64(0.56)},
			{Rate: pointer.FromFloat64(15.55)},
		}
		expected := errorsTest.WithPointerSource(structureValidator.ErrorValueNotValid(), "/1/rate")
		guardrails.ValidateBasalRateSchedule(schedule, guardRail, validator)
		errorsTest.ExpectEqual(validator.Error(), expected)
	})
})
