package main

import (
	"time"

	"gorm.io/gorm"
)

var (
	db     *gorm.DB
	dbPath string
	TOTAL  = "TOTAL"

	isFlushChart = true
)

const (
	// start to end get api data
	endT   = 20 * time.Minute
	startT = 60 * time.Minute
	// diff data to alarm if < data/2
	ratio = 2
	// cron run get api data
	cronTime        = "0 */5 * * * *"
	CranTimeOneHour = "7 7 */1 * * *"
)

var (
	Url        = "https://p.cdn.10086.cn"
	timeout    = 15 * time.Second
	tenant_id  = "1ogcrgmq-gycy38lfu552b2c3"
	tenant_key = "0AZuaLFJ7rakIe6F"
	mailTo     = []string{"wull@yipeng888.com", "wd@yipeng888.com", "yinye@yipeng888.com"}
)

type AuthRequestPostDate struct {
	DateTime      string   `json:"datetime"`
	Authorization PostAuth `json:"authorization"`
}

type PostAuth struct {
	Tenant_id string `json:"tenant_id"`
	Sign      string `json:"sign"`
}

type Cmcc struct {
	Token        string
	Domain       string
	ApiToken     string
	ApiStatistic string
	TenantDomain string
	StatisticArgs
}

type StatisticArgs struct {
	Domain     string // 要查的域名
	Detail     int    // 1查所有地区，包含0，0查全国
	Start      string // 时间
	End        string // 时间
	DomainSum  bool   // 域名带宽汇总
	IpProtocol string // 0，ipv4；1，ipv6；all，v4+v6
}

type ResponseStatistic struct {
	Domain string              `json:"domain"`
	Data   []ResponseProvinces `json:"data"`
}

type ResponseProvinces struct {
	Time      string     `json:"time"`
	Provinces []AreaList `json:"provinces"`
}

type AreaList struct {
	Area  string  `json:"area"`
	Value float64 `json:"value"`
}
type StatisticBw struct {
	Time      string  `json:"time" gorm:"type:varchar(20) COMMENT '时间'"`
	Value     float64 `json:"value" gorm:"type:decimal(30,10) COMMENT '值'"`
	Provinces string  `json:"provinces" gorm:"type:varchar(10) COMMENT '地市';uniqueIndex:prov_unix"`
	UnixTime  int64   `json:"unix_time,omitempty" gorm:"type:int(64);uniqueIndex:prov_unix"`
}

type ProvincesName struct {
	gorm.Model
	Provinces string `json:"provinces" gorm:"type:varchar(10) COMMENT '地市';uniqueIndex:provinces"`
	Name      string `json:"name" gorm:"type:varchar(20) COMMENT '地市名';"`
}

type CmccDomain struct {
	gorm.Model
	Domain   string `json:"domain" gorm:"type:varchar(50);uniqueIndex:domain_ips"`
	Isp      string `json:"isp" gorm:"type:varchar(50);uniqueIndex:domain_ips"`
	Supplier string `json:"supplier" gorm:"type:varchar(80)"`
}

type AlarmCfg struct {
	StartTime interface{}
	UsePlug   string
	Body      string
	Total     []float64
}

type AlarmTable struct {
	gorm.Model
	Host   string `json:"host,omitempty" gorm:"type:varchar(20);uniqueIndex:keySyn"`
	Ticket string `json:"ticket" gorm:"type:varchar(50)  COMMENT '标识';uniqueIndex:keySyn"`

	Status   int    `json:"status" gorm:"type:int(8) COMMENT '0:ok,1:problem,2:delete'"`
	Hit      int    `json:"hit" gorm:"type:int(8) COMMENT '触发后一小时命中次数'"`
	Body     string `json:"body" gorm:"type:varchar(256)"`
	UsePlugs string `json:"use_plugs" gorm:"type:varchar(10) COMMENT 'mail,dingtalk,wachat'"`
}

type AlarmOldNew struct {
	OldDB *AlarmTable
	New   *AlarmTable
}
