package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
)

var (
	fd       *os.File
	pid_file string
	config   = make(map[string]string)
	rootCtx  context.Context
	cancel   context.CancelFunc
	flagS    = flag.String("s", "", "send a 'stop' or 'reload' message to the Main Progress")
	flagC    = flag.String("c", "./config.ini", "use another setting file path, default: ./config.ini")
)

func main() {
	var err error
	runtime.GOMAXPROCS(runtime.NumCPU())
	flag.Parse()

	process_name := os.Args[0]
	pid_file = process_name + ".pid"

	// 如果有启动参数
	if len(os.Args[1:]) > 0 {
		if *flagS != "" {
			switch *flagS {
			case "run xxx":
			default:
				exitWithError(errors.New("wrong param usage"))
				return
			}
		}

		if *flagC != "" {
			p, _ := filepath.Abs(*flagC)
			if p_exist := checkFileExist(p); !p_exist {
				fmt.Println("config file", p, "does not exist")
				os.Exit(1)
			}

			pid := getPidFromFile()
			if pid > 0 {
				// reload config
				//loadSettings(p)
				fmt.Println(pid)
				syscall.Kill(pid, syscall.SIGUSR2)
			}
		}
	}

	fd, err = os.OpenFile(pid_file, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0400)
	defer fd.Close()

	if err != nil {
		fmt.Println("program is running...Exit")
		os.Exit(1)
	}

	loadSettings("")
	pid := os.Getpid()
	fd.Write([]byte(strconv.Itoa(pid)))

	// main progress here
	go consoleReceiveCommand()

	signalHandle()
}

func exitUnlockPid() {
	err := os.Remove(pid_file)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("delete pid success")
	}
	os.Exit(0)
}

func exitWithError(err error) {
	log.Println(err)
	os.Exit(1)
}

func checkFileExist(filename string) bool {
	var exist bool = true
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		exist = false
	}
	return exist
}

func signalHandle() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGHUP, syscall.SIGUSR2)

	for {
		sig := <-ch

		switch sig {
		case syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
			// graceful stop
			log.Println("received signal Exit")
			signal.Stop(ch)
			fd.Close()
			//cancel() // end mainProgress
			exitUnlockPid()
			log.Println("bye bye")
			return
		case syscall.SIGHUP, syscall.SIGUSR2:
			// graceful startProgress
			log.Println("received signal restart")
			//cancel() // end mainProgress

			log.Println("all workers canceled")
		}
	}
}

func loadSettings(config_file string) {
	config["logfile"] = "./log.txt"
}

func getPidFromFile() int {
	file, err := os.Open(pid_file)
	defer file.Close()

	if err != nil {
		panic(err)
	}

	pid_bytes := make([]byte, 8)
	_, err = file.Read(pid_bytes)
	if err != nil {
		panic(err)
	}

	pid_str := string(pid_bytes)
	pid_str = strings.Trim(pid_str, string(byte(0)))
	pid, err := strconv.Atoi(pid_str)
	if err != nil {
		return 0
	}
	return pid
}
