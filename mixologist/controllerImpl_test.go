package mixologist_test

import (
	"github.com/cloudendpoints/mixologist/fakes"
	. "github.com/cloudendpoints/mixologist/mixologist"
	gn "github.com/onsi/ginkgo"
	g "github.com/onsi/gomega"
	sc "google/api/servicecontrol/v1"
	"sync"
)

var _ = gn.Describe("ControllerImpl", func() {
	var (
		cfg = ServicesConfig{}
		reg = map[string]CheckerBuilder{
			"fakechecker": fakes.NewCheckerBuilder("fakechecker", nil),
		}
		checkerMgr, _ = NewCheckerManager(reg, &cfg)
		ctrl          = NewControllerImpl(checkerMgr)
	)

	gn.Describe("Given: NewControllerImpl()", func() {
		gn.Context("when: called", func() {
			gn.It("then: return a valid ControllerImpl object", func() {
				g.Expect(ctrl.ReportQueue()).ShouldNot(g.BeNil())
			})
		})
	})
	gn.Describe("Given: Report()", func() {
		var (
			req0 = &sc.ReportRequest{}
		)
		gn.Context("when: called with a valid Report object", func() {
			gn.It("then: should deliver object to queue", func() {
				var wg sync.WaitGroup
				var req *sc.ReportRequest
				wg.Add(1)
				go func() {
					req = <-ctrl.ReportQueue()
					wg.Done()
				}()

				resp, err := ctrl.Report(nil, req0)
				wg.Wait()
				g.Expect(err).To(g.BeNil())
				g.Expect(resp.GetReportErrors()).Should(g.BeNil())
				g.Expect(req).Should(g.Equal(req0))
			})
		})
	})
	gn.Describe("Given: Check()", func() {
		var (
			operationID = "CHECK_OPRN"
			req0        = &sc.CheckRequest{
				Operation: &sc.Operation{
					OperationId: operationID,
				},
			}
		)
		gn.Context("when: called with a valid Check object", func() {
			gn.It("then: should always Succeed and return no error", func() {
				resp, err := ctrl.Check(nil, req0)
				g.Expect(err).To(g.BeNil())
				g.Expect(resp.GetCheckErrors()).Should(g.BeEmpty())
				g.Expect(resp.OperationId).Should(g.Equal(operationID))
			})
		})
	})

})
