package config

import (
	"bufio"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"rrdtool"
	"strconv"
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

type MsgPostquery struct {
	Ratio float64 `form:"ratio" json:"ratio" xml:"ratio"  binding:"required"`
}

func createRrd(srcrrd, filename, savefile, dstcsv string, ratio float64) (string, error) {

	csvhandler := &rrdtool.Csvhandler{
		Headler:   []rrdtool.CsvHeadle{},
		TimeValue: make(map[int]rrdtool.CsvTimeValue),
	}
	csvhandler.Dbfile = srcrrd    // dbfile
	csvhandler.Pngfile = savefile // pngfile
	csvhandler.SouCsv = filename  // soucsv
	csvhandler.DstCsv = dstcsv    // dstcsv

	// data source file cvs
	err := csvhandler.GetFileCsv(csvhandler.SouCsv)
	if err != nil {
		log.Println(err)
		return csvhandler.Pngfile, err
	}
	err = csvhandler.CreateRRD(csvhandler.Dbfile, csvhandler.StartTime, csvhandler.LenDS, 300)
	if err != nil {
		log.Printf("createrrd:%v", err)
		return csvhandler.Pngfile, err
	}
	err = csvhandler.UpdateRRD(csvhandler.Dbfile)
	if err != nil {
		log.Printf("UpdateRRD:%v", err)
		return csvhandler.Pngfile, err
	}
	info, err := rrd.Info(csvhandler.Dbfile)
	if err != nil {
		log.Println(err)
		return csvhandler.Pngfile, err
	}
	csvhandler.Rrd.Info = info

	// 从rrd中取95值
	typeStruct := rrdtool.Get95th(csvhandler, ratio)
	err = csvhandler.CreateGrapher(csvhandler.Dbfile, csvhandler.Pngfile, csvhandler.StartTime, csvhandler.EndTime, typeStruct, ratio)
	if err != nil {
		log.Printf("CreateGrapher:%v", err)
		return csvhandler.Pngfile, err
	}
	err = saveCsv(csvhandler.DstCsv, csvhandler, typeStruct)
	if err != nil {
		log.Printf("saveCsv:%v", err)
		return csvhandler.Pngfile, err
	}

	log.Printf("start_unix:%v, end_unix:%v\n", csvhandler.StartUnix, csvhandler.EndUnix)
	log.Printf("time:%v,end:%v,count:%v", csvhandler.StartTime, csvhandler.EndTime, len(csvhandler.TimeValue))
	return csvhandler.Pngfile, err
}

func saveCsv(csvfilename string, csvhead *rrdtool.Csvhandler, tempStruct *rrdtool.FetchRrd) error {
	file, err := os.Create(csvfilename)
	if err != nil {
		log.Println(err)
		return err
	}
	defer file.Close()
	writer := csv.NewWriter(file)
	writer.Write([]string{"\xEF\xBB\xBF"})
	writer.Comma = ','

	head := [][]string{}
	for _, vmap := range csvhead.Headler {
		tmp := []string{}
		for kk, v := range vmap.HeadlerMap {
			tmp = append(tmp, kk)

			//  匹配95值
			if len(v) > 1 {
				for _, vv := range tempStruct.Rsult95th {
					tmp95 := fmt.Sprintf("%f", vv)
					tmp = append(tmp, tmp95)
				}
			} else {
				tmp = append(tmp, v...)
			}

			head = append(head, tmp)
		}
	}
	writer.WriteAll(head)
	// writer.Write([]string{"\n"})

	for i := 0; i < tempStruct.XCount; i++ {
		xRow := []string{}
		xRow = append(xRow, tempStruct.ResTimeList[i].Format("2006-01-02 15:04:05"))
		for j := 0; j < tempStruct.YCount; j++ {
			xRow = append(xRow, fmt.Sprint(tempStruct.ResValueList[i][j]))
		}
		// log.Printf("count:x-%v,y-%v", tempStruct.XCount, i)
		writer.Write(xRow)

	}

	writer.Flush()
	return nil
}
func createZip(savezip string, fileZip []string) ([]string, error) {
	err := rrdtool.Zip(savezip, fileZip)
	if err != nil {
		return fileZip, err
	}
	return fileZip, nil
}

// 配置upload，返回下载二进制流
func downOctetStream(c *gin.Context, downfile, dir string) error {

	f, err := os.OpenFile(dir+"/"+downfile, os.O_RDONLY, os.ModePerm)
	if err != nil {
		return err
	}
	defer f.Close()
	buf := bufio.NewReader(f)
	type_b, err := buf.Peek(512)
	if err != nil {
		return err
	}

	c.Writer.Header().Set("Content-type", http.DetectContentType(type_b)) //设置文件格式

	fileinfo, err := f.Stat()
	if err != nil {
		return err
	}

	//"application/octet-stream"
	c.Writer.Header().Set("Content-Length", strconv.FormatInt(fileinfo.Size(), 10))

	if ok, _ := strconv.ParseBool("true"); ok {
		c.Writer.Header().Add("Content-Disposition", "attachment; filename="+downfile) //下载文件名,不设置网页直接打开PDF.jpg.png 等格式
	}

	buf_b := make([]byte, 1024*1024) //发送大小
	for {
		n, err := buf.Read(buf_b)
		if err == io.EOF || n == 0 {
			break
		}

		c.Writer.Write(buf_b[:n])
	}
	return nil
}

func upload(c *gin.Context) {
	var msgQuery MsgPostquery
	if err := c.ShouldBind(&msgQuery); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if msgQuery.Ratio < 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"status": "paras err"})
		return
	} else if msgQuery.Ratio == 0 {
		msgQuery.Ratio = 1
	}

	dir := "./config/website/tmp"
	fileList := []string{}

	form, err := c.MultipartForm()
	if err != nil {
		c.String(http.StatusBadRequest, fmt.Sprintf("get err %s", err.Error()))
		return
	}
	// 获取所有文件
	files := form.File["files"]

	isZip := false
	// 遍历所有文件
	for _, file := range files {
		// 逐个存
		// if err := c.SaveUploadedFile(file, file.Filename); err != nil {
		src := fmt.Sprintf("%s/%s", dir, file.Filename)

		if err := c.SaveUploadedFile(file, src); err != nil {
			c.String(http.StatusBadRequest, fmt.Sprintf("upload err %s", err.Error()))
			return
		}

		// 生成图片并返回

		if strings.HasSuffix(src, ".csv") {
			_, err := createRrd(src+".rrd", src, src+".png", src+"_dst.csv", msgQuery.Ratio)
			if err != nil {
				log.Printf("create rrd graph err %s", err.Error())
				c.String(http.StatusBadRequest, fmt.Sprintf("create rrd graph err %s", err.Error()))
				return
			}
			isZip = true
		}

		// save filename
		fileList = append(fileList, src+".rrd", src, src+".png", src+"_dst.csv")
	}

	if isZip {
		log.Println(fileList)

		// 多文件压缩zip
		dstZip := time.Now().Format("20060102T150405")
		_, err = createZip(dir+"/"+dstZip+".zip", fileList)
		if err != nil {
			log.Printf("createZip err %s", err.Error())
			c.String(http.StatusBadRequest, fmt.Sprintf("createZip err %s", err.Error()))
			return
		}

		//返回zip流
		err = downOctetStream(c, dstZip+".zip", dir)
		if err != nil {
			log.Printf("downOctetStream err %s", err.Error())
			c.String(http.StatusBadRequest, fmt.Sprintf("downOctetStream err %s", err.Error()))
			return
		}

		return
	}
	// c.File(png_file)

	c.String(200, fmt.Sprintf("upload ok %d files", len(files)))

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
		v.POST("/upload", LimitHandler(lmt), func(c *gin.Context) {
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
			return
		}
	}()
	log.Println("listen web port:", c.String("port"))

	// 优雅Shutdown（或重启）服务
	quit := make(chan os.Signal, 2)
	signal.Notify(quit, os.Interrupt) // syscall.SIGKILL
	<-quit
	log.Println("Shutdown Server ...")
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}
	select {
	case <-ctx.Done():
	case <-time.After(5 * time.Second):
		log.Println("timeout")
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
