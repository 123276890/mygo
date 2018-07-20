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
	"io/ioutil"
	"errors"
	"regexp"

	"github.com/PuerkitoBio/goquery"
	"github.com/astaxie/beego/orm"
	"github.com/tidwall/gjson"
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

			getAutoHomeBrand(b.Url, b.Name)
			//time.Sleep(time.Millisecond * 1000)

			// update 品牌
			brand_query := &Brand{Brand_name:b.Name}
			err = o.Read(brand_query,"brand_name")

			if err != nil {
				// 数据库未找到该品牌
				//logger.Record("No such brand in DB:",b.Name)
				//下载并保存品牌logo
				logo_file_path, err := downloadBrandLogo(b)
				if err != nil {
					logger.Record("Error when downloading",b.Name,"logo: ",err)
				}
				brand_query.Brand_initial = b.Cap
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
			ss := b.getSeries()
			for _, s := range ss {
				if s.Status == "停产" || s.Status == "停售" {
					continue
				}
				series_query := &CarSeries{Series_name:s.Name,Brand_id:brand_query.Brand_id}
				err = o.Read(series_query, "series_name","brand_id")

				if err != nil {
					// 数据库未查到该车系
					series_query.Series_id = s.AutoHomeSid
					sid, err := o.Insert(series_query)
					if err != nil {
						logger.Record("Error when insert a new CarSeries:", err)
					} else {
						logger.Record("New CarSeries insert into DB Success:",sid)
					}
				}
				//更新车系主页 和 车系参数主页
				if series_query.Series_home == "" || series_query.Series_config == "" {
					series_query.Series_home = s.Url
					series_query.Series_config = s.Settings
					series_query.Status = s.Status
					num, err := o.Update(series_query,"series_home","series_config","status")
					if err != nil {
						logger.Record("Error when update CarSeries home page:",err)
					} else {
						logger.Record("Update CarSeries home page success",num)
					}
				}

				//
				for _, c := range s.Cars {
					car_info, err := fetchCarInfo(c)
					if err != nil {
						logger.Record("Error when fetch car's info:",c.Name,err)
						continue
					}
					query_car := &CarCrawl{Type_id:c.Aid,Car_name:c.Name,Series_id:s.AutoHomeSid,Series_name:s.Name,Brand_name:b.Name,Manufacturer:s.Manufacture.Name}
					err = o.Read(query_car, "type_id")
					if err != nil {

					}
					query_car.Actual_brake, _ = car_info["base"]["actual_brake"]
				}
			} // end for series
		}// end for brands
	}
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
	words_rune := []rune(b.Name)
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
	if checkFileExist(filename) {
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

func fetchCarInfo(c *Car) (map[string]map[string]string, error) {
	info := map[string]map[string]string{}
	var (
		charset string
		schemes string
		host string
		err error
	)

	resp, err := http.Get(c.Settings)
	defer resp.Body.Close()
	if err != nil {
		logger.Record("Error: goqueryGet http.Get:", err)
		return info, err
	}

	if resp.StatusCode != 200 {
		return info, errors.New("Network Error!")
	}

	if content_type, ok := resp.Header["Content-Type"]; ok {
		pair := strings.SplitN(content_type[0], "=", 2)
		charset = pair[1]
		pair = nil
	}

	u, err := url.Parse(c.Settings)
	if err != nil {
		logger.Record("Error: goqueryGet ParseUrl:", err)
		return info, err
	}
	schemes = u.Scheme
	host = u.Host

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		logger.Record("Error: goqueryGet Err:", err)
		return info, err
	}

	html, _ := doc.Html()
	doc = nil
	resp.Body.Close()
	html = ChineseToUtf(html, charset)
	//fmt.Println(html)
	// 基本参数组
	pos_start := strings.Index(html, "var config =")
	if pos_start <= 0 {
		return info, errors.New("Error when try to find inner color")
	}
	str_base := html[pos_start:]
	pos_end := strings.IndexByte(str_base, '\n')
	if pos_end > len(str_base) {
		return info, errors.New("Error when try to find inner color: position end is out of str length")
	}
	str_base = str_base[13:pos_end - 1]
	//fmt.Println(str_base)

	// 选项配置参数组
	pos_start = strings.Index(html, "var option =")
	if pos_start <= 0 {
		return info, errors.New("Error when try to find inner color")
	}
	str_option := html[pos_start:]
	pos_end = strings.IndexByte(str_option, '\n')
	if pos_end > len(str_option) {
		return info, errors.New("Error when try to find inner color: position end is out of str length")
	}
	str_option = str_option[13:pos_end - 1]
	//fmt.Println(str_option)

	// 外观颜色json
	pos_start = strings.Index(html, "var color =")
	if pos_start <= 0 {
		return info, errors.New("Error when try to find inner color")
	}
	str_color := html[pos_start:]
	pos_end = strings.IndexByte(str_color, '\n')
	if pos_end > len(str_color) {
		return info, errors.New("Error when try to find inner color: position end is out of str length")
	}
	str_color = str_color[12:pos_end - 1]
	//fmt.Println(str_color)

	// 内饰颜色json
	pos_start = strings.Index(html, "var innerColor =")
	if pos_start <= 0 {
		return info, errors.New("Error when try to find inner color")
	}
	str_inner := html[pos_start:]
	pos_end = strings.IndexByte(str_inner, '\n')
	if pos_end > len(str_inner) {
		return info, errors.New("Error when try to find inner color: position end is out of str length")
	}
	str_inner = str_inner[16:pos_end - 1]
	//fmt.Println(str_inner)

	result := gjson.Get(str_base, "result.paramtypeitems")

	if result.Exists() {
		items := result.Array()
		for _, item := range items {
			item_name := item.Get("name").String()
			if item_name == "基本参数" {
				base_params := item.Get("paramitems").Array()
				info["base"] = make(map[string]string)

				for _, v := range base_params {
					name := v.Get("name").String()
					//能源类型
					if strings.Contains(name, "能源类型") {
						v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
							if value.Get("specid").String() == strconv.Itoa(c.Aid) {
								info["base"]["energy_type"] = replaceDashAsNullString(value.Get("value").String())
								return true
							}
							return false
						})
					}
					//上市时间
					if strings.Contains(name, "上市") {
						v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
							if value.Get("specid").String() == strconv.Itoa(c.Aid) {
								info["base"]["market_time"] = replaceDashAsNullString(value.Get("value").String())
								return true
							}
							return false
						})
					}
					//工信部纯电续驶里程
					if strings.Contains(name, "工信部纯电续驶里程") {
						v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
							if value.Get("specid").String() == strconv.Itoa(c.Aid) {
								info["base"]["e_mileage"] = replaceDashAsNullString(value.Get("value").String())
								return true
							}
							return false
						})
					}
					//变速箱
					if strings.Contains(name, "变速箱") {
						v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
							if value.Get("specid").String() == strconv.Itoa(c.Aid) {
								info["base"]["gearbox"] = replaceDashAsNullString(value.Get("value").String())
								return true
							}
							return false
						})
					}

					id := v.Get("id").Int()
					switch id {
					case 295:	//最大功率
						v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
							if value.Get("specid").String() == strconv.Itoa(c.Aid) {
								info["base"]["max_power"] = replaceDashAsNullString(value.Get("value").String())
								return true
							}
							return false
						})
					case 571:	//最大扭矩
						v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
							if value.Get("specid").String() == strconv.Itoa(c.Aid) {
								info["base"]["max_torque"] = replaceDashAsNullString(value.Get("value").String())
								return true
							}
							return false
						})
					case 555:	//发动机
						v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
							if value.Get("specid").String() == strconv.Itoa(c.Aid) {
								info["base"]["engine"] = replaceDashAsNullString(value.Get("value").String())
								return true
							}
							return false
						})
					case 222:	//长*宽*高
						v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
							if value.Get("specid").String() == strconv.Itoa(c.Aid) {
								info["base"]["car_size"] = replaceDashAsNullString(value.Get("value").String())
								return true
							}
							return false
						})
					case 281:	//车身结构
						v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
							if value.Get("specid").String() == strconv.Itoa(c.Aid) {
								info["base"]["car_struct"] = replaceDashAsNullString(value.Get("value").String())
								return true
							}
							return false
						})
					case 267:	//最高车速
						v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
							if value.Get("specid").String() == strconv.Itoa(c.Aid) {
								info["base"]["max_speed"] = replaceDashAsNullString(value.Get("value").String())
								return true
							}
							return false
						})
					case 225:	//官方100加速
						v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
							if value.Get("specid").String() == strconv.Itoa(c.Aid) {
								info["base"]["official_speedup"] = replaceDashAsNullString(value.Get("value").String())
								return true
							}
							return false
						})
					case 272:	//实测100加速
						v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
							if value.Get("specid").String() == strconv.Itoa(c.Aid) {
								info["base"]["actual_speedup"] = replaceDashAsNullString(value.Get("value").String())
								return true
							}
							return false
						})
					case 273:	//实测100制动
						v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
							if value.Get("specid").String() == strconv.Itoa(c.Aid) {
								info["base"]["actual_brake"] = replaceDashAsNullString(value.Get("value").String())
								return true
							}
							return false
						})
					case 271:	//工信部综合油耗
						v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
							if value.Get("specid").String() == strconv.Itoa(c.Aid) {
								info["base"]["gerenal_fueluse"] = replaceDashAsNullString(value.Get("value").String())
								return true
							}
							return false
						})
					case 243:	//实测油耗
						v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
							if value.Get("specid").String() == strconv.Itoa(c.Aid) {
								info["base"]["actual_fueluse"] = replaceDashAsNullString(value.Get("value").String())
								return true
							}
							return false
						})
					case 274:	//整车质保
						v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
							if value.Get("specid").String() == strconv.Itoa(c.Aid) {
								quality := value.Get("value").String()
								pat := `<[^>]+>`
								re := regexp.MustCompile(pat)
								found := re.Split(quality, -1)
								if len(found) == 5 {
									info["base"]["quality_guarantee"] = found[0] + "年或" + found[2] + "万" + found[4]
								}
								return true
							}
							return false
						})
					} // end of switch id
				}// end for base_params
			} // end of 基本参数

			if item_name == "车身" {
				body_params := item.Get("paramitems").Array()
				info["body"] = make(map[string]string)

				for _, v := range body_params {
					id := v.Get("id").Int()

					switch id {
					case 275:	//长度(mm)
						v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
							if value.Get("specid").String() == strconv.Itoa(c.Aid) {
								info["body"]["length"] = replaceDashAsNullString(value.Get("value").String())
								return true
							}
							return false
						})
					case 276:	//宽度(mm)
						v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
							if value.Get("specid").String() == strconv.Itoa(c.Aid) {
								info["body"]["width"] = replaceDashAsNullString(value.Get("value").String())
								return true
							}
							return false
						})
					case 277:	//高度(mm)
						v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
							if value.Get("specid").String() == strconv.Itoa(c.Aid) {
								info["body"]["height"] = replaceDashAsNullString(value.Get("value").String())
								return true
							}
							return false
						})
					case 132:	//轴距(mm)
						v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
							if value.Get("specid").String() == strconv.Itoa(c.Aid) {
								info["body"]["shaft_distance"] = replaceDashAsNullString(value.Get("value").String())
								return true
							}
							return false
						})
					case 278:	//前轮距(mm)
						v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
							if value.Get("specid").String() == strconv.Itoa(c.Aid) {
								info["body"]["front_wheels_gap"] = replaceDashAsNullString(value.Get("value").String())
								return true
							}
							return false
						})
					case 638:	//后轮距(mm)
						v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
							if value.Get("specid").String() == strconv.Itoa(c.Aid) {
								info["body"]["back_wheels_gap"] = replaceDashAsNullString(value.Get("value").String())
								return true
							}
							return false
						})
					case 279:	//最小离地间隙(mm)
						v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
							if value.Get("specid").String() == strconv.Itoa(c.Aid) {
								info["body"]["min_ground"] = replaceDashAsNullString(value.Get("value").String())
								return true
							}
							return false
						})
					case 280:	//整备质量(kg)
						v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
							if value.Get("specid").String() == strconv.Itoa(c.Aid) {
								info["body"]["total_weight"] = replaceDashAsNullString(value.Get("value").String())
								return true
							}
							return false
						})
					case 281:	//车身结构
						v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
							if value.Get("specid").String() == strconv.Itoa(c.Aid) {
								info["body"]["body_struct"] = replaceDashAsNullString(value.Get("value").String())
								return true
							}
							return false
						})
					case 282:	//车门数
						body_params[8].Get("valueitems").ForEach(func(key, value gjson.Result) bool {
							if value.Get("specid").String() == strconv.Itoa(c.Aid) {
								info["body"]["doors"] = replaceDashAsNullString(value.Get("value").String())
								return true
							}
							return false
						})
					case 283:	//座位数
						v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
							if value.Get("specid").String() == strconv.Itoa(c.Aid) {
								info["body"]["seats"] = replaceDashAsNullString(value.Get("value").String())
								return true
							}
							return false
						})
					case 284:	//油箱容积(L)
						v.Get("valueitems").ForEach(func(key, value gjson.Result) bool {
							if value.Get("specid").String() == strconv.Itoa(c.Aid) {
								info["body"]["fuel_vol"] = replaceDashAsNullString(value.Get("value").String())
								return true
							}
							return false
						})
					case 285:	//行李厢容积(L)
						body_params[11].Get("valueitems").ForEach(func(key, value gjson.Result) bool {
							if value.Get("specid").String() == strconv.Itoa(c.Aid) {
								info["body"]["cargo_vol"] = replaceDashAsNullString(value.Get("value").String())
								return true
							}
							return false
						})
					}// end of switch
				}// end for body_params
			} // end of 车身

			if item_name == "发动机" {
				engine_params := item.Get("paramitems").Array()
				info["engine"] = make(map[string]string)

				for _, v := range engine_params {
					id := v.Get("id").Int()

					switch id {
					case 570:	//发动机型号
					case 287:	//排量(mL)
					case 640:	//进气形式
					case 289:	//气缸排列形式
					case 290:	//气缸个数
					case 291:	//每缸气门数
					case 182:	//压缩比
					case 641:	//配气机构
					case 181:	//缸径(mm)
					case 293:	//行程(mm)
					case 294:	//最大马力(Ps)
					case 295:	//最大功率(kW)
					case 296:	//最大功率转速(rpm)
					case 571:	//最大扭矩(N·m)
					case 642:	//最大扭矩转速(rpm)
					case 643:	//发动机特有技术
					case 572:	//燃料形式
					case 573:	//燃油标号
					case 574:	//供油方式
					case 575:	//缸盖材料
					case 576:	//缸体材料
					case 577:	//环保标准
					}
				}// end for engine_params
			} // end of 发动机

			if item_name == "电动机" {
				motor_params := item.Get("paramitems").Array()
				info["motor"] = make(map[string]string)

				for _, v := range motor_params {
					id := v.Get("id").Int()

					switch id {
					case 0:
					}// end switch
				}// end for motor_params
			} // end of 电动机

			if item_name == "变速箱" {
				gearboxes := item.Get("paramitems").Array()
				info["gearbox"] = make(map[string]string)

				for _, v := range gearboxes {
					id := v.Get("id").Int()

					switch id {
					case 559:	//挡位个数
					case 221:	//变速箱类型
					case 1072:	//简称
					}// end switch
				}// end for gearboxes
			} // end of 变速箱

			if item_name == "底盘转向" {
				underpan := item.Get("paramitems").Array()
				info["underpan"] = make(map[string]string)

				for _, v := range underpan {
					id := v.Get("id").Int()

					switch id {
					case 395:	//驱动方式
					case 578:	//前悬架类型
					case 579:	//后悬架类型
					case 510:	//助力类型
					case 223:	//车体结构
					}// end switch
				}// end for underpan
			} // end of 底盘转向

			if item_name == "车轮制动" {
				brake_params := item.Get("paramitems").Array()
				info["brake"] = make(map[string]string)

				for _, v := range brake_params {
					id := v.Get("id").Int()

					switch id {
					case 511:	//前制动器类型
					case 512:	//后制动器类型
					case 513:	//驻车制动类型
					case 580:	//前轮胎规格
					case 581:	//后轮胎规格
					case 515:	//备胎规格
					}// end switch
				}// end for brake_params
			} // end of 车轮制动
		}
	}// end of Base result.Exists()

	result = gjson.Get(str_option, "result.configtypeitems")

	if result.Exists() {
		items := result.Array()

		for _, item := range items {
			item_name := item.Get("name").String()

			if item_name == "主/被动安全装备" {
				if item.Get("configitems").IsArray() {
					secure_params := item.Get("configitems").Array()
					info["secure"] = make(map[string]string)

					for _,v := range secure_params {
						id := v.Get("id").Int()

						switch id {
						case 1082:	//主/副驾驶座安全气囊
						}
					} // end for range secure_params
				}
			} // end of 主/被动安全装备

			if item_name == "外部/防盗配置" {
				if item.Get("configitems").IsArray() {
					guard_params := item.Get("configitems").Array()
					info["guard"] = make(map[string]string)

					for _,v := range guard_params {
						id := v.Get("id").Int()

						switch id {
						case 583:	//电动天窗
						}
					} // end for range guard_params
				}
			} // end of 外部/防盗配置

			if item_name == "内部配置" {
				if item.Get("configitems").IsArray() {
					inside_params := item.Get("configitems").Array()
					info["inside"] = make(map[string]string)

					for _,v := range inside_params {
						name := v.Get("name").String()
						if name == "皮质方向盘" {

						}

						id := v.Get("id").Int()

						switch id {
						case 1085:	//方向盘调节
						}
					} // end for range inside_params
				}
			} // end of 内部配置

			if item_name == "座椅配置" {
				if item.Get("configitems").IsArray() {
					seat_params := item.Get("configitems").Array()
					info["seat"] = make(map[string]string)

					for _,v := range seat_params {
						id := v.Get("id").Int()

						switch id {
						case 592:	//运动风格座椅
						}
					} // end for range seat_params
				}
			} // end of 座椅配置

			if item_name == "多媒体配置" {
				if item.Get("configitems").IsArray() {
					media_params := item.Get("configitems").Array()
					info["media"] = make(map[string]string)

					for _,v := range media_params {
						id := v.Get("id").Int()

						switch id {
						case 607:	//GPS导航系统
						}
					} // end for range media_params
				}
			} // end of 多媒体配置

			if item_name == "灯光配置" {
				if item.Get("configitems").IsArray() {
					light_params := item.Get("configitems").Array()
					info["light"] = make(map[string]string)

					for _,v := range light_params {
						id := v.Get("id").Int()

						switch id {
						case 453:	//车内氛围灯
						}
					} // end for range light_params
				}
			} // end of 灯光配置

			if item_name == "玻璃/后视镜" {
				if item.Get("configitems").IsArray() {
					glass_params := item.Get("configitems").Array()
					info["glass"] = make(map[string]string)

					for _,v := range glass_params {
						id := v.Get("id").Int()

						switch id {
						case 623:	//车窗防夹手功能
						}
					} // end for range glass_params
				}
			} // end of 玻璃/后视镜

			if item_name == "空调/冰箱" {
				if item.Get("configitems").IsArray() {
					air_params := item.Get("configitems").Array()
					info["air"] = make(map[string]string)

					for _,v := range air_params {
						id := v.Get("id").Int()

						switch id {
						case 1097:	//空调控制方式
						}
					} // end for range air_params
				}
			} // end of 空调/冰箱
		}// end for range items
	}// end of Options result.Exists()

	fmt.Println(schemes,host,charset)
	return info, nil
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

		// 补齐车型名称中未含车系名称
		if !strings.HasPrefix(carName, series_name) {
			carName = series_name + " " + carName
		}

		/* 补齐车型名称中未含品牌名称
		if !strings.HasPrefix(carName, brand_name) {
			carName = brand_name + " " + carName
		}
		*/

		carUrl, ok := div.Find("a").Attr("href")
		if ok {
			if !strings.HasPrefix(carUrl, schemes + ":") {
				carUrl = schemes + ":" + carUrl
			}

			if strings.Contains(carUrl, "#") {
				carUrl = strings.SplitN(carUrl, "#", 2)[0]
			}
		}

		car := NewCar(carId, carName, carUrl)

		pa := div.Parent().Next().Next()
		if pa.HasClass("interval01-list-guidance") {
			price := pa.Find("div").Text()
			price = strings.TrimSpace(price)
			car.SetPrice(ChineseToUtf(price, charset))
		}

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
	brand_logo, ok := imgNode.Attr("src")
	if ok {
		if !strings.HasPrefix(brand_logo, schemes+":") {
			brand_logo = schemes + ":" + brand_logo
		}
		brands[brand_name].Img = brand_logo
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
					}

					if !strings.HasPrefix(series_link, schemes+"://"+host) {
						series_link = schemes + "://" + host + series_link
					}
					if strings.Contains(series_link, "#") {
						series_link = strings.SplitN(series_link, "#", 2)[0]
					}

					series := NewSeries(sid, s_name, s_status, series_link)
					series.Manufacture = manuf
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

func replaceDashAsNullString(s string) string {
	if s == "-" {
		return ""
	}
	return s
}