package record

import (
	"bufio"
	"fmt"
	"os"
	"sync"
)

func init() {
	successFile, err := os.Create("record/success.txt")
	if err != nil {
		fmt.Printf("error reading response body: %v", err)
		return
	}
	defer successFile.Close()

	// Create or open a file
	failFile, err := os.Create("record/fail.txt")
	if err != nil {
		fmt.Printf("error reading response body: %v", err)
		return
	}
	defer failFile.Close()
}

// WriteSuccessJob 输出正确的job
func WriteSuccessJob(jobs chan string, mutex *sync.Mutex, num *int) {
	// Open file
	file, err := os.OpenFile("record/success.txt", os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		fmt.Printf("an error occurred while opening the file: %v", err)
		return
	}
	defer file.Close()

	// Creating a writer
	writer := bufio.NewWriter(file)

	// Write data line by line
	for {
		job, ok := <-jobs
		if !ok {
			break
		}
		fmt.Printf("============success phone{%v}=============\n", job)
		_, err = writer.WriteString(job + "\n")
		if err != nil {
			fmt.Printf("write error: %v", err)
			return
		}

		// Refresh buffer
		err = writer.Flush()
		if err != nil {
			fmt.Printf("flush error: %v", err)
			return
		}
		mutex.Lock()
		*num--
		mutex.Unlock()
	}
}

// WriteFailJob 输出失败的Job
func WriteFailJob(jobs chan string, mutex *sync.Mutex, num *int) {
	// Open file
	file, err := os.OpenFile("record/fail.txt", os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		fmt.Printf("an error occurred while opening the file: %v", err)
		return
	}
	defer file.Close()

	// Creating a writer
	writer := bufio.NewWriter(file)

	// Write data line by line
	for {
		job, ok := <-jobs
		if !ok {
			break
		}
		fmt.Printf("============fail job{%v}=============\n", job)
		_, err = writer.WriteString(job + "\n")
		//_, err = fmt.Fprintf(writer, job)
		if err != nil {
			fmt.Printf("write error: %v", err)
			return
		}

		// Refresh buffer
		err = writer.Flush()
		if err != nil {
			fmt.Printf("flush error: %v", err)
			return
		}
		mutex.Lock()
		*num--
		mutex.Unlock()
	}
}
