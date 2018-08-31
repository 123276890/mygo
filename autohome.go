package main

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
	"path/filepath"
	"os"
	"io/ioutil"
	"errors"
	"regexp"
	"reflect"

	"github.com/PuerkitoBio/goquery"
	"github.com/astaxie/beego/orm"
	"github.com/tidwall/gjson"
)

func JobUpdateBrandLogo(brand_names []string) {
	var err error
	if len(brand_names) == 0 {
		return
	}
	brand_index_url := "https://car.autohome.com.cn/AsLeftMenu/As_LeftListNew.ashx?typeId=1 &brandId=0 &fctId=0 &seriesId=0"
	if brands.IsEmpty() {
		getAutoHomeBrands(brand_index_url)
	}

	for _, brand_name := range brand_names {
		if brand, ok := brands[brand_name]; ok {
			o := orm.NewOrm()
			o.Using("default")

			updateAutohomeBrandLogo(brand)

			brand_query := &Brand{Brand_name: brand.GetName()}
			err = o.Read(brand_query, "brand_name")

			if err == nil {
				brand_logo := brand_query.Brand_logo
				if brand_logo == "" {
					brand_query.Brand_logo, err = downloadBrandLogo(brand)

					if err != nil {
						logger.Record("Error When Downloading Brand Logo:", err)
						return
					}

					_, err := o.Update(brand_query, "brand_logo")
					if err != nil {
						logger.Record("Error When Updating Brand Logo to DB:", brand.GetName(), err)
					} else {
						logger.Record("Updating Brand Logo to DB Success!")
					}
				}
			}
		}
	}
}

func JobGetAutoHomeBrands() {
	excepts, ok := config["crawler"]["excepts"]
	if !ok {
		return
	}

	var err error
	o := orm.NewOrm()
	o.Using("default")

	brand_index_url := "https://car.autohome.com.cn/AsLeftMenu/As_LeftListNew.ashx?typeId=1 &brandId=0 &fctId=0 &seriesId=0"
	getAutoHomeBrands(brand_index_url)

	if !brands.IsEmpty() {
		for _, brand := range brands {
			if strings.Contains(excepts, brand.GetName()) {
				continue
			}

			getAutoHomeBrand(brand)
			//time.Sleep(time.Millisecond * 1000)

			// update 品牌
			brand_query := &Brand{Brand_name: brand.GetName()}
			err = o.Read(brand_query,"brand_name")

			if err != nil {
				// 数据库未找到该品牌
				//logger.Record("No such brand in DB:",brand.Name)
				//下载并保存品牌logo
				logo_file_path, err := downloadBrandLogo(brand)
				if err != nil {
					logger.Record("Error when downloading", brand.GetName(),"logo: ",err)
				}
				brand_query.Brand_initial = brand.Cap
				brand_query.Brand_logo = logo_file_path

				insert_id, err := o.Insert(brand_query)
				if err != nil {
					logger.Record("Error when insert a new Brand:",err)
				}
				brand_query.Brand_id = int(insert_id)
				logger.Record("New brand insert into DB Success:",insert_id)
			} else {
				logger.Record(brand_query.Brand_name,"found in DB",brand_query)
			}

			// update 车系
			ss := brand.getSeries()
			for _, s := range ss {
				if s.GetStatus() == "停产" || s.GetStatus() == "停售" {
					continue
				}
				series_query := &CarSeries{Series_name:s.GetName(),Brand_id:brand_query.Brand_id}
				err = o.Read(series_query, "series_name","brand_id")

				if err != nil {
					// 数据库未查到该车系
					series_query.Series_id = s.GetSid()
					sid, err := o.Insert(series_query)
					if err != nil {
						logger.Record("Error when insert a new CarSeries:", err)
						continue
					} else {
						logger.Record("New CarSeries insert into DB Success:",sid)
					}
				}
				//更新车系主页 和 车系参数主页
				if series_query.Series_home == "" || series_query.Series_config == "" {
					/*if series_query.Series_id == 370 {
						fmt.Println(s.GetStatus())
					}*/
					series_query.Series_home = s.GetHomeUrl()
					series_query.Series_config = s.GetConfigUrl()
					series_query.Status = s.GetStatus()
					num, err := o.Update(series_query,"series_home","series_config","status")
					if err != nil {
						logger.Record("Error when update CarSeries home page:",err)
					} else {
						logger.Record("Update CarSeries home page success",num)
					}
				}

				for _, car_crawl := range s.CarCrawls {
					exist := o.QueryTable(car_crawl).Filter("type_id", car_crawl.Type_id).Exist()
					if !exist {
						// 补齐车型表 品牌id
						car_crawl.SetBrandId(brand_query.Brand_id)
						car_id, err := o.Insert(car_crawl)
						if err != nil {
							logger.Record("Error when insert a new CarCrawl:",err,car_crawl.Car_name)
						} else {
							logger.Record("New CarCrawl insert into DB Success, car_id =",car_id)
						}
					}
				}// end of for range s.CarCrawls
			} // end for series
		}// end for brands
	}
}

func updateAutohomeBrandLogo(brand *AutoHomeBrand) {
	var (
		schemes		string
	)
	resp, err := xget(brand.GetUrl())
	if err != nil {
		logger.Record("Error: goqueryGet http.Get:", err)
		return
	}

	if resp.StatusCode != 200 {
		return
	}
	defer resp.Body.Close()

	u, err := url.Parse(brand.GetUrl())
	if err != nil {
		logger.Record("Error: goqueryGet ParseUrl:", err)
		return
	}
	schemes = u.Scheme

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		logger.Record("Error: goqueryGet Err:", err)
		return
	}

	logger.Record("开始抓取品牌图标:", brand.GetName())

	contBox := doc.Find(".contentright").Find(".contbox")
	imgNode := contBox.Find(".carbrand").Find(".carbradn-pic").Find("img")
	brand_logo, ok := imgNode.Attr("src")
	if ok {
		if !strings.HasPrefix(brand_logo, schemes+":") {
			brand_logo = schemes + ":" + brand_logo
		}
		brand.SetLogo(brand_logo)
	}
	return
}

func downloadBrandLogo(b *AutoHomeBrand) (string, error) {
	// 数据库记录的logo路径
	return_path := "/shop/brand/logo"
	// logo图片保存的真实路径 SHOPNC_ROOT /data/upload/shop/brand/logo
	logo_save_path := "/data/upload/shop/brand/logo"
	logo_save_path = filepath.Join(SHOPNC_ROOT,logo_save_path)

	response, err := http.Get(b.Img)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	if !checkFileExist(logo_save_path) {
		os.Mkdir(logo_save_path, 0755)
	}
	// 将品牌中文转换为英文缩写
	pinyin := ""
	words_rune := []rune(b.GetName())
	for _, v := range words_rune {
		s := string(v)
		p, ok := PinyinMap[s]
		if ok {
			pinyin += string(p[0])
		}
	}
	if pinyin == "" {
		return "", errors.New("Error: Can not generate logo image's file name!")
	}

	var extension string
	src, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	}
	filetype := http.DetectContentType(src)
	switch filetype {
	case "image/jpeg": extension = ".jpg"
	case "image/png": extension = ".png"
	case "image/gif": extension = ".gif"
	default:
		return "", errors.New("Error: Not a image file")
	}

	filename := pinyin + extension

	// 如果已存在同名文件
	if checkFileExist(filepath.Join(logo_save_path,filename)) {
		filename = reNameSameFileName(filename, logo_save_path)
	}

	dst, err := os.Create( filepath.Join(logo_save_path,filename) )
	if err != nil {
		return "", err
	}
	dst.Write(src)
	return_path = filepath.Join(return_path,filename)
	logger.Record("Download and save logo file success:", filepath.Join(logo_save_path,filename))

	return return_path, nil
}

