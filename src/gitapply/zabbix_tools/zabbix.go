package zabbix_tools

import (
	"database/sql"
	"fmt"
	. "glb_config"
	"log"
	"strings"
	. "utils"

	"github.com/urfave/cli/v2"
)

func GetItemidArray(db *sql.DB, resp Respond, sqlStr string) (itemid string, err error) {
	sqlstr := fmt.Sprintf(sqlStr, resp.Host, resp.IntName, resp.InOrOut)

	rows, err := db.Query(sqlstr)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&itemid)
		if err != nil {
			return
		}
		// log.Println(itemid)
	}
	if err = rows.Err(); err != nil {
		return
	}

	return
}
func GetValueForHistoryUnit(db *sql.DB, resp Respond, sqlStr, itemid string) (dataDb DataForHistory, err error) {
	dataDb.Itemid = itemid
	rows, err := db.Query(sqlStr, itemid, resp.StartTime, resp.EndTime)
	if err != nil {
		return dataDb, err
	}
	for rows.Next() {
		var data DataMapForHistory
		err = rows.Scan(&data.Clock, &data.Value)
		if err != nil {
			return dataDb, err
		}
		// log.Printf("itemid:%v,%v,%v", itemid, unix_t, value)
		dataDb.Datas = append(dataDb.Datas, data)
	}
	if err = rows.Err(); err != nil {
		return dataDb, err
	}

	return
}

// 转成每5分钟值.取5分钟内的值相加求平均，并没有要求源数据为1分钟。5分钟内没值则空。
func TurnItemidArray(resp Respond) (turndata DataForHistory, err error) {
	// time.Now().Format(time.RFC3339)
	cron := int64(300)
	// retNm := uint(5)
	RecordQ := NewSliceQueue(5)
	var sum, num, clockTmp int64

	max := len(resp.Itemids.Datas)
	turndata.Itemid = resp.Itemids.Itemid
	datamap := resp.Itemids.Datas
	isKeep := true
	for i := 0; i < max; i++ {
		var dmap DataMapForHistory
		// log.Println(time.Unix(datamap.Clock, 0).Format(timeLayout))

		if clockTmp == 0 {
			clockTmp = datamap[i].Clock
		} else {
			// 后面的时间 一定大于前面
			if datamap[i].Clock < clockTmp {
				return turndata, fmt.Errorf("Time disorder of data")
			}
		}
		// 数据大于等于2。判断在cron秒内的，则返回true，否返回false.
		RecordQ.Enqueue(&FlagInfo{AbnormalTime: datamap[i].Clock})

		// within five minutes to plus, but assign other value
		if RecordQ.IsWithinFiveMinutes(cron, "AbnormalTime") {
			sum += datamap[i].Value
			num++
			// log.Println(datamap[i].Clock, datamap[i].Value, sum)
			isKeep = false
		} else {
			// 当前时间在5分钟外,队列清空后，重新加下队列。
			RecordQ.Enqueue(&FlagInfo{AbnormalTime: datamap[i].Clock})
			isKeep = true
		}

		if i == max-1 {
			// 最后一个值
			isKeep = true
		}
		if num == 0 {
			continue
		}

		if isKeep {
			// 当前时间在5分钟外,用上一个值的时间赋值
			tmp := datamap[i].Clock
			if i-1 >= 0 && i != max-1 {
				tmp = datamap[i-1].Clock
			}
			dmap.Clock = tmp
			dmap.Value = sum / num
			sum = datamap[i].Value
			num = 1
			// log.Println(dmap.Value)
			turndata.Datas = append(turndata.Datas, dmap)
		}

	}

	return
}

func TurnMinutes(ctx *cli.Context) error {
	log.Println("TurnMinutes: 5 minutes average")
	var resp Respond

	// ifHCInOctets or ifHCOutOctets
	resp.Host = ctx.String("host")                   // "S6880-48S4Q"
	resp.IntName = "%" + ctx.String("intname") + "%" // "%10GE1/0/8%"
	resp.InOrOut = "%" + ctx.String("inout") + "%"   // "%ifHCInOctets%"
	resp.StartTime = ctx.String("start")             // "2023/05/01 00:00:00"
	resp.EndTime = ctx.String("end")                 // "2023/05/02 00:09:00"

	db, err := InitDb(Cfg.Dbstring)
	if err != nil {
		return err
	}

	itemid, err := GetItemidArray(db, resp, SelectItemidForHostInt)
	if err != nil {
		log.Printf("ERROR: GetItemidArray sql: %s,resp:%v err:%v", SelectItemidForHostInt, resp, err)
		return err
	}
	if itemid == "" {
		return fmt.Errorf("itemid is null: %s", itemid)
	}
	dataArr, err := GetValueForHistoryUnit(db, resp, SelectValueForHistoryUnit+ConditionTimeToTime, itemid)
	if err != nil {
		log.Printf("ERROR: GetValueForHistoryUnit sql: %s,resp:%v err:%v", SelectValueForHistoryUnit+ConditionTimeToTime, resp, err)
		return err
	}
	resp.Itemids = dataArr

	turnData, err := TurnItemidArray(resp)
	if err != nil {
		return err
	}
	resp.TurnItemids = turnData

	// for _, v := range resp.TurnItemids.Datas {
	// 	fmt.Printf("%v %v\n", UnixToTimeString(v.Clock), v.Value)
	// }
	// log.Println(resp)
	runes := `_`
	port := strings.Replace(ctx.String("intname"), "/", runes, -1)
	filename := fmt.Sprintf("%s-%s-%s.csv", resp.Host, port, ctx.String("inout"))
	err = SaveCsv("/tmp/one_"+filename, resp.Itemids.Datas)
	if err != nil {
		return err
	}
	err = SaveCsv("/tmp/five_"+filename, resp.TurnItemids.Datas)
	if err != nil {
		return err
	}
	return nil
}
