package main

import (
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"

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

func initDbAndLog() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	dir, err := homedir.Dir()
	if err != nil {
		log.Fatal("get home dir failed:", err)
	}
	logdir := path.Join(dir, "log/cmcc_alarm.log")
	// log.Println(logdir)
	logFile, err := os.OpenFile(logdir, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0777)
	twoWrite := io.MultiWriter(logFile, os.Stdout)
	if err != nil {
		log.Printf("failed to open %v, discarding any log.", logdir)
		log.SetOutput(ioutil.Discard)
	} else {
		log.SetOutput(twoWrite)
	}
	// defer logFile.Close()

	CreateSQLiteDb()
	initInfoToDb()
	// if err != nil {
	// 	if errMySQL, ok := err.(*mysql.MySQLError); ok {
	// 		switch errMySQL.Number {
	// 		case 1062:
	// 			log.Println("ignore db Duplicate errors")
	// 			// TODO handle Error 1062: Duplicate entry '%s' for key %d
	// 		}
	// 	}
	// }
}
