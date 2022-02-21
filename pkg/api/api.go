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

type paginationParams struct {
	offset int
	limit  int
}

type requestParams struct {
	path       string
	method     string
	pagination paginationParams
}

func (c *APIVelibClient) doRequest(params *requestParams) (*http.Response, error) {
	req, err := http.NewRequest(params.method, c.Endpoint+params.path, nil)
	if err != nil {
		return nil, err
	}

	req.AddCookie(&http.Cookie{
		Name:  "BEARER",
		Value: *c.Token,
	})

	q := req.URL.Query()
	if params.pagination.limit != 0 {
		q.Add("limit", strconv.Itoa(params.pagination.limit))
	}
	if params.pagination.offset != 0 {
		q.Add("offset", strconv.Itoa(params.pagination.offset))
	}
	req.URL.RawQuery = q.Encode()

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

var velibAPIUserStatsPath = "/getAllInfosUser"

type VelibUserStatsResponse struct {
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

func (c *APIVelibClient) GetUserStats() (*VelibUserStatsResponse, error) {
	httpResp, err := c.doRequest(&requestParams{
		path:   velibAPIUserStatsPath,
		method: http.MethodGet,
	})
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	var velibResp VelibUserStatsResponse
	err = json.NewDecoder(httpResp.Body).Decode(&velibResp)
	if err != nil {
		return nil, err
	}

	return &velibResp, nil
}

var velibAPIRidesListPath = "/getCourseList"

type VelibUserRideRequest struct {
	Limit  int
	Offset int
}
type VelibUserRidesResponse struct {
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

func (c *APIVelibClient) GetUserRides(req *VelibUserRideRequest) (*VelibUserRidesResponse, error) {

	httpResp, err := c.doRequest(&requestParams{
		path:   velibAPIRidesListPath,
		method: http.MethodGet,
		pagination: paginationParams{
			limit:  req.Limit,
			offset: req.Offset,
		},
	})
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	var velibResp VelibUserRidesResponse
	err = json.NewDecoder(httpResp.Body).Decode(&velibResp)
	if err != nil {
		return nil, err
	}

	return &velibResp, nil
}
