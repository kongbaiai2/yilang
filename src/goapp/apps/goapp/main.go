package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"runtime"
	"time"

	"github.com/kongbaiai2/yilang/goapp/internal/callothers/cacti_proxy"
	"github.com/kongbaiai2/yilang/goapp/internal/global"
	"github.com/kongbaiai2/yilang/goapp/internal/routers"
	"github.com/kongbaiai2/yilang/goapp/internal/task"
	"github.com/kongbaiai2/yilang/goapp/pkg/middleware"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

var (
	buildGitRevision = ""
	buildTimestamp   = ""
	tagGitVersion    = ""
)

func profilling(addr string) {
	if addr != "" {
		panic(http.ListenAndServe(addr, nil))
	}
}

func setDefaultConfig(v *viper.Viper) {
	if ok := v.IsSet("system.httpAddr"); !ok {
		viper.Set("system.httpAddr", 80)
	}

	if ok := v.IsSet("system.ginMode"); !ok {
		viper.Set("system.ginMode", "release")
	}

	//if need, we can bind env variable: v.AutomaticEnv()
}

func initLog() {
	zapConf := middleware.ZapConf{
		Level:       global.CONFIG.Zap.Level,
		Path:        global.CONFIG.Zap.Path,
		Format:      global.CONFIG.Zap.Format,
		Prefix:      global.CONFIG.Zap.Prefix,
		EncodeLevel: global.CONFIG.Zap.EncodeLevel,
	}

	logrotateConf := middleware.LogrotateConf{
		MaxSize:    global.CONFIG.LogRotate.MaxSize,
		MaxBackups: global.CONFIG.LogRotate.MaxBackups,
		MaxAges:    global.CONFIG.LogRotate.MaxAges,
		Compress:   global.CONFIG.LogRotate.Compress,
	}

	global.LOG = middleware.Zap(zapConf, logrotateConf)

	zapConf = middleware.ZapConf{
		Level:       global.CONFIG.Zap.Level,
		Path:        global.CONFIG.Zap.PathDb,
		Format:      global.CONFIG.Zap.Format,
		Prefix:      global.CONFIG.Zap.Prefix,
		EncodeLevel: global.CONFIG.Zap.EncodeLevel,
	}
	global.DBLOG = middleware.Zap(zapConf, logrotateConf)

}

func initDefaultParameters() {
	initLog()
	// global.DB = model.InitDBAndLog(global.CONFIG.Mysql, global.DBLOG)
	// global.REDIS = cache.Redis()
	// global.SqlDB, _ = global.DB.DB()
	// global.Cacti = cacti_proxy.NewCactiClient(global.CONFIG.CactiCfg.BaseURL, global.CONFIG.CactiCfg.Username, global.CONFIG.CactiCfg.Password)
	cacti_proxy.Cacti = &cacti_proxy.CactiOptions{}
	cacti_proxy.Cacti.SetConfig(cacti_proxy.CactiConfig{
		URL:      global.CONFIG.CactiCfg.BaseURL,
		Username: global.CONFIG.CactiCfg.Username,
		Password: global.CONFIG.CactiCfg.Password,
	}) //.LoginCacti()

}

func execTask() {
	// 0 0 0 3 1 *
	// task.SendMail()
	// task.GenDataAndGraph()

	task.AddFunc(global.CONFIG.Crantab, task.GenDataAndGraph)
	task.Start()

}

func main() {
	configfile := flag.String("f", "./config.yaml", "config file path")
	pAppVersion := flag.Bool("v", false, "show version")
	flag.Parse()

	if *pAppVersion {
		fmt.Printf("app:%s, goversion:%s, git:%s, time:%s, tag:%s", "goapp", runtime.Version(), buildGitRevision, buildTimestamp, tagGitVersion)
		return
	}

	// viper support parse yaml、toml、json file
	v := middleware.ViperParseConf(*configfile)
	setDefaultConfig(v)
	if err := v.Unmarshal(&global.CONFIG); err != nil {
		log.Fatal(err.Error())
	}

	go profilling(global.CONFIG.System.PprofPort)

	// 初始化默认参数
	initDefaultParameters()

	global.LOG.Infof("go tools is restarting...")
	execTask()

	gin.SetMode(global.CONFIG.System.GinMode)
	router := routers.NewRouter()
	s := &http.Server{
		Addr:           global.CONFIG.System.HttpPort,
		Handler:        router,
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	global.LOG.Infof("app:%s, goversion:%s, git:%s, time:%s, tag:%s", "goapp", runtime.Version(), buildGitRevision, buildTimestamp, tagGitVersion)

	if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		fmt.Printf("s.ListenAndServe err: %v", err)
	}
}
