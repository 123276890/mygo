package main

import (
	"path/filepath"
	"strconv"
)

const (
	DS = string(filepath.Separator)
)

type Series struct {
	Name string
	Cars map[string]string
}

type Manufacture struct {
	Name   string
	Url    string
	Series *Series
}

type Brand struct {
	Name         string
	Url          string
	Img          string
	Cap          string
	Nums         int
	Manufactures map[string]*Manufacture
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
	return "品牌: " + b.Name + "\n" +
		"链接: " + b.Url + "\n" +
		"图标: " + b.Img + "\n" +
		"首字母: " + b.Cap + "\n" +
		"车辆总数: " + strconv.Itoa(b.Nums) + "\n" +
		"厂商: " + strconv.Itoa(len(b.Manufactures)) + "家\n"
}
