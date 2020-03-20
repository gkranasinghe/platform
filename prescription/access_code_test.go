package prescription_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/prescription"
)

var _ = Describe("GenerateAccessCode", func() {
	It("generates an alphanumeric code", func() {
		code := prescription.GenerateAccessCode()
		Expect(code).To(MatchRegexp("^[A-Z0-9]+$"))
	})

	It("generates a code with length of 6 characters", func() {
		code := prescription.GenerateAccessCode()
		Expect(code).To(HaveLen(6))
	})
})
