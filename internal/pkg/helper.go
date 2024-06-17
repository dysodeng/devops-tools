package pkg

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"sync"
)

// ExecCmd 执行系统命令
func ExecCmd(cmd *exec.Cmd) error {
	cmd.Stdin = os.Stdin

	var wg sync.WaitGroup
	wg.Add(2)

	// 捕获标准输出
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Println("INFO: ", err)
		os.Exit(1)
	}
	readout := bufio.NewReader(stdout)
	go func() {
		defer wg.Done()
		PrintOutput(readout)
	}()

	// 捕获标准错误
	stderr, err := cmd.StderrPipe()
	if err != nil {
		fmt.Println("ERROR: ", err)
		os.Exit(1)
	}
	readerr := bufio.NewReader(stderr)
	go func() {
		defer wg.Done()
		PrintOutput(readerr)
	}()

	// 执行命令
	err = cmd.Run()
	if err != nil {
		return err
	}
	wg.Wait()

	return nil
}

func PrintOutput(reader *bufio.Reader) {
	var sumOutput string
	outputBytes := make([]byte, 200)

	for {
		n, err := reader.Read(outputBytes)
		if err != nil {
			break
		}
		output := string(outputBytes[:n])
		fmt.Print(output)
		sumOutput += output
	}
}

// CheckNetworkFileExists 检查网络文件是否存在
func CheckNetworkFileExists(url string) bool {
	resp, err := http.Head(url) // 发送HEAD请求获取HTTP头信息
	if err != nil {
		// 处理错误情况
		return false
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	statusCode := resp.StatusCode
	switch statusCode {
	case http.StatusOK:
		fallthrough
	case http.StatusPartialContent:
		return true
	default:
		return false
	}
}
