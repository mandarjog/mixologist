package main

import (
	"flag"
	"github.com/golang/glog"
	"net/http"
	"somnacin-internal/mixologist/mixologist"
	"strconv"
)

func main() {
	var port int
	flag.IntVar(&port, "port", mixologist.Port, "port")
	controller := mixologist.NewControllerImpl()
	handler := mixologist.NewHandler(controller)

	reportConsumer := mixologist.NewPrometheusReporter()
	reportConsumer.SetReportQueue(controller.ReportQueue)
	flag.Parse()

	handler.AddHandler(reportConsumer.GetPrefixAndHandler())
	addr := ":" + strconv.Itoa(port)
	srv := http.Server{
		Addr:    addr,
		Handler: handler,
	}
	glog.Info("Starting Server on " + addr)
	go reportConsumer.Start()
	err := srv.ListenAndServe()
	if err != nil {
		glog.Exitf("Unable to start server " + err.Error())
	}
}
