package task

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/kongbaiai2/yilang/goapp/internal/api/v1/cacti"
	"github.com/kongbaiai2/yilang/goapp/internal/api/v1/mail"
	"github.com/kongbaiai2/yilang/goapp/internal/global"
	"github.com/kongbaiai2/yilang/goapp/pkg/errcode"
	"github.com/kongbaiai2/yilang/goapp/pkg/utils"
	"github.com/robfig/cron/v3"
)

var Cron *cron.Cron
var Cvs_dir string = "cvs"

func init() {
	loc, _ := time.LoadLocation("Asia/Shanghai") // 或用 time.FixedZone("CST", 8*60*60)
	Cron = cron.New(cron.WithLocation(loc), cron.WithSeconds())
	os.MkdirAll(Cvs_dir, os.ModePerm)

}

func AddFunc(cran_taime string, f func()) (cron.EntryID, error) {
	return Cron.AddFunc(cran_taime, f)
}
func Start() {
	global.LOG.Infoln("cron start...")
	Cron.Start()
}

func GenMonthCvs(paras []cacti.GetPercentMonthlyRequest) (string, error) {
	// datavalue := [][]cacti.DataValueResult{}
	str := ""
	for _, para := range paras {
		month_tmp, err := GetPercentMonthly(para.GraphID, para.MonthAgo, para.IsDown)
		if err != nil {
			global.LOG.Errorln(err)
			return "", err
		}
		// datavalue = append(datavalue, month_tmp)
		if para.GraphID == 889 {
			str += fmt.Sprintf("Tx month %s: %.2f\n<p>", month_tmp[0].Data, month_tmp[0].Value)
		} else if para.GraphID == 1166 {
			str += fmt.Sprintf("youyuan month %s: %.2f\n<p>", month_tmp[0].Data, month_tmp[0].Value)
		} else if para.GraphID == 1091 {
			str += fmt.Sprintf("xiantong month %s: %.2f\n<p>", month_tmp[0].Data, month_tmp[0].Value)
		}
	}

	// total_month := MergeByDate(datavalue)

	// csv := MergedRows(total_month).ToCSVString()
	// uuid := utils.GenUUID32()[:8]
	// filename := fmt.Sprintf("%s/month_%s.csv", Cvs_dir, uuid)
	// if err := os.WriteFile(filename, []byte(csv), 0644); err != nil {
	// 	global.LOG.Errorln(err)
	// 	return "", err
	// }
	// global.LOG.Infoln("CSV save ", filename)
	return str, nil
}

func GenDayCvs(paras []cacti.GetPercentMonthlyRequest) (string, error) {
	datavalue := [][]cacti.DataValueResult{}

	for _, para := range paras {
		day_tmp, err := GetPercentEverDay(para.GraphID, para.MonthAgo, para.IsDown)
		if err != nil {
			global.LOG.Errorln(err)
			return "", err
		}
		datavalue = append(datavalue, day_tmp)
	}

	total_day := MergeByDate(datavalue)

	csv := MergedRows(total_day).ToCSVString()
	uuid := utils.GenUUID32()[:8]
	filename := fmt.Sprintf("%s/day_%s.csv", Cvs_dir, uuid)
	if err := os.WriteFile(filename, []byte(csv), 0644); err != nil {
		global.LOG.Errorln(err)
		return "", err
	}
	global.LOG.Infoln("CSV save ", filename)
	return filename, nil
}

func GenDataAndGraph() {
	// tx: 889
	// youyuan: 1166
	// xiantong: 1091

	month_lists := []cacti.GetPercentMonthlyRequest{
		{GraphID: 889, MonthAgo: 1, IsDown: true},
		{GraphID: 1166, MonthAgo: 1, IsDown: true},
		{GraphID: 1091, MonthAgo: 1, IsDown: true},
	}
	str, err := GenMonthCvs(month_lists)
	if err != nil {
		global.LOG.Errorln(err)
		return
	}

	day_lists := []cacti.GetPercentMonthlyRequest{
		{GraphID: 1166, MonthAgo: 1, IsDown: false},
	}
	f_name2, err := GenDayCvs(day_lists)
	if err != nil {
		global.LOG.Errorln(err)
		return
	}

	img, err := getFileForDir(global.CONFIG.CactiCfg.ImgPath)
	if err != nil {
		global.LOG.Errorln(err)
		return
	}
	img = append(img, f_name2)
	err = SendMail(img, str)
	if err != nil {
		global.LOG.Errorln(err)
		return
	}

	opt := cacti.DeleteDirRequest{
		Dir: Cvs_dir,
	}
	cacti.DeleteDirWork(&opt)
	opt.Dir = global.CONFIG.CactiCfg.ImgPath
	cacti.DeleteDirWork(&opt)

}

