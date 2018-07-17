package main

import (
	"testing"
)

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