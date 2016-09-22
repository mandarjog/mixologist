package main

import (
	"flag"
	"github.com/golang/glog"
	"net/http"
	"somnacin-internal/mixologist/mixologist"
	"somnacin-internal/mixologist/mixologist/rc/statsd"
	"strconv"
	"strings"

	// Needed for init()
	_ "somnacin-internal/mixologist/mixologist/rc/logsAdapter"
	_ "somnacin-internal/mixologist/mixologist/rc/prometheus"
)

var (
	config mixologist.Config

	// Mixologist commandline flags
	port       = flag.Int("port", mixologist.Port, "Port exposed for ServiceControl RPCs")
	nConsumers = flag.Int("nConsumers", mixologist.NConsumers, "Number of consumers for request processing")

	// Metrics backend flags
	reportConsumers = flag.String("report_consumers", "prometheus,statsd,mixologist.io/consumers/logsAdapter", "Comma-separated list of canonical names for report consumers")
	loggingBackends = flag.String("logging_backends", "", "Comma-separated list of canonical names for logging export backends. If left empty, the default logging backend will be used (if enabled).")
)

func init() {
	// Statsd configuration flags
	flag.StringVar(&statsd.Config.Addr, "statsd_addr", "statsd:8125", "Address (host:port) for a statsd backend; used only when statsd is being used for metrics export")

	// Logging configuration flags
	flag.BoolVar(&config.Logging.UseDefault, "use_default_logger", true, "Toggles default logging (std{out|err})")
}

func main() {
	flag.Parse()
	config.ReportConsumers = strings.Split(*reportConsumers, ",")
	config.Logging.Backends = strings.Split(*loggingBackends, ",")

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
