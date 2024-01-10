package core

import (
	"fmt"

	"concurrentrix/config"
	"concurrentrix/core/tools"
)

func ThreadSend(saveJobs *SaveJobsStr) {
	go func() {
		for {
			job := <-saveJobs.DealingJobs
			xRay, err := GetXRay(tools.GetPhoneSha256(job))
			if err != nil {
				fmt.Print(err.Error())
				return
			}
			fmt.Printf("============ START SEND PHONE{%v}=============\n", job)
			err = smsSend(job, xRay, saveJobs)
			if err != nil {
				if config.SendTimes == 0 {
					saveJobs.FailJobs <- job
				} else {
					ap := &AgainJobsStr{
						Job: job,
					}
					ap.Frequency++
					saveJobs.AgainJobs <- ap
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

func AgainSend(saveJobs *SaveJobsStr) {
	go func() {
		for {
			agPh := <-saveJobs.AgainJobs
			MutexAgainNum.Lock()
			AgainNum--
			MutexAgainNum.Unlock()
			xRay, err := GetXRay(tools.GetPhoneSha256(agPh.Job))
			if err != nil {
				fmt.Print(err.Error())
				return
			}
			if agPh.Frequency == config.SendTimes {
				saveJobs.FailJobs <- agPh.Job
			} else {
				fmt.Printf("============ {%d} AGAIN START SEND PHONE{%v}=============\n", agPh.Frequency, agPh.Job)
				err = smsSend(agPh.Job, xRay, saveJobs)
				if err != nil {
					fmt.Printf("%v\n", err)
					saveJobs.AgainJobs <- &AgainJobsStr{
						Job:       agPh.Job,
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
