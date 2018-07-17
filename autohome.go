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
	config["crawler"] = make(map[string]string)
	config["crawler"]["excepts"] = "欧宝,"

	excepts, ok := config["crawler"]["excepts"]
	if !ok {
		return
	}

	brand_index_url := "https://car.autohome.com.cn/AsLeftMenu/As_LeftListNew.ashx?typeId=1 &brandId=0 &fctId=0 &seriesId=0"
	getAutoHomeBrands(brand_index_url)

	if !brands.IsEmpty() {
		fmt.Println(brands.Count())
		for _, v := range brands {
			if strings.Contains(excepts, v.Name) {
				continue
			}
			getAutoHomeBrand(v.Url, v.Name)
			time.Sleep(time.Millisecond * 1000)
		}
	}
}

func fetchConfigUrl(sUrl string) (string, bool) {
	var charset string
	var schemes string
	var host string

	resp, err := http.Get(sUrl)
	defer resp.Body.Close()
	if err != nil {
		logger.Record("Error: goqueryGet http.Get:", err)
		return "", false
	}

	if resp.StatusCode != 200 {
		return "", false
	}

	if content_type, ok := resp.Header["Content-Type"]; ok {
		pair := strings.SplitN(content_type[0], "=", 2)
		charset = pair[1]
		pair = nil
	}

	u, err := url.Parse(sUrl)
	if err != nil {
		logger.Record("Error: goqueryGet ParseUrl:", err)
		return "", false
	}
	schemes = u.Scheme
	host = u.Host

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		logger.Record("Error: goqueryGet Err:", err)
		return "", false
	}

	configLink := doc.Find(".content").Find(".cartab-title").Find(".fn-right").Find("a").Eq(2)
	c := configLink.Text()
	c = strings.TrimSpace(ChineseToUtf(c, charset))
	if c == "配置" {
		link, ok := configLink.Attr("href")
		if ok {
			if !strings.HasPrefix(link, schemes + "://" + host) {
				link = schemes + "://" + host + link
			}

			if strings.Contains(link, "#") {
				link = strings.SplitN(link, "#", 2)[0]
			}

			return link, true
		}
	}
	return "", false
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

	logger.Record("Brand crawl Start:", brand_name)
	t_start := time.Now()

	contBox := doc.Find(".contentright").Find(".contbox")
	imgNode := contBox.Find(".carbrand").Find(".carbradn-pic").Find("img")
	imgSrc, ok := imgNode.Attr("src")
	if ok {
		if !strings.HasPrefix(imgSrc, schemes+":") {
			imgSrc = schemes + ":" + imgSrc
		}
		brands[brand_name].Img = imgSrc
	}

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
		//厂商旗下车系
		dd := dl.Find("dd").Find(".list-dl-text")
		serieses := make(map[string]*Series)
		dd.Each(func(j int, dldiv *goquery.Selection) {
			sLinks := dldiv.Find("a")
			sLinks.Each(func(k int, slink *goquery.Selection) {
				series_link, ok := slink.Attr("href")
				if ok {
					var s_status = ""

					s_name := slink.Text()
					s_name = ChineseToUtf(strings.TrimSpace(s_name), charset)
					if strings.Contains(s_name, "(") {
						pair := strings.SplitN(s_name, "(", 2)
						s_name = strings.TrimSpace(pair[0])
						s_status = strings.TrimSuffix(pair[1], ")")
					}

					if !strings.HasPrefix(series_link, schemes+"://"+host) {
						series_link = schemes + "://" + host + series_link
					}
					if strings.Contains(series_link, "#") {
						series_link = strings.SplitN(series_link, "#", 2)[0]
					}

					series := &Series{Name: s_name, Status: s_status, Url: series_link}
					// 拉取 车系配置详情链接
					settingsUrl, ok := fetchConfigUrl(series_link)
					if ok {
						series.Settings = settingsUrl
					}
					//TODO
					//拉取车系旗下车型名称

					serieses[s_name] = series
				}
			})
		})

		manuf.Series = serieses
		serieses = nil
		factors[mName] = manuf
		manuf = nil
	})
	brands[brand_name].Manufactures = factors
	factors = nil
	elapsed := time.Since(t_start)
	logger.Record("Brand crawl End:", brand_name, "[Runtime:", float64(elapsed.Nanoseconds())/1e6, "ms]")
	logger.Record(brands[brand_name], "[Runtime:", float64(elapsed.Nanoseconds())/1e6, "ms]")
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