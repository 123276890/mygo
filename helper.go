package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"bytes"
	"strings"
	"bufio"
	"io"
	"errors"
	//crand "crypto/rand"
	"math/rand"
	"strconv"
	"time"

	"github.com/axgle/mahonia"
)

type logManager struct{
	*log.Logger
	O			*os.File
}

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
		l = &logManager{log.New(logFileWriter, "", log.LstdFlags),logFileWriter}
		l.Println("\n")
		l.Record("Program Start...")
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

func exitWithError(err error) {
	logger.Record("Program fatal on Exit:",err)
	os.Exit(1)
}

func loadSettings() (map[string]map[string]string) {
	var configFile = "./config.ini"
	fmt.Println("load settings...")
	settings := make(map[string]map[string]string)
	if checkFileExist(configFile) == false {
		err := errors.New("config.ini file not found under current path")
		exitWithError(err)
	}

	file, err := os.Open(configFile)
	if err != nil {
		exitWithError(err)
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	section := "default"
	for {
		line, _, err := reader.ReadLine()
		//文件末尾
		if err == io.EOF {
			break
		}

		//空行跳过
		if bytes.Equal(line, []byte("")) {
			continue
		}
		line = bytes.TrimSpace(line)
		//跳过注释行
		if bytes.HasPrefix(line, []byte("#")) {
			//fmt.Println("line",i,"has #")
			continue
		}

		if bytes.HasPrefix(line, []byte("[")) && bytes.HasSuffix(line, []byte("]")) {
			section = strings.TrimSpace(string(line[1:len(line) - 1]))
			section = strings.ToLower(section)
			settings[section] = make(map[string]string)
		} else {
			str := string(line)
			if strings.Contains(str, "=") {
				pair := strings.SplitN(str, "=",2)
				key := strings.TrimSpace(pair[0])
				val := strings.TrimSpace(pair[1])
				if _, isset := settings[section]; isset {
					settings[section][key] = val
				}
			}
		}
	}
	return settings
}

type buffer []byte

func NewBuffer() *buffer {
	return &buffer{}
}

func (b *buffer) Write(p []byte) (int, error) {
	*b = append(*b, p...)
	return len(*b), nil
}

func loadPinyinMap() (map[string]string) {
	pinyins := make(map[string]string)
	repeats := map[string]string{}
	var (
		arr []string
		key string
		pinyin string
	)

	f, err := os.Open("pinyin.txt")
	if err != nil {
		logger.Record(err)
		return pinyins
	}
	defer f.Close()

	f_repeat, err := os.Open("repeats.txt")
	if err != nil {
		logger.Record(err)
		return pinyins
	}
	defer f_repeat.Close()

	r_repeat := bufio.NewReader(f_repeat)
	for {
		l, err := r_repeat.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			break
		}

		l = strings.TrimSpace(l)
		if l == "" {
			continue
		}
		arr = strings.Split(l, " ")
		key = arr[0]
		pinyin = arr[1]
		repeats[key] = pinyin
	}
	arr = nil

	buf := bufio.NewReader(f)
	for {
		line, err := buf.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			logger.Record(err)
			break
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		arr = strings.Split(line, " ")
		key = arr[0]
		pinyin = arr[1]

		if v, isset := repeats[key]; isset {
			pinyins[key] = v
		} else {
			pinyins[key] = pinyin
		}
	}
	return pinyins
}

func reNameSameFileName(filename string, path string) (string){
	var f string
	rand.Seed(time.Now().UnixNano())
	if strings.Contains(filename, ".") {
		pos := strings.LastIndex(filename, ".")
		f = filename[:pos] + strconv.Itoa(rand.Intn(10)) + filename[pos:]
	} else {
		f = filename + strconv.Itoa(rand.Intn(10))
	}

	if checkFileExist(filepath.Join(path,f)) {
		return reNameSameFileName(f, path)
	} else {
		return f
	}
}