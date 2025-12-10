package routers

import (
	"net/http"

	"github.com/kongbaiai2/yilang/goapp/internal/api/v1/cacti"
	"github.com/kongbaiai2/yilang/goapp/internal/global"

	"github.com/gin-gonic/gin"
)

func NewRouter() *gin.Engine {
	//g := gin.Default()
	g := gin.New()

	g.StaticFile("/", "./index.html")
	g.Static("/img", "img")
	g.GET("/status.default", func(c *gin.Context) {
		if global.ProcessExit {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"message": "service unavailable",
			})
		} else {
			c.JSON(http.StatusOK, gin.H{
				"message": "pong",
			})
		}
	})

	g.Use(GinLogger(), GinRecovery(true))

	// if global.CONFIG.System.Auth {
	// 	g.Use(Auth())
	// }

	/*************************************************************************/
	/***********     OPS API, provided to the Luban platform       ***********/
	/*************************************************************************/
	apiGroup := g.Group("/api")

	resApi := apiGroup.Group("/cacti")
	{
		resApi.GET("/GetPercentMonthly", cacti.GetPercentMonthly)
		resApi.GET("/GetPercentEveryDay", cacti.GetPercentEveryDay)
		resApi.GET("/DescribeImage", cacti.DescribeImage)
		resApi.GET("/DeleteDir", cacti.DeleteDir)

	}
	return g
}
