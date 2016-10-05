package mixologist_test

import (
	gn "github.com/onsi/ginkgo"
	g "github.com/onsi/gomega"
	sc "google/api/servicecontrol/v1"
	"github.com/cloudendpoints/mixologist/fakes"
	. "github.com/cloudendpoints/mixologist/mixologist"
)

var _ = gn.Describe("ReportConsumerManager", func() {
	var (
		name0      = "testRC0"
		rcbuilder0 = fakes.NewBuilder(name0, nil)
		name1      = "testRC1"
		rcbuilder1 = fakes.NewBuilder(name1, nil)
		rqChan     chan *sc.ReportRequest
		rcMgr      *ReportConsumerManagerImpl
	)
	gn.Describe("Given: NewReportConsumerManager()", func() {
		gn.BeforeEach(func() {
			rqChan = make(chan *sc.ReportRequest)
			ReportConsumerRegistry = make(map[string]ReportConsumerBuilder)
			RegisterReportConsumer(name0, rcbuilder0)
			RegisterReportConsumer(name1, rcbuilder1)
			rcMgr = NewReportConsumerManager(rqChan, ReportConsumerRegistry, Config{ReportConsumers: []string{"name", name0, name1}})
		})
		gn.Context("when: called with correct params", func() {
			gn.It("then: returns an initialized Manager", func() {
				g.Expect(rcMgr.GetPrefixAndHandlers()).NotTo(g.BeEmpty())
			})
			gn.It("then: consumes and delivers messages to all consumers", func() {
				var (
					req0      = &sc.ReportRequest{}
					req1      = &sc.ReportRequest{}
					consumer0 = rcbuilder0.Consumer
					consumer1 = rcbuilder1.Consumer
					req       = []*sc.ReportRequest{req0, req1}
				)
				g.Expect(rcMgr.GetPrefixAndHandlers()).NotTo(g.BeNil())
				rcMgr.Start(2)

				// send 2 messages down the pipe
				for i := 0; i < len(req); i++ {
					rqChan <- req[i]
				}
				// ensures async is taken care of

				g.Eventually(func() []*sc.ReportRequest {
					return consumer0.GetMessages()
				}).Should(g.HaveLen(len(req)))

				g.Eventually(func() []*sc.ReportRequest {
					return consumer1.GetMessages()
				}).Should(g.HaveLen(len(req)))

				g.Expect(consumer0.GetMessages()).Should(g.Equal(req))
				g.Expect(consumer1.GetMessages()).Should(g.Equal(req))

			})
		})
	})
})
