package api_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/auth/service"
	"github.com/tidepool-org/platform/auth/service/api"
	testService "github.com/tidepool-org/platform/auth/service/test"
	testRest "github.com/tidepool-org/platform/test/rest"
)

var _ = Describe("StatusGet", func() {
	var response *testRest.ResponseWriter
	var request *rest.Request
	var svc *testService.Service
	var rtr *api.Router

	BeforeEach(func() {
		response = testRest.NewResponseWriter()
		request = testRest.NewRequest()
		svc = testService.NewService()
		var err error
		rtr, err = api.NewRouter(svc)
		Expect(err).ToNot(HaveOccurred())
		Expect(rtr).ToNot(BeNil())
	})

	AfterEach(func() {
		svc.Expectations()
		response.Expectations()
	})

	Context("StatusGet", func() {
		It("panics if response is missing", func() {
			Expect(func() { rtr.StatusGet(nil, request) }).To(Panic())
		})

		It("panics if request is missing", func() {
			Expect(func() { rtr.StatusGet(response, nil) }).To(Panic())
		})

		Context("with service status", func() {
			var sts *service.Status

			BeforeEach(func() {
				sts = &service.Status{}
				svc.StatusOutputs = []*service.Status{sts}
				response.WriteJsonOutputs = []error{nil}
			})

			It("returns successfully", func() {
				rtr.StatusGet(response, request)
				Expect(response.WriteJsonInputs).To(HaveLen(1))
				Expect(response.WriteJsonInputs[0]).To(Equal(sts))
			})
		})
	})
})