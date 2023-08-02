package cmcc_api

import (
	"fmt"
	"log"
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
	if prov[0].UnixTime < startTime {
		the_time, err := time.ParseInLocation("2006-01-02 15:04:05", prov[0].Time, time.Local)
		if err != nil {
			return err
		}
		ticket := "not get cmcc api data: "
		body := fmt.Sprintf("befor %v minute,dbtime:%v,curtime:%v<br>", start, the_time, time.Now())
		warnerChan <- NewAlarmDb(host, ticket, domain, body, "mail", 1)
	}

	i := 0
	if prov[i].UnixTime-prov[i+1].UnixTime != 300 {
		body := fmt.Sprintf("\n\r%v to %v<br>", prov[i].UnixTime, prov[i+1].UnixTime)
		ticket := "cmcc api data lack: "
		warnerChan <- NewAlarmDb(host, ticket, domain, body, "mail", 1)
	}
	// 骤降，小于上个数的Threshold比例
	if prov[i].Value < prov[i+1].Value*float64(alert.Threshold) {
		body := fmt.Sprintf("%v to %v, \r\n<br>Hit:%v<%v * %v", prov[i].Time, prov[i+1].Time, prov[i].Value, prov[i+1].Value, alert.Threshold)
		ticket := "cmcc api data fall: "
		warnerChan <- NewAlarmDb(host, ticket, domain, body, "mail", 1)
	}

	return nil
}
func AlarmFall() {
	for {
		for domain, alertcfg := range mapAlertCfg {
			// 取db最新值汇总,Provicnes=TOTAL,start为开始时间
			TotalProvicnes, err := StatisticBwSelectProvinces("TOTAL", domain, 2)
			if err != nil {
				log.Println(domain, " get statistic bw data failed", err)
			}
			diffFallData(alertcfg, TotalProvicnes, 30*time.Minute, domain, alertcfg.MailWarn.Host)
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
	// 查ticket，host告警是否存在，更新alarm信息
	a.OldDB = a.getAlarmInfoFromDb(ticket, host)
	if a.OldDB.Ticket == ticket && a.OldDB.Host == host {
		hit = a.OldDB.Hit + 1
	} else {
		hit = 1
		a.OldDB = a.New
	}

	// 60分钟前未触发的告警，则删除
	interval := 60 * time.Minute
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

	log.Printf("Alarm triggered for the %d count, host:%v,ticket:%v,body:%v,plugs:%v", hit, host, ticket, body, usePlugs)
	err := AlarmTableReplace(1, hit, host, body, ticket, usePlugs)
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
	// 小于2，100以内 10的倍数不削峰，100以外50倍数不削
	if a.OldDB.Hit < 2 {
		return false
	}
	if a.OldDB.Hit%10 > 100 {
		return a.OldDB.Hit%50 == 0
	}
	if a.OldDB.Hit%10 == 0 {
		return false
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
	m_cli := &M_cli{
		From:        mcfg.MailFrom,
		To:          mcfg.MailTo,
		Subject:     a.New.Host,
		Body:        time.Now().Format(time.RFC3339) + "<br>" + a.New.Ticket + "<br>" + a.New.Body,
		Enclosure:   "", // 附件
		Smtp_domain: mcfg.SmtpDomain,
		Port:        mcfg.Port,
		User:        mcfg.User,
		Password:    mcfg.Password,
	}
	log.Printf("send to:%v, body:%v", mcfg.MailTo, a.New.Body)

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
