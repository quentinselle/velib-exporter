package main

import (
	"flag"
	"fmt"
	"net/http"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/qselle/velib-exporter/pkg/api"
	"github.com/sirupsen/logrus"
)

var (
	namespace     = "velib"
	distanceTotal = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "distance_total",
		Help:      "Distance total in Velib in meters",
	})
	distanceElectrical = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "distance_electrical",
		Help:      "Distance total in electrical Velib in meters",
	})
	distanceMechanical = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "distance_mechanical",
		Help:      "Distance total in mechanical Velib in meters",
	})
	tripNumber = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "trip_number",
		Help:      "Number of Velib trips",
	})
	tripAverageDuration = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "trip_average_duration",
		Help:      "Velib trip average duration in minutes",
	})
	tripHighestDistance = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "trip_highest_distance",
		Help:      "Velib trip highest distance in meters",
	})
	totalSavedCarbonDioxide = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "co2_total_saved",
		Help:      "Total CO2 saved by using Velib in grams",
	})
	fetchingErrors = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "fetching_errors",
		Help:      "Counter of errors while scrapping velib-metropole.fr",
	})
)

func main() {
	token := flag.String("token", "", "Velib API token")
	address := flag.String("address", "127.0.0.1", "Exporter listening address")
	port := flag.String("port", "5050", "Exporter listening port")
	debug := flag.Bool("debug", false, "Debug mode")
	flag.Parse()

	if *token == "" {
		logrus.Fatalf("missing -token flag")
	}

	if *debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	prometheus.MustRegister(distanceTotal)
	prometheus.MustRegister(distanceElectrical)
	prometheus.MustRegister(distanceMechanical)
	prometheus.MustRegister(tripNumber)
	prometheus.MustRegister(tripAverageDuration)
	prometheus.MustRegister(tripHighestDistance)
	prometheus.MustRegister(totalSavedCarbonDioxide)

	c := api.NewAPIVelibClient(token)

	s := gocron.NewScheduler(time.UTC)
	s.Every(1).Hour().Do(updateProm, c)
	s.StartAsync()

	http.Handle("/metrics", promhttp.Handler())
	listenAddr := fmt.Sprintf("%s:%s", *address, *port)
	logrus.Infof("Beginning to serve metrics on %s", listenAddr)
	err := http.ListenAndServe(listenAddr, nil)
	if err != nil {
		logrus.Fatal(err)
	}
}

func updateProm(c *api.APIVelibClient) {
	stats, err := c.GetUsersStats()
	if err != nil {
		fetchingErrors.Inc()
		logrus.WithError(err).Error("Error while scrapping Velib users statistics")
		return
	}

	logrus.WithFields(logrus.Fields{
		"distance_total":        stats.DistanceTotal,
		"distance_electrical":   stats.DistanceElectrical,
		"distance_mechanical":   stats.DistanceMechanical,
		"trip_number":           stats.TripNumber,
		"trip_average_duration": stats.TripAverageDuration,
		"trip_highest_distance": stats.TripHighestDistance,
		"co2_saved_total":       stats.TotalSavedCarbonDioxide,
	}).Info("Updating velib-exporter gauge")

	distanceTotal.Set(float64(stats.DistanceTotal))
	distanceElectrical.Set(float64(stats.DistanceElectrical))
	distanceMechanical.Set(float64(stats.DistanceMechanical))
	tripNumber.Set(float64(stats.TripNumber))
	tripAverageDuration.Set(float64(stats.TripAverageDuration))
	tripHighestDistance.Set(float64(stats.TripHighestDistance))
	totalSavedCarbonDioxide.Set(float64(stats.TotalSavedCarbonDioxide))
}
