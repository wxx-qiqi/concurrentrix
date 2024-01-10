package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sync"
	"time"

	"concurrentrix/core/tools"
)

var (
	MutexSendJobs sync.Mutex
	MutexAgainNum sync.Mutex
	SendJobs      = 0
	AgainNum      = 0
)

func smsSend(phone, xRay string, saveJobs *SaveJobsStr) error {
	// Constructing the request header
	headers := map[string]string{
		"Connection":           "close",
		"X-Ray":                xRay,
		"X-Grab-Device-ID":     generateRandomDriverID(),
		"alias_sessionid":      "2312271214-0yu1VjiyG9gadd" + xRay,
		"x-request-id":         fmt.Sprintf("%s", time.Now().UnixNano()),
		"X-Grab-Device-Model":  "Pixel 4 XL",
		"Authorization":        "grabsecure " + tools.GetPhoneSha256s(phone),
		"Accept-Language":      "zh-CN;q=1.0, en-US;q=0.9, en;q=0.8",
		"X-Translate-Language": "",
		"User-Agent":           "Grab/5.285.0 (Android 10; Build 62826273)",
		"Content-Type":         "application/json; charset=UTF-8",
		"Host":                 "api.grab.com",
		"accept-encoding":      "gzip",
	}

	// Constructing the request body
	payload := map[string]interface{}{
		"method":                  "SMS",
		"countryCode":             "cn",
		"phoneNumber":             phone,
		"templateID":              "pax_android_production",
		"numDigits":               6,
		"deviceID":                generateRandomDriverID(),
		"deviceManufacturer":      "Google",
		"deviceModel":             "Pixel 4 XL",
		"cellularOperator":        "",
		"locale":                  "en-KW",
		"retryCount":              0,
		"eligibleForAccountHints": "false",
		"adrID":                   nil,
		"adrIMEI":                 nil,
		"adrIMSI":                 nil,
		"adrMEID":                 nil,
		"adrSERIAL":               nil,
		"adrUDID":                 nil,
		"advertisingID":           nil,
		"scenario":                "signup",
	}

	// Convert request body to JSON
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		err = fmt.Errorf("JSON encoding error:%v", err)
		return err
	}

	// Send a POST request
	req, err := http.NewRequest("POST", smsSendUrl, bytes.NewBuffer(jsonPayload))
	if err != nil {
		err = fmt.Errorf("create request error:%v", err)
		return err
	}

	// Setting the request header
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// Setting up the proxy
	proxyURL := fmt.Sprintf("http://%s:%s@%s/", username, password, tunnel)
	parsedProxyURL, err := url.Parse(proxyURL)
	if err != nil {
		err = fmt.Errorf("parsing proxy URL error: %v", err)
		return err
	}
	transport := &http.Transport{Proxy: http.ProxyURL(parsedProxyURL)}
	client := &http.Client{Transport: transport}

	// Send request
	response, err := client.Do(req)
	if err != nil {
		err = fmt.Errorf("send request error: %v", err)
		return err
	}
	defer response.Body.Close()

	// Read the response body
	body, _ := ioutil.ReadAll(response.Body)
	if err != nil {
		err = fmt.Errorf("error reading response body: %v", err)
		return err
	}

	// Printing response results
	fmt.Printf("Response results: %s\n", body)
	if response.StatusCode == http.StatusOK {
		if bytes.Contains(body, []byte("challengeID")) {
			saveJobs.SuccessJobs <- phone
			fmt.Printf("phone:%s send success\n", phone)
		} else {
			ap := &AgainJobsStr{
				Job: phone,
			}
			ap.Frequency++
			saveJobs.AgainJobs <- ap
			MutexAgainNum.Lock()
			AgainNum++
			MutexAgainNum.Unlock()
			fmt.Printf("phone:%s send error。%s\n", phone, body)
		}
	} else {
		err = fmt.Errorf("phone:%s send error。Status Code: %d\n", phone, response.StatusCode)
	}
	return err
}
