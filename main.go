package main

import (
	"flag"
	"github.com/golang/glog"
	"net/http"
	"somnacin-internal/mixologist/mixologist"
)

func main() {
	flag.Parse()
	handler := mixologist.NewHandler(&mixologist.ControllerImpl{})
	srv := http.Server{
		Addr:    mixologist.Port,
		Handler: handler,
	}
	err := srv.ListenAndServe()
	if err != nil {
		glog.Exitf("Unable to start server " + err.Error())
	}
}
