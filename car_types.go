package main

import (
	"path/filepath"
	"strconv"
)

const (
	DS = string(filepath.Separator)
)

type Series struct {
	Name		string				`车系名称`
	Status 		string				`停产停售状态`
	Url    		string				`车系主页`
	Settings	string				`车系配置参数主页`
	Cars  		map[string]string	`车型`
}

type Manufacture struct {
	Name   string					`厂商名称`
	Url    string					`厂商主页`
	Series map[string]*Series		`旗下车系`
}

type Brand struct {
	Name         string						`品牌名称`
	Url          string						`品牌主页`
	Img          string						`品牌标志`
	Cap          string						`品牌首字母`
	Nums         int						`车型总数（包含已停售已停产）`
	Manufactures map[string]*Manufacture	`旗下厂商`
}

type Brands map[string]*Brand

func (b *Brands) IsEmpty() bool {
	if len(*b) > 0 {
		return false
	}
	return true
}

func (b *Brands) Count() int {
	return len(*b)
}

func (b Brand) String() string {
	baseStr := "品牌: " + b.Name + "\n" +
		"链接: " + b.Url + "\n" +
		"图标: " + b.Img + "\n" +
		"首字母: " + b.Cap + "\n" +
		"车辆总数: " + strconv.Itoa(b.Nums) + "\n" +
		"厂商: " + strconv.Itoa(len(b.Manufactures)) + "家\n"
	buf := NewBuffer()
	buf.Write([]byte(baseStr))
	mInfo := "[\n"
	for _, m := range b.Manufactures {
		mInfo += "	{厂商:" + m.Name + ",链接:" + m.Url + ",车系:[\n"
		for _, s := range m.Series {
			mInfo += "		{S:" + s.Name + ",状态:" + s.Status + ",链接:" + s.Url + ",\n"
			mInfo += "		参数:" + s.Settings + "},\n"
		}
		mInfo += "	]},\n"
		buf.Write([]byte(mInfo))
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