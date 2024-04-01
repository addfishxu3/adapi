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
)

// enum
const ( //Gender
	M Gender = "M"
	F Gender = "F"
)

type Gender string

const ( //Country
	TW Country = "TW"
	JP Country = "JP"
)

type Country string

const ( //Platform
	android Platform = "android"
	ios     Platform = "ios"
	web     Platform = "web"
)

type Platform string
type conditions struct {
	AgeStart int        `json:"ageStart"`
	AgeEnd   int        `json:"ageEnd"`
	Gender   Gender     `json:"gender"`
	Country  []Country  `json:"country"`
	Platform []Platform `json:"platform"`
}

func main() {
	r := gin.Default()
	r.RedirectFixedPath = true //可接受uri大小寫不同
	//Admin API
	r.POST("/api/v1/ad", ad)
	//Public API
	r.GET("/api/v1/ad/get", public)
	r.Run(":8080") // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
func ad(c *gin.Context) {

	//參數,日期驗證
	//條件
	//不限制預設""，查詢用||col=""；多值連字串，查詢用like；

	type adCreate struct {
		Title      string       `json:"title"`
		StartAt    time.Time    `json:"startAt"`
		EndAt      time.Time    `json:"endAt"`
		Conditions []conditions `json:"conditions"`
	}
	var ad1 adCreate

	title := c.PostForm("title")

	timeformat := "2006-01-02T15:04:05Z"
	startAt, err := time.Parse(timeformat, c.PostForm("startat"))
	checkErr(err)
	endAt, err := time.Parse(timeformat, c.PostForm("endat"))
	checkErr(err)

	ageStart, err := strconv.Atoi(c.DefaultPostForm("agestart", "1")) //預設1
	checkErr(err)
	ageEnd, err := strconv.Atoi(c.DefaultPostForm("ageend", "100"))
	checkErr(err)
	if !(ageStart >= 1 && ageStart <= 100) { //負數
		ageStart = 1 //預設1
	}
	if !(ageEnd >= 1 && ageEnd <= 100) {
		ageEnd = 100
	}

	var gender Gender = "" //不限制
	getgender := c.PostForm("gender")
	if getgender == "M" || getgender == "F" {
		gender = Gender(getgender)
	}

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
	//下面處理不限制

	db, err := sql.Open("mysql", "root:@(127.0.0.1:3306)/sys?charset=utf8")
	checkErr(err)
	defer db.Close()

	var sameCreateDate int
	errSD := db.QueryRow("SELECT COUNT(*) FROM sys.adinfo WHERE date(createAt)=date(NOW())").Scan(&sameCreateDate)
	checkErr(errSD)

	if sameCreateDate < 3000 { //每天產生廣告不超過3000

		var activeAd int
		errAA := db.QueryRow("SELECT COUNT(*) FROM sys.adinfo WHERE startAt<NOW()<endAt").Scan(&activeAd)
		checkErr(errAA)

		if activeAd < 1000 { //同時間的活躍廣告(startat<now)不超過1000 //or 找最快endat 如果endat<輸入的startat就ok
			stmt, err := db.Prepare("INSERT adinfo SET title=?,createAt=NOW(),startAt=?,endAt=?,ageStart=?,ageEnd=?,gender=?,country=?,platform=?")
			checkErr(err)

			res, err := stmt.Exec(title, startAt, endAt, ageStart, ageEnd, gender, countrystr.String(), platformstr.String())
			checkErr(err)
			//con,plat存連續字串
			ra, err := res.RowsAffected()
			checkErr(err)
			fmt.Println("insert", ra, "row(s).")

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
			fmt.Println("同時間的廣告超過1000")
		}
	} else {
		fmt.Println("每天產生廣告超過3000")
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
				go search(age, gender, country, platform, limit, offset, c, ch)
			}

		}
	}

	<-ch
}

func search(age int, gender string, country string, platform string, limit int, offset int, c *gin.Context, ch chan string) {
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
	c.Data(200, "application/json", j)
	ch <- "D"
}

//admin產生廣告(同時間的不超過1000/每天產生不超過3000)>存redis db>public看條件(需要輸入錯誤處理)取廣告>寫test
//github readme 寫設計理念
