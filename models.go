package main

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
	Car_name			string
	Series_id			int
	Series_name			string
	Brand_id			int
	Brand_name			string
	//Country_id			int
	Country_str			string
	//Produce_type		int
	Produce_type_str	string
	Manufacturer		string
	//Car_level			int
	Car_level_str		string
	//Market_price		float64
	Market_price_str	string
	//Energy_type			int
	Energy_type_str		string
	Market_time			string
	Engine				string
	Gearbox				string
	Car_size			string
	Car_struct			string
	Max_speed			string
	Official_speedup	string
	Actual_speedup		string
	Actual_brake		string
	Actual_fueluse		string
	Gerenal_fueluse		string
	Actual_ground		string
	Quality_guarantee	string
	Max_power			string
	Max_torque			string
	E_mileage			string
}