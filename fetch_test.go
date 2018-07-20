package main

import (
	"testing"
	"os"
	"bufio"
	"io"
	"strings"
	"regexp"
)

func Test_Regexp(t *testing.T) {
	pat := `<[^>]+>`
	str := "三<span class='hs_kw6_configEQ'></span>10<span class='hs_kw1_configEQ'></span>公里"
	reg := regexp.MustCompile(pat)
	found := reg.Split(str, -1)
	t.Log(found)
}

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

func Test_FetchCarType(t *testing.T) {
	// 威途X35 2017款 RQ5026XXYEVH0 45kWh		ID:1005463		https://www.autohome.com.cn/spec/1005463
	// 帕萨特 2017款 280TSI DSG尊雅版		ID:29314	https://www.autohome.com.cn/spec/29314/
	var (
		car_id int
		car_name string
		car_url string
	)
	//car_id, car_name, car_url = 1005463, "威途X35 2017款 RQ5026XXYEVH0 45kWh", "https://www.autohome.com.cn/spec/1005463"
	car_id, car_name, car_url = 29314, "帕萨特 2017款 280TSI DSG尊雅版", "https://www.autohome.com.cn/spec/29314/"
	car := NewCar(car_id,car_name,car_url)
	car.SetPrice("18.99万")
	info, err := fetchCarInfo(car)
	if err != nil {

	}
	t.Log(info)
}

func Test_getAutoHomeBrand(t *testing.T) {
	// 大众: 1	欧宝: 59
	sUrl := "https://car.autohome.com.cn/price/brand-59.html"
	brand_name := "欧宝"
	brands["欧宝"] = &AutoHomeBrand{Name:brand_name}

	getAutoHomeBrand(sUrl, brand_name)
	t.Log(brands["欧宝"])
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