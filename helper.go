package main

import (
	"fmt"
	"github.com/axgle/mahonia"
	"log"
	"os"
	"path/filepath"
	"sync"
)

type logManager struct{ *log.Logger }

func initLogManager(filename string) *logManager {
	var once sync.Once
	var l *logManager
	once.Do(func() {
		var err error
		var logFileWriter *os.File
		realpath, err := filepath.Abs(filepath.Dir(filename))
		if err != nil {
			log.Fatal("Failed:", realpath)
			return
		}
		basename := filepath.Base(filename)
		logfile := filepath.Join(realpath, basename)
		fmt.Println("opening logfile:", logfile)
		fmt.Println("tail -f ", logfile)

		if checkFileExist(logfile) == true {
			logFileWriter, err = os.OpenFile(logfile, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
		} else {
			logFileWriter, err = os.Create(logfile)
		}

		if err != nil {
			log.Fatalln("Failed:", err)
			return
		}
		l = &logManager{log.New(logFileWriter, "[DEBUG]", log.LstdFlags)}
		l.Record("\n")
	})
	return l
}

func (l *logManager) Record(v ...interface{}) {
	fmt.Println(v...)
	l.Logger.Println(v...)
}

func ChineseToUtf(src string, srcCode string) string {
	if srcCode == "gb2312" {
		srcCode = "gbk"
	}
	srcCoder := mahonia.NewDecoder(srcCode)
	if srcCoder == nil {
		logger.Record("Error: Could not create Decoder for", srcCode)
		return ""
	}
	srcResult := srcCoder.ConvertString(src)
	targetCoder := mahonia.NewDecoder("utf-8")
	_, data, _ := targetCoder.Translate([]byte(srcResult), true)
	result := string(data)
	return result
}

func checkFileExist(filename string) bool {
	var exist bool = true
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		exist = false
	}
	return exist
}
