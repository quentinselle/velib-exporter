package api

import (
	"encoding/json"
	"net/http"
	"time"
)

type APIVelibClient struct {
	Token      *string
	HTTPClient *http.Client
	Endpoint   string
}

func NewAPIVelibClient(token *string) *APIVelibClient {
	return &APIVelibClient{
		Token: token,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		Endpoint: "https://www.velib-metropole.fr/webapi/private",
	}
}

var velibAPIUserStatsPath = "/getAllInfosUser"

type velibUserStatsResponse struct {
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

type VelibUserStats struct {
	DistanceTotal           int
	DistanceElectrical      int
	DistanceMechanical      int
	TripNumber              int
	TripAverageDuration     int
	TripHighestDistance     int
	TotalSavedCarbonDioxide float64
}

func (c *APIVelibClient) GetUsersStats() (*VelibUserStats, error) {
	req, err := http.NewRequest(http.MethodGet, c.Endpoint+velibAPIUserStatsPath, nil)
	if err != nil {
		return nil, err
	}

	req.AddCookie(&http.Cookie{
		Name:  "BEARER",
		Value: *c.Token,
	})

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var velibResp velibUserStatsResponse
	err = json.NewDecoder(resp.Body).Decode(&velibResp)
	if err != nil {
		return nil, err
	}

	return &VelibUserStats{
		DistanceTotal:      velibResp.GeneralDetails.CustomerIndicators.DistanceGlobalCounter,
		DistanceElectrical: velibResp.GeneralDetails.CustomerIndicators.DistanceElectricalCounter,
		DistanceMechanical: velibResp.GeneralDetails.CustomerIndicators.DistanceGlobalCounter -
			velibResp.GeneralDetails.CustomerIndicators.DistanceElectricalCounter,
		TripNumber:              velibResp.GeneralDetails.CustomerIndicators.TripCounter,
		TripAverageDuration:     velibResp.GeneralDetails.CustomerIndicators.TripAverageDuration,
		TripHighestDistance:     velibResp.GeneralDetails.CustomerIndicators.TripHighestDistance,
		TotalSavedCarbonDioxide: velibResp.GeneralDetails.CustomerIndicators.GlobalSavedCarbonDioxide,
	}, nil
}
