package routers

import (
	v1 "dcproject/routers/api/v1"

	"github.com/gin-gonic/gin"

	"time"

	"github.com/gin-contrib/cache"
	"github.com/gin-contrib/cache/persistence"
)

func SetupRoute() *gin.Engine {
	r := gin.Default()
	
	//可接受uri大小寫不同
	r.RedirectFixedPath = true 
	store := persistence.NewRedisCache("localhost:6379", "", time.Hour)

	//Admin API
	r.POST("/api/v1/ad", v1.Ad)
	//Public API
	r.GET("/api/v1/ad/get", cache.CachePage(store, time.Hour, v1.Public))
	return r
}
