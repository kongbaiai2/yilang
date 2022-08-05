package main

import (
	"log"
	"os"
	"runtime"
	"time"

	"github.com/urfave/cli/v2"
)

func PrintStack() {
	var buf [4096]byte
	n := runtime.Stack(buf[:], false)
	log.Printf("==> %s", string(buf[:n]))

}

// func initLog() {
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
// }

func main() {
	defer func() { //必须要先声明defer，否则不能捕获到panic异常
		if err := recover(); err != nil {
			log.Printf("[ERROR] catch panic:%s", err)
			PrintStack()
		}
	}()

	// initLog()

	t := time.Now()

	app := cli.NewApp()
	app.Name = "go tools"
	app.Version = "0.1.0"
	app.Usage = "some tools for golang"
	app.Description = "This is how we describe greet the app"
	app.Authors = []*cli.Author{
		{Name: "yilang", Email: "kongbaiai2@126.com"},
	}
	app.Commands = AddSendMail(app.Commands)
	// app.Action = func(c *cli.Context) error {
	// 	println("Greetings")
	// 	return nil
	// }

	app.Run(os.Args)

	log.Println(time.Now().Sub(t))
}
