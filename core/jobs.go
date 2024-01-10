package core

import (
	"sync"
)

type AgainJobsStr struct {
	Job       string
	Frequency int
}

type SaveJobsStr struct {
	DealingJobs  chan string
	SuccessJobs  chan string
	FailJobs     chan string
	AgainJobs    chan *AgainJobsStr
	MutexUseJobs sync.Mutex
}

func NewSaveJobsStr(size int) *SaveJobsStr {
	return &SaveJobsStr{
		DealingJobs:  make(chan string, size),
		SuccessJobs:  make(chan string, size),
		FailJobs:     make(chan string, size),
		AgainJobs:    make(chan *AgainJobsStr, size),
		MutexUseJobs: sync.Mutex{},
	}
}

func (s *SaveJobsStr) Stop() {
	close(s.FailJobs)
	close(s.SuccessJobs)
	close(s.AgainJobs)
	close(s.DealingJobs)
}
