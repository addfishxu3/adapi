package main

import (
	"encoding/json"
	"errors"
	"fmt"

	"strings"
	"time"

	"net/http"
	"strconv"



	"github.com/gin-gonic/gin"
	"database/sql"

	"github.com/biter777/countries"
	_ "github.com/go-sql-driver/mysql"

	"github.com/gin-contrib/cache"
	"github.com/gin-contrib/cache/persistence"
)

type Gender string
const ( 
	M Gender = "M"
	F Gender = "F"
)

type Country string

type Platform string
const ( 
	android Platform = "android"
	ios     Platform = "ios"
	web     Platform = "web"
)


type conditions struct {
	AgeStart int        `json:"ageStart"`
	AgeEnd   int        `json:"ageEnd"`
	Gender   Gender     `json:"gender"`
	Country  []Country  `json:"country"`
	Platform []Platform `json:"platform"`
}

func main() {
	r := gin.Default()
	//可接受uri大小寫不同
	r.RedirectFixedPath = true 
	//redis cahche(1 hour)
	store := persistence.NewRedisCache("localhost:6379", "", time.Hour)
	//Admin API
	r.POST("/api/v1/ad", ad)
	//Public API
	r.GET("/api/v1/ad/get", cache.CachePage(store, time.Hour, public))
	r.Run(":8080")
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
func ad(c *gin.Context) {


	type adCreate struct {
		Title      string       `json:"title"`
		StartAt    time.Time    `json:"startAt"`
		EndAt      time.Time    `json:"endAt"`
		Conditions []conditions `json:"conditions"`
	}
	var ad1 adCreate

	title := c.PostForm("title")

	//驗證startAt、endAt的格式、時間邏輯
	timeformat := "2006-01-02T15:04:05Z"
	startAt, err := time.Parse(timeformat, c.PostForm("startat"))
	checkErr(err)
	endAt, err := time.Parse(timeformat, c.PostForm("endat"))
	checkErr(err)
	if startAt.After(endAt) {
		err := errors.New("時間錯誤")
		checkErr(err)
	}

	//驗證ageStart、ageEnd的格式、負數內容，預設1~100
	ageStart, err := strconv.Atoi(c.DefaultPostForm("agestart", "1")) //預設1
	checkErr(err)
	ageEnd, err := strconv.Atoi(c.DefaultPostForm("ageend", "100"))//預設100
	checkErr(err)
	if !(ageStart >= 1 && ageStart <= 100) { 
		ageStart = 1 
	}
	if !(ageEnd >= 1 && ageEnd <= 100) {
		ageEnd = 100
	}
	
	//驗證gender的內容
	var gender Gender = "" //不限制
	getgender := c.PostForm("gender")
	if getgender == "M" || getgender == "F" {
		gender = Gender(getgender)
	}

	//驗證country的內容，多值連字串存入
	var countrystr strings.Builder
	var country []Country = []Country{}
	getcountry := c.PostFormArray("country")
	for i := 0; i < len(getcountry); i++ {
		if countries.ByName(getcountry[i]).IsValid() {
			country = append(country, Country(Country(getcountry[i])))
			countrystr.WriteString(getcountry[i])
		}
	}
	if countrystr.Len() == 0 {
		countrystr.WriteString("") //不限制
	}

	//驗證platform的內容，多值連字串存入
	var platformstr strings.Builder
	var platform []Platform = []Platform{}
	getplatform := c.PostFormArray("platform")
	for i := 0; i < len(getplatform); i++ {
		p := Platform(getplatform[i])
		if p == web || p == ios || p == android {
			platform = append(platform, Platform(Platform(getplatform[i])))
			platformstr.WriteString(getplatform[i])
		}
	}
	if platformstr.Len() == 0 {
		countrystr.WriteString("") //不限制
	}

	//連結DB(mysql)
	db, err := sql.Open("mysql", "root:@(127.0.0.1:3306)/sys?charset=utf8")
	checkErr(err)
	defer db.Close()
	
	//設定每天產生廣告不超過3000
	var sameCreateDate int
	errSD := db.QueryRow("SELECT COUNT(*) FROM sys.adinfo WHERE date(createAt)=date(NOW())").Scan(&sameCreateDate)
	checkErr(errSD)

	if sameCreateDate < 3000 { 

		//設定同時間的活躍廣告(startat<now)不超過1000
		var activeAd int
		errAA := db.QueryRow("SELECT COUNT(*) FROM sys.adinfo WHERE  now() between startAt and endAt").Scan(&activeAd)
		checkErr(errAA)

		if activeAd < 1000 { 
			//新增廣告
			stmt, err := db.Prepare("INSERT adinfo SET title=?,createAt=NOW(),startAt=?,endAt=?,ageStart=?,ageEnd=?,gender=?,country=?,platform=?")
			checkErr(err)
			res, err := stmt.Exec(title, startAt, endAt, ageStart, ageEnd, gender, countrystr.String(), platformstr.String())
			checkErr(err)
			r, err := res.RowsAffected()
			checkErr(err)
			fmt.Println("insert", r, "row(s).")
			
			//顯示廣告內容
			var con = []conditions{
				{
					AgeStart: ageStart,
					AgeEnd:   ageEnd,
					Gender:   gender,
					Country:  country,
					Platform: platform,
				},
			}
			ad1 = adCreate{
				Title:      title,
				StartAt:    startAt,
				EndAt:      endAt,
				Conditions: con,
			}
			c.JSON(200, ad1)
		} else {
			c.String(200,"同時間的廣告超過1000")
		}
	} else {
		c.String(200,"每天產生廣告超過3000")
	}
}

func public(c *gin.Context) {
	//查詢條件都要給
	ch := make(chan string)
	//if query沒有給值,空白==""
	var err error
	var offset int = 0 //offset預設0
	if c.Query("offset") != "" {
		offset, err = strconv.Atoi(c.Query("offset")) //跳過筆數//int,>=0
		checkErr(err)
	}

	var limit int = 5
	if c.Query("limit") != "" {
		limit, err = strconv.Atoi(c.Query("limit")) //取多少筆(1~100,def 5)//int,1-100
		checkErr(err)
	}

	var age int = 0 //判斷數字而已
	if c.Query("age") != "" {
		//有age值
		age, err = strconv.Atoi(c.Query("age"))
		checkErr(err)
		if age < 1 || age > 100 {
			age = 0
		}
	}

	var gender string = ""
	if c.Query("gender") != "" {
		if Gender(c.Query("gender")) == M || Gender(c.Query("gender")) == F {
			gender = c.Query("gender")
		}
	}

	var country string = ""
	if c.Query("country") != "" {
		if countries.ByName(c.Query("country")).IsValid() {
			country = "%" + c.Query("country") + "%"
		}
	}

	var platform string = ""
	if c.Query("platform") != "" {
		if Platform(c.Query("platform")) == web || Platform(c.Query("platform")) == ios || Platform(c.Query("platform")) == android {
			platform = "%" + c.Query("platform") + "%"
		}
	}

	if offset >= 0 { //檢查負數
		if limit >= 1 && limit <= 100 { //檢查1~100
			if age != 0 && gender != "" && country != "" && platform != "" { //需有值
				//查詢資料
				//帶入指令
				go search(age, gender, country, platform, limit, offset, ch)
			}

		}
	}

	c.Data(200, "application/json", <-ch)
}

func search(age int, gender string, country string, platform string, limit int, offset int, ch chan []byte) {
	db, err := sql.Open("mysql", "root:@(127.0.0.1:3306)/sys?charset=utf8")
	checkErr(err)
	defer db.Close()

	//拼接query不安全，用?取代
	rows, err := db.Query("SELECT title, endAt FROM sys.adinfo WHERE startAt<NOW() && endAt>NOW() && ageStart<? && ageEnd>? && (gender=? || gender=\"\") &&(country LIKE ? || country=\"\") &&(platform LIKE ? || platform =\"\") ORDER BY endAt ASC LIMIT ? OFFSET ?",
		age, age, gender, country, platform, limit, offset)
	checkErr(err)

	type adinfo struct {
		Title string `json:"title"`
		EndAt string `json:"endAt"`
	}
	var adlist []adinfo
	getItems := make(map[string][]adinfo)
	for rows.Next() {
		var title string
		var endAt string
		err = rows.Scan(&title, &endAt)
		checkErr(err)
		//json
		ad := adinfo{
			Title: title,
			EndAt: endAt,
		}
		adlist = append(adlist, ad)
	}
	getItems["items"] = adlist

	j, _ := json.Marshal(getItems)
	ch <- j
}

//admin產生廣告(同時間的不超過1000/每天產生不超過3000)>存redis db>public看條件(需要輸入錯誤處理)取廣告>寫test
//github readme 寫設計理念
