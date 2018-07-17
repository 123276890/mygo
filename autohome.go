package main

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
	"path/filepath"
	"os"
	"io"

	"github.com/PuerkitoBio/goquery"
	"github.com/astaxie/beego/orm"
)

var (
	brands = make(AutoHomeBrands)
	logger = initLogManager("./log.txt")
)

func JobGetAutoHomeBrands() {
	excepts, ok := config["crawler"]["excepts"]
	if !ok {
		return
	}

	var err error
	o := orm.NewOrm()
	o.Using("default")
	/*nums, err := o.QueryTable("lrlz_brand").All(&db_brands)
	if err != nil {
		logger.Record("DB Query Error:",err)
		return
	}
	fmt.Println("Query result nums:",nums)*/

	brand_index_url := "https://car.autohome.com.cn/AsLeftMenu/As_LeftListNew.ashx?typeId=1 &brandId=0 &fctId=0 &seriesId=0"
	getAutoHomeBrands(brand_index_url)

	if !brands.IsEmpty() {
		fmt.Println(brands.Count())
		for _, b := range brands {
			if strings.Contains(excepts, b.Name) {
				continue
			}
			brand := Brand{Brand_name:b.Name}
			err = o.Read(&brand,"brand_name")

			if err != nil {
				// 数据库未找到该品牌
				logger.Record("No such brand in DB:",b.Name)
				//下载并保存品牌logo
				savepath, err := downloadBrandLogo(b)
				if err != nil {
					logger.Record("Error when downloading",b.Name,"logo: ",err)
				}
				brand.Brand_initial = b.Cap
				brand.Brand_logo = savepath

				brand_id, err := o.Insert(brand)
				if err != nil {
					logger.Record("Error when insert into DB:",err)
				}
				logger.Record("New brand insert into DB:",brand_id)

			} else {
				logger.Record(brand.Brand_name,"found in DB",brand)
			}

			//getAutoHomeBrand(v.Url, v.Name)
			//time.Sleep(time.Millisecond * 1000)
		}
	}
}

func downloadBrandLogo(b *AutoHomeBrand) (string, error) {
	logo_save_path := "/shop/brand/logo"
	logo_save_path = filepath.Join(SHOPNC_ROOT,logo_save_path)

	response, err := http.Get(b.Img)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	if !checkFileExist(logo_save_path) {
		os.Mkdir(logo_save_path, 0755)
	}
	//TODO 将品牌中文转换为英文缩写
	filename := ""
	dst, err := os.Create(logo_save_path + filename)
	if err != nil {
		return "", err
	}
	io.Copy(dst, response.Body)

	return logo_save_path, nil
}

func fetchSeriesInfo(sUrl string, series_name string, brand_name string) (map[string]interface{}, bool) {
	var charset string
	var schemes string
	var host string
	var err error

	found := false
	ret := make(map[string]interface{})

	resp, err := http.Get(sUrl)
	defer resp.Body.Close()
	if err != nil {
		logger.Record("Error: goqueryGet http.Get:", err)
		return nil, false
	}

	if resp.StatusCode != 200 {
		return nil, false
	}

	if content_type, ok := resp.Header["Content-Type"]; ok {
		pair := strings.SplitN(content_type[0], "=", 2)
		charset = pair[1]
		pair = nil
	}

	u, err := url.Parse(sUrl)
	if err != nil {
		logger.Record("Error: goqueryGet ParseUrl:", err)
		return nil, false
	}
	schemes = u.Scheme
	host = u.Host

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		logger.Record("Error: goqueryGet Err:", err)
		return nil, false
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

			ret["url"] = link
			found = true
		}
	}

	cars := make(map[string]*Car)
	carNodes := doc.Find(".content").Find("#divSeries").Find(".interval01-list-cars-infor")
	carNodes.Each(func(i int, div *goquery.Selection) {
		var carId int
		carId_str, ok	:= div.Find("p").Eq(0).Attr("id")
		if ok {
			if strings.HasPrefix(carId_str,"p") {
				carId_str = carId_str[1:]
			}
			carId, err = strconv.Atoi(carId_str)
			if err != nil {
				logger.Record("Error when convert carid string to int, carid string=",carId_str)
			}
		}
		carName := div.Find("a").Text()
		carName = ChineseToUtf(strings.TrimSpace(carName), charset)

		if !strings.HasPrefix(carName, series_name) {
			carName = series_name + " " + carName
		}

		if !strings.HasPrefix(carName, brand_name) {
			carName = brand_name + " " + carName
		}

		carUrl, ok := div.Find("a").Attr("href")
		if ok {
			if !strings.HasPrefix(carUrl, schemes + ":") {
				carUrl = schemes + ":" + carUrl
			}

			if strings.Contains(carUrl, "#") {
				carUrl = strings.SplitN(carUrl, "#", 2)[0]
			}
		}

		car := &Car{Aid:carId,Name:carName, Url:carUrl}
		cars[carName] = car
	})
	ret["cars"] = cars
	return ret, found
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

	logger.Record("AutoHomeBrand crawl Start:", brand_name)
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
		manuf := NewManufacture()

		mHref, ok := dt.Find("a").Attr("href")
		if ok {
			if !strings.HasPrefix(mHref, schemes+"://"+host) {
				mHref = schemes + "://" + host + mHref
				pair := strings.SplitN(mHref, "#", 2)
				mHref = pair[0]
				manuf.SetUrl(mHref)
			}
		}
		mName := dt.Find("a").Text()
		if mName != "" {
			mName = ChineseToUtf(mName, charset)
			mName = strings.TrimSpace(mName)
			manuf.SetName(mName)
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

					series := NewSeries(s_name, s_status, series_link)
					// 拉取 车系配置详情链接
					seriesInfo, ok := fetchSeriesInfo(series_link, s_name, brand_name)
					if ok {
						series.SetSettings(seriesInfo["url"].(string))
						series.SetCars(seriesInfo["cars"].(map[string]*Car))
					}

					serieses[s_name] = series
				}
			})
		})

		manuf.SetSeries(serieses)
		serieses = nil
		factors[mName] = manuf
		manuf = nil
	})
	brands[brand_name].Manufactures = factors
	factors = nil
	elapsed := time.Since(t_start)
	logger.Record("AutoHomeBrand crawl End:", brand_name, "[Runtime:", float64(elapsed.Nanoseconds())*1e-6, "ms]")
	logger.Record(brands[brand_name], "[Runtime:", float64(elapsed.Nanoseconds())*1e-6, "ms]")
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
						brands[brand_name] = NewAutoHomeBrand(brand_name, link, l)

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