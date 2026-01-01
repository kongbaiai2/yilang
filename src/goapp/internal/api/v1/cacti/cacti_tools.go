package cacti

import (
	"context"
	"fmt"
	"log"
	"math"
	"os"
	"time"

	"github.com/kongbaiai2/yilang/goapp/internal/callothers/cacti_proxy"
	"github.com/kongbaiai2/yilang/goapp/internal/global"
	"github.com/kongbaiai2/yilang/goapp/pkg/runner"
)

var chinaLoc *time.Location

func init() {
	var err error
	chinaLoc, err = time.LoadLocation("Asia/Shanghai")
	if err != nil {
		log.Fatal("Failed to load Asia/Shanghai timezone:", err)
	}
	os.MkdirAll(global.CONFIG.CactiCfg.ImgPath, os.ModePerm)
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

type DataValueResult struct {
	Data  string
	Value float64
}

func ProcessMonthly(localGraphID, month_ago int, isDown bool) (string, float64, error) {
	start, end := getPreviousMonthRange(month_ago)
	monthStr := time.Unix(start, 0).In(chinaLoc).Format("200601")

	filenamePrefix := fmt.Sprintf("month_%d_%s", localGraphID, monthStr)
	cacti_proxy.Cacti.FlushLogin() // 刷新登录，上次未登出情况下，本次登录不检验了。
	c := cacti_proxy.Cacti
	graph := cacti_proxy.Graph{}
	graph.Set(localGraphID, start, end, filenamePrefix, isDown)

	p95, err := c.Do(&graph)
	if err != nil {
		global.LOG.Errorf("Error processing monthly data: %v", err)
		return monthStr, 0, err
	}

	return monthStr, math.Round(p95/1000000*100) / 100, nil
}

func ProcessDaily(localGraphID, month_ago int, isDown bool) (day_str []DataValueResult, err error) {
	// 获取上个月第一天和最后一天
	now := time.Now().In(chinaLoc)
	days := []time.Time{}
	if month_ago == 0 {
		// 取最近24小时
		lastDay := now.AddDate(0, 0, -1)
		days = append(days, lastDay)
		isDown = true
	} else { // ago := MonthAgo(now, month_ago)
		firstOfThisMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, chinaLoc)
		firstOfLastMonth := firstOfThisMonth.AddDate(0, -month_ago, 0)

		year := firstOfLastMonth.Year()
		month := firstOfLastMonth.Month()

		days = getAllDaysInMonth(year, month)
	}

	cacti_proxy.Cacti.FlushLogin()
	c := cacti_proxy.Cacti

	// var err error
	// day_str := []string{}
	tasks := []runner.TaskItem{}
	for _, day := range days {
		graph := cacti_proxy.Graph{}

		filenamePrefix := fmt.Sprintf("everyday_%d_%s", localGraphID, day.Format("20060102"))

		start := day.Unix()
		end := day.Add(24*time.Hour - time.Second).Unix()

		graph.Set(localGraphID, start, end, filenamePrefix, isDown)
		tasks = append(tasks, runner.NewTaskItem("api-call", &runner.ContextualVoidTask{
			Fn: func(ctx context.Context) error {
				p95, tmp_err := c.Do(&graph)
				if tmp_err != nil {
					global.LOG.Errorln(tmp_err)
					return tmp_err
				}
				every_day := DataValueResult{
					Data:  day.Format("20060102"),
					Value: math.Round(p95/1000000*100) / 100,
				}
				day_str = append(day_str, every_day)
				return nil
			},
		}, runner.TaskPolicy{
			MaxRetry:    1,
			TaskTimeout: 30 * time.Second,
		}))

		// day_str = append(day_str, fmt.Sprintf("%s: %.2f 95th", day.Format("20060102"), p95/1000000))

		// time.Unix(start, 0).In(chinaLoc).Format("20060102 150405")
	}
	TaskResults := ProcessDailyTask(tasks)
	for _, taskResult := range TaskResults {
		if taskResult.Err != nil {
			global.LOG.Warnf("%+v", taskResult)
		}
	}
	global.LOG.Info(day_str)

	// if err != nil {
	// 	global.LOG.Errorf("Error processing daily data: %v", err)
	// 	return
	// }
	// global.LOG.Errorf("success, day p95: \n%v ", strings.Join(day_str, ",\n"))
	return

}

func ProcessDailyTask(tasks []runner.TaskItem) []runner.TaskResult {
	return runner.DefaultRunConcurrency(tasks)
}

func MonthAgo(t time.Time, n int) time.Time {
	return t.AddDate(0, -n, 0)
}

// getPreviousMonthRange 返回上个月的 start (00:00:00) 和 end (23:59:59) 的 Unix 时间戳（UTC）
func getPreviousMonthRange(month_ago int) (start, end int64) {
	now := time.Now().In(chinaLoc) // 当前北京时间

	// ago := MonthAgo(now, month_ago)
	// 本月 1 号 00:00:00（北京时间）
	firstOfThisMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, chinaLoc)
	// 上个月 1 号（北京时间）
	firstOfLastMonth := firstOfThisMonth.AddDate(0, -month_ago, 0)
	// 上个月最后一天 23:59:59（北京时间）
	lastOfLastMonth := firstOfThisMonth.Add(-time.Second)

	return firstOfLastMonth.Unix(), lastOfLastMonth.Unix()
}

// getAllDaysInMonth 返回某年某月所有日期（按中国时间）
func getAllDaysInMonth(year int, month time.Month) []time.Time {
	first := time.Date(year, month, 1, 0, 0, 0, 0, chinaLoc)
	var days []time.Time
	for d := first; d.Month() == month; d = d.AddDate(0, 0, 1) {
		days = append(days, d)
	}
	return days
}
