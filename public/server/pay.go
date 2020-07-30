package main

import (
	"bufio"
	"fmt"
	"os/exec"

	"golang.org/x/text/encoding/simplifiedchinese"
)

func getOutputDirectly(name string, args ...string) (output []byte) {
	cmd := exec.Command(name, args...)
	output, err := cmd.Output() // 等到命令执行完, 一次性获取输出
	if err != nil {
		panic(err)
	}
	output, err = simplifiedchinese.GB18030.NewDecoder().Bytes(output)
	if err != nil {
		panic(err)
	}
	return
}

func getOutputContinually(name string, args ...string) (output chan []byte) {
	cmd := exec.Command(name, args...)
	fmt.Println(cmd)
	output = make(chan []byte, 10240)
	defer close(output)

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		panic(err)
	}
	defer stdoutPipe.Close()

	go func() {
		scanner := bufio.NewScanner(stdoutPipe)
		for scanner.Scan() { // 命令在执行的过程中, 实时地获取其输出
			data, err := simplifiedchinese.GB18030.NewDecoder().Bytes(scanner.Bytes()) // 防止乱码
			if err != nil {
				fmt.Println("transfer error with bytes:", scanner.Bytes())
				continue
			}

			fmt.Printf("%s\n", string(data))
		}
	}()

	if err := cmd.Run(); err != nil {
		panic(err)
	}
	return
}

func main() {
	var path string;
	path = "D:\\www\\liuzhongsheng\\public\\vod\\public\\server\\video\\1.mkv"
	output1 := getOutputDirectly("ffmpeg","-re","-i",path,"-f","flv","rtmp://127.0.0.1:1935/live/124")
	fmt.Printf("%s\n", output1)
	// 不断输出, 直到结束
	
	output2 := getOutputContinually("ffmpeg","-re","-i",path,"-f","flv","rtmp://127.0.0.1:1935/live/123")
	for o := range output2 {
		fmt.Printf("%s\n", o)
	}
}