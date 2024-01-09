package main

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"concurrentrix/config"
	"concurrentrix/record"
)

var (
	SuccessPhones = make(chan string, config.ChanNum)
	FailPhones    = make(chan string, config.ChanNum)
	AgainPhones   = make(chan *AgainSendPool, config.ChanNum)
	MutexSendJobs sync.Mutex
	MutexUseJobs  sync.Mutex
	MutexAgainNum sync.Mutex
	SendJobs      = 0
	AgainNum      = 0
)

func main() {
	fmt.Printf("============ START TASK =============\n")
	phones, err := readPhone()
	if err != nil {
		fmt.Print(err.Error())
		return
	}

	useJobs := len(phones)
	go record.WriteSuccessPhone(SuccessPhones, &MutexUseJobs, &useJobs)
	go record.WriteFailPhone(FailPhones, &MutexUseJobs, &useJobs)

	phoneJobs := make(chan string)

	for i := 0; i < config.Work; i++ {
		threadSend(phoneJobs)
		againSend()
	}

	var lastPhone string
	for _, phone := range phones {
		phoneJobs <- phone
		lastPhone = phone
	}

	go func() {
		signalChan := make(chan os.Signal, 1)
		signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
		<-signalChan
		signal.Stop(signalChan)
		fmt.Printf("\n received interrupt signal. exiting... \n")
		// Use the lastPhone variable here
		fmt.Printf("Last phone received: %s\n", lastPhone)
		close(SuccessPhones)
		close(FailPhones)
		close(AgainPhones)
		close(phoneJobs)
		os.Exit(1)
	}()

	for {
		if SendJobs == len(phones) && useJobs == 0 && AgainNum == 0 {
			close(SuccessPhones)
			close(FailPhones)
			close(AgainPhones)
			close(phoneJobs)
			fmt.Printf("Last phone received: %s\n", lastPhone)
			os.Exit(1)
		}
	}
}

func threadSend(jobs chan string) {
	go func() {
		for {
			phone := <-jobs
			xRay, err := getXRay(getPhoneSha256(phone))
			if err != nil {
				fmt.Print(err.Error())
				return
			}
			fmt.Printf("============ START SEND PHONE{%v}=============\n", phone)
			err = smsSend(phone, xRay)
			if err != nil {
				if config.SendTimes == 0 {
					FailPhones <- phone
				} else {
					ap := &AgainSendPool{
						phone: phone,
					}
					ap.Frequency++
					AgainPhones <- ap
					MutexAgainNum.Lock()
					AgainNum++
					MutexAgainNum.Unlock()
				}
				fmt.Printf("%v\n", err)
			}
			MutexSendJobs.Lock()
			SendJobs++
			MutexSendJobs.Unlock()
		}
	}()
}

type AgainSendPool struct {
	phone     string
	Frequency int
}

func againSend() {
	go func() {
		for {
			agPh := <-AgainPhones
			MutexAgainNum.Lock()
			AgainNum--
			MutexAgainNum.Unlock()
			xRay, err := getXRay(getPhoneSha256(agPh.phone))
			if err != nil {
				fmt.Print(err.Error())
				return
			}
			if agPh.Frequency == config.SendTimes {
				FailPhones <- agPh.phone
			} else {
				fmt.Printf("============ {%d} AGAIN START SEND PHONE{%v}=============\n", agPh.Frequency, agPh.phone)
				err = smsSend(agPh.phone, xRay)
				if err != nil {
					fmt.Printf("%v\n", err)
					AgainPhones <- &AgainSendPool{
						phone:     agPh.phone,
						Frequency: agPh.Frequency + 1,
					}
					MutexAgainNum.Lock()
					AgainNum++
					MutexAgainNum.Unlock()
				}
			}
		}
	}()
}

func readPhone() (phones []string, err error) {
	file, err := os.Open("used_nuber.txt")
	if err != nil {
		err = fmt.Errorf(err.Error())
		return
	}
	defer file.Close()

	// Creating a Scanner with the bufio package
	scanner := bufio.NewScanner(file)

	// Read the contents of a file line by line
	for scanner.Scan() {
		line := scanner.Text()
		// 处理前后空格
		trimmedLine := strings.TrimSpace(line)
		phones = append(phones, trimmedLine)
	}

	// Checking for scanning errors
	err = scanner.Err()
	if err != nil {
		err = fmt.Errorf("error while scanning a document: %v", err)
		return
	}
	return
}

// Coded as Base64
func getPhoneSha256(strToHash string) string {
	// First SHA-256 hash
	sha256Hash := sha256.New()
	sha256Hash.Write([]byte(strToHash))
	hashedStr := sha256Hash.Sum(nil)
	hashedStrHex := hex.EncodeToString(hashedStr)

	// Second SHA-256 hash
	sha256Hash = sha256.New()
	sha256Hash.Write([]byte(hashedStrHex))
	digest := sha256Hash.Sum(nil)
	hexDigest := hex.EncodeToString(digest)

	return hexDigest
}

// Gain Authorization
func getPhoneSha256s(strToHash string) string {
	sha256Hash := sha256.New()
	sha256Hash.Write([]byte(strToHash))
	hashedStr := sha256Hash.Sum(nil)
	hashedStrHex := hex.EncodeToString(hashedStr)
	return hashedStrHex
}

var smsSendUrl = "https://api.grab.com/grabid/v1/phone/otp"
var getXRayUrl = "http://154.221.17.198:5612/business/invoke?group=huoyu_test&action=nb&arg1="
var tunnel = "i635.kdltps.com:15818"
var username = "t10407739755674"
var password = "op1un18b"

type xRayStr struct {
	ClientID string
	Data     string
	Message  interface{}
}

func getXRay(enco string) (string, error) {
	resp, err := http.Get(getXRayUrl + enco)
	if err != nil {
		err = fmt.Errorf("http get err: %v", err)
		return "", err
	}
	defer resp.Body.Close()

	// Read the response body
	var result *xRayStr
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&result)
	if err != nil {
		err = fmt.Errorf("decode err: %v", err)
		return "", err
	}

	return result.Data, nil
}

func generateRandomDriverID() string {
	// Setting the random number seed
	rand.Seed(time.Now().UnixNano())

	// Generate random numbers between 1000 and 9999
	randomNumber := rand.Intn(9000) + 1000

	// Generate driver_id with a fixed suffix
	driverID := fmt.Sprintf("%d%s", randomNumber, "c6509c1db2d1")

	return driverID
}

func smsSend(phone, xRay string) error {
	// Constructing the request header
	headers := map[string]string{
		"Connection":           "close",
		"X-Ray":                xRay,
		"X-Grab-Device-ID":     generateRandomDriverID(),
		"alias_sessionid":      "2312271214-0yu1VjiyG9gadd" + xRay,
		"x-request-id":         fmt.Sprintf("%s", time.Now().UnixNano()),
		"X-Grab-Device-Model":  "Pixel 4 XL",
		"Authorization":        "grabsecure " + getPhoneSha256s(phone),
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
			SuccessPhones <- phone
			fmt.Printf("phone:%s send success\n", phone)
		} else {
			ap := &AgainSendPool{
				phone: phone,
			}
			ap.Frequency++
			AgainPhones <- ap
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
