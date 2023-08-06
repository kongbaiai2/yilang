package cmcc_api

import (
	"fmt"
	"runtime"
	"time"
)

var (
	warnerChan chan Alarmer
	MinuteFive int64 = 300
)

func diffFallData(alert alertCfg, prov []StatisticBw, start time.Duration, domain, host string) error {
	startTime := (time.Now().Add(-start).Unix() / MinuteFive) * MinuteFive
	if len(prov) < 2 {
		return nil
	}
	ticket := ""
	body := ""
	isalarm := false
	if prov[0].UnixTime < startTime {
		// the_time, err := time.ParseInLocation("2006-01-02 15:04:05", prov[0].Time, time.Local)
		// if err != nil {
		// 	return err
		// }
		ticket += "没数据:"
		body += fmt.Sprintf(" %v 分钟之前没数据,\n\r<br>最后一次时间:%v,最后一次数据:%v\n\r<br>", start, prov[0].Time, prov[0].Value)
		isalarm = true
	}

	i := 0
	if prov[i].UnixTime-prov[i+1].UnixTime != 300 {
		body += fmt.Sprintf("\n\r<br>当前时间:%v, 前一个时间:%v\n\r<br>", prov[i].Time, prov[i+1].Time)
		ticket += "间隔5分钟:"
		isalarm = true
	}
	// 骤降，小于上个数的Threshold比例
	if prov[i].Value < prov[i+1].Value*float64(alert.Threshold) {
		body += fmt.Sprintf("当前时间:%v \r\n<br>前一个时间:%v, \r\n<br>当前值:%v 小于 前一个值:%v * %v\n\r<br>", prov[i].Time, prov[i+1].Time, prov[i].Value, prov[i+1].Value, alert.Threshold)
		ticket += "骤降:"
		isalarm = true
	}
	if isalarm {
		warnerChan <- NewAlarmDb(host, ticket, domain, body, "mail", 1)
		log.Printf("alearm domain:%s,host:%s,data:%v", domain, host, prov)
		isalarm = false
	}
	return nil
}
func AlarmFall() {
	time.Sleep(10 * time.Second)
	for {
		for domain, alertcfg := range mapAlertCfg {
			// 取db最新值汇总,Provicnes=TOTAL,start为开始时间
			TotalProvicnes, err := StatisticBwSelectProvinces("TOTAL", domain, 2)
			if err != nil {
				log.Println(domain, " get statistic bw data failed", err)
			}
			t := time.Duration(alertcfg.DelayMinute+30) * time.Minute

			diffFallData(alertcfg, TotalProvicnes, t, domain, alertcfg.MailWarn.Host)
		}
		time.Sleep(5 * time.Minute)
	}
}

type Alarmer interface {
	GetPlug() string
	GetDomain() string
	UseMail(mcfg mailCfg) error
	UseDingTalk() error
	UseWeChat() error
	AlarmPeakLimit() bool
}

func NewAlarmDb(host, ticket, domain, body, usePlugs string, status int) Alarmer {
	return &AlarmOldNew{New: &AlarmTable{Host: host, Ticket: ticket, Domain: domain, Body: body, UsePlugs: usePlugs, Status: status}}
}

func (a *AlarmOldNew) getAlarmInfoFromDb(ticket, host string) *AlarmTable {
	altab, err := AlarmTableWhere(ticket, host)
	if err != nil {
		log.Println(err)
		return &altab
	}
	return &altab
}

