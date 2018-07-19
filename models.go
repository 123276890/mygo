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
}