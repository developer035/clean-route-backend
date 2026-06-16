package api

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/sadityakumar9211/clean-route-backend/models"
	"github.com/spf13/viper"
)

type Post struct {
	FPMVec []float64 `json:"fpm_vec"`
}

func GetPredictedPm25(df []models.FeatureVector) ([]float64, error) {

	var awsModelEndpoint string
	var awsModelEndpointError bool

	if os.Getenv("RAILWAY") == "true" {
		awsModelEndpoint = os.Getenv("AWS_MODEL_ENDPOINT")
	} else {
		awsModelEndpoint, awsModelEndpointError = viper.Get("AWS_MODEL_ENDPOINT").(string)
		if !awsModelEndpointError {
			log.Fatalf("Invalid type assertion")
		}
	}

	// fmt.Println("The Query url is: ", awsModelEndpoint)

	jsonData, err := json.Marshal(df)
	checkErrNil(err)

	r, err := http.NewRequest("POST", awsModelEndpoint, bytes.NewBuffer(jsonData))
	checkErrNil(err)

	r.Header.Add("Content-Type", "application/json")

	// fmt.Println("Just before making the request...")
	client := &http.Client{}
	resp, err := client.Do(r)
	checkErrNil(err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	checkErrNil(err)

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("ML endpoint error: status=%d body=%s", resp.StatusCode, string(body))
	}

	// fmt.Println("After the Request...")
	post := &Post{}

	err = json.Unmarshal(body, post)
	checkErrNil(err)

	// fmt.Println("++++++++++++++ Inside the get predicted pm2.5 ++++++++++++++++")
	// fmt.Println("Acutal: ", inputFeatures.IPM, "Predicted: ", post.FPM)
	return post.FPMVec, nil
}
