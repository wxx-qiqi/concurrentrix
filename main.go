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
	"concurrentrix/record"
)

func main() {
	fmt.Printf("============ START TASK =============\n")
	jobs, err := readJob()
	if err != nil {
		fmt.Print(err.Error())
		return
	}

	useJobsNum := len(jobs)
	var lastJob string
	saveJobs := core.NewSaveJobsStr(config.ChanNum)

	go record.WriteSuccessJob(saveJobs.SuccessJobs, &saveJobs.MutexUseJobs, &useJobsNum)
	go record.WriteFailJob(saveJobs.FailJobs, &saveJobs.MutexUseJobs, &useJobsNum)
	go func() {
		signalChan := make(chan os.Signal, 1)
		signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
		<-signalChan
		signal.Stop(signalChan)
		fmt.Printf("\n received interrupt signal. exiting... \n")
		// Use the lastJob variable here
		fmt.Printf("Last job received: %s\n", lastJob)
		saveJobs.Stop()
		os.Exit(1)
	}()

	for i := 0; i < config.Work; i++ {
		core.ThreadSend(saveJobs)
		core.AgainSend(saveJobs)
	}

	for _, job := range jobs {
		saveJobs.DealingJobs <- job
		lastJob = job
	}

	for {
		if core.SendJobs == len(jobs) && useJobsNum == 0 && core.AgainNum == 0 {
			saveJobs.Stop()
			fmt.Printf("Last job received: %s\n", lastJob)
			os.Exit(1)
		}
	}
}

func readJob() (jobs []string, err error) {
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
		jobs = append(jobs, trimmedLine)
	}

	// Checking for scanning errors
	err = scanner.Err()
	if err != nil {
		err = fmt.Errorf("error while scanning a document: %v", err)
		return
	}
	return
}
