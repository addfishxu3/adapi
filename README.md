# Go Gin 廣告投放服務
使用`gin`框架，撰寫api服務

## 如何運行

**需要**
+ Mysql
+ Redis

**準備**

創建一個`gin`資料庫，建立儲存廣告資料的SQL(http)

**運行**
````
$ cd $GOPATH/src/dcproject

$ go run main.go
````
專案資料及運行的API

````
[GIN-debug] [WARNING] Running in "debug" mode. Switch to "release" mode in production.
 - using env:   export GIN_MODE=release
 - using code:  gin.SetMode(gin.ReleaseMode)

[GIN-debug] POST   /api/v1/ad                --> main/dcproject.ad (3 handlers)
[GIN-debug] GET    /api/v1/ad/get            --> main/dcproject.setupRoute.CachePage.func4 (3 handlers)

[GIN-debug] Listening and serving HTTP on :8080
````

## 單元測試

**測試**
````
$ cd $GOPATH/src/dcproject

$ go test
````


