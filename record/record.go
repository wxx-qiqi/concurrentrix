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
		fmt.Printf("读取响应体错误: %v", err)
		return
	}
	defer successFile.Close()

	// 创建或打开文件
	failFile, err := os.Create("record/fail.txt")
	if err != nil {
		fmt.Printf("读取响应体错误: %v", err)
		return
	}
	defer failFile.Close()
}

// WriteSuccessPhone 输出正确的phone
func WriteSuccessPhone(phones chan string, mutex *sync.Mutex, jobs *int) {
	// 打开文件
	file, err := os.OpenFile("record/success.txt", os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		fmt.Printf("打开文件时发生错误: %v", err)
		return
	}
	defer file.Close()

	// 创建写入器
	writer := bufio.NewWriter(file)

	// 逐行写入数据
	for {
		phone, ok := <-phones
		if !ok {
			break
		}
		fmt.Printf("============success phone{%v}=============\n", phone)
		_, err = writer.WriteString(phone + "\n")
		if err != nil {
			fmt.Printf("写入错误: %v", err)
			return
		}

		// 刷新缓冲区
		err = writer.Flush()
		if err != nil {
			fmt.Printf("flush错误: %v", err)
			return
		}
		mutex.Lock()
		*jobs--
		mutex.Unlock()
	}
}

// WriteFailPhone 输出失败的phone
func WriteFailPhone(phones chan string, mutex *sync.Mutex, jobs *int) {
	// 打开文件
	file, err := os.OpenFile("record/fail.txt", os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		fmt.Printf("打开文件时发生错误: %v", err)
		return
	}
	defer file.Close()

	// 创建写入器
	writer := bufio.NewWriter(file)

	// 逐行写入数据
	for {
		phone, ok := <-phones
		if !ok {
			break
		}
		fmt.Printf("============fail phone{%v}=============\n", phone)
		_, err = writer.WriteString(phone + "\n")
		//_, err = fmt.Fprintf(writer, phone)
		if err != nil {
			fmt.Printf("写入错误: %v", err)
			return
		}

		// 刷新缓冲区
		err = writer.Flush()
		if err != nil {
			fmt.Printf("flush错误: %v", err)
			return
		}
		mutex.Lock()
		*jobs--
		mutex.Unlock()
	}
}
