package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"syscall"
)

func consoleReceiveCommand() {
	showHelp()
	console := bufio.NewReader(os.Stdin)
	fmt.Println("Input a Command:")
	for {
		fmt.Print(">>>")
		input, err := console.ReadString('\n')
		if err != nil {
			fmt.Println(err)
			return
		}

		command := strings.TrimSuffix(string(input), "\n")
		command = strings.ToLower(command)
		//fmt.Println("command:",command)
		switch command {
		case "get brands":
			JobGetAutoHomeBrands()
		case "exit":
			pid := getPidFromFile()
			if pid > 0 {
				syscall.Kill(pid, syscall.SIGQUIT)
				exitUnlockPid()
			}
			return
		case "help":
			showHelp()
		default:
		}
	}
}

func showHelp() {
	help := "Usage:" + "\n" +
		"get brands: " + "\n" +
		"			抓取汽车之家品牌数据" + "\n" +
		"help: \n" +
		"			查看帮助" + "\n" +
		"exit: \n" +
		"			退出" + "\n"
	fmt.Println(help)
}

func startWorkers(ctx context.Context) {

}
