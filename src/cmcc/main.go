package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/robfig/cron"
)

func setCron() {
	log.Println("cmcc running success")
	c := cron.New()
	c.AddFunc(cronTime, doCronTask)
	c.AddFunc(CranTimeOneHour, doCronTaskOneHour)
	c.Start()

}
func doCronTask() {
	// 对外接口刷新chart的标记
	isFlushChart = true
	err := cmcc(endT, startT)
	if err != nil {
		log.Println(err)
	}
}
func doCronTaskOneHour() {
	err := cmcc(1*time.Hour, 3*time.Hour)
	if err != nil {
		log.Println(err)
	}
}
func Hello(c *gin.Context) {
	c.String(200, "hello %s", "world")
}
func runGin() {
	r := gin.Default()

	r.LoadHTMLGlob("html/*")
	r.GET("/line", func(c *gin.Context) {
		// every 5 minutes set true in func doCronTask()
		if isFlushChart {
			isFlushChart = false
			err := getData()
			if err != nil {
				c.String(404, "get db data err %s", err)
			}
		}

		c.HTML(http.StatusOK, "createChartLine.html", gin.H{
			"result": c.Param("content"),
		})
	})
	r.GET("/hello", Hello)
	r.Run(":7000")
}
func main() {
	defer func() {
		log.Println("exit")
	}()
	initDbAndLog()

	// 监听告警事件
	go AlarmListenDb()
	// 骤降告警
	go AlarmFall()
	// 定时执行任务
	setCron()

	runGin()
}
