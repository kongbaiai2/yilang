package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/go-echarts/go-echarts/v2/types"

	_ "github.com/go-sql-driver/mysql"
)

// func Query(db *sql.DB, dbInfo string, query string) (result map[int]map[string]string, err error) {
// 	rows, err := db.Query(query)
// 	if err != nil {
// 		return result, err
// 	}
// 	defer rows.Close()
// 	cols, _ := rows.Columns()               //返回所有列
// 	vals := make([][]byte, len(cols))       //这里表示一行所有列的值，用[]byte表示
// 	scans := make([]interface{}, len(cols)) //这里表示一行填充数据
// 	for k, _ := range vals {
// 		scans[k] = &vals[k]
// 	}
// 	i := 0
// 	result = make(map[int]map[string]string)
// 	for rows.Next() {
// 		rows.Scan(scans...)            //填充数据
// 		row := make(map[string]string) //每行数据
// 		for k, v := range vals {       //把vals中的数据复制到row中
// 			key := cols[k]
// 			//fmt.Println(string(v)) //这里把[]byte数据转成string
// 			row[key] = string(v)

// 		}
// 		result[i] = row
// 		i++
// 	}
// 	rows.Close()

// 	return result, nil
// }

func stringToInt(s string) (int, error) {
	toInit, err := strconv.Atoi(s)
	if err != nil {
		return -1, err
	}
	return toInit, nil

}
func stringToInt64(s string) (int64, error) {
	toInit64, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return -1, err
	}
	return toInit64, err
}

// func create line
func generateLineItems() []opts.LineData {
	items := make([]opts.LineData, 0)
	for i := 0; i < 7; i++ {
		items = append(items, opts.LineData{Value: rand.Intn(300)})
	}
	return items
}

func createChartLine(datas map[string][]opts.LineData, xItems []string) error {
	// p1 := charts.NewPage()
	Formattercustoms := []string{"{b}"}

	// create a new line instance
	line := charts.NewLine()

	// Put data into instance
	line.SetXAxis(xItems)
	for name, data := range datas {
		line.AddSeries(name, data)
	}

	total := len(datas) - 1
	for i := 0; i < total; i++ {
		Formattercustoms = append(Formattercustoms, fmt.Sprintf("{a%d}: {c%d} {a%d}: {c%d}", i, i, total, total))
		total--
	}
	// set some global options like Title/Legend/ToolTip or anything else
	line.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{PageTitle: "kongbaiai2@126.com", Theme: types.ThemeWesteros, Height: "600px", Width: "1000px"}),
		charts.WithTitleOpts(opts.Title{
			Title: "珠海移动cdn",
			Top:   "0%",
			Left:  "50%",
		}),
		charts.WithLegendOpts(opts.Legend{Show: true,
			Type: "scroll",
			Top:  "3%",
		}),
		charts.WithTooltipOpts(opts.Tooltip{
			Show: true,
			// TriggerOn: "mousemove",
			Formatter: strings.Join(Formattercustoms, "<br/>"),
			Trigger:   "axis",
		}),
	)

	line.SetSeriesOptions(charts.WithLineChartOpts(opts.LineChart{ShowSymbol: false}))

	f1, err := os.Create("./html/createChartLine.html")
	if err != nil {
		return err
	}
	line.Render(f1)
	return nil
}

func bwDataConvertLineItems(datas map[string][]StatisticBw) (mapLineData map[string][]opts.LineData, xitems []string) {
	mapLineData = make(map[string][]opts.LineData, 0)
	xInt := true
	for k, v := range datas {
		yitems := make([]opts.LineData, 0)
		for _, items := range v {
			yitems = append(yitems, opts.LineData{Value: strconv.FormatFloat(items.Value/1024/1024, 'f', 2, 64)})
			if xInt {
				xitems = append(xitems, items.Time)
			}
		}
		xInt = false
		mapLineData[k] = yitems
	}
	return mapLineData, xitems
}

func getData() error {
	now := time.Now()
	unix_time := now.Add(-time.Duration(24) * time.Hour).Unix()
	proNames, err := ProvincesNameSelect()
	if err != nil {
		return err
	}

	all := map[string][]StatisticBw{}
	for _, proName := range proNames {
		data, err := StatisticBwSelectWhereAsc(proName.Provinces, unix_time)
		if err != nil {
			log.Println(err)
			return err
		}

		if len(data) < 282 {
			log.Printf("name:%s,len:%d", proName.Name, len(data))
			continue
		}
		all[proName.Name] = data
	}
	mapLineData, xItems := bwDataConvertLineItems(all)

	err = createChartLine(mapLineData, xItems)
	if err != nil {
		return err
	}
	return nil
}
