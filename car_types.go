package main

import (
	"path/filepath"
	"strconv"
)

const (
	DS = string(filepath.Separator)
)

type Car struct {
	Aid			int					`汽车之家 ID`
	Name		string				`车型名称`
	Url			string				`车型配置参数主页`
}

type Series struct {
	AutoHomeSid	int					`汽车之家车系ID`
	Name		string				`车系名称`
	Status 		string				`停产停售状态`
	Url    		string				`车系主页`
	Settings	string				`车系配置参数主页`
	Cars  		map[string]*Car		`车型`
}

type Manufacture struct {
	Name   string					`厂商名称`
	Url    string					`厂商主页`
	Series map[string]*Series		`旗下车系`
}

type AutoHomeBrand struct {
	Name         string						`品牌名称`
	Url          string						`品牌主页`
	Img          string						`品牌标志`
	Cap          string						`品牌首字母`
	Nums         int						`车型总数（包含已停售已停产）`
	Manufactures map[string]*Manufacture	`旗下厂商`
}

type AutoHomeBrands map[string]*AutoHomeBrand

func (b *AutoHomeBrands) IsEmpty() bool {
	if len(*b) > 0 {
		return false
	}
	return true
}

func (b *AutoHomeBrands) Count() int {
	return len(*b)
}

func NewAutoHomeBrand(brand_name, brand_homepage, brand_capital string) *AutoHomeBrand {
	b := &AutoHomeBrand{Name:brand_name,Url:brand_homepage,Cap:brand_capital}
	return b
}

func NewManufacture() *Manufacture {
	m := &Manufacture{}
	return m
}

func (m *Manufacture) SetName(mName string) {
	m.Name = mName
}

func (m *Manufacture) SetUrl(sUrl string) {
	m.Url = sUrl
}

func (m *Manufacture) SetSeries(ss map[string]*Series) {
	m.Series = ss
}

func NewSeries(series_name, series_status, s_url string) *Series {
	s := &Series{Name:series_name, Status:series_status, Url:s_url}
	return s
}

func (s *Series) SetSettings(settings_url string) {
	s.Settings = settings_url
}

func (s *Series) SetCars(cars map[string]*Car) {
	s.Cars = cars
}

func (s *Series) SetName(sName string) {
	s.Name = sName
}

func (s *Series) SetUrl(sUrl string) {
	s.Url = sUrl
}

func (b AutoHomeBrand) getSeries() ([]*Series) {
	if len(b.Manufactures) <= 0 {

	}

	var s []*Series
	for _, m := range b.Manufactures {
		for _, v := range m.Series {
			s = append(s, v)
		}
	}
	return s
}

func (b AutoHomeBrand) String() string {
	baseStr := "品牌: " + b.Name + "\n" +
		"链接: " + b.Url + "\n" +
		"图标: " + b.Img + "\n" +
		"首字母: " + b.Cap + "\n" +
		"车辆总数: " + strconv.Itoa(b.Nums) + "\n" +
		"厂商: " + strconv.Itoa(len(b.Manufactures)) + "家\n"
	buf := NewBuffer()
	buf.Write([]byte(baseStr))
	str := "[\n"
	buf.Write([]byte(str))
	for _, m := range b.Manufactures {
		str = "	{厂商:" + m.Name + ", 链接:" + m.Url + ", 车系:[\n"
		buf.Write([]byte(str))
		for _, s := range m.Series {
			str = "		{S:" + s.Name + ",SID:" + strconv.Itoa(s.AutoHomeSid) + ", 状态:" + s.Status + ", 链接:" + s.Url + ",\n"
			str += "		参数:" + s.Settings + "},\n"
			buf.Write([]byte(str))
			for _, c := range s.Cars {
				str := "		AutoHomeId:" + strconv.Itoa(c.Aid) + ", Car Name:" + c.Name + "\n"
				buf.Write([]byte(str))
			}
		}
		str = "	]},\n"
		buf.Write([]byte(str))
	}
	buf.Write([]byte("]"))
	return string(*buf)
	/*return "品牌: " + b.Name + "\n" +
	"链接: " + b.Url + "\n" +
	"图标: " + b.Img + "\n" +
	"首字母: " + b.Cap + "\n" +
	"车辆总数: " + strconv.Itoa(b.Nums) + "\n" +
	"厂商: " + strconv.Itoa(len(b.Manufactures)) + "家\n"*/
}