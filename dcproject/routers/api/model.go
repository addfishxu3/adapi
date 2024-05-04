package api

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

var Pool *sqlx.DB

//Gender
type Gender string

const ( 
	M Gender = "M"
	F Gender = "F"
)

//Country
type Country string

//Platform
type Platform string

const ( 
	Android Platform = "android"
	Ios     Platform = "ios"
	Web     Platform = "web"
)

//Conditions
type Conditions struct {
	AgeStart int        `json:"ageStart"`
	AgeEnd   int        `json:"ageEnd"`
	Gender   Gender     `json:"gender"`
	Country  []Country  `json:"country"`
	Platform []Platform `json:"platform"`
}

