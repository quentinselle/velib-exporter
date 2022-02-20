package main

import (
	"flag"
	"fmt"
	"net/http"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

func main() {
	token := flag.String("token", "", "Velib API token")
	address := flag.String("address", "127.0.0.1", "Exporter listening address")
	port := flag.String("port", "5050", "Exporter listening port")
	debug := flag.Bool("debug", false, "Debug mode")
	flag.Parse()

	if *token == "" {
		logrus.Fatalf("Missing velib-metropole.fr API token\n\tPlease use -token")
	}

	if *debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	v := NewVelibExporterService(token)

	s := gocron.NewScheduler(time.UTC)
	s.Every(1).Hour().Do(v.updateProm)
	s.StartAsync()

	http.Handle("/metrics", promhttp.Handler())
	listenAddr := fmt.Sprintf("%s:%s", *address, *port)
	logrus.Infof("Beginning to serve metrics on %s", listenAddr)
	err := http.ListenAndServe(listenAddr, nil)
	if err != nil {
		logrus.Fatal(err)
	}
}
