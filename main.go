package main

import (
	"flag"
	"github.com/golang/glog"
	"net/http"
	"somnacin-internal/mixologist/mixologist"
	"somnacin-internal/mixologist/mixologist/rc/prometheus"
	"strconv"
)

func main() {
	var port int
	var nConsumers int
	flag.IntVar(&port, "port", mixologist.Port, "port")
	flag.IntVar(&nConsumers, "nConsumers", mixologist.NConsumers, "nConsumers")
	controller := mixologist.NewControllerImpl()
	flag.Parse()
	rcMgr := mixologist.NewReportConsumerManager(controller.ReportQueue, mixologist.ReportConsumerRegistry, []string{prometheus.Name})
	handler := mixologist.NewHandler(controller, rcMgr.GetPrefixAndHandlers())
	addr := ":" + strconv.Itoa(port)
	srv := http.Server{
		Addr:    addr,
		Handler: handler,
	}
	rcMgr.Start(nConsumers)
	glog.Info("Starting Server on " + addr)
	err := srv.ListenAndServe()
	if err != nil {
		glog.Exitf("Unable to start server " + err.Error())
	}
}
