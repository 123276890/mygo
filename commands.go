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
		case "get all":
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
	fmt.Println(HELP_TEXT)
}

func startWorkers(ctx context.Context) {

}
