package core

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"concurrentrix/internal"
)

var (
	smsSendUrl = "https://api.grab.com/grabid/v1/phone/otp"
	getXRayUrl = "http://154.221.17.198:5612/business/invoke?group=huoyu_test&action=nb&arg1="
	tunnel     = "i635.kdltps.com:15818"
	username   = "t10407739755674"
	password   = "op1un18b"
)

func generateRandomDriverID() string {
	// Setting the random number seed
	rand.Seed(time.Now().UnixNano())

	// Generate random numbers between 1000 and 9999
	randomNumber := rand.Intn(9000) + 1000

	// Generate driver_id with a fixed suffix
	driverID := fmt.Sprintf("%d%s", randomNumber, "c6509c1db2d1")

	return driverID
}

func GetXRay(enco string) (string, error) {
	resp, err := http.Get(getXRayUrl + enco)
	if err != nil {
		err = fmt.Errorf("http get err: %v", err)
		return "", err
	}
	defer resp.Body.Close()

	// Read the response body
	var result *internal.XRayStr
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&result)
	if err != nil {
		err = fmt.Errorf("decode err: %v", err)
		return "", err
	}

	return result.Data, nil
}
