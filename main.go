package main

import (
	"flag"
	"fmt"
	"net/http"
	"strconv"
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
	tripAverageDistance = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "trip_average_distance",
		Help:      "Velib trip average distance in meters",
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
	trip = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "trip",
		Help:      "Velib trip details",
	}, []string{"start_date", "end_date", "average_speed", "co2_saved_total"})
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
	prometheus.MustRegister(tripAverageDistance)
	prometheus.MustRegister(tripHighestDistance)
	prometheus.MustRegister(totalSavedCarbonDioxide)
	prometheus.MustRegister(fetchingErrors)
	prometheus.MustRegister(trip)

	c := api.NewAPIVelibClient(token)

	s := gocron.NewScheduler(time.UTC)
	s.Every(30).Minute().Do(updateUsersStats, c)
	s.Every(60).Minute().Do(updateUsersLastRides, c)
	s.StartAsync()

	http.Handle("/metrics", promhttp.Handler())
	listenAddr := fmt.Sprintf("%s:%s", *address, *port)
	logrus.Infof("Beginning to serve metrics on %s", listenAddr)
	err := http.ListenAndServe(listenAddr, nil)
	if err != nil {
		logrus.Fatal(err)
	}
}

func updateUsersStats(c *api.APIVelibClient) {
	stats, err := c.GetUserStats()
	if err != nil {
		fetchingErrors.Inc()
		logrus.WithError(err).Error("Error while scrapping Velib users statistics")
		return
	}

	logrus.WithFields(logrus.Fields{
		"distance_total":      stats.GeneralDetails.CustomerIndicators.DistanceGlobalCounter,
		"distance_electrical": stats.GeneralDetails.CustomerIndicators.DistanceElectricalCounter,
		"distance_mechanical": stats.GeneralDetails.CustomerIndicators.DistanceGlobalCounter -
			stats.GeneralDetails.CustomerIndicators.DistanceElectricalCounter,
		"trip_number":           stats.GeneralDetails.CustomerIndicators.TripCounter,
		"trip_average_duration": stats.GeneralDetails.CustomerIndicators.TripAverageDuration,
		"trip_average_distance": stats.GeneralDetails.CustomerIndicators.DistanceGlobalCounter / stats.GeneralDetails.CustomerIndicators.TripCounter,
		"trip_highest_distance": stats.GeneralDetails.CustomerIndicators.TripHighestDistance,
		"co2_saved_total":       stats.GeneralDetails.CustomerIndicators.GlobalSavedCarbonDioxide,
	}).Info("Updating velib-exporter user gauge")

	distanceTotal.Set(float64(stats.GeneralDetails.CustomerIndicators.DistanceGlobalCounter))
	distanceElectrical.Set(float64(stats.GeneralDetails.CustomerIndicators.DistanceElectricalCounter))
	distanceMechanical.Set(float64(stats.GeneralDetails.CustomerIndicators.DistanceGlobalCounter -
		stats.GeneralDetails.CustomerIndicators.DistanceElectricalCounter))
	tripNumber.Set(float64(stats.GeneralDetails.CustomerIndicators.TripCounter))
	tripAverageDuration.Set(float64(stats.GeneralDetails.CustomerIndicators.TripAverageDuration))
	tripAverageDistance.Set(float64(stats.GeneralDetails.CustomerIndicators.DistanceGlobalCounter / stats.GeneralDetails.CustomerIndicators.TripCounter))
	tripHighestDistance.Set(float64(stats.GeneralDetails.CustomerIndicators.TripHighestDistance))
	totalSavedCarbonDioxide.Set(stats.GeneralDetails.CustomerIndicators.GlobalSavedCarbonDioxide)
}

func updateUsersLastRides(c *api.APIVelibClient) {
	rides, err := c.GetUserRides(&api.VelibUserRideRequest{
		Limit:  20,
		Offset: 0,
	})

	if err != nil {
		fetchingErrors.Inc()
		logrus.WithError(err).Error("Error while scrapping Velib users statistics")
		return
	}

	for _, ride := range rides.WalletOperations {

		logrus.WithFields(logrus.Fields{
			"start_date":      ride.StartDate,
			"end_date":        ride.EndDate,
			"average_speed":   ride.Parameter3.AverageSpeed,
			"distance":        ride.Parameter3.Distance,
			"co2_saved_total": ride.Parameter3.SavedCarbonDioxide,
		}).Info("Updating velib-exporter ride gauge")

		distance, err := strconv.ParseFloat(ride.Parameter3.Distance, 64)
		if err != nil {
			logrus.Error("Invalid distance, skipping row")
			continue
		}
		trip.With(prometheus.Labels{
			"start_date":      fmt.Sprintf("%d", ride.StartDate),
			"end_date":        fmt.Sprintf("%d", ride.EndDate),
			"average_speed":   fmt.Sprintf("%f", ride.Parameter3.AverageSpeed),
			"co2_saved_total": fmt.Sprintf("%f", ride.Parameter3.SavedCarbonDioxide),
		}).Set(distance)
	}
}
