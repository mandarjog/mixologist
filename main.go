package main

import (
	"flag"
	"github.com/cloudendpoints/mixologist/mixologist"
	cfg "github.com/cloudendpoints/mixologist/mixologist/config"
	"github.com/cloudendpoints/mixologist/mixologist/rc/statsd"
	"github.com/golang/glog"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	// Needed for init()
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
)

func init() {
	// Statsd configuration flags
	flag.StringVar(&statsd.Config.Addr, "statsd_addr", "statsd:8125", "Address (host:port) for a statsd backend; used only when statsd is being used for metrics export")

	// Logging configuration flags
	flag.BoolVar(&config.Logging.UseDefault, "use_default_logger", true, "Toggles default logging (std{out|err})")

	flag.StringVar(&config.WhiteListBackEnd, "whitelist_url", "https://gist.githubusercontent.com/mandarjog/c38f4a992cc5d470ad763e70eca709b9/raw/", "json/yml file with whitelist")
}

func main() {
	flag.Parse()
	config.ReportConsumers = strings.Split(*reportConsumers, ",")
	config.Checkers = strings.Split(*checkers, ",")
	config.Logging.Backends = strings.Split(*loggingBackends, ",")
	osc := cfg.ServicesConfig{}

	if data, err := ioutil.ReadFile(*configFile); err != nil {
		glog.Errorf("Unable to read %s:  %s", *configFile, err)
	} else {
		yaml.Unmarshal(data, &osc)
	}

	checkerMgr, _ := mixologist.NewCheckerManager(mixologist.CheckerRegistry, osc)
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
	err := srv.ListenAndServe()
	if err != nil {
		glog.Exitf("Unable to start server " + err.Error())
	}
}
