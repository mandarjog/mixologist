package mixologist_test

import (
	gn "github.com/onsi/ginkgo"
	g "github.com/onsi/gomega"
	"somnacin-internal/mixologist/fakes"
	. "somnacin-internal/mixologist/mixologist"
)

var _ = gn.Describe("registry", func() {
	var (
		name      = "testRC"
		rcbuilder = fakes.NewBuilder(name, nil)
	)
	gn.Describe("given: RegisterReportConsumer()", func() {
		gn.Context("when: passed a valid registration request", func() {
			gn.BeforeEach(func() {
				ReportConsumerRegistry = make(map[string]ReportConsumerBuilder)
				RegisterReportConsumer(name, rcbuilder)
			})
			gn.It("then: should add the specified consumer", func() {
				g.Expect(ReportConsumerRegistry).ShouldNot(g.BeEmpty())
				g.Expect(ReportConsumerRegistry[name]).To(g.Equal(rcbuilder))
			})
			gn.It("then: should re-add the specified consumer if its the same", func() {
				g.Expect(RegisterReportConsumer(name, rcbuilder)).To(g.Succeed())
			})
			gn.It("then: should fail re-add the specified consumer if its NOT same", func() {
				g.Expect(RegisterReportConsumer(name, fakes.NewBuilder(name, nil))).NotTo(g.Succeed())
			})
		})
	})
})