func GetPercentMonthly(graphID int, monthAgo int, isDown bool) ([]cacti.DataValueResult, error) {
	opt := cacti.GetPercentMonthlyRequest{
		GraphID:  graphID,
		MonthAgo: monthAgo,
		IsDown:   isDown,
	}

	err, res := cacti.GetPercentMonthlyWork(&opt)
	if err != errcode.StatusSuccess {
		global.LOG.Errorln(err)
		return nil, err
	}

	return res.Values, nil
}

func GetPercentEverDay(graphID int, monthAgo int, isDown bool) ([]cacti.DataValueResult, error) {
	opt := cacti.GetPercentEveryDayRequest{
		GraphID:  graphID,
		MonthAgo: monthAgo,
		IsDown:   isDown,
	}

	err, res := cacti.GetPercentEveryDayWork(&opt)
	if err != errcode.StatusSuccess {
		global.LOG.Errorln(err)
		return nil, err
	}

	return res.Values, nil
}

func SendMail(filename []string, body string) error {
	opt := mail.SendMailRequest{
		Smtp: mail.SmtpConfig{
			Host:     global.CONFIG.Mail.Smtp.Host,
			Port:     global.CONFIG.Mail.Smtp.Port,
			Username: global.CONFIG.Mail.Smtp.Username,
			Password: os.Getenv("SMTP_PASSWORD")},
		Header: mail.MailHeader{
			From:    global.CONFIG.Mail.Header.From,
			To:      global.CONFIG.Mail.Header.To,
			Subject: global.CONFIG.Mail.Header.Subject,
			Body:    global.CONFIG.Mail.Header.Body + "\n<p>" + body,
			Attach:  filename,
		},
	}

	err := mail.SendMailWork(&opt)
	if err != errcode.StatusSuccess {
		return err
	}
	global.LOG.Infoln("SendMail success")

	return nil
}

// MergeByDate 按 Data 字段合并多个 []DataValueResult，返回按日期排序的矩阵
// 返回：[]MergedRow，其中 MergedRow.Data 是日期，MergedRow.Values[i] 是第 i 个 ret 在该日期的值（缺失为 0）
func MergeByDate(rets [][]cacti.DataValueResult) []MergedRow {
	// Step 1: 收集所有日期，并构建 date -> [val1, val2, ...] 映射
	dateToValues := make(map[string][]float64)
	allDates := make(map[string]bool)

	// 初始化每个日期对应 len(rets) 个槽位，填 0.0
	for i := range rets {
		for _, r := range rets[i] {
			allDates[r.Data] = true
			if _, exists := dateToValues[r.Data]; !exists {
				dateToValues[r.Data] = make([]float64, len(rets))
			}
			dateToValues[r.Data][i] = r.Value
		}
	}

	// Step 2: 收集所有唯一日期并排序（按字典序即 YYYYMMDD 升序）
	var dates []string
	for d := range allDates {
		dates = append(dates, d)
	}
	sort.Strings(dates) // ✅ "20251101" < "20251102" < ...

	// Step 3: 构建结果
	var result []MergedRow
	for _, date := range dates {
		values := dateToValues[date]
		// 如果某 ret 没这个 date，则 values[i] 已是 0.0（初始化时未覆盖的位置保持 0）
		result = append(result, MergedRow{Data: date, Values: values})
	}

	return result
}

// MergedRow 是合并后的单行结果
type MergedRow struct {
	Data   string
	Values []float64 // 长度 == len(rets)
}

type MergedRows []MergedRow

// ✅ 现在可以为 MergedRows 定义方法！
func (mrs MergedRows) ToCSVString() string {
	if len(mrs) == 0 {
		return "Date\n"
	}

	// 表头
	headers := []string{"Date"}
	for i := 1; i <= len(mrs[0].Values); i++ {
		headers = append(headers, fmt.Sprintf("Value%d", i))
	}

	var lines []string
	lines = append(lines, joinCSV(headers))

	for _, r := range mrs {
		row := []string{r.Data}
		for _, v := range r.Values {
			row = append(row, strconv.FormatFloat(v, 'f', 1, 64))
		}
		lines = append(lines, joinCSV(row))
	}

	return joinLines(lines)
}

// 工具函数：CSV 字段转义（简单版：无逗号/换行时可省略引号）
func joinCSV(fields []string) string {
	for i, f := range fields {
		if contains(f, ",", "\n", "\"") {
			fields[i] = "\"" + strings.ReplaceAll(f, "\"", "\"\"") + "\""
		}
	}
	return strings.Join(fields, ",")
}

func joinLines(lines []string) string {
	return strings.Join(lines, "\n")
}

func contains(s string, chars ...string) bool {
	for _, c := range chars {
		if strings.Contains(s, c) {
			return true
		}
	}
	return false
}

func getFileForDir(dir string) ([]string, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var images []string
	for _, file := range files {
		if !file.IsDir() && (filepath.Ext(file.Name()) == ".png" || filepath.Ext(file.Name()) == ".jpg") {
			// 构造图片的相对路径
			imagePath := dir + "/" + file.Name()
			images = append(images, imagePath)
		}
	}
	return images, nil
}
