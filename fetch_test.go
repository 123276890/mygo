package main

import (
	"testing"
	"regexp"
	"fmt"
	"reflect"
)

func Test_Regexp(t *testing.T) {
	pat := `<[^>]+>`
	str := "三<span class='hs_kw6_configEQ'></span>10<span class='hs_kw1_configEQ'></span>公里"
	reg := regexp.MustCompile(pat)
	found := reg.Split(str, -1)
	//wg.Wait()
	t.Log(found)
}

func Test_PinYinSuoXie(t *testing.T) {
	//str := "Icona"
	//str := "马自达"
	str := "广汽集团"
	pinyin := ""
	words_rune := []rune(str)
	for _, v := range words_rune {
		s := string(v)
		p, ok := PinyinMap[s]
		if ok {
			pinyin += string(p[0])
		}
	}
	t.Log(pinyin)
}

func Test_reNameSameFileName(t *testing.T) {
	filename := "dn.jpg"
	path := "/Users/a2/work/shopnc/data/upload/shop/brand/logo"

	result := reNameSameFileName(filename, path)
	t.Log(result)
}

func Test_ReadPinyinMap(t *testing.T) {
	pyMap := loadPinyinMap()
	t.Log(len(pyMap))
}

/*func Test_ConvertPinyinFile(t *testing.T) {
	f, err := os.Open("googlepinyin.txt")
	if err != nil {
		t.Fatal(err.Error())
		return
	}
	defer f.Close()

	w, err := os.OpenFile("pinyin.txt", os.O_APPEND|os.O_WRONLY|os.O_CREATE, os.ModePerm)
	if err != nil {
		t.Fatal("Error on openning output file:", err)
	}
	defer w.Close()

	r := bufio.NewReader(f)
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			} else {
				t.Fatal(err)
			}
			break
		}
		line = strings.TrimSpace(line)
		arr := strings.Split(line, " ")
		if len([]rune(arr[0])) > 1 {
			continue
		}
		//output_line := arr[0] + " " + arr[3] + "\n"
		//_, err = w.Write([]byte(output_line))
		if err != nil {
			t.Fatal("Error when write to output file:", err)
			break
		}
	}
	t.Log("output success")
}*/

func Test_UpdateBrandLogo(t *testing.T) {
	var brand_names []string

	brand_names = []string{"宝骏","东风","福田","丰田","福特","哈飞","华泰","恒天","海马","红旗","捷豹","江铃","金龙","金旅","雷诺","林肯","领克","双龙","三菱","赛麟","蔚来"}
	JobUpdateBrandLogo(brand_names)
	t.Log("Done")
}

func Test_getAutoHomeBrand(t *testing.T) {
	// 大众: 1	欧宝: 59
	var (
		url string
		brand_name string
		brand_cap string
	)
	//brand_name, url, brand_cap = "欧宝", "https://car.autohome.com.cn/price/brand-59.html", "O"
	brand_name, url, brand_cap = "北京", "https://car.autohome.com.cn/price/brand-27.html", "B"

	brand := NewAutoHomeBrand(brand_name, url, brand_cap)

	brands[brand_name] = brand

	getAutoHomeBrand(brand)
	t.Log(brand)
}

func Test_FetchSeriesInfo(t *testing.T) {
	var (
		sid			int
		s_name		string
		brand_name	string
		sUrl		string
		mName		string
	)
	//sid, s_name, brand_name, mName,sUrl = 528, "帕萨特", "大众", ""一汽-大众奥迪"", "https://car.autohome.com.cn/price/series-528.html"
	sid, s_name, brand_name, mName,sUrl = 4441, "WEY P8", "WEY", "长城汽车", "https://car.autohome.com.cn/price/series-4441.html"

	b := NewAutoHomeBrand(brand_name,"","")
	manufacture := NewManufacture(mName)

	s := NewSeries(sid, s_name, "", sUrl)
	s.SetBrand(b).SetManufacture(manufacture)

	ret, ok := fetchSeriesInfo(s)
	if ok {
		t.Log(ret)
	} else {
		t.Fatal("Can not fetch anything!")
	}
}

/*func Test_FetchCarInfo(t *testing.T) {
	// 威途X35 2017款 RQ5026XXYEVH0 45kWh		ID:1005463		https://www.autohome.com.cn/spec/1005463
	// 帕萨特 2017款 280TSI DSG尊雅版		ID:29314	https://www.autohome.com.cn/spec/29314/
	var (
		car_id int
		car_name string
		car_url string
	)
	//car_id, car_name, car_url = 1005463, "威途X35 2017款 RQ5026XXYEVH0 45kWh", "https://www.autohome.com.cn/spec/1005463"
	car_id, car_name, car_url = 29314, "帕萨特 2017款 280TSI DSG尊雅版", "https://www.autohome.com.cn/spec/29314/"
	//car_id, car_name, car_url = 32018, "奥迪A4L 2018款 30周年年型 45 TFSI quattro 个性运动版", "https://www.autohome.com.cn/spec/32018/"
	car := NewCar(car_id,car_name,car_url)
	car.SetPrice("18.99万")
	info, err := fetchCarInfo(car)
	if err != nil {

	}
	t.Log(len(info))
}*/

func Test_FetchCarInfo2(t *testing.T) {
	var surl string
	surl = "https://car.autohome.com.cn/config/series/528.html"
	info, err := fetchCarInfo(surl)
	if err != nil {
		fmt.Println("Error:",err)
		return
	}
	fmt.Println(info)
}

