package main

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

var (
	brands = make(Brands)
	logger = initLogManager("./log.txt")
)

func JobGetAutoHomeBrands() {
	brand_index_url := "https://car.autohome.com.cn/AsLeftMenu/As_LeftListNew.ashx?typeId=1 &brandId=0 &fctId=0 &seriesId=0"
	getAutoHomeBrands(brand_index_url)

	if !brands.IsEmpty() {
		fmt.Println(brands.Count())
		for _, v := range brands {
			getAutoHomeBrand(v.Url, v.Name)
			time.Sleep(time.Second * 1)
		}
	}
}

func getAutoHomeBrand(brandUrl string, brand_name string) {
	var charset string
	var schemes string
	var host string

	resp, err := http.Get(brandUrl)
	defer resp.Body.Close()
	if err != nil {
		logger.Record("Error: goqueryGet http.Get:", err)
		return
	}

	if resp.StatusCode != 200 {
		return
	}

	if content_type, ok := resp.Header["Content-Type"]; ok {
		pair := strings.SplitN(content_type[0], "=", 2)
		charset = pair[1]
		pair = nil
	}

	u, err := url.Parse(brandUrl)
	if err != nil {
		logger.Record("Error: goqueryGet ParseUrl:", err)
		return
	}
	schemes = u.Scheme
	host = u.Host

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		logger.Record("Error: goqueryGet Err:", err)
		return
	}

	contBox := doc.Find(".contentright").Find(".contbox")
	imgNode := contBox.Find(".carbrand").Find(".carbradn-pic").Find("img")
	imgSrc, ok := imgNode.Attr("src")
	if ok {
		if !strings.HasPrefix(imgSrc, schemes+":") {
			imgSrc = schemes + ":" + imgSrc
		}
		brands[brand_name].Img = imgSrc
	}

	logger.Record("Brand crawl Start:", brand_name)
	contNode := contBox.Find(".carbradn-cont").Find(".list-dl")
	factors := make(map[string]*Manufacture)

	contNode.Each(func(i int, dl *goquery.Selection) {
		dt := dl.Find("dt")
		manuf := &Manufacture{}

		mHref, ok := dt.Find("a").Attr("href")
		if ok {
			if !strings.HasPrefix(mHref, schemes+"://"+host) {
				mHref = schemes + "://" + host + mHref
				pair := strings.SplitN(mHref, "#", 2)
				mHref = pair[0]
				manuf.Url = mHref
			}
		}
		mName := dt.Find("a").Text()
		if mName != "" {
			mName = ChineseToUtf(mName, charset)
			mName = strings.TrimSpace(mName)
			manuf.Name = mName
		}

		factors[mName] = manuf
	})
	brands[brand_name].Manufactures = factors
	logger.Record("Brand crawl End:", brand_name)
	logger.Record(brands[brand_name])
}

func getAutoHomeBrands(sUrl string) {
	var charset string
	var schemes string
	var host string

	resp, err := http.Get(sUrl)
	defer resp.Body.Close()
	if err != nil {
		logger.Record("Error: goqueryGet http.Get:", err)
		return
	}

	if resp.StatusCode != 200 {
		return
	}

	if content_type, ok := resp.Header["Content-Type"]; ok {
		pair := strings.SplitN(content_type[0], "=", 2)
		charset = pair[1]
		pair = nil
	}

	u, err := url.Parse(sUrl)
	if err != nil {
		logger.Record("Error: goqueryGet ParseUrl:", err)
		return
	}
	schemes = u.Scheme
	host = u.Host

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		logger.Record("Error: goqueryGet Err:", err)
		return
	}

	contentNode := doc.Find(".cartree-letter")
	contentNode.Each(func(i int, s *goquery.Selection) {
		l := s.Text()
		l = strings.TrimSpace(ChineseToUtf(l, charset))

		lBrand := s.Next().Find("li")
		lBrand.Each(func(j int, li *goquery.Selection) {
			linkNode := li.Find("a")
			link, ok := linkNode.Attr("href")
			if ok {
				if !strings.HasPrefix(link, schemes+"://"+host) {
					link = schemes + "://" + host + link
				}

				link = strings.TrimSpace(link)
				brand_html, _ := linkNode.Html()
				var brand_name = ""

				if strings.Contains(brand_html, "</i>") {
					pair := strings.SplitN(brand_html, "</i>", 2)
					brand_html = pair[1]
					pair = nil

					if strings.Contains(brand_html, "<em>") {
						pair := strings.SplitN(brand_html, "<em>", 2)
						brand_name = pair[0]

						brand_name = strings.TrimSpace(ChineseToUtf(brand_name, charset))
						brands[brand_name] = &Brand{Name: brand_name, Url: link, Cap: l}

						em := pair[1]
						pair = nil
						em = strings.TrimPrefix(em, "(")
						em = strings.TrimSuffix(em, ")</em>")
						num, _ := strconv.Atoi(em)
						brands[brand_name].Nums = num
					}
				}
			}
			linkNode = nil
		})
	})
	contentNode = nil
	doc = nil
	logger.Record("init brands done,", "brands total:", len(brands))
}
