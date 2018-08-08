package main

import (
	"path/filepath"
	"strconv"
)

const (
	DS = string(filepath.Separator)
)

/*type Car struct {
	Aid			int					`汽车之家 ID`
	Name		string				`车型名称`
	homeUrl			string				`车型主页`
	Series_id			int
	configUrl	string
	Price		string
	Quality_guarantee	string
	brand_id	int
}*/

type Series struct {
	Sid        int               `汽车之家车系ID`
	Name       string            `车系名称`
	status     string            `停产停售状态`
	homeUrl    string            `车系主页`
	configUrl  string            `车系配置参数主页`
	*AutoHomeBrand
	*Manufacture
	CarCrawls  map[int]*CarCrawl `车型`
}

type Manufacture struct {
	Name   string					`厂商名称`
	Url    string					`厂商主页`
	Series map[string]*Series		`旗下车系`
}

type AutoHomeBrand struct {
	Name         string                  `品牌名称`
	Url          string                  `品牌主页`
	Img          string                  `品牌标志`
	Cap          string                  `品牌首字母`
	nums         int                     `车型总数（包含已停售已停产）`
	Manufactures map[string]*Manufacture `旗下厂商`
}

type AutoHomeBrands map[string]*AutoHomeBrand

func NewAutoHomeBrand(brand_name, brand_homepage, brand_capital string) *AutoHomeBrand {
	return &AutoHomeBrand{Name:brand_name,Url:brand_homepage,Cap:brand_capital}
}

func NewManufacture(v ...interface{}) *Manufacture {
	m := &Manufacture{}
	if len(v) > 0 {
		if name, ok := v[0].(string); ok {
			m.Name = name
		}
	}
	return m
}

func NewSeries(sid int, series_name, series_status, home_url string) *Series {
	return &Series{Sid:sid, Name:series_name, status:series_status, homeUrl:home_url}
}

func NewSeriesById(sid int) *Series {
	return &Series{Sid:sid}
}

func (b *AutoHomeBrands) IsEmpty() bool {
	if len(*b) > 0 {
		return false
	}
	return true
}

func (b *AutoHomeBrands) Count() int {
	return len(*b)
}

// brand funcs
func (b *AutoHomeBrand) SetName(name string) (*AutoHomeBrand) {
	b.Name = name
	return b
}

func (b *AutoHomeBrand) SetUrl(url string) (*AutoHomeBrand) {
	b.Url = url
	return b
}

func (b *AutoHomeBrand) SetLogo(logo_url string) (*AutoHomeBrand) {
	b.Img = logo_url
	return b
}

func (b *AutoHomeBrand) SetTotalNums(nums int) (*AutoHomeBrand) {
	b.nums = nums
	return b
}

func (b *AutoHomeBrand) SetManufs(manufs map[string]*Manufacture) (*AutoHomeBrand) {
	b.Manufactures = manufs
	return b
}

func (b *AutoHomeBrand) GetName() (string) {
	return b.Name
}

func (b *AutoHomeBrand) GetUrl() (string) {
	return b.Url
}

func (b *AutoHomeBrand) GetTotalNums() (int) {
	return b.nums
}

func (b *AutoHomeBrand) GetManufs() (map[string]*Manufacture) {
	return b.Manufactures
}

func (b AutoHomeBrand) getSeries() (s []*Series) {
	if len(b.GetManufs()) <= 0 {
		return
	}

	for _, m := range b.GetManufs() {
		for _, v := range m.Series {
			s = append(s, v)
		}
	}
	return
}

// manufacture funcs
func (m *Manufacture) SetName(mName string) {
	m.Name = mName
}

func (m *Manufacture) SetUrl(sUrl string) {
	m.Url = sUrl
}

func (m *Manufacture) SetSeries(ss map[string]*Series) {
	m.Series = ss
}

func (m *Manufacture) GetName() (string) {
	return m.Name
}

// series funcs
func (s *Series) SetSettings(settings_url string) (*Series) {
	s.configUrl = settings_url
	return s
}

func (s *Series) SetCars(cars map[int]*CarCrawl) (*Series) {
	s.CarCrawls = cars
	return s
}

func (s *Series) SetName(sName string) (*Series) {
	s.Name = sName
	return s
}

func (s *Series) SetHomeUrl(sUrl string) (*Series) {
	s.homeUrl = sUrl
	return s
}

func (s *Series) SetStatus(status string) (*Series) {
	s.status = status
	return s
}

func (s *Series) SetBrand(b *AutoHomeBrand) (*Series) {
	s.AutoHomeBrand = b
	return s
}

func (s *Series) SetManufacture(m *Manufacture) (*Series) {
	s.Manufacture = m
	return s
}

func (s *Series) GetSid() (int) {
	return s.Sid
}

func (s *Series) GetName() (string) {
	return s.Name
}

func (s *Series) GetBrandName() (string) {
	return s.AutoHomeBrand.GetName()
}

func (s *Series) GetManufactureName() (string) {
	return s.Manufacture.GetName()
}

func (s *Series) GetHomeUrl() (string) {
	return s.homeUrl
}

func (s *Series) GetConfigUrl() (string) {
	return s.configUrl
}

func (s *Series) GetStatus() (string) {
	return s.status
}

func (b AutoHomeBrand) String() string {
	baseStr := "品牌: " + b.GetName() + "\n" +
		"链接: " + b.GetUrl() + "\n" +
		"图标: " + b.Img + "\n" +
		"首字母: " + b.Cap + "\n" +
		"车辆总数: " + strconv.Itoa(b.GetTotalNums()) + "\n" +
		"厂商: " + strconv.Itoa(len(b.GetManufs())) + "家\n"
	buf := NewBuffer()
	buf.Write([]byte(baseStr))
	str := "[\n"
	buf.Write([]byte(str))
	for _, m := range b.GetManufs() {
		str = "	{厂商:" + m.Name + ", 链接:" + m.Url + ", 车系:[\n"
		buf.Write([]byte(str))
		for _, s := range m.Series {
			str = "		{S:" + s.GetName() + ",SID:" + strconv.Itoa(s.GetSid()) + ", 状态:" + s.GetStatus() + ", 链接:" + s.GetHomeUrl() + ",\n"
			str += "		参数:" + s.GetConfigUrl() + "},\n"
			buf.Write([]byte(str))
			for _, c := range s.CarCrawls {
				str := "		AutoHomeId:" + strconv.Itoa(c.Type_id) + ", Car Name:" + c.Car_name + ",\n"
				buf.Write([]byte(str))
				str = "			Sid:" + strconv.Itoa(c.Series_id) + ", 车系:" + c.Series_name + ",Bid:" + strconv.Itoa(c.Brand_id) + ", 品牌:" + c.Brand_name + ",\n"
				buf.Write([]byte(str))
				str = "			Config:" + c.settings + ",\n"
				buf.Write([]byte(str))
			}
		}
		str = "	]},\n"
		buf.Write([]byte(str))
	}
	buf.Write([]byte("]"))
	return string(*buf)
}