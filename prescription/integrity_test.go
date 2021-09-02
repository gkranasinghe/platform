package prescription_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"

	"github.com/tidepool-org/platform/data/blood/glucose"

	"github.com/tidepool-org/platform/data/types/settings/pump"
	"github.com/tidepool-org/platform/prescription"
	"github.com/tidepool-org/platform/prescription/test"
)

var _ = Describe("Integrity hash", func() {
	Context("NewIntegrityAttributesFromRevisionCreate", func() {
		var revisionCreate *prescription.RevisionCreate

		BeforeEach(func() {
			revisionCreate = test.RandomRevisionCreate()
		})

		It("sets all integrity attributes correctly", func() {
			attrs := prescription.NewIntegrityAttributesFromRevisionCreate(*revisionCreate)
			matchAttributes := MatchAllFields(Fields{
				"DataAttributes": Equal(revisionCreate.DataAttributes),
				"CreatedUserID":  Equal(revisionCreate.CreatedUserID),
			})
			Expect(attrs).To(matchAttributes)
		})
	})

	Context("NewIntegrityAttributesFromRevision", func() {
		var revision *prescription.Revision

		BeforeEach(func() {
			revision = test.RandomRevision()
		})

		It("sets all integrity attributes correctly", func() {
			attrs := prescription.NewIntegrityAttributesFromRevision(*revision)
			matchAttributes := MatchAllFields(Fields{
				"DataAttributes": Equal(revision.Attributes.DataAttributes),
				"CreatedUserID":  Equal(revision.Attributes.CreatedUserID),
			})
			Expect(attrs).To(matchAttributes)
		})
	})

	Context("MustGenerateIntegrityHash", func() {
		var revision *prescription.Revision
		var attrs prescription.IntegrityAttributes

		BeforeEach(func() {
			revision = test.RandomRevision()
		})

		It("generates hash with the expected algorithm", func() {
			attrs = prescription.NewIntegrityAttributesFromRevision(*revision)
			hash := prescription.MustGenerateIntegrityHash(attrs)
			Expect(hash.Algorithm).To(Equal("JCSSHA512"))
		})

		It("generates hash with length of 128", func() {
			attrs = prescription.NewIntegrityAttributesFromRevision(*revision)
			hash := prescription.MustGenerateIntegrityHash(attrs)
			Expect(hash.Hash).To(HaveLen(128))
		})

		It("generates a different hash for different attributes", func() {
			attrs = prescription.NewIntegrityAttributesFromRevision(*revision)
			first := prescription.MustGenerateIntegrityHash(attrs)
			revision = test.RandomRevision()
			attrs = prescription.NewIntegrityAttributesFromRevision(*revision)
			second := prescription.MustGenerateIntegrityHash(attrs)
			Expect(first.Hash).ToNot(Equal(second.Hash))
		})

		It("generates correct hash for the fixture", func() {
			attrs = test.IntegrityAttributes
			hash := prescription.MustGenerateIntegrityHash(attrs)
			Expect(hash.Algorithm).To(Equal("JCSSHA512"))
			Expect(hash.Hash).To(Equal(test.ExpectedHash))
		})
	})

	Context("Prescription settings constants", func() {
		It("don't change", func() {
			// When this test fails it means that some of the constants used in prescriptions have changed,
			// which will cause prescription integrity checks to fail. Those constant shouldn't be changed
			// in normal circumstances.
			Expect(prescription.StateActive).To(Equal("active"))
			Expect(prescription.StateClaimed).To(Equal("claimed"))
			Expect(prescription.StateSubmitted).To(Equal("submitted"))
			Expect(prescription.StateDraft).To(Equal("draft"))
			Expect(prescription.StateExpired).To(Equal("expired"))
			Expect(prescription.StateInactive).To(Equal("inactive"))
			Expect(prescription.StatePending).To(Equal("pending"))
			Expect(prescription.AccountTypePatient).To(Equal("patient"))
			Expect(prescription.AccountTypeCaregiver).To(Equal("caregiver"))
			Expect(prescription.CalculatorMethodTotalDailyDose).To(Equal("totalDailyDose"))
			Expect(prescription.CalculatorMethodTotalDailyDoseAndWeight).To(Equal("totalDailyDoseAndWeight"))
			Expect(prescription.CalculatorMethodWeight).To(Equal("weight"))
			Expect(prescription.SexFemale).To(Equal("female"))
			Expect(prescription.SexMale).To(Equal("male"))
			Expect(prescription.SexUndisclosed).To(Equal("undisclosed"))
			Expect(prescription.TherapySettingInitial).To(Equal("initial"))
			Expect(prescription.TherapySettingTransferPumpSettings).To(Equal("transferPumpSettings"))
			Expect(prescription.TrainingInModule).To(Equal("inModule"))
			Expect(prescription.TrainingInPerson).To(Equal("inPerson"))
			Expect(prescription.UnitKg).To(Equal("kg"))
			Expect(prescription.UnitLbs).To(Equal("lbs"))
			Expect(pump.BolusAmountMaximumUnitsUnits).To(Equal("Units"))
			Expect(pump.BasalRateMaximumUnitsUnitsPerHour).To(Equal("Units/hour"))
			Expect(glucose.MgdL).To(Equal("mg/dL"))
		})
	})
})