func fetchCarInfo(surl string) (ret map[int]*CarCrawl, err error) {
	var (
		charset string
		dict = make(map[string]map[int]string)
	)
	ret = make(map[int]*CarCrawl)

	resp, err := xget(surl)
	if err != nil {
		logger.Record("Error: goqueryGet http.Get:", err)
		return
	}

	if resp.StatusCode != 200 {
		err = errors.New("Network Error!")
		return
	}
	defer resp.Body.Close()

	if content_type, ok := resp.Header["Content-Type"]; ok {
		pair := strings.SplitN(content_type[0], "=", 2)
		charset = pair[1]
		pair = nil
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		logger.Record("Error: goqueryGet Err:", err)
		return
	}

	html, _ := doc.Html()
	doc = nil
	resp.Body.Close()
	html = ChineseToUtf(html, charset)

	pat := `<script>((?:.|\\n)*?)</script>`
	reg := regexp.MustCompile(pat)

	car_info_datas := reg.FindAllString(html, -1)
	var js_matches []string

	for _, strs := range car_info_datas {
		if strings.Index(strs, "try{document.") < 0 && len([]rune(strs)) > 500 {
			js_matches = append(js_matches, strs)
		}
	}

	for i, js := range js_matches {
		switch i {
		//case 0: dict["keylink"] = getAutoHomeDict(js)
		case 1: dict["config"] = getAutoHomeDict(js)
		case 2: dict["option"] = getAutoHomeDict(js)
		}
	}

	// 基本参数组
	pos_start := strings.Index(html, "var config =")
	if pos_start <= 0 {
		err = errors.New("Error when try to find inner color")
		return
	}
	str_base := html[pos_start:]
	pos_end := strings.IndexByte(str_base, '\n')
	if pos_end > len(str_base) {
		err = errors.New("Error when try to find inner color: position end is out of str length")
		return
	}
	str_base = str_base[13:pos_end - 1]

	// 选项配置参数组
	pos_start = strings.Index(html, "var option =")
	if pos_start <= 0 {
		err = errors.New("Error when try to find inner color")
		return
	}
	str_option := html[pos_start:]
	pos_end = strings.IndexByte(str_option, '\n')
	if pos_end > len(str_option) {
		err = errors.New("Error when try to find inner color: position end is out of str length")
		return
	}
	str_option = str_option[13:pos_end - 1]

	// 外观颜色json
	pos_start = strings.Index(html, "var color =")
	if pos_start <= 0 {
		err = errors.New("Error when try to find inner color")
		return
	}
	str_color := html[pos_start:]
	pos_end = strings.IndexByte(str_color, '\n')
	if pos_end > len(str_color) {
		err = errors.New("Error when try to find inner color: position end is out of str length")
		return
	}
	str_color = str_color[12:pos_end - 1]

	// 内饰颜色json
	pos_start = strings.Index(html, "var innerColor =")
	if pos_start <= 0 {
		err = errors.New("Error when try to find inner color")
		return
	}
	str_inner := html[pos_start:]
	pos_end = strings.IndexByte(str_inner, '\n')
	if pos_end > len(str_inner) {
		err = errors.New("Error when try to find inner color: position end is out of str length")
		return
	}
	str_inner = str_inner[16:pos_end - 1]

	//车型id抓取
	result := gjson.Get(str_base,"result.speclist")
	if result.Exists() {
		items := result.Array()
		for _, item := range items {
			key_car_id := item.Get("specid")
			if key_car_id.Exists() {
				car_id := int(key_car_id.Int())
				car := NewAutoHomeCar(car_id)
				ret[car_id] = car
			}
		}
	}
	//基本参数抓取
	result = gjson.Get(str_base, "result.paramtypeitems")

	if result.Exists() {
		items := result.Array()
		for _, item := range items {
			item_name := item.Get("name").String()

			if item_name == "基本参数" {
				base_params := item.Get("paramitems").Array()

				for _, v := range base_params {
					name := v.Get("name").String()
					//能源类型
					if strings.Contains(name, "能源类型") {
						v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
							dealAutoHomeItemValue(value, ret,"Energy_type_str",dict)
							return true
						})
					}
					//上市时间
					if strings.Contains(name, "上市") {
						v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
							dealAutoHomeItemValue(value, ret,"Market_time",dict)
							return true
						})
					}
					//工信部纯电续驶里程
					if strings.Contains(name, "工信部纯电续驶里程") {
						v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
							dealAutoHomeItemValue(value, ret,"E_mileage",dict)
							return true
						})
					}
					//变速箱
					if strings.Contains(name, "变速箱") {
						v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
							dealAutoHomeItemValue(value, ret,"Gearbox",dict)
							return true
						})
					}

					id := v.Get("id").Int()
					switch id {
					case 220:	//级别
						v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
							dealAutoHomeItemValue(value, ret,"Car_level_str",dict)
							return true
						})
					case 567:	//车型名称
						v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
							dealAutoHomeItemValue(value, ret,"Car_name",dict)
							return true
						})
					case 295:	//最大功率
						v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
							dealAutoHomeItemValue(value, ret,"Max_power",dict)
							return true
						})
					case 571:	//最大扭矩
						v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
							dealAutoHomeItemValue(value, ret,"Max_torque",dict)
							return true
						})
					case 555:	//发动机
						v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
							dealAutoHomeItemValue(value, ret,"Engine",dict)
							return true
						})
					case 222:	//长*宽*高
						v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
							dealAutoHomeItemValue(value, ret,"Car_size",dict)
							return true
						})
					case 281:	//车身结构
						v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
							dealAutoHomeItemValue(value, ret,"Car_struct",dict)
							return true
						})
					case 267:	//最高车速
						v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
							dealAutoHomeItemValue(value, ret,"Max_speed",dict)
							return true
						})
					case 225:	//官方100加速
						v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
							dealAutoHomeItemValue(value, ret,"Official_speedup",dict)
							return true
						})
					case 272:	//实测100加速
						v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
							dealAutoHomeItemValue(value, ret,"Actual_speedup",dict)
							return true
						})
					case 273:	//实测100制动
						v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
							dealAutoHomeItemValue(value, ret,"Actual_brake",dict)
							return true
						})
					case 271:	//工信部综合油耗
						v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
							dealAutoHomeItemValue(value, ret,"Gerenal_fueluse",dict)
							return true
						})
					case 243:	//实测油耗
						v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
							dealAutoHomeItemValue(value, ret,"Actual_fueluse",dict)
							return true
						})
					case 274:	//整车质保
						v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
							dealAutoHomeItemValue(value, ret,"Quality_guarantee",dict)
							return true
						})
					} // end of switch id
				}// end for base_params
			} // end of 基本参数

			if item_name == "车身" {
				body_params := item.Get("paramitems").Array()

				for _, v := range body_params {
					id := v.Get("id").Int()

					switch id {
					case 275:	//长度(mm)
						v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
							dealAutoHomeItemValue(value, ret,"Length",dict)
							return true
						})
					case 276:	//宽度(mm)
						v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
							dealAutoHomeItemValue(value, ret,"Width",dict)
							return true
						})
					case 277:	//高度(mm)
						v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
							dealAutoHomeItemValue(value, ret,"Height",dict)
							return true
						})
					case 132:	//轴距(mm)
						v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
							dealAutoHomeItemValue(value, ret,"Shaft_distance",dict)
							return true
						})
					case 278:	//前轮距(mm)
						v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
							dealAutoHomeItemValue(value, ret,"Front_wheels_gap",dict)
							return true
						})
					case 638:	//后轮距(mm)
						v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
							dealAutoHomeItemValue(value, ret,"Back_wheels_gap",dict)
							return true
						})
					case 279:	//最小离地间隙(mm)
						v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
							dealAutoHomeItemValue(value, ret,"Min_ground",dict)
							return true
						})
					case 280:	//整备质量(kg)
						v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
							dealAutoHomeItemValue(value, ret,"Total_weight",dict)
							return true
						})
					case 281:	//车身结构
						v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
							dealAutoHomeItemValue(value, ret,"Body_struct",dict)
							return true
						})
					case 282:	//车门数
						v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
							dealAutoHomeItemValue(value, ret,"Doors",dict)
							return true
						})
					case 283:	//座位数
						v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
							dealAutoHomeItemValue(value, ret,"Seats",dict)
							return true
						})
					case 284:	//油箱容积(L)
						v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
							dealAutoHomeItemValue(value, ret,"Fuel_vol",dict)
							return true
						})
					case 285:	//行李厢容积(L)
						v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
							dealAutoHomeItemValue(value, ret,"Cargo_vol",dict)
							return true
						})
					}// end of switch
				}// end for body_params
			} // end of 车身

			if item_name == "发动机" {
				engine_params := item.Get("paramitems").Array()

				for _, v := range engine_params {
					id := v.Get("id").Int()

					switch id {
					case 570:	//发动机型号
						v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
							dealAutoHomeItemValue(value, ret,"Engine_type",dict)
							return true
						})
					case 287:	//排量(mL)
						v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
							dealAutoHomeItemValue(value, ret,"Cc",dict)
							return true
						})
					case 640:	//进气形式
						v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
							dealAutoHomeItemValue(value, ret,"Air_intake",dict)
							return true
						})
					case 289:	//气缸排列形式
						v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
							dealAutoHomeItemValue(value, ret,"Cylinder_arrange",dict)
							return true
						})
					case 290:	//气缸个数
						v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
							dealAutoHomeItemValue(value, ret,"Cylinders",dict)
							return true
						})
					case 291:	//每缸气门数
						v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
							dealAutoHomeItemValue(value, ret,"Valves",dict)
							return true
						})
					case 182:	//压缩比
						v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
							dealAutoHomeItemValue(value, ret,"Compress_rate",dict)
							return true
						})
					case 641:	//配气机构
						v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
							dealAutoHomeItemValue(value, ret,"Valve_machanism",dict)
							return true
						})
					case 181:	//缸径(mm)
						v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
							dealAutoHomeItemValue(value, ret,"Cylinder_radius",dict)
							return true
						})
					case 293:	//行程(mm)
						v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
							dealAutoHomeItemValue(value, ret,"Stroke",dict)
							return true
						})
					case 294:	//最大马力(Ps)
						v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
							dealAutoHomeItemValue(value, ret,"Engine_hp",dict)
							return true
						})
					case 295:	//最大功率(kW)
						v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
							dealAutoHomeItemValue(value, ret,"Engine_power",dict)
							return true
						})
					case 296:	//最大功率转速(rpm)
						v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
							dealAutoHomeItemValue(value, ret,"Engine_rpm",dict)
							return true
						})
					case 571:	//最大扭矩(N·m)
						v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
							dealAutoHomeItemValue(value, ret,"Engine_torque",dict)
							return true
						})
					case 642:	//最大扭矩转速(rpm)
						v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
							dealAutoHomeItemValue(value, ret,"Torque_rpm",dict)
							return true
						})
					case 643:	//发动机特有技术
						v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
							dealAutoHomeItemValue(value, ret,"Tech_spec",dict)
							return true
						})
					case 572:	//燃料形式
						v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
							dealAutoHomeItemValue(value, ret,"Engine_energy",dict)
							return true
						})
					case 573:	//燃油标号
						v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
							dealAutoHomeItemValue(value, ret,"Roz",dict)
							return true
						})
					case 574:	//供油方式
						v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
							dealAutoHomeItemValue(value, ret,"Oil_drive",dict)
							return true
						})
					case 575:	//缸盖材料
						v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
							dealAutoHomeItemValue(value, ret,"Cylinder_cover",dict)
							return true
						})
					case 576:	//缸体材料
						v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
							dealAutoHomeItemValue(value, ret,"Cylinder_body",dict)
							return true
						})
					case 577:	//环保标准
						v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
							dealAutoHomeItemValue(value, ret,"Environmental_standard",dict)
							return true
						})
					}
				}// end for engine_params
			} // end of 发动机

			if item_name == "电动机" {
				//info["base"]["car_type"] = "1"	// 新能源车
				motor_params := item.Get("paramitems").Array()

				for _, v := range motor_params {
					name := v.Get("name").String()

					if name == "电机类型" {
						v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
							dealAutoHomeItemValue(value, ret,"Motor_type",dict)
							dealAutoHomeItemValue(value, ret,"Car_type",dict)	//标记为新能源车
							return true
						})
					}

					if name == "驱动电机数" {
						v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
							dealAutoHomeItemValue(value, ret,"Motor_num",dict)
							return true
						})
					}

					if name == "电机布局" {
						v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
							dealAutoHomeItemValue(value, ret,"Motor_arrange",dict)
							return true
						})
					}

					id := v.Get("id").Int()

					switch id {
					case 1325:	//电动机总功率
						v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
							dealAutoHomeItemValue(value, ret,"Motor_power",dict)
							return true
						})
					case 1326:	//电动机总扭矩
						v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
							dealAutoHomeItemValue(value, ret,"Engine_type",dict)
							return true
						})
					case 1327:	//前电动机最大功率
						v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
							dealAutoHomeItemValue(value, ret,"Motor_front_power",dict)
							return true
						})
					case 1328:	//前电动机最大扭矩
						v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
							dealAutoHomeItemValue(value, ret,"Motor_front_torque",dict)
							return true
						})
					case 1329:	//后电动机最大功率
						v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
							dealAutoHomeItemValue(value, ret,"Motor_back_power",dict)
							return true
						})
					case 1330:	//后电动机最大扭矩
						v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
							dealAutoHomeItemValue(value, ret,"Motor_back_torque",dict)
							return true
						})
					case 1013:	//工信部续航里程
						v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
							dealAutoHomeItemValue(value, ret,"Mileage",dict)
							return true
						})
					case 1124:	//电池容量
						v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
							dealAutoHomeItemValue(value, ret,"Bat_cap",dict)
							return true
						})
					}// end switch
				}// end for motor_params
			} // end of 电动机

			if item_name == "变速箱" {
				gearboxes := item.Get("paramitems").Array()

				for _, v := range gearboxes {
					id := v.Get("id").Int()

					switch id {
					case 559:	//挡位个数
						v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
							dealAutoHomeItemValue(value, ret,"Gears_num",dict)
							return true
						})
					case 221:	//变速箱类型
						v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
							dealAutoHomeItemValue(value, ret,"Gears_type",dict)
							return true
						})
					case 1072:	//简称
						v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
							dealAutoHomeItemValue(value, ret,"Gearbox_name",dict)
							return true
						})
					}// end switch
				}// end for gearboxes
			} // end of 变速箱

			if item_name == "底盘转向" {
				underpan := item.Get("paramitems").Array()
				//无 中央差速器结构字段
				for _, v := range underpan {
					id := v.Get("id").Int()

					switch id {
					case 395:	//驱动方式
						v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
							dealAutoHomeItemValue(value, ret,"Drive_method",dict)
							return true
						})
					case 578:	//前悬架类型
						v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
							dealAutoHomeItemValue(value, ret,"Susp_front_type",dict)
							return true
						})
					case 579:	//后悬架类型
						v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
							dealAutoHomeItemValue(value, ret,"Susp_back_type",dict)
							return true
						})
					case 510:	//助力类型
						v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
							dealAutoHomeItemValue(value, ret,"Assist_type",dict)
							return true
						})
					case 223:	//车体结构
						v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
							dealAutoHomeItemValue(value, ret,"Structure",dict)
							return true
						})
					case 412:	//四驱形式
						v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
							dealAutoHomeItemValue(value, ret,"Four_wheel_drive",dict)
							return true
						})
					case 415:	//中央差速器结构
						v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
							dealAutoHomeItemValue(value, ret,"Central_diff",dict)
							return true
						})
					}// end switch
				}// end for underpan
			} // end of 底盘转向

			if item_name == "车轮制动" {
				brake_params := item.Get("paramitems").Array()

				for _, v := range brake_params {
					id := v.Get("id").Int()

					switch id {
					case 511:	//前制动器类型
						v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
							dealAutoHomeItemValue(value, ret,"Front_brake",dict)
							return true
						})
					case 512:	//后制动器类型
						v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
							dealAutoHomeItemValue(value, ret,"Back_brake",dict)
							return true
						})
					case 513:	//驻车制动类型
						v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
							dealAutoHomeItemValue(value, ret,"Park_brake",dict)
							return true
						})
					case 580:	//前轮胎规格
						v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
							dealAutoHomeItemValue(value, ret,"Front_wheel_size",dict)
							return true
						})
					case 581:	//后轮胎规格
						v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
							dealAutoHomeItemValue(value, ret,"Back_wheel_size",dict)
							return true
						})
					case 515:	//备胎规格
						v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
							dealAutoHomeItemValue(value, ret,"Backup_wheel",dict)
							return true
						})
					}// end switch
				}// end for brake_params
			} // end of 车轮制动
		}
	}// end of Base result.Exists()

	//可选参数抓取
	result = gjson.Get(str_option, "result.configtypeitems")

	if result.Exists() {
		items := result.Array()

		for _, item := range items {
			item_name := item.Get("name").String()

			if item_name == "主/被动安全装备" {
				if item.Get("configitems").IsArray() {
					secure_params := item.Get("configitems").Array()
					// 疲劳驾驶提示id为0
					for _,v := range secure_params {
						id := v.Get("id").Int()

						switch id {
						case 1082:	//主/副驾驶座安全气囊
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"Seat_srs",dict)
								return true
							})
						case 421:	//前/后排侧气囊
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"Side_airbag",dict)
								return true
							})
						case 422:	//前/后排头部气囊(气帘)
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"Head_srs",dict)
								return true
							})
						case 423:	//膝部气囊
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"Knee_srs",dict)
								return true
							})
						case 551:	//胎压监测装置
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"Tire_pres_monitor",dict)
								return true
							})
						case 424:	//零胎压继续行驶
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"Zero_tire_pres",dict)
								return true
							})
						case 552:	//安全带未系提示
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"Unbelt_notice",dict)
								return true
							})
						case 1084:	//ISOFIX儿童座椅接口
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"Isofix",dict)
								return true
							})
						case 110:	//ABS防抱死
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"Anti_lock",dict)
								return true
							})
						case 125:	//制动力分配(EBD/CBC等)
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"Bfd",dict)
								return true
							})
						case 437:	//刹车辅助(EBA/BAS/BA等)
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"Bas",dict)
								return true
							})
						case 438:	//牵引力控制(ASR/TCS/TRC等)
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"Tcs",dict)
								return true
							})
						case 109:	//车身稳定控制(ESC/ESP/DSC等)
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"Stable_control",dict)
								return true
							})
						case 426:	//并线辅助
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"Bsa",dict)
								return true
							})
						case 788:	//车道偏离预警系统
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"Ldw",dict)
								return true
							})
						case 436:	//主动刹车/主动安全系统
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"Abs",dict)
								return true
							})
						case 637:	//夜视系统
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"Nvs",dict)
								return true
							})
						}
					} // end for range secure_params
				}
			} // end of 主/被动安全装备

			if item_name == "辅助/操控配置" {
				if item.Get("configitems").IsArray() {
					oper_params := item.Get("configitems").Array()
					//无法抓取: 电磁感应悬架
					for _,v := range oper_params {
						name := v.Get("name").String()
						if name == "上坡辅助" {
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"Hac",dict)
								return true
							})
						}

						id := v.Get("id").Int()

						switch id {
						case 1086:	//前/后驻车雷达
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"Radar",dict)
								return true
							})
						case 448:	//倒车视频影像
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"Reverse_video",dict)
								return true
							})
						case 473:	//全景摄像头
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"Panorama",dict)
								return true
							})
						case 445:	//定速巡航
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"Cruise_ctrl",dict)
								return true
							})
						case 446:	//自适应巡航
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"Self_adpt_cruise",dict)
								return true
							})
						case 472:	//自动泊车入位
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"Auto_park_in",dict)
								return true
							})
						case 334:	//发动机启停技术
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"Engine_start_stop",dict)
								return true
							})
						case 363:	//自动驻车
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"Auto_park",dict)
								return true
							})
						case 138:	//陡坡缓降
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"Hdc",dict)
								return true
							})
						case 399:	//可变悬架
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"Variable_susp",dict)
								return true
							})
						case 167:	//空气悬架
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"Air_susp",dict)
								return true
							})
						case 409:	//可变转向比
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"Vgrs",dict)
								return true
							})
						case 975:	//前桥限滑差速器/差速锁
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"Front_diff_lock",dict)
								return true
							})
						case 976:	//中央差速器锁止功能
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"Central_diff_lock",dict)
								return true
							})
						case 977:	//后桥限滑差速器/差速锁
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"Back_diff_lock",dict)
								return true
							})
						case 404:	//整体主动转向系统
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"Ads",dict)
								return true
							})
						}
					} // end for range oper_params
				}
			} // end of 辅助/操控配置

			if item_name == "外部/防盗配置" {
				if item.Get("configitems").IsArray() {
					guard_params := item.Get("configitems").Array()

					for _,v := range guard_params {
						name := v.Get("name").String()
						if name == "远程启动" {
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"Remote_start",dict)
								return true
							})
						}

						if name == "车顶行李架" {
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"Roof_rack",dict)
								return true
							})
						}

						if name == "感应后备厢" {
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"React_cargo",dict)
								return true
							})
						}

						id := v.Get("id").Int()

						switch id {
						case 583:	//电动天窗
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"E_sunroof",dict)
								return true
							})
						case 584:	//全景天窗
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"Pano_sunroof",dict)
								return true
							})
						case 585:	//运动外观套件
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"Sport_package",dict)
								return true
							})
						case 525:	//铝合金轮圈
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"Alloy_wheel",dict)
								return true
							})
						case 443:	//电动吸合门
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"E_suction_door",dict)
								return true
							})
						case 1122:	//侧滑门
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"Slide_door",dict)
								return true
							})
						case 452:	//电动后备厢
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"E_cargo",dict)
								return true
							})
						case 481:	//发动机电子防盗
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"Engine_e_guard",dict)
								return true
							})
						case 558:	//车内中控锁
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"E_ctrl_lock",dict)
								return true
							})
						case 582:	//遥控钥匙
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"Remote_key",dict)
								return true
							})
						case 431:	//无钥匙启动系统
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"Keyless_start",dict)
								return true
							})
						case 1066:	//无钥匙进入系统
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"Keyless_enter",dict)
								return true
							})
						}
					} // end for range guard_params
				}
			} // end of 外部/防盗配置

			if item_name == "内部配置" {
				if item.Get("configitems").IsArray() {
					inside_params := item.Get("configitems").Array()

					for _,v := range inside_params {
						name := v.Get("name").String()
						if name == "皮质方向盘" {
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"Leather_steering",dict)
								return true
							})
						}

						if name == "方向盘记忆" {
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"Steer_mem",dict)
								return true
							})
						}

						if name == "内置行车记录仪" {
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"Car_dvr",dict)
								return true
							})
						}

						if strings.Contains(name, "降噪") {
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"Anc",dict)
								return true
							})
						}

						if strings.Contains(name, "手机无线") {
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"Wireless_charge",dict)
								return true
							})
						}

						id := v.Get("id").Int()

						switch id {
						case 1085:	//方向盘调节
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"Steer_adjt",dict)
								return true
							})
						case 589:	//方向盘电动调节
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"Steer_e_adjt",dict)
								return true
							})
						case 444:	//多功能方向盘
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"Functional_steer",dict)
								return true
							})
						case 468:	//方向盘换挡
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"Steer_shift",dict)
								return true
							})
						case 1064:	//方向盘加热
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"Steer_heat",dict)
								return true
							})
						case 590:	//行车电脑显示屏
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"Computer_scr",dict)
								return true
							})
						case 471:	//HUD抬头数字显示
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"Hud",dict)
								return true
							})
						}
					} // end for range inside_params
				}
			} // end of 内部配置

			if item_name == "座椅配置" {
				if item.Get("configitems").IsArray() {
					seat_params := item.Get("configitems").Array()
					//无法抓取：副驾驶位后排可调节按钮、第二排独立座椅、可加热/制冷杯架
					for i,v := range seat_params {
						id := v.Get("id").Int()

						//座椅材质
						if i == 0 {
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"Seat_mat",dict)
								return true
							})
						}

						switch id {
						case 592:	//运动风格座椅
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"Sport_seat",dict)
								return true
							})
						case 639:	//座椅高低调节
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"Height_adjt",dict)
								return true
							})
						case 449:	//腰部支撑调节
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"Lumbar_support",dict)
								return true
							})
						case 593:	//肩部支撑调节
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"Shoulder_support",dict)
								return true
							})
						case 1087:	//主/副驾驶座电动调节
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"Seat_e_adjt",dict)
								return true
							})
						case 595:	//第二排靠背角度调节
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"Snd_backrest_adjt",dict)
								return true
							})
						case 596:	//第二排座椅移动
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"Snd_seat_mv",dict)
								return true
							})
						case 597:	//后排座椅电动调节
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"Back_seat_adjt",dict)
								return true
							})
						case 598:	//电动座椅记忆
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"E_seat_mem",dict)
								return true
							})
						case 1088:	//前/后排座椅加热
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"Seat_heat",dict)
								return true
							})
						case 1089:	//前/后排座椅通风
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"Seat_vent",dict)
								return true
							})
						case 1090:	//前/后排座椅按摩
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"Seat_masg",dict)
								return true
							})
						case 603:	//第三排座椅
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"Third_row_seat",dict)
								return true
							})
						case 1091:	//后排座椅放倒方式
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"Back_seat_down",dict)
								return true
							})
						case 1092:	//前/后中央扶手
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"Handrail",dict)
								return true
							})
						case 606:	//后排杯架
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"Back_cup_hold",dict)
								return true
							})
						}
					} // end for range seat_params
				}
			} // end of 座椅配置

			if item_name == "多媒体配置" {
				if item.Get("configitems").IsArray() {
					media_params := item.Get("configitems").Array()
					//无法抓取：外接音源接口
					for _,v := range media_params {
						name := v.Get("name").String()
						if name == "中控台彩色大屏尺寸" {
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"Colorful_scr_size",dict)
								return true
							})
						}

						if name == "手机互联/映射" {
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"Mobile_map",dict)
								return true
							})
						}

						if name == "车联网" {
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"Network",dict)
								return true
							})
						}

						if name == "220V/230V电源" {
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"Back_power_supply",dict)
								return true
							})
						}

						if name == "CD/DVD" {
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"Cddvd",dict)
								return true
							})
						}

						id := v.Get("id").Int()

						switch id {
						case 607:	//GPS导航系统
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"Gps",dict)
								return true
							})
						case 455:	//定位互动服务
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"Gps_interact",dict)
								return true
							})
						case 608:	//中控台彩色大屏
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"Colorful_scr",dict)
								return true
							})
						case 464:	//中控液晶屏分屏显示
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"Lcd_sep",dict)
								return true
							})
						case 609:	//蓝牙/车载电话
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"Blueteeth",dict)
								return true
							})
						case 610:	//车载电视
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"Television",dict)
								return true
							})
						case 611:	//后排液晶屏
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"Back_lcd",dict)
								return true
							})
						case 1212:	//扬声器品牌
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"Speaker_brand",dict)
								return true
							})
						case 618:	//扬声器数量
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"Speaker_num",dict)
								return true
							})
						}
					} // end for range media_params
				}
			} // end of 多媒体配置

			if item_name == "灯光配置" {
				if item.Get("configitems").IsArray() {
					light_params := item.Get("configitems").Array()

					for i, v := range light_params {
						//近光灯
						if i == 0 {
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"Low_beam",dict)
								return true
							})
						}
						//远光灯
						if i == 1 {
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"High_beam",dict)
								return true
							})
						}

						name := v.Get("name").String()
						if name == "LED日间行车灯" {
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"Led_beam",dict)
								return true
							})
						}

						if name == "转向头灯" {
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"Turn_head_light",dict)
								return true
							})
						}

						if strings.Contains(name, "应远近光") {
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"Adaptive_beam",dict)
								return true
							})
						}

						id := v.Get("id").Int()

						switch id {
						case 441:	//自动头灯
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"Head_light",dict)
								return true
							})
						case 1161:	//转向辅助灯
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"Turn_light",dict)
								return true
							})
						case 619:	//前雾灯
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"Front_fog_lamp",dict)
								return true
							})
						case 620:	//大灯高度可调
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"Light_height_adjt",dict)
								return true
							})
						case 621:	//大灯清洗装置
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"Light_clean_dev",dict)
								return true
							})
						case 453:	//车内氛围灯
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"Mood_light",dict)
								return true
							})
						}
					} // end for range light_params
				}
			} // end of 灯光配置

			if item_name == "玻璃/后视镜" {
				if item.Get("configitems").IsArray() {
					glass_params := item.Get("configitems").Array()

					for _,v := range glass_params {
						name := v.Get("name").String()
						if strings.Contains(name, "车窗一键") {
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"E_lift_window",dict)
								return true
							})
						}

						if name == "流媒体车内后视镜" {
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"Stream_media_rearview",dict)
								return true
							})
						}

						id := v.Get("id").Int()

						switch id {
						case 622:	//前/后电动车窗
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"Power_window",dict)
								return true
							})
						case 623:	//车窗防夹手功能
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"Anti_pinch_hand",dict)
								return true
							})
						case 624:	//防紫外线/隔热玻璃
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"Insulating_glass",dict)
								return true
							})
						case 625:	//后视镜电动调节
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"E_adjt_rearview",dict)
								return true
							})
						case 626:	//后视镜加热
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"Heat_rearview",dict)
								return true
							})
						case 1095:	//内/外后视镜自动防眩目
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"Dimming_mirror",dict)
								return true
							})
						case 628:	//后视镜电动折叠
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"Power_mirror",dict)
								return true
							})
						case 629:	//后视镜记忆
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"Mirror_mem",dict)
								return true
							})
						case 630:	//后风挡遮阳帘
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"Abat_vent",dict)
								return true
							})
						case 631:	//后排侧遮阳帘
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"Side_abat_vent",dict)
								return true
							})
						case 1063:	//后排侧隐私玻璃
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"Side_priv_glass",dict)
								return true
							})
						case 632:	//遮阳板化妆镜
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"Sun_shield",dict)
								return true
							})
						case 633:	//后雨刷
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"Back_wiper",dict)
								return true
							})
						case 454:	//感应雨刷
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"React_wiper",dict)
								return true
							})
						}
					} // end for range glass_params
				}
			} // end of 玻璃/后视镜

			if item_name == "空调/冰箱" {
				if item.Get("configitems").IsArray() {
					air_params := item.Get("configitems").Array()

					for _,v := range air_params {
						name := v.Get("name").String()
						if strings.Contains(name, "净化器") {
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"Air_cleaner",dict)
								return true
							})
						}

						id := v.Get("id").Int()

						switch id {
						case 1097:	//空调控制方式
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"Air_type",dict)
								return true
							})
						case 459:	//后排独立空调
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"Back_air",dict)
								return true
							})
						case 634:	//后座出风口
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"Back_outlet",dict)
								return true
							})
						case 463:	//温度分区控制
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"Temper_zone_ctrl",dict)
								return true
							})
						case 635:	//车内空气调节/花粉过滤
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"Air_adjt",dict)
								return true
							})
						case 636:	//车载冰箱
							v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
								dealAutoHomeItemValue(value, ret,"Car_fridge",dict)
								return true
							})
						}
					} // end for range air_params
				}
			} // end of 空调/冰箱
		}// end for range items
	}// end of Options result.Exists()
	return
}

