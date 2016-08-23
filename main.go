package main

import (
	"flag"
	"github.com/golang/glog"
	"net/http"
	"somnacin-internal/mixologist/mixologist"
	"strconv"
)

func main() {
	portPtr := flag.Int("port", mixologist.Port, "port")
	controller := mixologist.NewControllerImpl()
	handler := mixologist.NewHandler(controller)

	reportConsumer := mixologist.NewPrometheusReporter()
	reportConsumer.SetReportQueue(controller.ReportQueue)
	go reportConsumer.Start()

	flag.Parse()
	addr := ":" + strconv.Itoa(*portPtr)
	srv := http.Server{
		Addr:    addr,
		Handler: handler,
	}
	glog.Info("Starting Server on " + addr)
	err := srv.ListenAndServe()
	if err != nil {
		glog.Exitf("Unable to start server " + err.Error())
	}
}
