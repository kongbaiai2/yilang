package main

import (
	"config"
	"fmt"
	"gomail"
	"io"
	"io/ioutil"
	"log"
	"os"
	"rrdtool"
	"runtime"
	"time"

	"github.com/urfave/cli/v2"
)

var (
	buildGitRevison string
	buildTimestamp  string
)

func TimeOut() error {
	time2 := int64(1658854666)
	time1 := time.Now().Unix()
	// time.Sleep(1 * time.Second)
	// log.Printf("%#v, %v", time1, time2)

	if time1-time2 >= 31104000 {
		log.Printf("gt 5: %v", time1-time2)
		return fmt.Errorf("time out")
	}
	return nil
}
func PrintStack() {
	var buf [4096]byte
	n := runtime.Stack(buf[:], false)
	log.Printf("==> %s", string(buf[:n]))

}

func initLog() {
	// 	// logDir := flag.String("log", "/home/bjyipeng/gomail.log", "location log dir")
	// 	// flag.Parse()
	// 	logFile, err := os.OpenFile("/home/bjyipeng/gomail.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	// 	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	// 	if err != nil {
	// 		log.Println("failed to open ./gomail.log, discarding any log.")
	// 		log.SetOutput(ioutil.Discard)
	// 	} else {
	// 		log.SetOutput(logFile)
	// 	}
	os.MkdirAll("./logs", 0644)
	logFile, err := os.OpenFile("./logs/gotools.log", os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		log.Println("failed to open ./logs/gotools.log, discarding any log.")
		log.SetOutput(ioutil.Discard)
	} else {
		mw := io.MultiWriter(os.Stdout, logFile)
		log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
		log.SetOutput(mw)
	}

	defer logFile.Close()

}

func main() {
	err := TimeOut()
	if err != nil {
		return
	}

	defer func() { //必须要先声明defer，否则不能捕获到panic异常
		if err := recover(); err != nil {
			log.Printf("[ERROR] catch panic:%s", err)
			PrintStack()
		}
	}()

	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	// initLog()

	t := time.Now()

	app := cli.NewApp()
	app.Name = "gotools"
	app.Version = fmt.Sprintf("\nGitHEAD: %s\nBuidTime: %s", buildGitRevison, buildTimestamp)
	app.Usage = "some tools for golang"
	app.Description = "This is how we describe greet the app"
	app.Authors = []*cli.Author{
		{Name: "yilang", Email: "kongbaiai2@126.com"},
	}
	app.Commands = gomail.AddSendMail(app.Commands)
	app.Commands = rrdtool.AddRRDtool(app.Commands)
	app.Commands = config.AddGinRun(app.Commands)

	app.Run(os.Args)

	log.Println(time.Now().Sub(t))
}
