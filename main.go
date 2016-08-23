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

	flag.Parse()
	handler := mixologist.NewHandler(&mixologist.ControllerImpl{})
	addr := ":" + strconv.Itoa(*portPtr)
	glog.Info("Starting Server on " + addr)
	srv := http.Server{
		Addr:    addr,
		Handler: handler,
	}
	err := srv.ListenAndServe()
	if err != nil {
		glog.Exitf("Unable to start server " + err.Error())
	}
}
