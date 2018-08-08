package main

import "strconv"

type Brand struct {
	Brand_id 		int			`orm:"pk";pk:"auto"`
	Brand_name		string
	Brand_initial	string		`orm:"null"`
	Brand_logo		string		`orm:"null"`
}

type CarSeries struct {
	Series_id		int			`orm:"pk";pk:"auto"`
	Series_name		string
	Brand_id		int
	Series_home		string
	Series_config	string
	Status			string
}

type CarCrawl struct {
	Id					int			`orm:"pk";pk:"auto"`
	Type_id				int			`汽车之家车型id`
	settings			string
	Car_name			string								`车型名称`
	Car_type			int									`0/1是否显示电动机表`
	Series_id			int
	Series_name			string								`车系名称`
	Brand_id			int
	Brand_name			string								`品牌名称`
	//Country_id			int
	Country_str			string								`国家`
	//Produce_type		int
	Produce_type_str	string								`生产方式`
	Manufacturer		string								`厂商`
	//Car_level			int
	Car_level_str		string								`级别`
	//Market_price		float64
	Market_price_str	string								`厂商指导价(元)`
	//Energy_type			int
	Energy_type_str		string								`能源类型`
	Market_time			string								`上市时间`
	Engine				string								`发动机`
	Gearbox				string								`变速箱`
	Car_size			string								`长*宽*高(mm)`
	Car_struct			string								`车身结构`
	Max_speed			string								`最高车速(km/h)`
	Official_speedup	string								`官方0-100km/h加速(s)`
	Actual_speedup		string								`实测0-100km/h加速(s)`
	Actual_brake		string								`实测100-0km/h制动(m)`
	Actual_fueluse		string								`实测油耗(L/100km)`
	Gerenal_fueluse		string								`工信部综合油耗(L/100km)`
	Actual_ground		string								`实测离地间隙(mm)`
	Quality_guarantee	string								`整车质保`
	Max_power			string								`最大功率(kW)`
	Max_torque			string								`最大扭矩(N・m)`
	E_mileage			string								`工信部纯电续驶里程(km)`
	Length                 string                           `长度(mm)`
	Width                  string                           `宽度(mm)`
	Height                 string                           `高度(mm)`
	Shaft_distance         string                           `轴距(mm)`
	Front_wheels_gap       string                           `前轮距(mm)`
	Back_wheels_gap        string                           `后轮距(mm)`
	Min_ground             string                           `最小离地间隙(mm)`
	Body_struct            string                           `车身结构`
	Doors                  string                           `车门数(个)`
	Seats                  string                           `座位数(个)`
	Fuel_vol               string                           `油箱容积(L)`
	Cargo_vol              string                           `行李厢容积(L)`
	Open_type              string                           `后排车门开启方式`
	Cargo_size             string                           `货箱尺寸(mm)`
	Total_weight           string                           `整备质量(kg)`
	Carry_cap              string                           `最大载重质量(kg)`
	Engine_type            string                           `发动机型号`
	Cc                     string                           `排量(mL)`
	Air_intake             string                           `进气形式`
	Cylinder_arrange       string                           `气缸排列形式`
	Cylinders              string                           `气缸数(个)`
	Valves                 string                           `每缸气门数(个)`
	Compress_rate          string                           `压缩比`
	Valve_machanism        string                           `配气机构`
	Cylinder_radius        string                           `缸径(mm)`
	Stroke                 string                           `行程(mm)`
	Engine_hp              string                           `最大马力(Ps)`
	Engine_power           string                           `最大功率(kW)`
	Engine_rpm             string                           `最大功率转速(rpm)`
	Engine_torque          string                           `最大扭矩(N・m)`
	Torque_rpm             string                           `最大扭矩转速(rpm)`
	Tech_spec              string                           `发动机特有技术`
	Engine_energy          string                           `燃料形式`
	Roz                    string                           `燃油标号`
	Oil_drive              string                           `供油方式`
	Cylinder_cover         string                           `缸盖材料`
	Cylinder_body          string                           `缸体材料`
	Environmental_standard string                           `环保标准`
	Motor_type             string                           `电机类型`
	Motor_power            string                           `电动机总功率(kW)`
	Motor_torque           string                           `电动机总扭矩(N・m)`
	Motor_front_power      string                           `前电动机最大功率(kW)`
	Motor_front_torque     string                           `前电动机最大扭矩(N・m)`
	Motor_back_power       string                           `后电动机最大功率(kW)`
	Motor_back_torque      string                           `后电动机最大扭矩(N・m)`
	Sys_power              string                           `系统综合功率(kW)`
	Sys_torque             string                           `系统综合扭矩(N・m)`
	Motor_num              string                           `驱动电机数`
	Motor_arrange          string                           `电机布局`
	Bat_type               string                           `电池类型`
	Mileage                string                           `工信部续航里程(km)`
	Bat_cap                string                           `电池容量(kWh)`
	Bat_use                string                           `百公里耗电量(kWh/100km)`
	Bat_guarantee          string                           `电池组质保`
	Bat_charge             string                           `电池充电时间`
	Fast_charge            string                           `快充电量(%)`
	Charge_pile_price      string                           `充电桩价格`
	Gearbox_name           string                           `简称`
	Gears_num              string                           `挡位个数`
	Gears_type             string                           `变速箱类型`
	Drive_type             string                           `驾驶类型：手动，自动`
	Drive_method           string                           `驱动方式`
	Susp_front_type        string                           `前悬架类型`
	Susp_back_type         string                           `后悬架类型`
	Assist_type            string                           `助力类型`
	Structure              string                           `车体结构`
	Four_wheel_drive       string                           `四驱形式`
	Central_diff           string                           `中央差速器结构`
	Front_brake            string                           `前制动器类型`
	Back_brake             string                           `后制动器类型`
	Park_brake             string                           `驻车制动类型`
	Front_wheel_size       string                           `前轮胎规格`
	Back_wheel_size        string                           `后轮胎规格`
	Backup_wheel           string                           `备胎规格`
	Seat_srs               string                           `主/副驾驶座安全气囊`
	Side_airbag            string                           `前/后排侧气囊`
	Head_srs               string                           `前/后排头部气囊(气帘)`
	Knee_srs               string                           `膝部气囊`
	Tire_pres_monitor      string                           `胎压监测装置`
	Zero_tire_pres         string                           `零胎压继续行驶`
	Unbelt_notice          string                           `安全带未系提示`
	Isofix                 string                           `ISOFIX儿童座椅接口`
	Anti_lock              string                           `ABS防抱死`
	Bfd                    string                           `制动力分配(EBD/CBC等)`
	Bas                    string                           `刹车辅助(EBA/BAS/BA等)`
	Tcs                    string                           `牵引力控制(ASR/TCS/TRC等)`
	Stable_control         string                           `车身稳定控制(ESC/ESP/DSC等)`
	Bsa                    string                           `并线辅助`
	Ldw                    string                           `车道偏离预警系统`
	Abs                    string                           `主动刹车/主动安全系统`
	Nvs                    string                           `夜视系统`
	Tired_drive            string                           `疲劳驾驶提示`
	Radar                  string                           `前/后驻车雷达`
	Reverse_video          string                           `倒车视频影像`
	Panorama               string                           `全景摄像头`
	Cruise_ctrl            string                           `定速巡航`
	Self_adpt_cruise       string                           `自适应巡航`
	Auto_park_in           string                           `自动泊车入位`
	Engine_start_stop      string                           `发动机启停技术`
	Auto_drive             string                           `自动驾驶技术`
	Hac                    string                           `上坡辅助`
	Auto_park              string                           `自动驻车`
	Hdc                    string                           `陡坡缓降`
	Variable_susp          string                           `可变悬架`
	Air_susp               string                           `空气悬架`
	E_susp                 string                           `电磁感应悬架`
	Vgrs                   string                           `可变转向比`
	Front_diff_lock        string                           `前桥限滑差速器/差速锁`
	Central_diff_lock      string                           `中央差速器锁止功能`
	Back_diff_lock         string                           `后桥限滑差速器/差速锁`
	Ads                    string                           `整体主动转向系统`
	E_sunroof              string                           `电动天窗`
	Pano_sunroof           string                           `全景天窗`
	Sunroofs               string                           `多天窗`
	Sport_package          string                           `运动外观套件`
	Alloy_wheel            string                           `铝合金轮圈`
	E_suction_door         string                           `电动吸合门`
	Slide_door             string                           `侧滑门`
	E_cargo                string                           `电动后备厢`
	React_cargo            string                           `感应后备厢`
	Roof_rack              string                           `车顶行李架`
	Engine_e_guard         string                           `发动机电子防盗`
	E_ctrl_lock            string                           `车内中控锁`
	Remote_key             string                           `遥控钥匙`
	Keyless_start          string                           `无钥匙启动系统`
	Keyless_enter          string                           `无钥匙进入系统`
	Remote_start           string                           `远程启动`
	Leather_steering       string                           `皮质方向盘`
	Steer_adjt             string                           `方向盘调节`
	Steer_e_adjt           string                           `方向盘电动调节`
	Functional_steer       string                           `多功能方向盘`
	Steer_shift            string                           `方向盘换挡`
	Steer_heat             string                           `方向盘加热`
	Steer_mem              string                           `方向盘记忆`
	Computer_scr           string                           `行车电脑显示屏`
	Lcd_panel              string                           `全液晶仪表盘`
	Hud                    string                           `HUD抬头数字显示`
	Car_dvr                string                           `内置行车记录仪`
	Anc                    string                           `主动降噪`
	Wireless_charge        string                           `手机无线充电`
	Seat_mat               string                           `座椅材质`
	Sport_seat             string                           `运动风格座椅`
	Height_adjt            string                           `座椅高低调节`
	Lumbar_support         string                           `腰部支撑调节`
	Shoulder_support       string                           `肩部支撑调节`
	Seat_e_adjt            string                           `主/副驾驶座电动调节`
	Snd_backrest_adjt      string                           `第二排靠背角度调节`
	Snd_seat_mv            string                           `第二排座椅移动`
	Back_seat_adjt         string                           `后排座椅电动调节`
	Vice_adjt_btn          string                           `副驾驶位后排可调节按钮`
	E_seat_mem             string                           `电动座椅记忆`
	Seat_heat              string                           `前/后排座椅加热`
	Seat_vent              string                           `前/后排座椅通风`
	Seat_masg              string                           `前/后排座椅按摩`
	Snd_row_seat           string                           `第二排独立座椅`
	Third_row_seat         string                           `第三排座椅`
	Back_seat_down         string                           `后排座椅放倒方式`
	Handrail               string                           `前/后中央扶手`
	Back_cup_hold          string                           `后排杯架`
	Heat_cold_cup          string                           `可加热/制冷杯架`
	Gps                    string                           `GPS导航系统`
	Gps_interact           string                           `定位互动服务`
	Colorful_scr           string                           `中控台彩色大屏`
	Colorful_scr_size      string                           `中控台彩色大屏尺寸`
	Lcd_sep                string                           `中控液晶屏分屏显示`
	Blueteeth              string                           `蓝牙/车载电话`
	Mobile_map             string                           `手机互联/映射`
	Network                string                           `车联网`
	Television             string                           `车载电视`
	Back_lcd               string                           `后排液晶屏`
	Back_power_supply      string                           `220V/230V电源`
	External_audio         string                           `外接音源接口`
	Cddvd                  string                           `CD/DVD`
	Speaker_brand          string                           `扬声器品牌`
	Speaker_num            string                           `扬声器数量`
	Low_beam               string                           `近光灯`
	High_beam              string                           `远光灯`
	Led_beam               string                           `LED日间行车灯`
	Adaptive_beam          string                           `自适应远近光`
	Head_light             string                           `自动头灯`
	Turn_light             string                           `转向辅助灯`
	Turn_head_light        string                           `转向头灯`
	Front_fog_lamp         string                           `前雾灯`
	Light_height_adjt      string                           `大灯高度可调`
	Light_clean_dev        string                           `大灯清洗装置`
	Mood_light             string                           `车内氛围灯`
	Power_window           string                           `前/后电动车窗`
	E_lift_window          string                           `车窗一键升降`
	Anti_pinch_hand        string                           `车窗防夹手功能`
	Insulating_glass       string                           `防紫外线/隔热玻璃`
	E_adjt_rearview        string                           `后视镜电动调节`
	Heat_rearview          string                           `后视镜加热`
	Dimming_mirror         string                           `内/外后视镜自动防眩目`
	Stream_media_rearview  string                           `流媒体车内后视镜`
	Power_mirror           string                           `后视镜电动折叠`
	Mirror_mem             string                           `后视镜记忆`
	Abat_vent              string                           `后风挡遮阳帘`
	Side_abat_vent         string                           `后排侧遮阳帘`
	Side_priv_glass        string                           `后排侧隐私玻璃`
	Sun_shield             string                           `遮阳板化妆镜`
	Back_wiper             string                           `后雨刷`
	React_wiper            string                           `感应雨刷`
	Air_type               string                           `空调控制方式`
	Back_air               string                           `后排独立空调`
	Back_outlet            string                           `后座出风口`
	Temper_zone_ctrl       string                           `温度分区控制`
	Air_adjt               string                           `车内空气调节/花粉过滤`
	Air_cleaner            string                           `车载空气净化器`
	Car_fridge             string                           `车载冰箱`
}

func NewAutoHomeCar(aid int) (*CarCrawl) {
	c := &CarCrawl{Type_id:aid}
	c.settings = "https://car.autohome.com.cn/config/spec/" + strconv.Itoa(aid) + ".html"
	return c
}

func (c *CarCrawl) SetSeriesId(sid int) (*CarCrawl) {
	c.Series_id = sid
	return c
}

func (c *CarCrawl) SetPriceStr(priceStr string) (*CarCrawl) {
	c.Market_price_str = priceStr
	return c
}

func (c *CarCrawl) SetSeriesName(sname string) (*CarCrawl) {
	c.Series_name = sname
	return c
}

func (c *CarCrawl) SetBrandId(bid int) (*CarCrawl) {
	c.Brand_id = bid
	return c
}

func (c *CarCrawl) SetBrandName(bname string) (*CarCrawl) {
	c.Brand_name = bname
	return c
}

func (c *CarCrawl) SetManufacturer(mname string) (*CarCrawl) {
	c.Manufacturer = mname
	return c
}