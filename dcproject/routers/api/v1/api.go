package v1

import (
	api "dcproject/routers/api"

	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/biter777/countries"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"

	"log"
)

// Open Database MySQL
func init() {
	dsn := "root:@(localhost:3306)/sys?charset=utf8"
	var err error
	api.Pool, err = sqlx.Open("mysql", dsn)
	checkErr(err)
	api.Pool.SetMaxOpenConns(100)
	api.Pool.SetMaxIdleConns(20)
}

// Error handling
func checkErr(err error) {
	if err != nil {
		log.Panicln("err: ", err.Error())
	}
}

// Create AD
func Ad(c *gin.Context) {

	//驗證title有沒有內容
	title := c.PostForm("title")
	if title == "" {
		checkErr(errors.New("標題需有資料"))
	}

	//驗證startAt、endAt的格式、時間邏輯
	timeformat := "2006-01-02T15:04:05Z"
	startAt, err := time.Parse(timeformat, c.PostForm("startAt"))
	checkErr(err)
	endAt, err := time.Parse(timeformat, c.PostForm("endAt"))
	checkErr(err)
	if startAt.After(endAt) {
		checkErr(errors.New("時間錯誤"))
	}

	//驗證ageStart、ageEnd的格式、範圍，預設1~100
	ageStart, err := strconv.Atoi(c.DefaultPostForm("ageStart", "1")) //預設1
	checkErr(err)
	ageEnd, err := strconv.Atoi(c.DefaultPostForm("ageEnd", "100"))//預設100
	checkErr(err)
	if !(ageStart >= 1 && ageStart <= 100) { 
		ageStart = 1 
	}
	if !(ageEnd >= 1 && ageEnd <= 100) {
		ageEnd = 100
	}
	if ageStart > ageEnd {
		checkErr(errors.New("年齡錯誤"))
	}
	
	//驗證gender的內容
	var gender api.Gender = "" //""代表不限制
	getgender := c.PostForm("gender")
	if getgender == "M" || getgender == "F" {
		gender = api.Gender(getgender)
	}

	//驗證country的內容，多值連字串存入
	var countrystr strings.Builder
	var country []api.Country = []api.Country{}
	getcountry := c.PostFormArray("country")
	for i := 0; i < len(getcountry); i++ {
		if countries.ByName(getcountry[i]).IsValid() {
			country = append(country, api.Country(api.Country(getcountry[i])))
			countrystr.WriteString(getcountry[i])
		}
	}
	if countrystr.Len() == 0 {
		countrystr.WriteString("") //""代表不限制
	}

	//驗證platform的內容，多值連字串存入
	var platformstr strings.Builder
	var platform []api.Platform = []api.Platform{}
	getplatform := c.PostFormArray("platform")
	for i := 0; i < len(getplatform); i++ {
		p := api.Platform(getplatform[i])
		if p == api.Web || p == api.Ios || p == api.Android {
			platform = append(platform, api.Platform(api.Platform(getplatform[i])))
			platformstr.WriteString(getplatform[i])
		}
	}
	if platformstr.Len() == 0 {
		countrystr.WriteString("") //""代表不限制
	}

	//每天產生廣告不超過3000
	var sameCreateDate int
	err = api.Pool.Get(&sameCreateDate, "select COUNT(*) FROM sys.adinfo where date(createAt)=date(NOW())")
	checkErr(err)
	if sameCreateDate >= 3000 {
		checkErr(errors.New("今天產生廣告已超過3000"))
	}

	//同時間的活躍廣告(startat<now)不超過1000
	var activeAd int
	err = api.Pool.Get(&activeAd, "SELECT COUNT(*) FROM sys.adinfo WHERE now() between startAt and endAt")
	checkErr(err)
	if activeAd >= 1000 {
		checkErr(errors.New("同時間的廣告超過1000"))
	}

	//存入廣告資料
	sqlstr := "INSERT adinfo SET title=?,createAt=NOW(),startAt=?,endAt=?,ageStart=?,ageEnd=?,gender=?,country=?,platform=?"
	res, err := api.Pool.Exec(sqlstr, title, startAt, endAt, ageStart, ageEnd, gender, countrystr.String(), platformstr.String())
	checkErr(err)
	ra, err := res.RowsAffected()
	checkErr(err)
	fmt.Println("insert", ra, "row(s).")

	//顯示存入的廣告資料
	type adCreate struct {
		Title      string       `json:"title"`
		StartAt    time.Time    `json:"startAt"`
		EndAt      time.Time    `json:"endAt"`
		Conditions []api.Conditions `json:"conditions"`
	}
	var con = []api.Conditions{
		{
			AgeStart: ageStart,
			AgeEnd:   ageEnd,
			Gender:   gender,
			Country:  country,
			Platform: platform,
		},
	}
	ad := adCreate{
		Title:      title,
		StartAt:    startAt,
		EndAt:      endAt,
		Conditions: con,
	}
	c.JSON(200, ad)
}

func Public(c *gin.Context) {

	//驗證offset為數字、預設0、非負數
	offset, err := strconv.Atoi(c.DefaultQuery("offset","0")) 
	checkErr(err)
	if offset < 0 {
		checkErr(errors.New("offset error"))
	}

	//驗證limit為數字、預設5、範圍1~100
	limit, err := strconv.Atoi(c.DefaultQuery("limit","5")) 
	checkErr(err)
	if !(limit >= 1 && limit <= 100) {
		checkErr(errors.New("limit error"))
	}

	//驗證age為數字、範圍1~100
	age, err := strconv.Atoi(c.DefaultQuery("age", "0"))
	checkErr(err)
	if age < 1 || age > 100 {
		checkErr(errors.New("age error"))
	}

	//驗證gender內容(M或F)
	gender := c.Query("gender")
	if !(api.Gender(gender) == api.M || api.Gender(gender) == api.F) {
		checkErr(errors.New("gender error"))
	}

	//驗證country內容
	var country string
	if countries.ByName(c.Query("country")).IsValid() {
		country = "%" + c.Query("country") + "%"
	}else{
		checkErr(errors.New("country error"))
	}

	//驗證platform內容(web或ios或android)
	var platform string
	if api.Platform(c.Query("platform")) == api.Web || api.Platform(c.Query("platform")) == api.Ios || api.Platform(c.Query("platform")) == api.Android {
		platform = "%" + c.Query("platform") + "%"
	}else{
		checkErr(errors.New("platform error"))
	}

	//查詢廣告資料
	ch := make(chan []byte)
	go search(age, gender, country, platform, limit, offset, ch)

	c.Data(200, "application/json", <-ch)

}

func search(age int, gender string, country string, platform string, limit int, offset int, ch chan []byte) {

	//用?帶入查詢資料
	rows, err := api.Pool.Query("SELECT title, endAt FROM sys.adinfo WHERE startAt<NOW() && endAt>NOW() && ageStart<? && ageEnd>? && (gender=? || gender=\"\") &&(country LIKE ? || country=\"\") &&(platform LIKE ? || platform =\"\") ORDER BY endAt ASC LIMIT ? OFFSET ?",
		age, age, gender, country, platform, limit, offset)
	checkErr(err)

	type adinfo struct {
		Title string `json:"title"`
		EndAt string `json:"endAt"`
	}
	var adlist []adinfo
	getItems := make(map[string][]adinfo)
	
	//取出每列資料
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

	//結果存入channel
	j, _ := json.Marshal(getItems)
	ch <- j
}
