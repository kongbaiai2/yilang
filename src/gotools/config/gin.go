package config

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"rrdtool"
	"strings"
	"time"

	"github.com/didip/tollbooth"
	"github.com/didip/tollbooth/limiter"
	"github.com/ffhelicopter/tmm/api"
	"github.com/ffhelicopter/tmm/handler"
	"github.com/gin-gonic/gin"
	"github.com/urfave/cli/v2"
	"github.com/ziutek/rrd"
)

func createRrd(filename string) string {
	csvhandler := &rrdtool.Csvhandler{
		Headler:   []rrdtool.CsvHeadle{},
		TimeValue: make(map[int]rrdtool.CsvTimeValue),
	}
	csvhandler.Dbfile = "test.rrd"  // dbfile
	csvhandler.Pngfile = "test.png" // pngfile
	csvhandler.SouCsv = filename    // soucsv
	csvhandler.DstCsv = "dst.rrd"   // dstcsv

	// data source file cvs
	err := csvhandler.GetFileCsv(csvhandler.SouCsv)
	if err != nil {
		log.Println(err)
		return csvhandler.Pngfile
	}
	err = csvhandler.CreateRRD(csvhandler.Dbfile, csvhandler.StartTime, csvhandler.LenDS, 300)
	if err != nil {
		log.Printf("createrrd:%v", err)
	}
	err = csvhandler.UpdateRRD(csvhandler.Dbfile)
	if err != nil {
		log.Printf("UpdateRRD:%v", err)
	}
	info, err := rrd.Info(csvhandler.Dbfile)
	if err != nil {
		log.Println(err)
	}
	csvhandler.Rrd.Info = info

	// 从rrd中取95值
	typeStruct := rrdtool.Get95th(csvhandler, info)
	err = csvhandler.CreateGrapher(csvhandler.Dbfile, csvhandler.Pngfile, csvhandler.StartTime, csvhandler.EndTime, typeStruct)
	if err != nil {
		log.Printf("CreateGrapher:%v", err)
	}
	csvhandler.Rrd.Cdef = rrd.NewExporter()

	log.Printf("start_unix:%v, end_unix:%v\n", csvhandler.StartUnix, csvhandler.EndUnix)
	log.Printf("time:%v,end:%v,count:%v", csvhandler.StartTime, csvhandler.EndTime, len(csvhandler.TimeValue))
	return csvhandler.Pngfile
}
func upload(c *gin.Context) {

	form, err := c.MultipartForm()
	if err != nil {
		c.String(http.StatusBadRequest, fmt.Sprintf("get err %s", err.Error()))
	}
	// 获取所有图片
	files := form.File["files"]
	png_file := "test.png"
	// 遍历所有图片
	for _, file := range files {
		// 逐个存
		// if err := c.SaveUploadedFile(file, file.Filename); err != nil {
		dst := fmt.Sprintf("./config/website/tmp/%s", file.Filename)
		if err := c.SaveUploadedFile(file, dst); err != nil {
			c.String(http.StatusBadRequest, fmt.Sprintf("upload err %s", err.Error()))
			return
		}

		// 生成图片并返回
		if strings.Contains(dst, ".csv") {
			createRrd(dst)
		}

	}
	log.Println(png_file)
	c.File("./config/website/tmp/test.png")
	// c.String(200, fmt.Sprintf("upload ok %d files", len(files)))

}

func LimitHandler(lmt *limiter.Limiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		httpError := tollbooth.LimitByRequest(lmt, c.Writer, c.Request)
		if httpError != nil {
			c.Data(httpError.StatusCode, lmt.GetMessageContentType(), []byte(httpError.Message))
			c.Abort()
		} else {
			c.Next()
		}
	}
}

func GinRun(c *cli.Context) {
	router := gin.Default()
	// 静态资源加载，本例为css,js以及资源图片
	router.StaticFS("/public", http.Dir("./config/website/static"))
	router.StaticFile("/favicon.jpg", "./config/website/favicon.jpg")
	router.StaticFile("/uploads", "./config/website/tpl/upload.html")

	// 导入所有模板，多级目录结构需要这样写,tpl动态目录
	router.LoadHTMLGlob("./config/website/tpl/*")

	// 使用全局CORS中间件。
	// router.Use(Cors())
	// 即使是全局中间件，在use前的代码不受影响
	//rate-limit 限流中间件
	lmt := tollbooth.NewLimiter(5, nil)
	lmt.SetMessage("服务繁忙，请稍后再试...")

	// config/website分组
	v := router.Group("/")
	{
		v.GET("/index.html", LimitHandler(lmt), handler.IndexHandler)
		v.GET("/add.html", handler.AddHandler)
		v.POST("/postme.html", handler.PostmeHandler)
		v.POST("/upload", func(c *gin.Context) {
			upload(c)
		})
	}

	v1 := router.Group("/v1")
	{
		// 下面是群组中间的用法
		// v1.Use(Cors())
		// 单个中间件的用法
		// v1.GET("/user/:id/*action",Cors(), api.GetUser)
		// rate-limit
		v1.GET("/user/:id/*action", LimitHandler(lmt), api.GetUser)
		//v1.GET("/user/:id/*action", Cors(), api.GetUser)
		// AJAX OPTIONS ，下面是有关OPTIONS用法的示例
		// v1.OPTIONS("/users", OptionsUser)      // POST
		// v1.OPTIONS("/users/:id", OptionsUser)  // PUT, DELETE
	}

	// router.Run(":80")
	// 这样写就可以了，下面所有代码（go1.8+）是为了优雅处理重启等动作。
	srv := &http.Server{
		Addr:         ":" + c.String("port"),
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
	go func() {
		// 监听请求
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// 优雅Shutdown（或重启）服务
	quit := make(chan os.Signal, 2)
	signal.Notify(quit, os.Interrupt) // syscall.SIGKILL
	<-quit
	log.Println("Shutdown Server ...")
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}
	select {
	case <-ctx.Done():
	case <-time.After(5 * time.Second):
		fmt.Println("timeout")
	}
	log.Println("Server exiting")
}

func AddGinRun(goCom []*cli.Command) []*cli.Command {

	Command := &cli.Command{
		Name: "gin",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "port",
				Aliases: []string{"p"},
				Usage:   "specify gin listen port case:8080",
				Value:   "8080",
			},
		},
		Action: func(c *cli.Context) error {

			GinRun(c)
			return nil
		},
	}

	goCom = append(goCom, Command)
	return goCom
}
