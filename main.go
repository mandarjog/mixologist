package main

import (
	"flag"
	"net/http"
	"strconv"
	"strings"

	"github.com/cloudendpoints/mixologist/mixologist"
	"github.com/cloudendpoints/mixologist/mixologist/rc/statsd"
	"github.com/golang/glog"

	// Needed for init()
	_ "github.com/cloudendpoints/mixologist/mixologist/cp/block"
	_ "github.com/cloudendpoints/mixologist/mixologist/cp/whitelist"
	_ "github.com/cloudendpoints/mixologist/mixologist/rc/logsAdapter"
	_ "github.com/cloudendpoints/mixologist/mixologist/rc/prometheus"
)

var (
	config mixologist.Config

	// Mixologist commandline flags
	port       = flag.Int("port", mixologist.Port, "Port exposed for ServiceControl RPCs")
	nConsumers = flag.Int("nConsumers", mixologist.NConsumers, "Number of consumers for request processing")

	// Metrics backend flags
	reportConsumers = flag.String("report_consumers", "prometheus,statsd,mixologist.io/consumers/logsAdapter", "Comma-separated list of canonical names for report consumers")
	checkers        = flag.String("checkers", "whitelist,acl", "Comma-separated list of canonical names for report consumers")
	loggingBackends = flag.String("logging_backends", "", "Comma-separated list of canonical names for logging export backends. If left empty, the default logging backend will be used (if enabled).")
	configFile      = flag.String("config_file", "mixCfg.yml", "Yml config file")
	kubeconfig      = flag.String("kubeconfig", "", "Path to kubeconfig")
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
	config.Checkers = strings.Split(*checkers, ",")
	config.Logging.Backends = strings.Split(*loggingBackends, ",")
	osc := mixologist.ServicesConfig{}
	var err error
	var configMgr *mixologist.ConfigManager
	checkerMgr, _ := mixologist.NewCheckerManager(mixologist.CheckerRegistry, &osc)
	if configMgr, err = mixologist.NewConfigManager(*configFile, *kubeconfig); err != nil {
		glog.Exitf("Unable to start server " + err.Error())
	}
	configMgr.Register(checkerMgr)
	go configMgr.Loop()

	controller := mixologist.NewControllerImpl(checkerMgr)
	rcMgr := mixologist.NewReportConsumerManager(controller.ReportQueue(), mixologist.ReportConsumerRegistry, config)
	handler := mixologist.NewHandler(controller, rcMgr.GetPrefixAndHandlers())
	addr := ":" + strconv.Itoa(*port)
	srv := http.Server{
		Addr:    addr,
		Handler: handler,
	}
	rcMgr.Start(*nConsumers)
	glog.Info("Starting Server on " + addr)
	err = srv.ListenAndServe()
	if err != nil {
		glog.Exitf("Unable to start server " + err.Error())
	}
}
