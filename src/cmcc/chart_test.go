package main

import (
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"testing"

	"github.com/mitchellh/go-homedir"
)

func Test_chart(t *testing.T) {
	// CreateSQLiteDb()
	// initInfoToDb()
	initDbAndLogtest()
	// err := getData()

	// select unix_timestamp("2023-02-16 22:06:45");
	getData()

}

func initDbAndLogtest() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	dir, err := homedir.Dir()
	if err != nil {
		log.Fatal("get home dir failed:", err)
	}
	logdir := path.Join(dir, "log/test.log")
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
