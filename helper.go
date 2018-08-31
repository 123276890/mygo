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
	"math/rand"
	"strconv"
	"time"
	"net/url"
	"net/http"
	"io/ioutil"
	"encoding/json"

	"github.com/axgle/mahonia"
)

const (
	HELP_TEXT = `
Usage:
get all: 
		抓取汽车之家所有品牌数据
help:
		查看帮助
exit: 
		退出
`
)

type logManager struct{
	*log.Logger
	m			*sync.Mutex
	O			*os.File
}

func initLogManager() *logManager {
	var once sync.Once
	var l *logManager
	once.Do(func() {
		var err error
		var logFileWriter *os.File
		realpath, err := filepath.Abs(filepath.Dir(loggerFileName))
		if err != nil {
			log.Fatal("Failed:", realpath)
			return
		}
		basename := filepath.Base(loggerFileName)
		file_fullpath := filepath.Join(realpath, basename)
		fmt.Println("opening file_fullpath:", file_fullpath)
		fmt.Println("tail -f ", file_fullpath)

		if checkFileExist(file_fullpath) == true {
			logFileWriter, err = os.OpenFile(file_fullpath, os.O_APPEND|os.O_RDWR, os.ModeAppend)
		} else {
			logFileWriter, err = os.Create(file_fullpath)
		}

		if err != nil {
			log.Fatalln("Failed:", err)
			return
		}
		l = &logManager{}
		l.Logger = log.New(logFileWriter, "", log.LstdFlags)
		l.m = new(sync.Mutex)
		l.O = logFileWriter
		l.Println("\n")
		l.Record("Program Start...")
	})
	return l
}

func (l *logManager) Record(v ...interface{}) {
	l.m.Lock()
	defer l.m.Unlock()

	fmt.Println(v...)
	l.Logger.Println(v...)
}

func monitorLogFileSize() {
	for {
		file_info, err := logger.O.Stat()	//os.Stat(log_file)
		if err != nil {
			break
		}
		filesize := file_info.Size()
		if filesize > MAX_LOG_FILE_SIZE {
			wg.Add(1)
			logger_file_name := logger.O.Name()
			contents, err := ioutil.ReadFile(logger_file_name)
			if err != nil {
				logger.Record(err)
				continue
			}

			logger.m.Lock()

			backup_filename := strconv.Itoa(time.Now().Year()) + strconv.Itoa(int(time.Now().Month())) + strconv.Itoa(time.Now().Day()) + "_backup.log"
			backup_file, err := os.OpenFile(backup_filename, os.O_WRONLY | os.O_CREATE, os.ModePerm)
			if err != nil {
				goto errHandler
			}

			_, err = backup_file.Write(contents)
			if err != nil {
				goto errHandler
			}
			logger.O.Truncate(0)

		errHandler:
			logger.m.Unlock()
			backup_file.Close()
			wg.Done()
		}
		time.Sleep(time.Second * 10)
	}
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

func xget(urlstr string) (*http.Response, error){
	var (
		err error
	)
	method := "GET"
	u, err := url.Parse(urlstr)
	if err != nil {
		logger.Record(err)
		return nil, err
	}
	host := u.Host
	filename := host + ".cookie"
	f, err := os.OpenFile(filename, os.O_RDWR | os.O_CREATE, os.ModePerm)
	if err != nil {
		logger.Record(err)
		return nil, err
	}

	req, err := http.NewRequest(method, urlstr, nil)
	if err != nil {
		logger.Record(err)
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/67.0.3396.99 Safari/537.36")
	cookies_bytes, err := ioutil.ReadAll(f)
	if err != nil {
		if err != io.EOF {
			logger.Record(err)
			return nil, err
		}
	}

	cookies := loadCookies(cookies_bytes)
	for _, c := range cookies {
		req.AddCookie(c)
	}

	clnt := http.Client{}

	resp, err := clnt.Do(req)
	if err != nil {
		logger.Record(err)
		return nil, err
	}

	cookies_resp := resp.Cookies()
	for _, c := range cookies_resp {
		exist := false
		for _, v := range cookies {
			if c.Raw == v.Raw {
				exist = true
			}
		}

		if exist {
			continue
		}

		str, err := json.Marshal(c)
		if err != nil {
			log.Fatal(err)
			continue
		}
		str = append(str,'\n')
		io.WriteString(f, string(str))
	}
	return resp, nil
}

func loadCookies(cookies_bytes []byte) ([]*http.Cookie) {
	var cookies []*http.Cookie
	datas := bytes.Split(cookies_bytes, []byte("\n"))
	for _, data := range datas {
		if len(data) <= 0 {
			continue
		}
		r := bytes.NewReader(data)
		dc := json.NewDecoder(r)
		var cookie http.Cookie
		err := dc.Decode(&cookie)
		if err != nil {
			log.Fatal(err)
			continue
		}
		cookies = append(cookies, &cookie)
	}
	return cookies
}