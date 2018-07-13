package main

import (
	"bufio"
	"fmt"
	"bytes"
	"strings"
	"context"
	"os"
	"syscall"
)

func consoleReceiveCommand() {
	console := bufio.NewReader(os.Stdin)
	fmt.Println("Input a Command:")
	for {
		fmt.Print(">>>")
		input, _, err := console.ReadLine()
		if err != nil {
			fmt.Println(err)
			break
		}

		input = bytes.TrimSuffix(input, []byte("\n"))
		command := strings.ToLower(string(input))
		switch command {
		case "exit":
			pid := getPidFromFile()
			if pid > 0 {
				syscall.Kill(pid, syscall.SIGQUIT)
				os.Exit(0)
			}
			return
		case "help":
			showHelp()
		default:
		}
	}
}

func showHelp() {
	help := "Usage:" +
		"start:" +
		"stop:" +
		"exit:"
	fmt.Println(help)
}

func startWorkers(ctx context.Context) {

}
