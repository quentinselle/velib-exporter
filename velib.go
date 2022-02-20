package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

var (
	VelibAPIMetricsPath = "https://www.velib-metropole.fr/webapi/private/getAllInfosUser"

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

type VelibExporterService struct {
	Token      *string
	HTTPClient *http.Client
}

func NewVelibExporterService(token *string) VelibExporterService {
	prometheus.MustRegister(distanceTotal)
	prometheus.MustRegister(distanceElectrical)
	prometheus.MustRegister(distanceMechanical)
	prometheus.MustRegister(tripNumber)
	prometheus.MustRegister(tripAverageDuration)
	prometheus.MustRegister(tripHighestDistance)
	prometheus.MustRegister(totalSavedCarbonDioxide)

	return VelibExporterService{
		Token: token,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

type VelibResponse struct {
	GeneralDetails struct {
		CustomerIndicators struct {
			DistanceGlobalCounter     int     `json:"distanceGlobalCounter"`
			DistanceElectricalCounter int     `json:"distanceElectricalCounter"`
			TripCounter               int     `json:"tripCounter"`
			TripAverageDuration       int     `json:"tripAverageDuration"`
			TripHighestDistance       int     `json:"tripHighestDistance"`
			GlobalSavedCarbonDioxide  float64 `json:"globalSavedCarbonDioxide"`
		} `json:"customerIndicators"`
	} `json:"generalDetails"`
}

type velibStats struct {
	distanceTotal           int
	distanceElectrical      int
	distanceMechanical      int
	tripNumber              int
	tripAverageDuration     int
	tripHighestDistance     int
	totalSavedCarbonDioxide float64
}

func (v VelibExporterService) fetchVelibStats() (*velibStats, error) {
	req, err := http.NewRequest(http.MethodGet, VelibAPIMetricsPath, nil)
	if err != nil {
		return nil, err
	}

	req.AddCookie(&http.Cookie{
		Name:  "BEARER",
		Value: *v.Token,
	})

	logrus.WithFields(logrus.Fields{
		"Method": http.MethodGet,
		"URL":    VelibAPIMetricsPath,
	}).Debug("Requesting")
	resp, err := v.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	logrus.Debugf("Response status code %v", resp.StatusCode)

	var velibResp VelibResponse
	err = json.NewDecoder(resp.Body).Decode(&velibResp)
	if err != nil {
		return nil, err
	}
	logrus.Debugf("Response body %+v", velibResp)

	return &velibStats{
		distanceTotal:      velibResp.GeneralDetails.CustomerIndicators.DistanceGlobalCounter,
		distanceElectrical: velibResp.GeneralDetails.CustomerIndicators.DistanceElectricalCounter,
		distanceMechanical: velibResp.GeneralDetails.CustomerIndicators.DistanceGlobalCounter -
			velibResp.GeneralDetails.CustomerIndicators.DistanceElectricalCounter,
		tripNumber:              velibResp.GeneralDetails.CustomerIndicators.TripCounter,
		tripAverageDuration:     velibResp.GeneralDetails.CustomerIndicators.TripAverageDuration,
		tripHighestDistance:     velibResp.GeneralDetails.CustomerIndicators.TripHighestDistance,
		totalSavedCarbonDioxide: velibResp.GeneralDetails.CustomerIndicators.GlobalSavedCarbonDioxide,
	}, nil
}

func (v VelibExporterService) updateProm() {
	stats, err := v.fetchVelibStats()
	if err != nil {
		fetchingErrors.Inc()
		logrus.WithError(err).Error("Error while scrapping Velib statistics")
		return
	}

	logrus.WithFields(logrus.Fields{
		"distance_total":        stats.distanceTotal,
		"distance_electrical":   stats.distanceElectrical,
		"distance_mechanical":   stats.distanceMechanical,
		"trip_number":           stats.tripNumber,
		"trip_average_duration": stats.tripAverageDuration,
		"trip_highest_distance": stats.tripHighestDistance,
		"co2_saved_total":       stats.totalSavedCarbonDioxide,
	}).Info("Updating velib-exporter gauge")

	distanceTotal.Set(float64(stats.distanceTotal))
	distanceElectrical.Set(float64(stats.distanceElectrical))
	distanceMechanical.Set(float64(stats.distanceMechanical))
	tripNumber.Set(float64(stats.tripNumber))
	tripAverageDuration.Set(float64(stats.tripAverageDuration))
	tripHighestDistance.Set(float64(stats.tripHighestDistance))
	totalSavedCarbonDioxide.Set(float64(stats.totalSavedCarbonDioxide))
}