// 插入告警，存在更新时间和hit，不存在插入新的
func (a *AlarmOldNew) insertAlarmToDB() error {
	var (
		hit int
	)
	ticket := a.New.Ticket
	host := a.New.Host
	usePlugs := a.New.UsePlugs
	body := a.New.Body
	domain := a.New.Domain
	// 查ticket，host告警是否存在，更新alarm信息
	a.OldDB = a.getAlarmInfoFromDb(ticket, host)
	if a.OldDB.Ticket == ticket && a.OldDB.Host == host {
		hit = a.OldDB.Hit + 1
	} else {
		hit = 1
		a.OldDB = a.New
	}

	// 60分钟前未触发的告警，则删除
	interval := 180 * time.Minute
	now := time.Now()
	currT := now.Sub(a.OldDB.UpdatedAt)
	if a.OldDB.UpdatedAt != a.New.UpdatedAt && currT > interval {
		// 周期内没有触发，则告警删除
		log.Printf("%v not triggered,update cycle %v ;\nhost:%v,ticket:%v,body:%v,plugs:%v",
			currT, interval, a.OldDB.Host, a.OldDB.Ticket, a.OldDB.Body, a.OldDB.UsePlugs)
		err := AlarmTableDelete(a.OldDB.ID)
		if err != nil {
			log.Println(err)
		}
		hit = 1
	}

	log.Printf("Alarm triggered for the hit: %d, host:%v,ticket:%v,plugs:%v,body:%v", hit, host, ticket, usePlugs, body)
	err := AlarmTableReplace(1, hit, host, body, ticket, domain, usePlugs)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}
func (a *AlarmOldNew) AlarmPeakLimit() bool {
	err := a.insertAlarmToDB()
	if err != nil {
		return false
	}
	if a.New.Status == 0 {
		return true
	}
	// 削峰
	if a.OldDB.Hit < 4 {
		return false
	}
	if a.OldDB.Hit > 100 {
		return a.OldDB.Hit%50 != 0
	}
	if a.OldDB.Hit > 10 {
		return a.OldDB.Hit%10 != 0
	}
	return true
}

func (a *AlarmOldNew) GetPlug() string {
	return a.New.UsePlugs
}
func (a *AlarmOldNew) GetDomain() string {
	return a.New.Domain
}
func (a *AlarmOldNew) UseMail(mcfg mailCfg) error {
	body := fmt.Sprintf("时间：%s\r\n<br>告警项：%s\r\n<br>告警内容：%s\r\n<br>告警域名：%s\r\n<br>命中数：%d\r\n<br>",
		time.Now().Format(time.RFC3339), a.New.Ticket, a.New.Body, a.New.Domain, a.OldDB.Hit)
	m_cli := &M_cli{
		From:        mcfg.MailFrom,
		To:          mcfg.MailTo,
		Subject:     a.New.Host,
		Body:        body,
		Enclosure:   "", // 附件
		Smtp_domain: mcfg.SmtpDomain,
		Port:        mcfg.Port,
		User:        mcfg.User,
		Password:    mcfg.Password,
	}
	log.Printf("send to:%v, domain: %s", mcfg.MailTo, a.New.Domain)

	err := m_cli.SendMail()
	if err != nil {
		return err
	}

	return nil
}
func (a *AlarmOldNew) UseDingTalk() error {
	return nil
}
func (a *AlarmOldNew) UseWeChat() error {
	return nil
}

func AlarmListenDb() {
	defer func() { //必须要先声明defer，否则不能捕获到panic异常
		if err := recover(); err != nil {
			b := [4096]byte{}
			runtime.Stack(b[:], true)
			log.Printf("[ERROR] catch panic:%s,b:%v", err, string(b[:]))
		}
	}()

	ticker := time.NewTicker(24 * time.Hour)
	warnerChan = make(chan Alarmer, 10)
	for {
		// warnerChan <- NewAlarmDb("test1", "ticket test", "body test", "mail", 1)
		select {
		case alarm := <-warnerChan:

			// 收到告警后先削峰
			if alarm.AlarmPeakLimit() {
				continue
			}

			mailcfg := mapAlertCfg[alarm.GetDomain()].MailWarn

			switch alarm.GetPlug() {
			case "mail":
				// main body
				err := alarm.UseMail(mailcfg)
				if err != nil {
					log.Println(err)
				}
			case "dingtalk":
				alarm.UseDingTalk()
			case "wechat":
				alarm.UseWeChat()
			}
		case <-ticker.C:
			// 定期邮件，判断服务正常。
			warnerChan <- NewAlarmDb("珠海移动CDN数据API", "定期自检: ", "zhimage.guangdongyunchen.com", "是否发邮件正常，可忽略", "mail", 1)
		}
	}
}
