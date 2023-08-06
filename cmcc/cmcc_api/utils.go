package cmcc_api

import (
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"runtime"

	"github.com/sirupsen/logrus"

	"github.com/mitchellh/go-homedir"
	"gopkg.in/gomail.v2"
)

type M_cli struct {
	From        string
	To          []string
	Subject     string
	Body        string
	Enclosure   string
	Smtp_domain string
	Port        int
	User        string
	Password    string
}

func (c *M_cli) SendMail() error {
	m := gomail.NewMessage()
	m.SetHeader("From", c.From)
	m.SetHeader("To", c.To...)
	m.SetHeader("Subject", c.Subject)
	m.SetBody("text/html", c.Body)
	if c.Enclosure != "" {
		m.Attach(c.Enclosure)
	}

	d := gomail.NewDialer(c.Smtp_domain, c.Port, c.User, c.Password)
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	// Send the email to Bob, Cora and Dan.
	if err := d.DialAndSend(m); err != nil {
		log.Printf("Send the email failed: %v", err)
		return err
	}
	return nil
}

// Sha256加密
func Sha256(src string) string {
	// log.Println(src)
	m := sha256.New()
	m.Write([]byte(src))
	return hex.EncodeToString(m.Sum(nil))
}

func MapToJson(param map[string]interface{}) (string, error) {
	dataType, err := json.Marshal(param)
	if err != nil {
		return "", err
	}
	return string(dataType), nil
}

func JsonToMap(str string) map[string]interface{} {
	var tempMap map[string]interface{}
	err := json.Unmarshal([]byte(str), &tempMap)
	if err != nil {
		panic(err)
	}
	return tempMap
}

func initDbAndLog(file string) {
	InitLog(log, file)

	CreateMysqlDb(cfg.DbCfg)
	initInfoToDb()

}

//	logmap := log.WithField("common", "this is a common field")
//
// log.info() or logmap.info()
func InitLog(log *logrus.Logger, logname string) {
	// log.SetFormatter(&log.JSONFormatter{})
	// format = "2006-01-02 15:04:05.00"
	log.SetFormatter(&logrus.TextFormatter{ForceColors: true, TimestampFormat: "02 15:04:05", FullTimestamp: true,
		CallerPrettyfier: func(f *runtime.Frame) (function string, file string) {
			// f.Func.Name()
			file = path.Base(f.File)
			function = fmt.Sprintf("%s:%d", file, f.Line)
			return function, ""
		}})

	dir, err := homedir.Dir()
	if err != nil {
		log.Fatal("get home dir failed:", err)
	}
	logFile, err := os.OpenFile(path.Join(dir, "log", logname), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0777)
	twoWrite := io.MultiWriter(logFile, os.Stdout)
	if err != nil {
		log.Panic(err)
	}
	log.SetOutput(twoWrite)

	log.SetLevel(logrus.WarnLevel)

	log.SetReportCaller(true)
	log.Infof("write log to %s", path.Join(dir, "log", logname))

}

func getCallerInfo(skip int, isonly bool) (info string) {

	pc, file, lineNo, ok := runtime.Caller(skip)
	if !ok {
		info = "runtime.Caller() failed"
		return
	}
	funcName := runtime.FuncForPC(pc).Name()
	fileName := path.Base(file) // Base函数返回路径的最后一个元素
	if isonly {
		return file
	}
	return fmt.Sprintf("%s, %s:%d ", funcName, fileName, lineNo)

}
