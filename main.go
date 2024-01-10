package main

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"concurrentrix/config"
	"concurrentrix/core"
	"concurrentrix/core/tools"
	"concurrentrix/internal"
	"concurrentrix/record"
)

func main() {
	fmt.Printf("============ START TASK =============\n")
	phones, err := readPhone()
	if err != nil {
		fmt.Print(err.Error())
		return
	}

	useJobs := len(phones)
	go record.WriteSuccessPhone(core.SuccessPhones, &core.MutexUseJobs, &useJobs)
	go record.WriteFailPhone(core.FailPhones, &core.MutexUseJobs, &useJobs)

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
		close(core.SuccessPhones)
		close(core.FailPhones)
		close(core.AgainPhones)
		close(phoneJobs)
		os.Exit(1)
	}()

	for {
		if core.SendJobs == len(phones) && useJobs == 0 && core.AgainNum == 0 {
			close(core.SuccessPhones)
			close(core.FailPhones)
			close(core.AgainPhones)
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
			xRay, err := core.GetXRay(tools.GetPhoneSha256(phone))
			if err != nil {
				fmt.Print(err.Error())
				return
			}
			fmt.Printf("============ START SEND PHONE{%v}=============\n", phone)
			err = core.SmsSend(phone, xRay)
			if err != nil {
				if config.SendTimes == 0 {
					core.FailPhones <- phone
				} else {
					ap := &internal.AgainSendPool{
						Phone: phone,
					}
					ap.Frequency++
					core.AgainPhones <- ap
					core.MutexAgainNum.Lock()
					core.AgainNum++
					core.MutexAgainNum.Unlock()
				}
				fmt.Printf("%v\n", err)
			}
			core.MutexSendJobs.Lock()
			core.SendJobs++
			core.MutexSendJobs.Unlock()
		}
	}()
}

func againSend() {
	go func() {
		for {
			agPh := <-core.AgainPhones
			core.MutexAgainNum.Lock()
			core.AgainNum--
			core.MutexAgainNum.Unlock()
			xRay, err := core.GetXRay(tools.GetPhoneSha256(agPh.Phone))
			if err != nil {
				fmt.Print(err.Error())
				return
			}
			if agPh.Frequency == config.SendTimes {
				core.FailPhones <- agPh.Phone
			} else {
				fmt.Printf("============ {%d} AGAIN START SEND PHONE{%v}=============\n", agPh.Frequency, agPh.Phone)
				err = core.SmsSend(agPh.Phone, xRay)
				if err != nil {
					fmt.Printf("%v\n", err)
					core.AgainPhones <- &internal.AgainSendPool{
						Phone:     agPh.Phone,
						Frequency: agPh.Frequency + 1,
					}
					core.MutexAgainNum.Lock()
					core.AgainNum++
					core.MutexAgainNum.Unlock()
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
