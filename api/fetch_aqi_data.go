package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"

	waqi "github.com/sadityakumar9211/clean-route-backend/models/waqi"
	"github.com/spf13/viper"
)

func FetchAQIData(location []float64, delayCode uint8) (float64, error) {
	baseUrl := "https://api.waqi.info/feed/geo:" + fmt.Sprintf("%f;%f/?", location[1], location[0])

	var waqiAccessToken string
	var waqiAccessTokenError bool
	if os.Getenv("RAILWAY") == "true" {
		waqiAccessToken = os.Getenv("WAQI_API_KEY")
	} else {
		waqiAccessToken, waqiAccessTokenError = viper.Get("WAQI_API_KEY").(string)
		if !waqiAccessTokenError {
			log.Fatalf("Invalid type assertion")
		}
	}

	params := url.Values{}
	params.Add("token", waqiAccessToken)

	waqiUrl := baseUrl + params.Encode()

	resp, err := http.Get(waqiUrl)
	checkErrNil(err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	checkErrNil(err)

	// First, check if the response status is ok before unmarshalling data field
	var statusCheck struct {
		Status string      `json:"status"`
		Data   interface{} `json:"data"` // Use interface{} to accept both object and string
	}

	err = json.Unmarshal([]byte(body), &statusCheck)
	if err != nil {
		log.Fatal("Error while unmarshling JSON: ", err)
	}

	if statusCheck.Status != "ok" {
		// Extract error message if data is a string
		errMsg := "WAQI response is not 'OK' but: " + statusCheck.Status
		if dataStr, ok := statusCheck.Data.(string); ok {
			errMsg = "WAQI API Error: " + dataStr
		}
		log.Printf("WAQI fetch failed for location [lat=%f lon=%f]: %s", location[1], location[0], errMsg)
		return 0, errors.New(errMsg)
	}

	// Now safely unmarshal into the full AQIData struct
	var waqiResponse waqi.APIResponse
	err = json.Unmarshal([]byte(body), &waqiResponse)
	if err != nil {
		log.Fatal("Error while unmarshling AQI data JSON: ", err)
	}

	// Currently not utilizing the delayCode - no forecasting till now.
	// Will have to to update from this part onwards to include:
	/*
		1. API call to fetch the weather data from darksky or similary apis
			- relative humidity
			- wind speed
			- wind direction
			- temperature
			- we need these current and forecasted both the values
	*/
	// 2. We need to make api call to AWS Sagemaker and get the forecasted aqi value.

	if waqiResponse.Status != "ok" {
		log.Printf("WAQI fetch failed for location [lat=%f lon=%f]: status=%s", location[1], location[0], waqiResponse.Status)
		return 0, errors.New("WAQI response is not 'OK' but: " + waqiResponse.Status)
	} else {
		return waqiResponse.Data.IAQI["pm25"].V, nil
	}
}

func checkErrNil(err error) {
	if err != nil {
		log.Fatal("Error: ", err)
	}
}
