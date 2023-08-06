package cmcc_api

import (
	"fmt"

	"net/http"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/gin-gonic/gin"
	"github.com/robfig/cron"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type cronJob struct {
	ctx *TenantItem
}
type cronJobHour struct {
	ctx *TenantItem
}

func setCron() {
	isAlarm := true
	if isAlarm {

		// 监听告警事件
		go AlarmListenDb()

		// 骤降告警,触发告警
		go AlarmFall()

		isAlarm = false
	}

	mapAlertCfg = make(map[string]alertCfg, 5)
	c := cron.New()
	for _, ctxPtr := range cfg.CustomList {
		if ctxPtr.IsDisable {
			log.Println(ctxPtr.Tenant.Domain, "is disable.")
			continue
		}
		ctx := ctxPtr
		mapAlertCfg[ctx.Tenant.Domain] = ctx.GetData.Alert
		// isAlarm := true
		cronjob := &cronJob{
			ctx: &ctx,
		}
		cronjobhour := &cronJobHour{
			ctx: &ctx,
		}
		c.AddJob(ctx.GetData.CronTime, cronjob)
		c.AddJob(ctx.GetData.CranTimeOneHour, cronjobhour)

		CmccDomainAdd(ctx.Tenant.Domain, "cmcc", "zhuhaiyidong")
		cronjob.Run()

	}

	c.Start()
	log.Println("cmcc running success")
}
func (c *cronJob) Run() {
	doCronTask(c.ctx)
}
func doCronTask(ctx *TenantItem) {
	// 对外接口刷新chart的标记
	isFlushChart = true
	// 取start end之间的数据。这里是一小时前 到 半小时前的数据
	startT := time.Duration(ctx.GetData.Alert.DelayMinute) * time.Minute
	endT := startT + 30*time.Minute
	err := ctx.cmcc(startT, endT)
	if err != nil {
		log.Println(err)
	}
}
func (c *cronJobHour) Run() {
	doCronTaskOneHour(c.ctx)
}
func doCronTaskOneHour(ctx *TenantItem) {
	err := ctx.cmcc(1*time.Hour, 3*time.Hour)
	if err != nil {
		log.Println(err)
	}
}
func Hello(c *gin.Context) {
	c.String(200, "hello %s", "world")
}
func runGin() {
	GenerateUuid()
	r := gin.Default()

	r.LoadHTMLGlob("html/*")
	r.GET("/line", func(c *gin.Context) {
		// every 5 minutes set true in func doCronTask()
		if isFlushChart {
			isFlushChart = false
			err := getData()
			if err != nil {
				c.String(404, "get db data err %s", err)
			}
		}

		c.HTML(http.StatusOK, "createChartLine.html", gin.H{
			"result": c.Param("content"),
		})
	})
	r.GET("/auto", Auto_sw_cmds)
	r.GET("/hello", Hello)
	r.Run(":" + cfg.ListenPort)
}

// func Runmain() {
// 	defer func() {
// 		log.Println("exit")
// 	}()
// 	initDbAndLog()

// 	// 监听告警事件
// 	go AlarmListenDb()
// 	// 骤降告警
// 	go AlarmFall()
// 	// 定时执行任务
// 	setCron()

// 	runGin()
// }

func GetConfig(cfg *Config, cfgname, suffix string, dirArr ...string) {

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
		panic(fmt.Errorf("fatal error config file: %w", err))
	}

	err = viper.Unmarshal(cfg)
	if err != nil {
		panic(err)
	}

	// 监听配置文件变更
	viper.OnConfigChange(func(e fsnotify.Event) {
		log.Warn("Config file changed:", e.Name)
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

type Config struct {
	ListenPort string       `mapstructure:"listen_port"`
	LogLevel   uint32       `mapstructure:"log_level"`
	DbCfg      string       `mapstructure:"db_config"`
	Licence    string       `mapstructure:"licence"`
	CustomList []TenantItem `mapstructure:"custom_list"`
	AutoLimits []AutoLimit  `mapstructure:"auto_limit"`
}
type AutoLimit struct {
	IsDisable   bool     `mapstructure:"is_disable"`
	SwIpList    []string `mapstructure:"sw_ip_list"`
	Description string   `mapstructure:"description"`
	Cmd         string   `mapstructure:"cmd"`
	UndoCmd     string   `mapstructure:"undo_cmd"`
	Encrypt     string   `mapstructure:"encrypt"`
	User        string   `mapstructure:"user"`
}
type TenantItem struct {
	Tenant    tenantCfg  `mapstructure:"tenant"`
	GetData   getDataCfg `mapstructure:"get_data"`
	Chart     chartCfg   `mapstructure:"chart"`
	IsDisable bool       `mapstructure:"is_disable"`
}
type tenantCfg struct {
	Domain       string `mapstructure:"domain"`
	Url          string `mapstructure:"url"`
	TenantId     string `mapstructure:"tenant_id"`
	TenantKey    string `mapstructure:"tenant_key"`
	TokenApi     string `mapstructure:"token_api"`
	ApiStatistic string `mapstructure:"api_statistic"`
}
type getDataCfg struct {
	CronTime        string   `mapstructure:"cron_time"`
	CranTimeOneHour string   `mapstructure:"cran_time_one_hour"`
	Alert           alertCfg `mapstructure:"alert"`
}
type alertCfg struct {
	MailWarn    mailCfg `mapstructure:"mail_warn"`
	Threshold   float64 `mapstructure:"threshold"`
	DelayMinute int     `mapstructure:"delay_minute"`
}
type mailCfg struct {
	Host       string   `mapstructure:"host"`
	MailFrom   string   `mapstructure:"mail_from"`
	MailTo     []string `mapstructure:"mail_to"`
	SmtpDomain string   `mapstructure:"smtp_domain"`
	Port       int      `mapstructure:"port"`
	User       string   `mapstructure:"user"`
	Password   string   `mapstructure:"password"`
}
type chartCfg struct {
	ShowTime string `mapstructure:"show_hour"`
}

var cfg Config
var key []byte
var mapAlertCfg map[string]alertCfg
var log *logrus.Logger
var glb_uuid string

func GetCfg() *Config {
	return &cfg
}
func NewRun(keysyn []byte) {
	log = logrus.New()
	key = keysyn
	t := time.Now()
	defer func() {
		log.Println(time.Since(t))
	}()
	// read config
	GetConfig(&cfg, "cmcc", "json", ".", "./config", "../config")
	// log.Printf("%#v", cfg)

	// init log
	initDbAndLog("cmcc_alarm.log")

	// 定时执行任务
	setCron()

	runGin()
}
