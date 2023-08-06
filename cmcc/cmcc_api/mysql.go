package cmcc_api

import (
	"math/rand"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func init() {
	rand.Seed(time.Now().Unix())
	// dir, err := homedir.Dir()
	// if err != nil {
	// 	log.Fatal("get home dir failed:", err)
	// }
	// dbPath = path.Join(dir, ".felix/sqlite.db")
	// dbPath = "cmcc:cmccZH@123.com@tcp(localhost:3306)/cmcc?charset=utf8mb4&parseTime=True&loc=Local"
	// "user:pass@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
}

// 插入db，重复则替换
func initInfoToDb() {
	ProvincesNameReplace("SH", "上海95峰值")
	ProvincesNameReplace("ZJ", "浙江95峰值")
	ProvincesNameReplace("JX", "江西95峰值")
	ProvincesNameReplace("NM", "内蒙古95峰")
	ProvincesNameReplace("SC", "四川95峰值")
	ProvincesNameReplace("SD", "山东95峰值")
	ProvincesNameReplace("HA", "河南95峰值")
	ProvincesNameReplace("HN", "湖南95峰值")
	ProvincesNameReplace("JS", "江苏95峰值")
	ProvincesNameReplace("BJ", "北京95峰值")
	ProvincesNameReplace("AH", "安徽95峰值")
	ProvincesNameReplace("TJ", "天津95峰值")
	ProvincesNameReplace("CQ", "重庆95峰值")
	ProvincesNameReplace("HE", "河北95峰值")
	ProvincesNameReplace("SX", "山西95峰值")
	ProvincesNameReplace("LN", "辽宁95峰值")
	ProvincesNameReplace("JL", "吉林95峰值")
	ProvincesNameReplace("HL", "黑龙江95峰")
	ProvincesNameReplace("FJ", "福建95峰值")
	ProvincesNameReplace("HB", "湖北95峰值")
	ProvincesNameReplace("GD", "广东95峰值")
	ProvincesNameReplace("GX", "广西95峰值")
	ProvincesNameReplace("HI", "海南95峰值")
	ProvincesNameReplace("GZ", "贵州95峰值")
	ProvincesNameReplace("YN", "云南95峰值")
	ProvincesNameReplace("XZ", "西藏95峰值")
	ProvincesNameReplace("SN", "陕西95峰值")
	ProvincesNameReplace("GS", "甘肃95峰值")
	ProvincesNameReplace("QH", "青海95峰值")
	ProvincesNameReplace("NX", "宁夏95峰值")
	ProvincesNameReplace("XJ", "新疆95峰值")
	ProvincesNameReplace("OTHER", "其它")
	ProvincesNameReplace(TOTAL, "全国95峰值")
}

func CreateMysqlDb(dbPath string) {
	// log.Println("sql in:")
	// sqlite, err := gorm.Open(mysql.Open(dbPath), &gorm.Config{})
	sql, err := gorm.Open(mysql.New(mysql.Config{
		DSN:                       dbPath, // DSN data source name
		DefaultStringSize:         256,    // string 类型字段的默认长度
		DisableDatetimePrecision:  true,   // 禁用 datetime 精度，MySQL 5.6 之前的数据库不支持
		DontSupportRenameIndex:    true,   // 重命名索引时采用删除并新建的方式，MySQL 5.7 之前的数据库和 MariaDB 不支持重命名索引
		DontSupportRenameColumn:   true,   // 用 `change` 重命名列，MySQL 8 之前的数据库和 MariaDB 不支持重命名列
		SkipInitializeWithVersion: false,  // 根据当前 MySQL 版本自动配置
	}), &gorm.Config{})

	if err != nil {
		logrus.WithError(err).Fatalf("master fail to open its sqlite db in %s. please install master first.", dbPath)
		return
	}

	db = sql
	//TODO::optimize
	//db.DropTable("term_logs")
	// 创建表时添加后缀
	db.Set("gorm:table_options", "ENGINE=InnoDB")
	db.AutoMigrate(&StatisticBw{}, &CmccDomain{}, &AlarmTable{})
	// db.Logger.LogMode(logger.Error)

	// 不存在则创建
	if !db.Migrator().HasTable(&ProvincesName{}) {
		db.Set("gorm:table_options", "ENGINE=InnoDB").Migrator().CreateTable(&ProvincesName{})
	}

	// db.Migrator().CreateIndex(&CmccDomain{}, "domain") // AddUniqueIndex("pro_name", "name")
	// db.Model(&ProvincesName{}).CreateIndex()
	// db.Migrator().CreateIndex(&ProvincesName{}, "Name")

}

func writeBwToDb(resultStruct []ResponseStatistic, host string) error {
	for _, v := range resultStruct {
		domain := v.Domain
		num := 0
		num = len(v.Data)
		var sum float64
		for _, vv := range v.Data {
			sum = 0
			the_time, err := time.ParseInLocation("2006-01-02 15:04:05", vv.Time, time.Local)
			unixTime := the_time.Unix()
			if err != nil {
				return err
			}

			for _, data := range vv.Provinces {
				if data.Area == TOTAL {
					err := StatisticBwReplace(vv.Time, data.Area, data.Value, unixTime, domain)
					if err != nil {
						return err
					}
					sum += data.Value
				}
			}
		}

		// 数据为0直接告警
		if sum == 0 {
			warnerChan <- NewAlarmDb(host, "数据为零: ", v.Domain, v.Domain, "mail", 1)
		}
		// log.Printf("insert into: %v, total: %v,domain: %v", vv.Time, sum, v.Domain)
		log.Printf("insert into count: %d, total: %f,domain %s", num, sum, domain)
	}
	return nil
}

// insert into statistic_bws(time, value,provinces,unix_time) value()
func StatisticBwReplace(time, provinces string, value float64, unix_time int64, domain string) error {
	ins := &StatisticBw{Time: time, Provinces: provinces, Value: value, UnixTime: unix_time, Domain: domain}
	// 插入冲突时，用最新的覆盖。
	return db.Clauses(clause.OnConflict{UpdateAll: true}).Create(&ins).Error
	// return db.Create(ins).Error
}

func StatisticBwInsert(time, provinces string, value float64, unix_time int64) error {
	ins := &StatisticBw{Time: time, Provinces: provinces, Value: value, UnixTime: unix_time}
	// 插入冲突时，用最新的覆盖。
	return db.Create(ins).Error
}

func StatisticBwSelectProvinces(total, domain string, num int) ([]StatisticBw, error) {
	// ins := &StatisticBw{Time: time, Provinces: provinces, Value: value, UnixTime: unix_time}
	var resq []StatisticBw
	query := db.Order("updated_at")
	if total != "" {
		query = db.Where("Provinces = ?", total).Where("Domain = ?", domain).Order("unix_time desc").Limit(num)
	}
	err := query.Find(&resq).Error
	return resq, err
}
func StatisticBwSelectWhereAsc(total string, unix int64) ([]StatisticBw, error) {
	// ins := &StatisticBw{Time: time, Provinces: provinces, Value: value, UnixTime: unix_time}
	var resq []StatisticBw
	query := db.Order("updated_at")
	if total != "" {
		query = db.Where("Provinces = ?", total).Where("unix_time > ?", unix).Order("unix_time").Limit(282)
	}
	err := query.Find(&resq).Error
	return resq, err
}

func ProvincesNameReplace(provinces, name string) error {
	ins := &ProvincesName{Name: name, Provinces: provinces}
	// 插入冲突时，用最新的覆盖。
	return db.Clauses(clause.OnConflict{UpdateAll: true}).Create(&ins).Error
	// return db.Create(ins).Error
}

func ProvincesNameSelect() ([]ProvincesName, error) {
	var resq []ProvincesName
	err := db.Find(&resq).Error

	return resq, err
}

func AlarmTableReplace(status, hit int, host, body, key, domain, usePlugs string) error {
	ins := &AlarmTable{Status: status, Hit: hit, Host: host, Body: body, Ticket: key, Domain: domain, UsePlugs: usePlugs}
	// 插入冲突时，用最新的覆盖。

	return db.Clauses(clause.OnConflict{UpdateAll: true}).Create(&ins).Error
	// return db.Create(ins).Error
}

func AlarmTableWhere(key, host string) (AlarmTable, error) {
	var resq AlarmTable
	k := &AlarmTable{Ticket: key, Host: host}
	query := db.Where(k)
	err := query.Find(&resq).Error

	return resq, err
}
func AlarmTableDelete(id uint) error {

	// err := db.Exec("delete from alarm_tables where id = ?", id).Error
	err := db.Unscoped().Where("id = ?", id).Delete(&AlarmTable{}).Error

	return err
}
