package main

import (
	"flag"
	"github.com/golang/glog"
	"net/http"
	"somnacin-internal/mixologist/mixologist"
	"somnacin-internal/mixologist/mixologist/rc/statsd"
	"strconv"
	"strings"

	// Needed for init(), as prometheus has no configuration at the moment
	_ "somnacin-internal/mixologist/mixologist/rc/prometheus"
)

var (
	config mixologist.Config

	// Mixologist commandline flags
	port       = flag.Int("port", mixologist.Port, "Port exposed for ServiceControl RPCs")
	nConsumers = flag.Int("nConsumers", mixologist.NConsumers, "Number of consumers for request processing")

	// Metrics backend flags
	metricsBackends = flag.String("metrics_backends", "prometheus,statsd", "Comma-separated list of canonical names for metrics export backends")
)

func init() {
	// Statsd configuration flags
	flag.StringVar(&statsd.Config.Addr, "statsd_addr", "statsd:8125", "Address (host:port) for a statsd backend; used only when statsd is being used for metrics export")
}

func main() {
	flag.Parse()
	config.Metrics.Backends = strings.Split(*metricsBackends, ",")

	controller := mixologist.NewControllerImpl()
	rcMgr := mixologist.NewReportConsumerManager(controller.ReportQueue(), mixologist.ReportConsumerRegistry, config)
	handler := mixologist.NewHandler(controller, rcMgr.GetPrefixAndHandlers())
	addr := ":" + strconv.Itoa(*port)
	srv := http.Server{
		Addr:    addr,
		Handler: handler,
	}
	rcMgr.Start(*nConsumers)
	glog.Info("Starting Server on " + addr)
	err := srv.ListenAndServe()
	if err != nil {
		glog.Exitf("Unable to start server " + err.Error())
	}
}