func fetchSeriesInfo(series *Series) (map[string]interface{}, bool) {
	var (
		charset string
		schemes string
		host string
		err error
	)

	found := false
	ret := make(map[string]interface{})

	resp, err := xget(series.GetHomeUrl())
	if err != nil {
		logger.Record("Error: goqueryGet http.Get:", err)
		return nil, false
	}

	if resp.StatusCode != 200 {
		return nil, false
	}
	defer resp.Body.Close()

	if content_type, ok := resp.Header["Content-Type"]; ok {
		pair := strings.SplitN(content_type[0], "=", 2)
		charset = pair[1]
		pair = nil
	}

	u, err := url.Parse(series.GetHomeUrl())
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

	//cars := make(map[string]*CarCrawl)
	var car_crawls map[int]*CarCrawl

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

			// 抓取车辆配置信息
			ret["url"] = link
			car_crawls, err = fetchCarInfo(link)
			if err != nil {
				logger.Record("Error when fetch car info:",err)
			}
			found = true
		}
	}

	if found == false {
		return nil, found
	}

	// 抓取厂商指导价
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

		if car_crawl, ok := car_crawls[carId]; ok {
			pa := div.Parent().Next().Next()
			if pa.HasClass("interval01-list-guidance") {
				price := pa.Find("div").Text()
				price = strings.TrimSpace(price)
				if charset != "utf8" {
					price = ChineseToUtf(price, charset)
				}
				car_crawl.SetPriceStr(price)
			}

			if series.status == "进口" {
				car_crawl.Produce_type_str = "进口"
			} else {
				car_crawl.Produce_type_str = "国产"
			}
		}
	})
	//设置车系id，车系名称，品牌名称，厂商
	for _, car := range car_crawls {
		car.SetSeriesId(series.GetSid()).
			SetSeriesName(series.GetName()).
			SetBrandName(series.GetBrandName()).
			SetManufacturer(series.GetManufactureName())
	}
	ret["cars"] = car_crawls
	return ret, found
}

