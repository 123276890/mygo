package main

import (
	"testing"
	"os"
	"bufio"
	"io"
	"strings"
)

func Test_PinYinSuoXie(t *testing.T) {
	//str := "Icona"
	//str := "马自达"
	str := "广汽集团"
	pinyin := ""
	words_rune := []rune(str)
	for _, v := range words_rune {
		s := string(v)
		p, ok := PinyinMap[s]
		if ok {
			pinyin += string(p[0])
		}
	}
	t.Log(pinyin)
}

func Test_reNameSameFileName(t *testing.T) {
	filename := "dn.jpg"
	path := "/Users/a2/work/shopnc/data/upload/shop/brand/logo"

	result := reNameSameFileName(filename, path)
	t.Log(result)
}

func Test_ReadPinyinMap(t *testing.T) {
	pyMap := loadPinyinMap()
	t.Log(len(pyMap))
}

func Test_ConvertPinyinFile(t *testing.T) {
	f, err := os.Open("googlepinyin.txt")
	if err != nil {
		t.Fatal(err.Error())
		return
	}
	defer f.Close()

	w, err := os.OpenFile("pinyin.txt", os.O_APPEND|os.O_WRONLY|os.O_CREATE, os.ModePerm)
	if err != nil {
		t.Fatal("Error on openning output file:", err)
	}
	defer w.Close()

	r := bufio.NewReader(f)
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			} else {
				t.Fatal(err)
			}
			break
		}
		line = strings.TrimSpace(line)
		arr := strings.Split(line, " ")
		if len([]rune(arr[0])) > 1 {
			continue
		}
		output_line := arr[0] + " " + arr[3] + "\n"
		_, err = w.Write([]byte(output_line))
		if err != nil {
			t.Fatal("Error when write to output file:", err)
			break
		}
	}
	t.Log("output success")
}

func Test_getAutoHomeBrand(t *testing.T) {
	/*sUrl := ""
	brand_name := ""
	ret := getAutoHomeBrand(sUrl, brand_name)*/
}

func Test_FetchSeriesInfo(t *testing.T) {
	sUrl := "https://car.autohome.com.cn/price/series-528.html"
	ret, ok := fetchSeriesInfo(sUrl, "帕萨特", "大众")
	if ok {
		t.Log(ret)
	} else {
		t.Fatal("Can not fetch anything!")
	}
}