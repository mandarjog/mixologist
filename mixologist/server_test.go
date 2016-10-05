package mixologist_test

import (
	"bytes"
	"errors"
	"github.com/golang/protobuf/proto"
	gn "github.com/onsi/ginkgo"
	g "github.com/onsi/gomega"
	sc "google/api/servicecontrol/v1"
	"io"
	"net/http"
	"net/http/httptest"
	"github.com/cloudendpoints/mixologist/fakes"
	. "github.com/cloudendpoints/mixologist/mixologist"
	"github.com/cloudendpoints/mixologist/testutils"
	"strings"
)

var _ = gn.Describe("Server", func() {
	const (
		hosturl = "http://mytest"
	)
	gn.Describe("Given: ServerHTTP()", func() {
		const (
			prefix        = "/mymetrics"
			serviceName   = "service007"
			servicePrefix = "/services/v1/" + serviceName
		)
		var (
			phi         = []*PrefixAndHandler{fakes.BuildPrefixAndHandler(prefix)}
			ctrl        = fakes.NewController()
			hndlr       *Handler
			w           *httptest.ResponseRecorder
			operationId string
		)
		gn.BeforeEach(func() {
			ctrl = fakes.NewController()
			hndlr = NewHandler(ctrl, phi)
			w = httptest.NewRecorder()
			operationId = fakes.UUID()
		})
		gn.Context("when: called with :check request", func() {
			gn.It("then: Should deliver the message to contrller.Check() ", func() {
				rqpb := testutils.CreateCheck(
					&testutils.ExpectedCheck{
						ServiceName:   serviceName,
						OperationName: "getfiles",
						OperationId:   operationId,
					})
				rqbytes, err := proto.Marshal(&rqpb)
				g.Expect(err).Should(g.BeNil())
				req := httptest.NewRequest("POST", servicePrefix+CheckSuffix, bytes.NewReader(rqbytes))

				hndlr.ServeHTTP(w, req)

				g.Expect(w.Code).Should(g.Equal(http.StatusOK))
				resp := &sc.CheckResponse{}
				proto.Unmarshal(w.Body.Bytes(), resp)
				g.Expect(resp.OperationId).Should(g.Equal(operationId))
				g.Expect(proto.Equal(&rqpb, ctrl.SpyCR)).To(g.BeTrue())
			})
		})
		gn.Context("when: called with :report request", func() {
			gn.It("then: Should deliver the message to contrller.Report() ", func() {
				rqpb := testutils.CreateReport(
					&testutils.ExpectedReport{
						ApiName:     serviceName,
						ApiMethod:   "getfiles",
						OperationId: operationId,
					})
				rqbytes, err := proto.Marshal(&rqpb)
				g.Expect(err).Should(g.BeNil())
				req := httptest.NewRequest("POST", servicePrefix+ReportSuffix, bytes.NewReader(rqbytes))

				hndlr.ServeHTTP(w, req)

				g.Expect(w.Code).Should(g.Equal(http.StatusOK))
				resp := &sc.ReportResponse{}
				proto.Unmarshal(w.Body.Bytes(), resp)
				g.Expect(proto.Equal(&rqpb, ctrl.SpyRR)).To(g.BeTrue())

			})
		})
		gn.Context("when: called with :check request and controller.Check returns error", func() {
			gn.It("then: returns StatusInternalServerError ", func() {
				rqpb := testutils.CreateCheck(
					&testutils.ExpectedCheck{
						ServiceName:   serviceName,
						OperationName: "getfiles",
						OperationId:   operationId,
					})
				rqbytes, err := proto.Marshal(&rqpb)
				g.Expect(err).Should(g.BeNil())

				req := httptest.NewRequest("POST", servicePrefix+CheckSuffix, bytes.NewReader(rqbytes))
				ctrl.PlantedError = errors.New("Check Returned Error")
				hndlr.ServeHTTP(w, req)
				g.Expect(w.Code).Should(g.Equal(http.StatusInternalServerError))
				g.Expect(w.Body.String()).Should(g.Equal(ctrl.PlantedError.Error()))

			})
		})
		gn.Context("when: called with :report request and controller.Report returns error", func() {
			gn.It("then: returns StatusInternalServerError ", func() {
				rqpb := testutils.CreateReport(
					&testutils.ExpectedReport{
						ApiName:     serviceName,
						ApiMethod:   "getfiles",
						OperationId: operationId,
					})
				rqbytes, err := proto.Marshal(&rqpb)
				g.Expect(err).Should(g.BeNil())

				req := httptest.NewRequest("POST", servicePrefix+ReportSuffix, bytes.NewReader(rqbytes))
				ctrl.PlantedError = errors.New("Report Returned Error")
				hndlr.ServeHTTP(w, req)
				g.Expect(w.Code).Should(g.Equal(http.StatusInternalServerError))
				g.Expect(w.Body.String()).Should(g.Equal(ctrl.PlantedError.Error()))
			})
		})

		gn.Context("when: called with :check request and controller.Check returns error", func() {
			gn.It("then: returns StatusInternalServerError ", func() {
				req := httptest.NewRequest("POST", servicePrefix+CheckSuffix, strings.NewReader("BAD_DECODE"))
				hndlr.ServeHTTP(w, req)
				g.Expect(w.Code).Should(g.Equal(http.StatusInternalServerError))
				g.Expect(w.Body.String()).Should(g.Equal("unexpected EOF"))
			})
		})
		gn.Context("when: called with :Report request and controller.Report returns error", func() {
			gn.It("then: returns StatusInternalServerError ", func() {
				req := httptest.NewRequest("POST", servicePrefix+ReportSuffix, strings.NewReader("BAD_DECODE"))
				hndlr.ServeHTTP(w, req)
				g.Expect(w.Code).Should(g.Equal(http.StatusInternalServerError))
				g.Expect(w.Body.String()).Should(g.Equal("unexpected EOF"))
			})
		})
		gn.Context("when: called with :Report request and Unable to read data", func() {
			gn.It("then: returns StatusInternalServerError ", func() {
				injectedErr := errors.New("Injected: Unable to Read data")
				hndlr = NewHandler(ctrl, phi, ReadHTTPBody(func(r io.Reader) (msg []byte, err error) {
					return nil, injectedErr
				}))
				req := httptest.NewRequest("POST", servicePrefix+ReportSuffix, strings.NewReader("BAD_DECODE"))
				hndlr.ServeHTTP(w, req)
				g.Expect(w.Code).Should(g.Equal(http.StatusInternalServerError))
				g.Expect(w.Body.String()).Should(g.Equal(injectedErr.Error()))
			})
		})
		gn.Context("when: called with :Report request and Unable to Marshal return data", func() {
			gn.It("then: returns StatusInternalServerError ", func() {
				injectedErr := errors.New("Injected: Unable to Marshal data")
				hndlr = NewHandler(ctrl, phi, Marshal(func(pb proto.Message) (buf []byte, err error) {
					return nil, injectedErr
				}))
				rqpb := testutils.CreateReport(
					&testutils.ExpectedReport{
						ApiName:     serviceName,
						ApiMethod:   "getfiles",
						OperationId: operationId,
					})
				rqbytes, err := proto.Marshal(&rqpb)
				g.Expect(err).Should(g.BeNil())
				req := httptest.NewRequest("POST", servicePrefix+ReportSuffix, bytes.NewReader(rqbytes))

				hndlr.ServeHTTP(w, req)
				g.Expect(w.Code).Should(g.Equal(http.StatusInternalServerError))
				g.Expect(w.Body.String()).Should(g.Equal(injectedErr.Error()))
			})
		})
		gn.Context("when: called with uri matching a handler", func() {
			gn.It("then: Should call consumer handler with exact prefix uri", func() {
				req := httptest.NewRequest("GET", prefix, nil)
				hndlr.ServeHTTP(w, req)
				g.Expect(w.Code).Should(g.Equal(http.StatusOK))
				g.Expect(w.Body.String()).Should(g.Equal(prefix))
			})
			gn.It("then: Should call consumer handler with uri containing prefix", func() {
				req := httptest.NewRequest("GET", prefix+"/suffix", nil)
				hndlr.ServeHTTP(w, req)
				g.Expect(w.Code).Should(g.Equal(http.StatusOK))
				g.Expect(w.Body.String()).Should(g.Equal(prefix))
			})

		})
		gn.Context("Error cases", func() {
			gn.Context("when: called with NON-POST Request", func() {
				gn.It("then: returns mehod not allowed", func() {
					req := httptest.NewRequest("GET", "/NonExistent/"+prefix, nil)
					hndlr.ServeHTTP(w, req)
					g.Expect(w.Code).Should(g.Equal(http.StatusMethodNotAllowed))
				})
			})
			gn.Context("when: called with POST without :check or :report", func() {
				gn.It("then: returns 404", func() {
					req := httptest.NewRequest("POST", "/NonExistent/"+prefix, nil)
					hndlr.ServeHTTP(w, req)
					g.Expect(w.Code).Should(g.Equal(http.StatusNotFound))
				})
			})
		})
	})
})
