package glb_config

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/fsnotify/fsnotify"
	_ "github.com/go-sql-driver/mysql"
	"github.com/spf13/viper"
)

func init() {
	GetConfig(&Cfg, "glb_config", "json", ".", "./glb_config", "../glb_config")
}

func InitDb(dbstring string) (*sql.DB, error) {
	// log.Println(info)
	// dbString := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&loc=Local", info.User, info.Passwd, info.Host, info.Port, info.Name)
	db, err := sql.Open("mysql", dbstring)
	if err != nil {
		log.Println(err)
		return db, err
	}
	if err = db.Ping(); err != nil {
		return db, err
	}
	db.SetMaxOpenConns(16)
	db.SetMaxIdleConns(16)

	// rows, err := db.Query("show tables")
	// if err != nil {
	// 	log.Println(err)
	// 	return db, err
	// }
	// defer rows.Close()

	// for rows.Next() {
	// 	var ret string
	// 	err = rows.Scan(&ret)
	// 	if err != nil {
	// 		return db, err
	// 	}
	// 	log.Println("ret:", ret)
	// }
	// if err = rows.Err(); err != nil {
	// 	return db, err
	// }
	return db, nil
}

func GetConfig(cfg *config, cfgname, suffix string, dirArr ...string) {

	viper := viper.New()
	// viper.SetDefault("key2", "value2")
	// viper.SetConfigFile("./config.yaml")

	viper.SetConfigName(cfgname) // 配置文件名,不需要后缀名
	viper.SetConfigType(suffix)  // 配置文件格式json
	for _, dir := range dirArr {
		viper.AddConfigPath(dir) // 查找配置文件的路径
		// viper.AddConfigPath("./config")     // 查找配置文件的路径
		// viper.AddConfigPath("$HOME/yilang") // 查找配置文件的路径
		// viper.AddConfigPath(".")            // 查找配置文件的路径
	}

	err := viper.ReadInConfig() // 查找并读取配置文件
	if err != nil {             // 处理错误
		panic(fmt.Errorf("Fatal error config file: %w \n", err))
	}

	err = viper.Unmarshal(cfg)
	if err != nil {
		panic(err)
	}

	// 监听配置文件变更
	viper.OnConfigChange(func(e fsnotify.Event) {
		log.Println("Config file changed:", e.Name)
		err = viper.Unmarshal(cfg)
		if err != nil {
			panic(err)
		}
	})
	viper.WatchConfig()

	// log.Println(viper.Get("url"))
	// log.Println(viper.GetString("ratio"))

	// 是否存在user
	// if viper.IsSet("user") {
	// 	fmt.Println("key user is not exists")
	// }
	// 打印所有
	// m := viper.AllSettings()
	// fmt.Println(m)
}
