package api

import (
	"encoding/json"
	"net/http"
	"strconv"
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

var (
	velibAPIUserStatsPath = "/getAllInfosUser"
	velibAPIRidesListPath = "/getCourseList"
)

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

func (c *APIVelibClient) GetUserStats() (*VelibUserStats, error) {
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

type velibUserRidesResponse struct {
	Paging struct {
		TotalNumberOfRecords int `json:"totalNumberOfRecords"`
	} `json:"paging"`
	WalletOperations []struct {
		StartDate  int64 `json:"startDate"`
		EndDate    int64 `json:"endDate"`
		Parameter3 struct {
			Distance           string  `json:"DISTANCE"`
			AverageSpeed       float64 `json:"AVERAGE_SPEED"`
			SavedCarbonDioxide float64 `json:"SAVED_CARBON_DIOXIDE"`
		} `json:"parameter3"`
	} `json:"walletOperations"`
}

type VelibUserRide struct {
	StartDate          int64
	EndDate            int64
	Distance           float64
	AverageSpeed       float64
	SavedCarbonDioxide float64
}

func (c *APIVelibClient) GetUserRides() ([]*VelibUserRide, error) {
	var velibUserRides []*VelibUserRide

	offset, limit := 0, 10
	for {
		req, err := http.NewRequest(http.MethodGet, c.Endpoint+velibAPIRidesListPath, nil)
		if err != nil {
			return nil, err
		}

		req.AddCookie(&http.Cookie{
			Name:  "BEARER",
			Value: *c.Token,
		})

		q := req.URL.Query()
		q.Add("limit", "10")
		q.Add("offset", strconv.Itoa(offset))
		req.URL.RawQuery = q.Encode()

		resp, err := c.HTTPClient.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		var velibResp velibUserRidesResponse
		err = json.NewDecoder(resp.Body).Decode(&velibResp)
		if err != nil {
			return nil, err
		}

		for _, ride := range velibResp.WalletOperations {
			distance, err := strconv.ParseFloat(ride.Parameter3.Distance, 64)
			if err != nil {
				return nil, err
			}

			velibUserRide := &VelibUserRide{
				StartDate:          ride.StartDate,
				EndDate:            ride.EndDate,
				Distance:           distance,
				AverageSpeed:       ride.Parameter3.AverageSpeed,
				SavedCarbonDioxide: ride.Parameter3.SavedCarbonDioxide,
			}
			velibUserRides = append(velibUserRides, velibUserRide)
		}
		if (offset + limit) > velibResp.Paging.TotalNumberOfRecords {
			break
		}
		offset += limit
	}
	return velibUserRides, nil
}