func Test_DecodeAutoHomeDictionary(t *testing.T) {
	var (
		_ = `testsss`
		js = `<script>(function(Ka_){            var $style$ = Ka_.createElement('style');            if (Ka_.head) {                Ka_.head.appendChild($style$);            } else {                Ka_.getElementsByTagName('head')[0].appendChild($style$);            }            var $sheet$ = $style$.sheet;                                                                                      function $Innerhtml$ ($item$, $index$){                var $tempArray$ = $GetElementsByCss$($GetClassName$($item$));                for (x in $tempArray$) {                    $tempArray$[x].innerHTML = $index$;                    try {                        $tempArray$[x].currentStyle = '';                    } catch (e) {                    }                }            }                    var AD_=function(AD__){'return AD_';return AD__;};  var AJ_=function(){'AJ_';var _A=function(){return '构';}; return _A();};  var Ap_=function(){'return Ap_';return 'g';};  var BJ_=function(){'BJ_';var _B=function(){return 'p';}; return _B();};  var BN_=function(){'return BN_';return '6';};  var BS_=function(BS__){var _B=function(BS__){'return BS_';return BS__;}; return _B(BS__);};  var BW_=function(BW__){'return BW_';return BW__;};  var Bi_=function(){'Bi_';var _B=function(){return '97,';}; return _B();};  var CE_=function(CE__){'return CE_';return CE__;};  var CF_=function(){'return CF_';return ',26';};             function $GetWindow$ ()            {                return this[''+(function(){'return DG_';return 'w'})()+JZ_+uk_()+jW_()+Hu_()];            }              var CP_=function(){'CP_';var _C=function(){return '6;58,36';}; return _C();};  var CT_=function(){'CT_';var _C=function(){return '锁';}; return _C();};  var Cp_=function(){'return Cp_';return '导';};  var Ct_=function(){'return Ct_';return 'w';};  var Cy_=function(){'return Cy_';return '5';};  var Dw_=function(){'Dw_';var _D=function(){return 't';}; return _D();};  var EW_=function(){'return EW_';return '4';};  var Fb_=function(Fb__){'return Fb_';return Fb__;};  var Fc_=function(){'Fc_';var _F=function(){return ';';}; return _F();};  var GF_=function(GF__){var _G=function(GF__){'return GF_';return GF__;}; return _G(GF__);};  var GM_=function(){'return GM_';return '132;2';};  var Go_=function(){'return Go_';return '支放数';};            function $Split$ ($item$, $index$)           {                if ($item$) {                     return $item$[''+uA_()+FI_()+ak_('li')+Ya_()] ($index$);                } else {                    return '';                }            }              var HA_=function(){'HA_';var _H=function(){return '1;152';}; return _H();};  var He_=function(){'return He_';return '力功加';};  var Hh_=function(Hh__){'return Hh_';return Hh__;};  var Hu_=function(){'Hu_';var _H=function(){return 'ow';}; return _H();};  var Ia_=function(Ia__){'return Ia_';return Ia__;};  var Iw_=function(){'return Iw_';return '8;1';};  var JI_=function(JI__){var _J=function(JI__){'return JI_';return JI__;}; return _J(JI__);};  var JK_=function(JK__){var _J=function(JK__){'return JK_';return JK__;}; return _J(JK__);};  var KE_=function(){'return KE_';return '6';};  var Kb_=function(Kb__){var _K=function(Kb__){'return Kb_';return Kb__;}; return _K(Kb__);};  var Ki_=function(){'Ki_';var _K=function(){return 'd';}; return _K();};  var Kj_=function(Kj__){var _K=function(Kj__){'return Kj_';return Kj__;}; return _K(Kj__);};  var Kp_=function(){'Kp_';var _K=function(){return 't';}; return _K();};  var Lg_=function(Lg__){'return Lg_';return Lg__;};  var MR_=function(){'return MR_';return 'n';};  var NI_=function(NI__){var _N=function(NI__){'return NI_';return NI__;}; return _N(NI__);};  var NO_=function(){'NO_';var _N=function(){return '驶';}; return _N();};  var NT_=function(NT__){'return NT_';return NT__;};  var PI_=function(PI__){var _P=function(PI__){'return PI_';return PI__;}; return _P(PI__);};  var PM_=function(PM__){var _P=function(PM__){'return PM_';return PM__;}; return _P(PM__);};  var PU_=function(PU__){'return PU_';return PU__;};  var QU_=function(QU__){'return QU_';return QU__;};  var Rd_=function(){'return Rd_';return '140';};             function $GetElementsByCss$ ($item$) {                 return document.querySelectorAll($item$);            }              var Rm_=function(){'return Rm_';return 'e';};  var Te_=function(){'Te_';var _T=function(){return ',120;2,';}; return _T();};  var UE_=function(){'return UE_';return ';';};  var UF_=function(UF__){var _U=function(UF__){'return UF_';return UF__;}; return _U(UF__);};  var UO_=function(){'return UO_';return '0;126,7';};  var UZ_=function(){'UZ_';var _U=function(){return '20,33;0';}; return _U();};  var VR_=function(){'return VR_';return '1';};  var WT_=function(WT__){var _W=function(WT__){'return WT_';return WT__;}; return _W(WT__);};  var Wr_=function(Wr__){'return Wr_';return Wr__;};  var XI_=function(XI__){'return XI_';return XI__;};  var XP_=function(){'return XP_';return 't';};  var Xe_=function(){'Xe_';var _X=function(){return '燃';}; return _X();};  var Xq_=function(Xq__){'return Xq_';return Xq__;};  var Xs_=function(Xs__){var _X=function(Xs__){'return Xs_';return Xs__;}; return _X(Xs__);};  var Xx_=function(){'Xx_';var _X=function(){return '7';}; return _X();};  var YS_=function(){'return YS_';return ';16,5';};  var Yt_=function(){'return Yt_';return '风';};  var ZE_=function(ZE__){'return ZE_';return ZE__;};  var ZR_=function(){'ZR_';var _Z=function(){return '大天央';}; return _Z();};  var Zf_=function(Zf__){'return Zf_';return Zf__;};  var Zk_=function(){'return Zk_';return '1';};             var $ruleDict$ = '';              var aZ_=function(){'aZ_';var _a=function(){return ',';}; return _a();};  var ab_=function(){'return ab_';return '宽';};  var ak_=function(ak__){var _a=function(ak__){'return ak_';return ak__;}; return _a(ak__);};             var $rulePosList$ = '';                         function $FillDicData$ () {                  $ruleDict$ = $GetWindow$()[''+Ki_()+Rm_()+sf_()+'od'+(function(){'return kJ_';return 'e'})()+VZ_+Bf_()+UF_('om')+pB_()+nx_()+ry_()+XP_()](''+ZE_('中主')+Zf_('仪价')+EQ_()+(function(){'return hg_';return '供'})()+GF_('保倒')+Lg_('像儿')+wN_()+sS_+my_()+JC_()+pp_()+He_()+(function(){'return Qs_';return '动'})()+dH_()+xu_()+'叭号'+YC_+IG_+zW_()+jO_()+MO_()+mp_()+EU_()+ZR_()+jm_()+Kd_+ab_()+Cp_()+tV_('差并')+kU_()+'影径'+(function(){'return sp_';return '悬'})()+(function(){'return NW_';return '成'})()+cj_()+wE_()+xx_()+Go_()+Ao_()+(function(){'return FQ_';return '最机材'})()+AJ_()+rn_()+Ba_()+BW_('桥椅')+dr_()+tx_()+Ms_()+NI_('测液')+LA_()+Xe_()+hn_()+OA_()+sA_()+sm_()+(function(TW__){'return TW_';return TW__;})('矩碟')+pe_()+(function(Zl__){'return Zl_';return Zl__;})('积称')+vR_()+bB_()+(function(){'return Dj_';return '稳'})()+Jx_()+qV_()+(function(Ad__){'return Ad_';return Ad__;})('立童')+eR_()+(function(){'return qA_';return '线综缩'})()+GC_()+Aq_()+PF_()+(function(){'return pZ_';return '脑'})()+'节蓝'+aX_+gn_()+OH_()+Dh_()+rv_()+vV_+(function(){'return II_';return '轴进适'})()+da_()+'配量'+Lk_+CT_()+(function(ES__){'return ES_';return ES__;})('长门')+eu_+Ok_()+(function(){'return Zc_';return '音'})()+PU_('预频')+Yt_()+NO_()+Ly_()+$SystemFunction1$(''));                  $rulePosList$=$Split$(($SystemFunction1$('')+''+ha_()+nb_('38;64,')+Wx_()+uM_()+Yd_()+(function(){'return Cx_';return '04,'})()+MV_()+(function(){'return OQ_';return '9,151;8'})()+bm_+GY_()+(function(Qc__){'return Qc_';return Qc__;})(';69,37')+IJ_()+xs_()+zu_('140;15')+cy_()+(function(Eh__){'return Eh_';return Eh__;})(',1')+PE_()+'13,141'+mD_()+WJ_()+(function(){'return As_';return '9;82,'})()+pk_()+AH_()+cU_()+(function(){'return KF_';return '33,148,'})()+xN_+(function(){'return xv_';return '40,146;'})()+Qz_+rC_('9,90')+cO_+ts_()+(function(){'return HV_';return ','})()+vO_()+zS_()+Gm_()+Fo_+wo_()+zk_()+WT_(';5,8')+Ma_+vk_()+(function(){'return dS_';return '1'})()+(function(){'return zl_';return '41,'})()+(function(){'return ZS_';return '82;16,1'})()+Wh_()+fG_+(function(kE__){'return kE_';return kE__;})('47,51;')+'57,1'+yW_()+bc_()+tk_()+Yp_()+pR_()+KX_()+BS_('124,')+Ty_()+Bi_()+fQ_()+Bh_()+JW_()+wd_()+UX_+PM_('2;150,')+QU_('89')+Ed_+cZ_()+UZ_()+vI_()+EI_()+eL_('97,43;')+DT_()+JI_(';2')+ez_()+YY_()+hb_('1,14')+DC_+ec_()+Wr_(';155;6')+xW_()+gW_(',25;11')+(function(zq__){'return zq_';return zq__;})('0,82')+YS_()+BN_()+Hj_()+Du_()+cQ_()+(function(){'return WM_';return ','})()+(function(){'return yl_';return '51;'})()+yQ_()+(function(){'return uC_';return ',13'})()+lz_()+Te_()+Ph_()+UO_()+Iw_()+Xs_('6,')+KE_()+HA_()+PI_(',88;3;')+(function(){'return vJ_';return '1'})()+yr_()+(function(Uj__){'return Uj_';return Uj__;})(',132')+mK_()+(function(Ff__){'return Ff_';return Ff__;})('50,1')+lX_('15;60;')+vs_()+Ds_+DV_()+lU_()+yc_()+(function(){'return lf_';return '8,4'})()+IN_()+Fb_(',51;')+(function(){'return GE_';return '83,82;2'})()+Xq_('1,17')+UE_()+Xv_()+wv_()+Kb_('08;82,')+Ej_()+(function(la__){'return la_';return la__;})(',142;1')+Xm_()+NT_('18,55;')+rN_('38,62;')+qt_()+Fc_()+vC_()+(function(){'return iB_';return '7,129'})()+sn_()+(function(){'return ZU_';return '3'})()+(function(){'return lM_';return '2'})()+CF_()+Ri_+(function(){'return Ye_';return ',61;66,'})()+HB_()+EW_()+(function(){'return cW_';return ',11'})()+(function(){'return Yf_';return '2'})()+oh_()+Jn_()+tl_('1;')+xY_()+AD_('5;46,1')+iX_()+Cy_()+zK_()+CV_()+Xx_()+QW_()+Vh_()+(function(Bw__){'return Bw_';return Bw__;})('86;9')+'3,53'+Sq_()+ed_()+Tn_()+to_()+pr_()+aZ_()+VT_+CL_()+(function(fk__){'return fk_';return fk__;})('0,10')+(function(pz__){'return pz_';return pz__;})('7,20')+bC_()+YD_+(function(Dl__){'return Dl_';return Dl__;})('9,113,')+dl_+jK_()+fI_('137,82')+';29,10'+CP_()+Ga_+Yl_()+pT_+Ae_+Ry_()+VR_()+(function(){'return YX_';return ','})()+rM_()+XI_(',118;1')+gk_()+oQ_()+kp_()+eq_()+(function(){'return NC_';return '1'})()+NP_()+sT_()+ux_()+hT_()+tL_()+sX_()+FK_()+qx_()+'5,20,3'+(function(){'return OC_';return '3;24,'})()+eT_()+yV_()+Ra_+Hh_('3,10')+Wc_()+Rd_()+uZ_()+nn_()+Kj_('3;139,')+Zk_()+JK_('55;3')+Zn_()+GM_()+hV_()),$SystemFunction2$(';'));                  $imgPosList$=$Split$(('##imgPosList_jsFuns##'+$SystemFunction2$(';')),$SystemFunction1$(';'));                  $RenderToHTML$();                  return ';';            }              var bB_=function(){'bB_';var _b=function(){return '程';}; return _b();};  var bC_=function(){'return bC_';return ';153,12';};  var bQ_=function(){'bQ_';var _b=function(){return 'p';}; return _b();};  var dr_=function(){'return dr_';return '比气氙';};  var eL_=function(eL__){'return eL_';return eL__;};  var eR_=function(){'eR_';var _e=function(){return '箱';}; return _e();};  var ec_=function(){'ec_';var _e=function(){return '130,122';}; return _e();};  var ed_=function(){'return ed_';return '3';};  var fI_=function(fI__){'return fI_';return fI__;};  var gW_=function(gW__){var _g=function(gW__){'return gW_';return gW__;}; return _g(gW__);};  var gn_=function(){'gn_';var _g=function(){return '视警话';}; return _g();};  var hb_=function(hb__){'return hb_';return hb__;};  var hn_=function(){'hn_';var _h=function(){return '牙牵独';}; return _h();};  var jW_=function(){'jW_';var _j=function(){return 'd';}; return _j();};  var lU_=function(){'return lU_';return '9,102;1';};  var lX_=function(lX__){'return lX_';return lX__;};  var lz_=function(){'lz_';var _l=function(){return '5';}; return _l();};  var nb_=function(nb__){'return nb_';return nb__;};  var oh_=function(){'return oh_';return ';11,8';};  var pB_=function(){'pB_';var _p=function(){return 'pon';}; return _p();};  var pR_=function(){'pR_';var _p=function(){return ';40,111';}; return _p();};             function $RenderToHTML$ ()            {                 $InsertRuleRun$();            }              var pf_=function(){'return pf_';return 'e';};  var pp_=function(){'return pp_';return '前';};  var pr_=function(){'return pr_';return '5';};  var qV_=function(){'return qV_';return '窗';};  var rC_=function(rC__){'return rC_';return rC__;};  var rN_=function(rN__){var _r=function(rN__){'return rN_';return rN__;}; return _r(rN__);};  var sT_=function(){'sT_';var _s=function(){return '2;1,2';}; return _s();};             var $imgPosList$ = '';              var sm_=function(){'return sm_';return '盖盘真';};  var sn_=function(){'sn_';var _s=function(){return ';';}; return _s();};  var tL_=function(){'return tL_';return '7,84;';};  var tV_=function(tV__){'return tV_';return tV__;};  var tk_=function(){'return tk_';return '116,2';};  var tl_=function(tl__){'return tl_';return tl__;};  var ts_=function(){'return ts_';return '16;69';};  var ue_=function(ue__){'return ue_';return ue__;};  var uk_=function(){'return uk_';return 'n';};  var vO_=function(){'return vO_';return '4;157';};  var vk_=function(){'vk_';var _v=function(){return ';';}; return _v();};  var wo_=function(){'wo_';var _w=function(){return '85,11';}; return _w();};  var wv_=function(){'return wv_';return '4,1';};  var yW_=function(){'return yW_';return '8';};  var zW_=function(){'return zW_';return '喇';};  var zk_=function(){'zk_';var _z=function(){return '9';}; return _z();};  var zu_=function(zu__){'return zu_';return zu__;}; function AH_(){function _A(){return 'AH__';};if(_A()=='AH__'){ return '8,68;';}else{ return _A();}} function Ao_(){'return Ao_';return '整无晶';} function Aq_(){function _A(){return 'Aq__';};if(_A()=='Aq__'){ return '耗';}else{ return _A();}} function BP_(){'return BP_';return 'a';} function Ba_(){function _B(){return '标格';};if(_B()=='标格,'){ return 'Ba_';}else{ return _B();}} function Bf_(){function _B(){return 'Bf__';};if(_B()=='Bf__'){ return 'RIC';}else{ return _B();}} function Bh_(){function _B(){return 'Bh_';};if(_B()=='Bh__'){ return _B();}else{ return '21';}}             function $SystemFunction1$ ($item$)            {                 $ResetSystemFun$();                 if ($GetWindow$()[''+JL_()+UA_+bQ_()+zR_()+JD_()+Wf_()+gs_+pf_()] != undefined) {                     $GetWindow$()[''+JL_()+UA_+bQ_()+zR_()+JD_()+Wf_()+gs_+pf_()] = function(element, pseudoElt)                     {                         if (pseudoElt != undefined && typeof(pseudoElt) == 'string' && pseudoElt.toLowerCase().indexOf(':before') > -1) {                             var obj = {};obj.getPropertyValue = function (x) { return x; };return obj;                         } else {                             return window.hs_fuckyou(element, pseudoElt);                         }                     };                 }                 return $item$;            }             function CL_(){function _C(){return ';52,8';};if(_C()==';52,8'){ return ';52,8';}else{ return _C();}} function CV_(){function _C(){return 'CV__';};if(_C()=='CV__'){ return '101,98;';}else{ return _C();}} function DT_(){function _D(){return 'DT__';};if(_D()=='DT__'){ return '18,95';}else{ return _D();}} function DV_(){function _D(){return 'DV__';};if(_D()=='DV__'){ return '2;5';}else{ return _D();}} function Dh_(){function _D(){return 'Dh__';};if(_D()=='Dh__'){ return '质距车';}else{ return _D();}} function Du_(){'return Du_';return '5';} function EI_(){function _E(){return 'EI__';};if(_E()=='EI__'){ return ',120;';}else{ return _E();}} function EQ_(){function _E(){return '体';};if(_E()=='体'){ return '体';}else{ return _E();}} function EU_(){'return EU_';return '外';} function Ej_(){function _E(){return 'Ej__';};if(_E()=='Ej__'){ return '34;46';}else{ return _E();}} function FI_(){function _F(){return 'p';};if(_F()=='p'){ return 'p';}else{ return _F();}} function FK_(){function _F(){return '74,1';};if(_F()=='74,1,'){ return 'FK_';}else{ return _F();}} function FS_(){'return FS_';return 'u';} function GC_(){function _G(){return 'GC__';};if(_G()=='GC__'){ return '缸';}else{ return _G();}} function GY_(){function _G(){return '0,79';};if(_G()=='0,79,'){ return 'GY_';}else{ return _G();}} function Gm_(){function _G(){return ',56,76;';};if(_G()==',56,76;'){ return ',56,76;';}else{ return _G();}} function HB_(){function _H(){return 'HB_';};if(_H()=='HB__'){ return _H();}else{ return '65;9';}} function Hj_(){'return Hj_';return ',76;1';} function IJ_(){'return IJ_';return ';1';} function IN_(){function _I(){return 'IN_';};if(_I()=='IN__'){ return _I();}else{ return ';147';}} function Ix_(){function _I(){return 'e';};if(_I()=='e'){ return 'e';}else{ return _I();}} function JC_(){function _J(){return '列制';};if(_J()=='列制,'){ return 'JC_';}else{ return _J();}} function JD_(){function _J(){return 'dSt';};if(_J()=='dSt'){ return 'dSt';}else{ return _J();}} function JL_(){function _J(){return 'get';};if(_J()=='get'){ return 'get';}else{ return _J();}} function JW_(){function _J(){return ';118,';};if(_J()==';118,'){ return ';118,';}else{ return _J();}}            function $ChartAt$ ($item$)           {                 return $ruleDict$[''+Yy_()+ue_('rA')+Kp_()] (parseInt($item$));           }            function Jn_(){function _J(){return 'Jn__';};if(_J()=='Jn__'){ return '7,7';}else{ return _J();}} function Jw_(){function _J(){return 'Jw_';};if(_J()=='Jw__'){ return _J();}else{ return 'de';}} function Jx_(){'return Jx_';return '空';} function KK_(){function _K(){return 'loc';};if(_K()=='loc'){ return 'loc';}else{ return _K();}} function KX_(){function _K(){return ';56,76;';};if(_K()==';56,76;'){ return ';56,76;';}else{ return _K();}} function LA_(){function _L(){return '源滑热';};if(_L()=='源滑热'){ return '源滑热';}else{ return _L();}}             function $SuperInsertRule$ () {                if ($sheet$ !== undefined && $sheet$[''+Ia_('in')+Hz_+Ix_()+(function(fV__){'return fV_';return fV__;})('rt')+xQ_()+CE_('ul')+fN_]) {                    return true;                } else {                    return false;                }            }             function Ly_(){'return Ly_';return '驻驾高';} function MO_(){function _M(){return '囊';};if(_M()=='囊'){ return '囊';}else{ return _M();}} function MV_(){function _M(){return 'MV__';};if(_M()=='MV__'){ return '35,14';}else{ return _M();}} function Mn_(){'return Mn_';return 'V';} function Ms_(){function _M(){return 'Ms__';};if(_M()=='Ms__'){ return '油';}else{ return _M();}} function NP_(){function _N(){return 'NP__';};if(_N()=='NP__'){ return ',14';}else{ return _N();}} function Ni_(){'return Ni_';return 'l';} function OA_(){'return OA_';return '率环';} function OH_(){function _O(){return 'OH__';};if(_O()=='OH__'){ return '调';}else{ return _O();}} function Ok_(){function _O(){return 'Ok_';};if(_O()=='Ok__'){ return _O();}else{ return '限隙';}} function PE_(){function _P(){return 'PE_';};if(_P()=='PE__'){ return _P();}else{ return '7,';}}             function $GetLocationURL$ ()            {                return $GetWindow$()[''+KK_()+'at'+YE_()+MR_()][''+zh_+Tb_()+(function(){'return Vo_';return 'ef'})()];            }             function PF_(){function _P(){return 'PF__';};if(_P()=='PF__'){ return '胎';}else{ return _P();}} function Ph_(){'return Ph_';return '125,10';} function QW_(){'return QW_';return '0,144,2';} function Ry_(){'return Ry_';return ',154;13';} function Sq_(){'return Sq_';return ',17,6';} function Tb_(){function _T(){return 'Tb__';};if(_T()=='Tb__'){ return 'r';}else{ return _T();}} function Tn_(){function _T(){return 'Tn_';};if(_T()=='Tn__'){ return _T();}else{ return ',15;16';}} function Ty_(){function _T(){return '133,';};if(_T()=='133,,'){ return 'Ty_';}else{ return _T();}} function Vh_(){function _V(){return 'Vh__';};if(_V()=='Vh__'){ return '2;45,';}else{ return _V();}} function WJ_(){function _W(){return 'WJ_';};if(_W()=='WJ__'){ return _W();}else{ return '72,3';}} function Wc_(){function _W(){return '3;49,';};if(_W()=='3;49,'){ return '3;49,';}else{ return _W();}} function Wf_(){function _W(){return 'y';};if(_W()=='y'){ return 'y';}else{ return _W();}} function Wh_(){'return Wh_';return '35';} function Wx_(){function _W(){return '8';};if(_W()=='8'){ return '8';}else{ return _W();}} function Xm_(){function _X(){return '0;1';};if(_X()=='0;1'){ return '0;1';}else{ return _X();}} function Xv_(){function _X(){return 'Xv_';};if(_X()=='Xv__'){ return _X();}else{ return '12';}} function YE_(){function _Y(){return 'YE_';};if(_Y()=='YE__'){ return _Y();}else{ return 'io';}} function YY_(){'return YY_';return '6';} function Ya_(){function _Y(){return 'Ya__';};if(_Y()=='Ya__'){ return 't';}else{ return _Y();}} function Yd_(){function _Y(){return ';1';};if(_Y()==';1,'){ return 'Yd_';}else{ return _Y();}} function Yl_(){'return Yl_';return '33;1';} function Yp_(){function _Y(){return 'Yp__';};if(_Y()=='Yp__'){ return '8';}else{ return _Y();}} function Yy_(){function _Y(){return 'cha';};if(_Y()=='cha'){ return 'cha';}else{ return _Y();}} function ZL_(){function _Z(){return 'aul';};if(_Z()=='aul'){ return 'aul';}else{ return _Z();}} function Zn_(){function _Z(){return 'Zn_';};if(_Z()=='Zn__'){ return _Z();}else{ return '0,135,';}} function bc_(){'return bc_';return ';';} function cQ_(){'return cQ_';return '9';} function cU_(){function _c(){return 'cU__';};if(_c()=='cU__'){ return '1';}else{ return _c();}} function cZ_(){'return cZ_';return ';16,15,';} function cj_(){function _c(){return '扬扭';};if(_c()=='扬扭,'){ return 'cj_';}else{ return _c();}} function cy_(){function _c(){return ',20';};if(_c()==',20'){ return ',20';}else{ return _c();}} function dH_(){function _d(){return '助匙';};if(_d()=='助匙,'){ return 'dH_';}else{ return _d();}} function da_(){'return da_';return '通速';} function eT_(){'return eT_';return '1';} function eq_(){function _e(){return '9;13';};if(_e()=='9;13,'){ return 'eq_';}else{ return _e();}} function ez_(){function _e(){return 'ez_';};if(_e()=='ez__'){ return _e();}else{ return '8,143;';}} function fQ_(){function _f(){return 'fQ__';};if(_f()=='fQ__'){ return '1';}else{ return _f();}} function gk_(){function _g(){return '58,156;';};if(_g()=='58,156;'){ return '58,156;';}else{ return _g();}} function hT_(){'return hT_';return ';9';} function hV_(){function _h(){return '7;48';};if(_h()=='7;48,'){ return 'hV_';}else{ return _h();}} function ha_(){'return ha_';return '96,6;1';} function iH_(){'return iH_';return 'Vie';} function iX_(){'return iX_';return '0';} function jK_(){function _j(){return ';73,75;';};if(_j()==';73,75;'){ return ';73,75;';}else{ return _j();}} function jO_(){function _j(){return '器';};if(_j()=='器'){ return '器';}else{ return _j();}} function jm_(){'return jm_';return '头子';} function kU_(){'return kU_';return '度座引';} function kp_(){'return kp_';return '7';}             function $SystemFunction2$ ($item$)            {                 $ResetSystemFun$();                 if ($GetDefaultView$()) {                     if ($GetDefaultView$()[''+JL_()+UA_+bQ_()+zR_()+JD_()+Wf_()+gs_+pf_()] != undefined) {                          $GetDefaultView$()[''+JL_()+UA_+bQ_()+zR_()+JD_()+Wf_()+gs_+pf_()] = function(element, pseudoElt){                                 if (pseudoElt != undefined && typeof(pseudoElt) == 'string' && pseudoElt.toLowerCase().indexOf(':before') > -1) {                                     var obj = {};                                     obj.getPropertyValue = function(x) {                                         return x;                                     };                                     return obj;                        } else {                            return window.hs_fuckyou_dd(element, pseudoElt);                        }                    };                }             }            return $item$;           }            function mD_(){function _m(){return ';';};if(_m()==';'){ return ';';}else{ return _m();}} function mK_(){function _m(){return ';';};if(_m()==';'){ return ';';}else{ return _m();}} function mp_(){function _m(){return '地声备';};if(_m()=='地声备'){ return '地声备';}else{ return _m();}} function my_(){function _m(){return '分';};if(_m()=='分'){ return '分';}else{ return _m();}} function nP_(){'return nP_';return 'f';} function nn_(){function _n(){return '3';};if(_n()=='3'){ return '3';}else{ return _n();}} function nx_(){function _n(){return 'e';};if(_n()=='e'){ return 'e';}else{ return _n();}} function oQ_(){function _o(){return '16,';};if(_o()=='16,'){ return '16,';}else{ return _o();}} function pe_(){function _p(){return 'pe__';};if(_p()=='pe__'){ return '离';}else{ return _p();}} function pk_(){function _p(){return 'pk_';};if(_p()=='pk__'){ return _p();}else{ return '14';}} function qg_(){function _q(){return 'e';};if(_q()=='e'){ return 'e';}else{ return _q();}} function qt_(){function _q(){return 'qt_';};if(_q()=='qt__'){ return _q();}else{ return '31';}} function qx_(){function _q(){return 'qx__';};if(_q()=='qx__'){ return '31;30,1';}else{ return _q();}} function rM_(){function _r(){return '6;82';};if(_r()=='6;82,'){ return 'rM_';}else{ return _r();}} function rn_(){function _r(){return '架';};if(_r()=='架'){ return '架';}else{ return _r();}} function rv_(){function _r(){return 'rv__';};if(_r()=='rv__'){ return '转';}else{ return _r();}} function ry_(){function _r(){return 'n';};if(_r()=='n'){ return 'n';}else{ return _r();}} function sA_(){'return sA_';return '电皮';} function sX_(){function _s(){return 'sX__';};if(_s()=='sX__'){ return '109,44;';}else{ return _s();}} function sf_(){function _s(){return 'sf__';};if(_s()=='sf__'){ return 'c';}else{ return _s();}} function to_(){function _t(){return ',13';};if(_t()==',13'){ return ',13';}else{ return _t();}} function tx_(){'return tx_';return '池';} function uA_(){'return uA_';return 's';} function uM_(){function _u(){return ',42';};if(_u()==',42'){ return ',42';}else{ return _u();}}             function $InsertRule$ ($index$, $item$){                 $sheet$[''+Ia_('in')+Hz_+Ix_()+(function(fV__){'return fV_';return fV__;})('rt')+xQ_()+CE_('ul')+fN_]($GetClassName$($index$) + $RuleCalss1$()+'"' + $item$ + '" }', 0);                 var $tempArray$ = $GetElementsByCss$($GetClassName$($index$));                 for (x in $tempArray$) {                    try {                        $tempArray$[x].currentStyle = '';                    } catch (e) {                    }                  }            }             function uZ_(){function _u(){return ',';};if(_u()==','){ return ',';}else{ return _u();}} function ux_(){function _u(){return 'ux__';};if(_u()=='ux__'){ return '0';}else{ return _u();}} function vC_(){'return vC_';return '9';}             function $InsertRuleRun$ () {                for ($index$ = 0; $index$ < $rulePosList$.length; $index$++) {                    var $tempArray$ = $Split$($rulePosList$[$index$], ',');                    var $temp$ = '';                    for ($itemIndex$ = 0; $itemIndex$ < $tempArray$.length; $itemIndex$++) {                        $temp$ += $ChartAt$($tempArray$[$itemIndex$]) + '';                    }                    $InsertRule$($index$, $temp$);                }            }             function vI_(){function _v(){return ',41;37';};if(_v()==',41;37,'){ return 'vI_';}else{ return _v();}} function vR_(){function _v(){return '移';};if(_v()=='移'){ return '移';}else{ return _v();}} function vs_(){function _v(){return '123,';};if(_v()=='123,,'){ return 'vs_';}else{ return _v();}} function wE_(){'return wE_';return '指排接';} function wN_(){function _w(){return '元全';};if(_w()=='元全,'){ return 'wN_';}else{ return _w();}} function wd_(){function _w(){return 'wd_';};if(_w()=='wd__'){ return _w();}else{ return '99';}} function xQ_(){'return xQ_';return 'R';} function xW_(){'return xW_';return '2';} function xY_(){function _x(){return '91,8';};if(_x()=='91,8,'){ return 'xY_';}else{ return _x();}}             function $ResetSystemFun$ () {                if ($GetWindow$()[''+JL_()+UA_+bQ_()+zR_()+JD_()+Wf_()+gs_+pf_()] != undefined) {                    if (window.hs_fuckyou == undefined) {                        window.hs_fuckyou = $GetWindow$()[''+JL_()+UA_+bQ_()+zR_()+JD_()+Wf_()+gs_+pf_()];                    }                }                if ($GetDefaultView$()) {                    if ($GetDefaultView$()[''+JL_()+UA_+bQ_()+zR_()+JD_()+Wf_()+gs_+pf_()] != undefined) {                        if (window.hs_fuckyou_dd == undefined) {                            window.hs_fuckyou_dd = $GetDefaultView$()[''+JL_()+UA_+bQ_()+zR_()+JD_()+Wf_()+gs_+pf_()];                        }                    }                }            }             function xs_(){function _x(){return '34,';};if(_x()=='34,'){ return '34,';}else{ return _x();}} function xu_(){'return xu_';return '单压口';} function xx_(){function _x(){return 'xx__';};if(_x()=='xx__'){ return '控摄撑';}else{ return _x();}} function yQ_(){function _y(){return '145;30';};if(_y()=='145;30,'){ return 'yQ_';}else{ return _y();}} function yV_(){'return yV_';return '17,81;';} function yc_(){'return yc_';return '1';} function yr_(){function _y(){return '36';};if(_y()=='36,'){ return 'yr_';}else{ return _y();}} function zK_(){function _z(){return ';';};if(_z()==';'){ return ';';}else{ return _z();}} function zR_(){function _z(){return 'ute';};if(_z()=='ute'){ return 'ute';}else{ return _z();}} function zS_(){function _z(){return ',133;30';};if(_z()==',133;30'){ return ',133;30';}else{ return _z();}} var Ae_='7'; var DC_=';'; var Ds_='9';             function $GetCustomStyle$ ()            {                var $customstyle$ = '';                try {                    if (HS_GetCustomStyle) {                        $customstyle$ = HS_GetCustomStyle();                    } else {                        if (navigator.userAgent.indexOf('Windows NT 5') != -1) {                            $customstyle$ = 'margin-bottom:-4.8px;';                        } else {                            $customstyle$ = 'margin-bottom:-5px;';                        }                    }                } catch (e) {                }                return $customstyle$;            }                         function $GetDefaultView$ ()            {                return Ka_[''+Jw_()+nP_()+ZL_()+Dw_()+iH_()+Ct_()];            }             var Ed_=';67,7'; var Fo_='61,142;'; var Ga_=','; var Hz_='s'; var IG_='商'; var JZ_='i'; var Kd_='定实容';             function $GetClassName$ ($index$)            {                 return '.hs_kw' + $index$ + '_baikeOl';            }            function $RuleCalss1$ ()            {                return '::before { content:'            }             var Lk_='金钥铝'; var Ma_='5'; var Qz_='1'; var Ra_='2'; var Ri_=';30'; var UA_='Com'; var UX_=';77,1'; var VT_='120'; var VZ_='U'; var XV_='etP'; var YC_='合名后'; var YD_='8;54,8;'; var aX_='行表规'; var bm_='5,114;3'; var cO_=';'; var dl_='52,80'; var eu_='间'; var fG_=',132;'; var fN_='e'; var gs_='l'; var pT_='2'; var sS_='准'; var vV_='轮'; var xN_='68;49,1'; var zh_='h';             var Lw_= $FillDicData$('rI_'); var mI_='_'; var Pg_=';'; function lx_(){function _l(){return '332';};if(_l()=='332'){ return '332';}else{ return _l();}}  function RJ_(){function _R(){return 'RJ__';};if(_R()=='RJ__'){ return '802';}else{ return _R();}}  var XF_=function(XF__){var _X=function(XF__){'return XF_';return XF__;}; return _X(XF__);};  var Uh_=function(Uh__){var _U=function(Uh__){'return Uh_';return Uh__;}; return _U(Uh__);};})(document);</script>`
	)

	getAutoHomeDict(js)
}

func Test_Reflect(t *testing.T) {
	key := "brand_id"
	value_str := "宝马 X5"
	value_int := 123
	car := NewAutoHomeCar(111123)

	//r_type := reflect.TypeOf(car)
	r_value := reflect.ValueOf(car)
	if r_value.Kind() == reflect.Ptr {
		r_value = r_value.Elem()
	}

	if r_value.Kind() != reflect.Struct {
		return
	}

	field := r_value.FieldByName(key)
	if !field.CanSet() {
		fmt.Println(key,"can not set")
		return
	}
	switch field.Kind() {
	case reflect.String:
		field.SetString(value_str)
	case reflect.Int:
		field.SetInt(int64(value_int))
	}
	fmt.Println(r_value)
}