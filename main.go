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
	"sync"

	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql"
	"github.com/robertkrimen/otto"
)

const (
	MAX_LOG_FILE_SIZE = 1 << 20 * 50
)

var (
	fd       *os.File
	pid_file string

	config = loadSettings()
	PinyinMap = loadPinyinMap()
	SHOPNC_ROOT = config["crawler"]["shopnc_root"]
	loggerFileName = "./gocrawler.log"
	logger = initLogManager()
	wg sync.WaitGroup
	brands = make(AutoHomeBrands)
	vm = otto.New()

	rootCtx  context.Context
	cancel   context.CancelFunc
	flagS    = flag.String("s", "", "send a 'stop' or 'reload' message to the Main Progress")
	flagC    = flag.String("c", "./config.ini", "use another setting file path, default: ./config.ini")
)

func init() {
	db_type, ok := config["db"]["db_type"]
	if !ok {
		exitWithError(errors.New("can not get db_type config"))
	}

	db_host, ok := config["db"]["db_host"]
	if !ok {
		exitWithError(errors.New("can not get db_host config"))
	}

	db_port, ok := config["db"]["db_port"]
	if !ok {
		exitWithError(errors.New("can not get db_port config"))
	}

	db_user, ok := config["db"]["db_user"]
	if !ok {
		exitWithError(errors.New("can not get db_user config"))
	}

	db_passwd, ok := config["db"]["db_passwd"]
	if !ok {
		exitWithError(errors.New("can not get db_passwd config"))
	}

	db_name, ok := config["db"]["db_name"]
	if !ok {
		exitWithError(errors.New("can not get db_name config"))
	}

	prefix := "lrlz_"
	maxIdle := 30
	maxConn := 30
	dsn := db_user + ":" + db_passwd + "@tcp(" + db_host + ":" + db_port + ")/" + db_name + "?charset=utf8"
	orm.RegisterDataBase("default", db_type, dsn, maxIdle, maxConn)
	orm.RegisterModelWithPrefix(prefix, new(Brand), new(CarSeries), new(CarCrawl)) //TODO 注册model与数据表关联
	orm.Debug = true
	orm.DebugLog = orm.NewLog(logger.O)
	go monitorLogFileSize()
}

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

func signalHandle() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGHUP, syscall.SIGUSR2)

	for {
		sig := <-ch

		log.Println("received signal", sig)
		switch sig {
		case syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
			// graceful stop
			fd.Close()
			//cancel() // end mainProgress
			exitUnlockPid()
			log.Println("bye bye")
			signal.Stop(ch)
			return
		case syscall.SIGHUP, syscall.SIGUSR2:
			// graceful startProgress
			//cancel() // end mainProgress

			log.Println("all workers canceled")
		}
	}
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
