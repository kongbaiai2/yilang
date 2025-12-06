/*
 1. configure:
    via YAML file, e.g:
    ```
    database:
    foodb:
    driver: mysql_auto
    src: user:password@/dbname?charset=utf8&parseTime=True&loc=Local
    bardb:
    driver: postgres
    src: host=myhost user=gorm dbname=gorm sslmode=disable password=mypassword
    ```
    OR via env:
    export APPNAME_DATABASE_FOODB_DRIVER=mysql_auto
    export APPNAME_DATABASE_FOODB_SRC="user:password@/dbname?charset=utf8&parseTime=True&loc=Local"
    export APPNAME_DATABASE_BARDB_DRIVER=postgres
    export APPNAME_DATABASE_BARDB_SRC="host=myhost user=gorm dbname=gorm sslmode=disable password=mypassword"

 2. usage:
    ```

    db , err := model.GetDB("foodb")
    if err != nil {
    // handle error
    }
    defer model.Close("foodb")

    db.Find(&users)

    ```

    see more: http://jinzhu.me/gorm/
*/
package model

import (
	"fmt"
	"log"
	"strings"
	"sync/atomic"

	"github.com/kongbaiai2/yilang/goapp/internal/config"

	"go.uber.org/zap"
	"gopkg.in/go-playground/validator.v8"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	validate   = validator.New(&validator.Config{TagName: "validate"})
	mysql_auto atomic.Value
)

// GetMySQL return *grom.DB, if success.
/*
func GetMySQL() (db *gorm.DB) {
	if v := mysql_auto.Load(); v == nil || v.(*gorm.DB) == nil {
		var err error
		db, err = gorm.Open(mysql.Open(global.CONFIG.Mysql.Path), &gorm.Config{})
		if err != nil {
			log.Fatal(err.Error())
		}

		sqlDb, err := db.DB()
		if err != nil {
			log.Fatal(err.Error())
		}

		sqlDb.SetMaxIdleConns(global.CONFIG.Mysql.MaxIdleConns)
		sqlDb.SetMaxOpenConns(global.CONFIG.Mysql.MaxOpenConns)
		//db.SetLogger(global.LOG)

		mysql_auto.Store(db)
	} else {
		db = v.(*gorm.DB)
	}

	return
}
*/

type Writer struct {
	W_zap *zap.SugaredLogger
}

func (w Writer) Printf(format string, args ...interface{}) {
	if len(args) < 2 {
		return
	}
	if long_fileName, ok := args[0].(string); ok {
		if strings.Contains(long_fileName, ".go:") {
			sort := long_fileName[strings.LastIndex(long_fileName, "/")+1:]
			args[0] = sort
		}
	}

	msg := fmt.Sprintf(format, args...) + "\n"
	fmt.Println(msg) // 打印到控制台

	// 替换掉彩色打印符号
	msg = strings.ReplaceAll(msg, logger.Reset, "")
	msg = strings.ReplaceAll(msg, logger.Red, "")
	msg = strings.ReplaceAll(msg, logger.Green, "")
	msg = strings.ReplaceAll(msg, logger.Yellow, "")
	msg = strings.ReplaceAll(msg, logger.Blue, "")
	msg = strings.ReplaceAll(msg, logger.Magenta, "")
	msg = strings.ReplaceAll(msg, logger.Cyan, "")
	msg = strings.ReplaceAll(msg, logger.White, "")
	msg = strings.ReplaceAll(msg, logger.BlueBold, "")
	msg = strings.ReplaceAll(msg, logger.MagentaBold, "")
	msg = strings.ReplaceAll(msg, logger.RedBold, "")
	msg = strings.ReplaceAll(msg, logger.YellowBold, "")

	w.W_zap.Infof(msg) // 输出到文件
}

func InitDBAndLog(mySqlConf config.Mysql, dbLog *zap.SugaredLogger) *gorm.DB {
	gormlogger := logger.New(
		Writer{
			W_zap: dbLog, // io.Writer
		},
		logger.Config{
			LogLevel:                  logger.Info, // Log level
			Colorful:                  true,        // 允许彩色打印
			IgnoreRecordNotFoundError: true,        // Ignore ErrRecordNotFound error for logger

		},
	)
	var err error
	db, err := gorm.Open(mysql.Open(mySqlConf.Path), &gorm.Config{Logger: gormlogger})
	if err != nil {
		log.Fatal(err.Error())
	}

	sqlDb, err := db.DB()
	if err != nil {
		log.Fatal(err.Error())
	}

	sqlDb.SetMaxIdleConns(mySqlConf.MaxIdleConns)
	sqlDb.SetMaxOpenConns(mySqlConf.MaxOpenConns)

	return db
}

func InitDB(mySqlConf config.Mysql, filename string) *gorm.DB {
	var err error
	db, err := gorm.Open(mysql.Open(mySqlConf.Path), &gorm.Config{})
	if err != nil {
		log.Fatal(err.Error())
	}

	sqlDb, err := db.DB()
	if err != nil {
		log.Fatal(err.Error())
	}

	sqlDb.SetMaxIdleConns(mySqlConf.MaxIdleConns)
	sqlDb.SetMaxOpenConns(mySqlConf.MaxOpenConns)

	return db
}
