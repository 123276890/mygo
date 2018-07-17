package main

type Brand struct {
	Brand_id 		int			`orm:"auto;pk"`
	Brand_name		string
	Brand_initial	string		`orm:"null"`
	Brand_logo		string		`orm:"null"`
}