func getAutoHomeBrand(brand *AutoHomeBrand) {
	var charset string
	var schemes string
	var host string

	resp, err := xget(brand.GetUrl())
	if err != nil {
		logger.Record("Error: goqueryGet http.Get:", err)
		return
	}

	if resp.StatusCode != 200 {
		return
	}

	defer resp.Body.Close()

	if content_type, ok := resp.Header["Content-Type"]; ok {
		pair := strings.SplitN(content_type[0], "=", 2)
		charset = pair[1]
		pair = nil
	}

	u, err := url.Parse(brand.GetUrl())
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

	brand_name := brand.GetName()
	logger.Record("AutoHomeBrand crawl Start:", brand_name)
	t_start := time.Now()

	contBox := doc.Find(".contentright").Find(".contbox")
	imgNode := contBox.Find(".carbrand").Find(".carbradn-pic").Find("img")
	brand_logo, ok := imgNode.Attr("src")
	if ok {
		if !strings.HasPrefix(brand_logo, schemes+":") {
			brand_logo = schemes + ":" + brand_logo
		}
		brands[brand_name].SetLogo(brand_logo)
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
				s_name := ChineseToUtf(slink.Text(),charset)
				if ok && !strings.Contains(s_name, "停产") && !strings.Contains(s_name, "停售") {
					var s_status = ""

					s_name = strings.TrimSpace(s_name)
					if strings.Contains(s_name, "(") {
						if !strings.Contains(s_name, "进口") {
							pair := strings.SplitN(s_name, "(", 2)
							s_name = strings.TrimSpace(pair[0])
							s_status = strings.TrimSuffix(pair[1], ")")
						}
					}

					if strings.Contains(s_name, "进口") {
						s_status = "进口"
					}

					//抓取车系ID
					pos_series := strings.Index(series_link, "series-")
					pos_dot := strings.Index(series_link,".")
					sid_text := series_link[pos_series+7:pos_dot]
					if strings.Contains(sid_text, "-") {
						pos := strings.Index(sid_text, "-")
						sid_text = sid_text[:pos]
					}
					sid, err := strconv.Atoi(sid_text)
					if err != nil {
						logger.Record("Error when get Series ID:", s_name, err)
						return
					}

					if !strings.HasPrefix(series_link, schemes+"://"+host) {
						series_link = schemes + "://" + host + series_link
					}
					if strings.Contains(series_link, "#") {
						series_link = strings.SplitN(series_link, "#", 2)[0]
					}

					series := NewSeries(sid, s_name, s_status, series_link)
					series.SetBrand(brand).SetManufacture(manuf)
					//series.SetBrandName(brand_name)
					//series.Manufacture = manuf
					// 拉取 车系配置详情链接
					seriesInfo, ok := fetchSeriesInfo(series)
					if ok {
						series.SetSettings(seriesInfo["url"].(string))
						series.SetCars(seriesInfo["cars"].(map[int]*CarCrawl))
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
	brands[brand_name].SetManufs(factors)
	factors = nil
	elapsed := time.Since(t_start)
	logger.Record("AutoHomeBrand crawl End:", brand_name, "[Runtime:", float64(elapsed.Nanoseconds())*1e-6, "ms]")
	logger.Record(brands[brand_name], "[Runtime:", float64(elapsed.Nanoseconds())*1e-6, "ms]")
}

func getAutoHomeBrands(sUrl string) {
	var charset string
	var schemes string
	var host string

	resp, err := xget(sUrl)
	if err != nil {
		logger.Record("Error: goqueryGet http.Get:", err)
		return
	}

	if resp.StatusCode != 200 {
		return
	}
	defer resp.Body.Close()

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
		capital := s.Text()
		capital = strings.TrimSpace(ChineseToUtf(capital, charset))

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
						if strings.Contains(brand_name, "阿斯顿") && strings.Contains(brand_name, "马丁") {
							brand_name = "阿斯顿・马丁"
						}
						if strings.Contains(brand_name, "阿尔法") && strings.Contains(brand_name, "罗密欧") {
							brand_name = "阿尔法・罗密欧"
						}
						brands[brand_name] = NewAutoHomeBrand(brand_name, link, capital)

						em := pair[1]
						pair = nil
						em = strings.TrimPrefix(em, "(")
						em = strings.TrimSuffix(em, ")</em>")
						num, _ := strconv.Atoi(em)
						brands[brand_name].SetTotalNums(num)
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

func decodeJsFuncs(str string) (key, value string) {
	pos := strings.Index(str, "var")
	if pos > 0 {
		str = str[:pos]
	}
	key = strings.Split(str, "()")[0]
	key = strings.Replace(key, "function","",-1)
	key = strings.TrimSpace(key)

	if strings.HasSuffix(str, "function") {
		str = strings.TrimSuffix(str, "function")
	}
	str = strings.TrimSpace(str)

	v, err := vm.Run(str + " " + key + "();")
	if v.IsString() && err == nil {
		value = v.String()
	}
	return
}

func decodeJsVars(str string) (key, value string) {
	str = strings.Replace(str, "var", "", -1)
	str = strings.TrimSpace(str)
	if strings.Contains(str, "=") {
		pair := strings.SplitN(str, "=", 2)
		key = strings.TrimSpace(pair[0])
		value = strings.TrimSpace(pair[1])

		if strings.HasPrefix(value, "'") {
			value = strings.TrimPrefix(value, "'")
		}

		if strings.HasSuffix(value, "'") {
			value = strings.TrimSuffix(value, "'")
		}
	}
	return
}

func decodeJsVarfuncs(str string) (key, value string) {
	str = strings.Replace(str, "var", "", -1)
	str = strings.TrimSpace(str)
	if strings.Contains(str, "=") {
		pair := strings.SplitN(str, "=", 2)
		key = pair[0]
		if strings.Contains(pair[1], "function") {
			js := "var " +str + "; " + key + "();"
			v, err := vm.Run(js)
			if err == nil && v.IsString() {
				value = v.String()
				return
			}
		}
	}
	return
}

func replaceDashAsNullString(s string) string {
	if strings.Contains(s, "&nbsp;") {
		s = strings.Replace(s, "&nbsp;"," ", -1)
	}
	return s
}

func getAutoHomeDict(js string) (dict_slice map[int]string){
	var (
		map_strs   = make(map[string]string)
		dict string
		pattern string
		re_comp *regexp.Regexp
		matches []string
	)
	dict_slice = make(map[int]string)
	map_source := make(map[string]string)
	// 去掉尾部
	str := strings.Replace(js, "})(document);</script>", " function",-1)
	// 匹配functions
	pattern = `function\s(\S){0,2}_\(\)\s*\{.*?\}+\s+`
	re_comp = regexp.MustCompile(pattern)
	matches = re_comp.FindAllString(str, -1)
	for _, fc := range matches {
		key, value := decodeJsFuncs(fc)
		map_source[key] = fc
		if key != "" {
			map_strs[key] = value
		}
	}
	// 匹配 var申明
	pattern = `var\s?\S\S_=\s?'\S*'`
	re_comp = regexp.MustCompile(pattern)
	matches = re_comp.FindAllString(js, -1)
	for _, variable := range matches {
		key, value := decodeJsVars(variable)
		map_source[key] = variable
		if key != "" {
			map_strs[key] = value
		}
	}
	// 匹配 var functions
	pattern = `var\s?\S\S_=\s?function\s?\(\)\s?\{.*?return.*?return.*?\}`
	re_comp = regexp.MustCompile(pattern)
	matches = re_comp.FindAllString(js, -1)
	for _, varfunc := range matches {
		key, value := decodeJsVarfuncs(varfunc)
		map_source[key] = varfunc
		if key != "" {
			map_strs[key] = value
		}
	}
	//拼接字典
	pattern = `function\s*\$FillDicData\$\s*\(\)\s*?{.*?\$RenderToHTML`
	if is_match, _ := regexp.MatchString(pattern, js); is_match {
		re_comp = regexp.MustCompile(pattern)
		str_match := re_comp.FindString(js)

		if !strings.Contains(str_match, "$GetWindow$()") {
			return
		}
		position := strings.Index(str_match, "$GetWindow$()")
		str_tmp := str_match[position:]

		if !strings.Contains(str_tmp, "$rulePosList$") {
			return
		}
		position = strings.Index(str_tmp,"$rulePosList$")
		str_tmp = str_tmp[:position]

		if !strings.Contains(str_tmp, "]") {
			return
		}
		position = strings.Index(str_tmp, "]")
		str_tmp = str_tmp[position + 1:]
		strs_dict := strings.Split(str_tmp, "+")
		for i := 1; i < len(strs_dict); i++ {
			str_to_match := strs_dict[i]

			if is_match, _ := regexp.MatchString(`\(\'\S+\'\)`, str_to_match); is_match {
				// match (function(TW__){'return TW_';return TW__;})('矩碟')
				// or Zf_('仪价')
				re_tmp := regexp.MustCompile(`\(\'\S+\'\)`)
				tmp := re_tmp.FindString(str_to_match)
				tmp = strings.Replace(tmp, "(", "", -1)
				tmp = strings.Replace(tmp, ")", "", -1)
				tmp = strings.Replace(tmp, "'", "", -1)
				tmp = strings.TrimSpace(tmp)
				dict += tmp
			} else if is_match, _ := regexp.MatchString(`^\'\S+\'$`, str_to_match); is_match {
				// match 'od'
				tmp := strings.Replace(str_to_match, "'", "", -1)
				tmp = strings.TrimSpace(tmp)
				dict += tmp
			} else if is_match, _ := regexp.MatchString(`\(function\s{0,3}\(\)\{.*?return.*?return.*?\}\)`, str_to_match); is_match {
				// match (function(){'return kJ_';return 'e'})()
				re_tmp := regexp.MustCompile(`\(function\s{0,3}\(\)\{.*?return.*?return.*?\}\)`)
				str_matched := re_tmp.FindString(str_to_match)

				if is_match, _ := regexp.MatchString(`return\s?\'\S+\'`, str_matched); is_match {
					re_tmp := regexp.MustCompile(`return\s?\'\S+\'`)
					str_tmp := re_tmp.FindString(str_matched)

					tmp := strings.Replace(str_tmp, "return", "", -1)
					tmp = strings.Replace(tmp, "'", "", -1)
					tmp = strings.TrimSpace(tmp)
					dict += tmp
				}
			} else if is_match, _ := regexp.MatchString(`^\S{2}_\(\)$`,str_to_match); is_match {
				// match ha_()
				key := strings.Replace(str_to_match,"()","",-1)
				tmp := map_strs[key]
				dict += tmp
			} else if is_match, _ := regexp.MatchString(`^\S{2}_$`,str_to_match); is_match {
				// match bm_
				key := str_to_match
				dict += map_strs[key]
			} else if is_match, _ := regexp.MatchString(`\('([A-Z]|[a-z]|[0-9]|[,]|[']|[;]|[\u4e00-\u9fbb]){1,10}'\)`,str_to_match); is_match {
				re_tmp := regexp.MustCompile(`\('([A-Z]|[a-z]|[0-9]|[,]|[']|[;]|[\u4e00-\u9fbb]){1,10}'\)`)
				tmp := re_tmp.FindString(str_to_match)
				if len([]rune(tmp)) >= 2 {
					tmp = tmp[:2]
				}
				dict += tmp
			} else {
				dict += "X"
			}
		}
		//拼接坐标
		indexes := ""
		position = strings.Index(str_match, "$rulePosList$")
		str_tmp = str_match[position:]

		position = strings.Index(str_tmp, "$SystemFunction2$")
		str_tmp = str_tmp[:position - 2]
		str_tmp = strings.TrimSpace(str_tmp)
		strs_indexes := strings.Split(str_tmp,"+")
		for i := 1; i < len(strs_indexes); i++ {
			str_to_match := strs_indexes[i]
			tmp := ""

			if is_match, _ := regexp.MatchString(`\(\'\S+\'\)`, str_to_match); is_match {
				// match (function(Qc__){'return Qc_';return Qc__;})(';69,37')
				// or nb_('38;64,')
				re_tmp := regexp.MustCompile(`\(\'\S+\'\)`)
				tmp = re_tmp.FindString(str_to_match)
				tmp = strings.Replace(tmp, "(", "", -1)
				tmp = strings.Replace(tmp, ")", "", -1)
				tmp = strings.Replace(tmp, "'", "", -1)
				tmp = strings.TrimSpace(tmp)
				indexes += tmp
			} else if is_match, _ := regexp.MatchString(`^\'\S+\'$`, str_to_match); is_match {
				// match '13,141'
				tmp = strings.Replace(str_to_match, "'", "", -1)
				tmp = strings.TrimSpace(tmp)
				indexes += tmp
			} else if is_match, _ := regexp.MatchString(`\(function\s{0,3}\(\)\{.*?return.*?return.*?\}\)`, str_to_match); is_match {
				// match (function(){'return Cx_';return '04,'})()
				re_tmp := regexp.MustCompile(`\(function\s{0,3}\(\)\{.*?return.*?return.*?\}\)`)
				str_matched := re_tmp.FindString(str_to_match)

				if is_match, _ := regexp.MatchString(`return\s?\'\S+\'`, str_matched); is_match {
					re_tmp := regexp.MustCompile(`return\s?\'\S+\'`)
					str_tmp := re_tmp.FindString(str_matched)

					tmp = strings.Replace(str_tmp, "return", "", -1)
					tmp = strings.Replace(tmp, "'", "", -1)
					tmp = strings.TrimSpace(tmp)
					indexes += tmp
				}
			} else if is_match, _ := regexp.MatchString(`^\S{2}_\(\)$`,str_to_match); is_match {
				// match ha_()
				key := strings.Replace(str_to_match,"()","",-1)
				tmp = map_strs[key]
				indexes += tmp
			} else if is_match, _ := regexp.MatchString(`^\S{2}_$`,str_to_match); is_match {
				// match bm_
				key := str_to_match
				tmp = map_strs[key]
				indexes += tmp
			} else if strings.TrimSpace(str_to_match) == "''" {
				continue
			} else {
				indexes += "X"
			}
		}

		runes_dict := []rune(dict)
		items := strings.Split(indexes, ";")
		for i, str := range items {
			sbresult := ""
			if str == "" {
				continue
			}
			nums := strings.Split(str, ",")
			for _, num := range nums {
				index, err := strconv.Atoi(num)
				if err != nil {
					logger.Record("convert",num,"to int failed!",err)
					continue
				}

				if index < len(runes_dict) {
					s := string(runes_dict[index])
					sbresult += s
				}
			}
			dict_slice[i] = sbresult
		}
	}
	return
}

func replaceHtmlByDict(origin string, dict map[string]map[int]string) (value string) {
	value = origin
	pattern := `<span class='hs_kw.*?'></span>`
	if is_match, _ := regexp.MatchString(pattern, origin); is_match {
		re_tmp := regexp.MustCompile(pattern)
		str_matches := re_tmp.FindAllString(origin, -1)

		for _, str := range str_matches {
			tmp := strings.Replace(str, "<span class='hs_kw","", -1)
			tmp = strings.Replace(tmp, "'></span>", "", -1)

			pattern := `\d*`
			re_tmp := regexp.MustCompile(pattern)
			str_num := re_tmp.FindString(tmp)
			num, _ := strconv.Atoi(str_num)

			if strings.Contains(tmp, "config") {
				if num < len(dict["config"]) {
					str_to_replace := dict["config"][num]
					value = strings.Replace(value,str,str_to_replace,1)
				}
			} else if strings.Contains(tmp, "option") {
				if num < len(dict["option"]) {
					str_to_replace := dict["option"][num]
					value = strings.Replace(value,str,str_to_replace,1)
				}
			} else if strings.Contains(tmp, "keylink") {
				value = string(dict["keylink"][num])
			}
		}
	}
	return
}

func dealAutoHomeItemValue(item gjson.Result, info map[int]*CarCrawl, key string, dict map[string]map[int]string) {
	id := item.Get("specid")
	if !id.Exists() {
		return
	}
	car_id := int(id.Int())
	car, exist := info[car_id]
	if !exist {
		return
	}
	value := item.Get("value").String()
	if key == "Car_type" {
		value = "1"
	}
	if key == "Engine_type" {
		value = car.Engine
	}
	if value != "" {
		value = replaceHtmlByDict(value, dict)
	}
	value = replaceDashAsNullString(value)

	v := reflect.ValueOf(car)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return
	}

	field := v.FieldByName(key)
	if !field.CanSet() {
		logger.Record(key,"can not be set value!")
		return
	}

	switch field.Kind() {
	case reflect.String:
		field.SetString(value)
	case reflect.Int:
		val, _ := strconv.Atoi(value)
		field.SetInt(int64(val))
	case reflect.Uint:
		val, _ := strconv.Atoi(value)
		field.SetUint(uint64(val))
	}
	return
}