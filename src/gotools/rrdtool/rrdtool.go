package rrdtool

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/urfave/cli/v2"
	"github.com/ziutek/rrd"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

const (
	dbfile    = "./test.rrd"
	pngfile   = "./test_rrd1.png"
	soucsv    = "./test.csv"
	dstcsv    = "./save_test.csv"
	step      = 300
	heartbeat = 2 * step

	timeLayout = "2006-01-02 15:04:05" //转化所需模板
)

var (
	color1 = "00ff00"
	color2 = "0000ff"
	color3 = "ffD801"
	color4 = "00ffff"
	color5 = "e238ec"
)

type Csvhandler struct {
	Headler     []CsvHeadle          `json:"headlers"`
	TimeValue   map[int]CsvTimeValue `json:"time_values"`
	LenDS       int                  `json:"len"`
	ArrayDsName []string             `json:"array_ns_name"`
	Rrd         RRDer
	Dbfile      string
	Pngfile     string
	SouCsv      string
	DstCsv      string
	StartTime   time.Time
	StartStr    string
	StartUnix   int64
	EndTime     time.Time
	EndStr      string
	EndUnix     int64
	Ratio       float64
	// Other       []string `json:"other"`
}
type CsvHeadle struct {
	HeadlerMap map[string][]string `json:"heads"`
}
type CsvTimeValue struct {
	TimeValueMap map[int64][]string `json:"time_value"`
}
type RRDer struct {
	Create *rrd.Creator
	Update *rrd.Updater
	Graph  *rrd.Grapher
	Cdef   *rrd.Exporter
	Info   map[string]interface{}
}
type FetchRrd struct {
	XCount       int
	YCount       int
	ResValueList [][]float64
	ResTimeList  []time.Time
	Rsult95th    []float64
}

func StringToFloat64(arr []string) []float64 {
	retInt64 := []float64{}
	for _, row := range arr {
		if len(row) == 0 {
			continue
		}

		// int64Num, _ := strconv.ParseInt(vv[0], 10, 64)
		customFloat64, err := strconv.ParseFloat(row, 64)
		// customInt64, err := strconv.ParseInt(row, 10, 64)
		if err != nil {
			log.Printf("stringToFloat64 err:%v", err)
			return nil
		}
		retInt64 = append(retInt64, customFloat64)

	}

	return retInt64
}
func (c *Csvhandler) GetFileCsv(csv_file string) error {

	csvFile, err := os.Open(csv_file)
	if err != nil {
		return err
	}

	// UTF8BOM as chinese Languages
	reader := csv.NewReader(transform.NewReader(bufio.NewReader(csvFile), unicode.UTF8BOM.NewDecoder()))
	reader.Comma = ','
	reader.FieldsPerRecord = -1 //每行的列可以不相等
	// reader := csv.NewReader(bufio.NewReader(csvFile))

	i := 0
	j := 0
	m := 0
	n := 0

	for {
		line, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		typeMap := CsvTimeValue{
			TimeValueMap: make(map[int64][]string),
		}
		tempMap := CsvHeadle{
			HeadlerMap: make(map[string][]string),
		}
		if len(line) >= 2 {
			if theTime, err := time.ParseInLocation(timeLayout, line[0], time.Local); err == nil {
				typeMap.TimeValueMap[theTime.Unix()] = line[1:]

				if j == 0 {
					j = i
					c.LenDS = len(line) - 1
				}
				c.TimeValue[i] = typeMap

			} else if the2Time, err := time.ParseInLocation(timeLayout, line[1], time.Local); err == nil {
				// log.Printf("the2Time:%v", the2Time)
				if n == 0 {
					c.StartTime = the2Time
					c.StartStr = line[1]
					c.StartUnix = the2Time.Unix()
				}
				if n == 1 {
					c.EndTime = the2Time
					c.EndStr = line[1]
					c.EndUnix = the2Time.Unix()
				}
				n++

			} else {
				if m == 0 {
					m = i
				}
				tempMap.HeadlerMap[line[0]] = line[1:]
				c.Headler = append(c.Headler, tempMap)
			}
		}
		i++

	}

	return nil
}
func (c *Csvhandler) CreateRRD(dbfile string, starttime time.Time, len int, step uint) error {
	// Create

	// c := rrd.NewCreator(dbfile, time.Now(), step)
	c.Rrd.Create = rrd.NewCreator(dbfile, starttime, step)

	cc := c.Rrd.Create
	cc.RRA("LAST", 0.5, 1, 12000)
	cc.RRA("AVERAGE", 0.5, 1, 12000)
	cc.RRA("MAX", 0.5, 5, 12000)

	// GAUGE DERIVE COUNTER ABSOLUTE COMPUTE
	for i := 0; i < len; i++ {
		dsname := fmt.Sprintf("line_%d", i)
		cc.DS(dsname, "GAUGE", heartbeat, 0, "U")
		// cc.DS("in", "GAUGE", heartbeat, 0, "U")
		// cc.DS("out", "COUNTER", heartbeat, 0, "U")
		log.Printf("create DS:%v, name: %v", i, dsname)

		c.ArrayDsName = append(c.ArrayDsName, dsname)
	}

	err := cc.Create(true)
	if err != nil {
		return err
	}
	return nil
}

func (c *Csvhandler) UpdateRRD(dbfile string) error {
	// // Update
	// u := rrd.NewUpdater(dbfile)

	c.Rrd.Update = rrd.NewUpdater(dbfile)
	u := c.Rrd.Update
	dateArr := c.TimeValue

	for i := len(c.Headler); i <= len(dateArr)+len(c.Headler)+2; i++ {
		list := make([]interface{}, 0)
		for k, v := range dateArr[i].TimeValueMap {
			// arrInt64 := StringToFloat64(v)

			list = append(list, fmt.Sprintf("%d", k))
			for _, vv := range v {
				list = append(list, vv)
			}

			// log.Println(i, list)
			err := u.Update(list...)
			if err != nil {
				fmt.Println(err)
				return err
			}
		}
	}
	return nil
}

func DataGrapher(g *rrd.Grapher, dsname string, color string, sUnix, eUnix int64, value95 float64, ratio_f float64) {
	def_name := dsname
	vdef_last := fmt.Sprintf("%s,LAST", def_name)
	vdef_ave := fmt.Sprintf("%s,AVERAGE", def_name)
	vdef_max := fmt.Sprintf("%s,MAXIMUM", def_name)
	vdef_name_last := dsname + "last"
	vdef_name_ave := dsname + "ave"
	vdef_name_max := dsname + "max"

	// cdef_ratio := fmt.Sprintf("%s,%s,*", def_name, ratio_f)
	// cdef_name := fmt.Sprintf("%s_%s", dsname, ratio_f)
	// g.CDef(cdef_name, cdef_ratio) // 设置等比大小
	// g.Line(1, cdef_name, "ff0000", "95th\\: ")
	if ratio_f > 0 && ratio_f != 1 {
		cdef_ratio := fmt.Sprintf("%s,%f,*", def_name, ratio_f)
		cdef_name := fmt.Sprintf("%s_ratiof", dsname)

		g.Def(def_name, dbfile, dsname, "AVERAGE", "start="+strconv.FormatInt(sUnix, 10), "end="+strconv.FormatInt(eUnix, 10))

		g.CDef(cdef_name, cdef_ratio) // 设置等比大小
		g.VDef(vdef_name_last, fmt.Sprintf("%s,LAST", cdef_name))
		g.VDef(vdef_name_ave, fmt.Sprintf("%s,AVERAGE", cdef_name))
		g.VDef(vdef_name_max, fmt.Sprintf("%s,MAXIMUM", cdef_name))

		g.Line(1, cdef_name, color, dsname)

		g.GPrint(vdef_name_last, "Current\\: %4.2lf%s")
		g.GPrint(vdef_name_ave, "Average\\: %4.2lf%s")
		g.GPrint(vdef_name_max, "Maximum\\: %4.2lf%s")
		g.Comment("\\n")

		strValue95 := fmt.Sprintf("%v", value95)
		g.HRule(strValue95, color, "95th\\: ")
		print95 := fmt.Sprintf("%.4f", value95)
		g.Comment(print95)
		g.Comment("\\n")
	} else if ratio_f == 1 {

		g.Def(def_name, dbfile, dsname, "AVERAGE", "start="+strconv.FormatInt(sUnix, 10), "end="+strconv.FormatInt(eUnix, 10))
		g.VDef(vdef_name_last, vdef_last)
		g.VDef(vdef_name_ave, vdef_ave)
		g.VDef(vdef_name_max, vdef_max)
		// g.CDef()

		g.Line(1, def_name, color, dsname)
		g.GPrint(vdef_name_last, "Current\\: %4.2lf%s")
		g.GPrint(vdef_name_ave, "Average\\: %4.2lf%s")
		g.GPrint(vdef_name_max, "Maximum\\: %4.2lf%s")
		g.Comment("\\n")

		strValue95 := fmt.Sprintf("%v", value95)
		g.HRule(strValue95, color, "95th\\: ")
		print95 := fmt.Sprintf("%.4f", value95)
		g.Comment(print95)
		g.Comment("\\n")
	}
	// cdef_vname := fmt.Sprintf("%s,95,PERCENT", dsname) // 这里95值取值 PERCENTNAN PERCENT,不好用，自定义排序后算95值。
	// VDEF:perc95=mydata,95,PERCENT
	// cdef_vname := fmt.Sprintf("%s,95,PERCENT", def_name+"95") // 这里95值取值
	// line95th := dsname + "95th"

	// g.Def(def_name+"95", dbfile, value95, "MAX", "start="+strconv.FormatInt(sUnix, 10), "end="+strconv.FormatInt(eUnix, 10))
	// g.VDef(line95th, cdef_vname)
	// g.HRule(line95th, color, "95th\\: ")

	// g.GPrint(line95th, "%.6lf")
}
func (c *Csvhandler) CreateGrapher(dbfile, file_png string, start, end time.Time, line95 *FetchRrd, ratio float64) error {
	if ratio <= 0 {
		// log.Panic("errors: ratio <=0 ")
		return fmt.Errorf("errors: ratio <=0 ")
	}
	// Graph Headler
	// g := rrd.NewGrapher()
	c.Rrd.Graph = rrd.NewGrapher()
	g := c.Rrd.Graph
	g.SetTitle("Traffic\n")
	g.SetVLabel("bits per second")
	g.SetSize(700, 200)
	g.SetBase(uint(1000))
	g.SetSlopeMode()
	g.SetLowerLimit(0)
	FromTo := fmt.Sprintf("From %v To %v\\c", c.StartStr, c.EndStr)
	FromTo = strings.Replace(FromTo, ":", "\\:", -1)
	g.Comment(FromTo)
	g.Comment("  \\n")
	// g.SetWatermark("yipeng")
	// g.SetRightAxisLabel("label string")
	// g.SetImageFormat("PDF") // PNG|SVG|EPS|PDF|XML|XMLENUM|JSON|JSONTIME|CSV|TSV|SSV

	// Graph data
	color_auto := ""
	for i := 0; i < len(c.ArrayDsName); i++ {
		switch i % 5 {
		case 0:
			color_auto = color1
		case 1:
			color_auto = color2
		case 2:
			color_auto = color3
		case 3:
			color_auto = color4
		case 4:
			color_auto = color5
		}

		DataGrapher(g, c.ArrayDsName[i], color_auto, c.StartUnix, c.EndUnix, line95.Rsult95th[i], ratio)

	}

	// g.Def("v2", dbfile, "out", "AVERAGE")
	// g.VDef("last2", "v2,LAST")
	// g.VDef("avg2", "v2,AVERAGE")
	// g.VDef("max2", "v2,MAXIMUM")

	// g.Line(1, "v2", "0000ff", "var 2--")
	// g.GPrint("last2", "当前值\\: %lf")
	// g.GPrint("avg2", "平均值\\: %lf")
	// g.GPrint("max2", "最大值\\: %lf")
	// g.Comment("\\n")

	// g.PrintT("max1", "最大值\\: %lf")
	// g.Print("avg2", "avg2=%lf")

	// now := time.Now()

	// i, err := g.SaveGraph(file_png, now.Add(-20*time.Second), now)
	i, err := g.SaveGraph(file_png, start, end)
	log.Printf("grapher info: %+v\n", i)
	if err != nil {
		return err
	}

	return nil
}

// 归并排序，分治法。
func MergeSort(arr []float64) []float64 {

	length := len(arr)
	if length < 2 {
		return arr
	}
	middle := length / 2
	left := arr[0:middle]
	right := arr[middle:]
	return Merge(MergeSort(left), MergeSort(right))
}

func Merge(left []float64, right []float64) []float64 {
	var result []float64
	for len(left) != 0 && len(right) != 0 {
		if left[0] <= right[0] {
			result = append(result, left[0])
			left = left[1:]
		} else {
			result = append(result, right[0])
			right = right[1:]
		}
	}

	for len(left) != 0 {
		result = append(result, left[0])
		left = left[1:]
	}

	for len(right) != 0 {
		result = append(result, right[0])
		right = right[1:]
	}

	return result
}

func (c *Csvhandler) FetchRRD(dbfile string, start, end time.Time, ratio float64) (resStruct *FetchRrd, err error) {
	// Fetch
	// end := time.Unix(int64(inf["last_update"].(uint)), 0)
	// start := end.Add(-20 * step * time.Second)

	// end := time.Unix(1659877360, 0)
	// start := end.Add(-10)
	// fmt.Printf("Fetch Params:\n")
	// fmt.Printf("Start: %s\n", start)
	// fmt.Printf("End: %s\n", end)
	// fmt.Printf("Step: %s\n", step*time.Second)
	fetchRes, err := rrd.Fetch(dbfile, "AVERAGE", start, end, step*time.Second)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer fetchRes.FreeValues()
	fmt.Printf("FetchResult:\n")
	fmt.Printf("Start: %s\n", fetchRes.Start)
	fmt.Printf("End: %s\n", fetchRes.End)
	fmt.Printf("Step: %s\n", fetchRes.Step)
	for _, dsName := range fetchRes.DsNames {
		fmt.Printf("\t%s", dsName)
	}
	fmt.Printf("\n")

	row := 0
	resStruct = &FetchRrd{
		ResTimeList:  []time.Time{},
		ResValueList: [][]float64{},
	}

	for ti := fetchRes.Start.Add(fetchRes.Step); ti.Before(end) || ti.Equal(end); ti = ti.Add(fetchRes.Step) {
		resStruct.ResTimeList = append(resStruct.ResTimeList, ti)
		// fmt.Printf("%s / %d", ti, ti.Unix())
		rowsList := []float64{}
		for i := 0; i < len(fetchRes.DsNames); i++ {
			v := fetchRes.ValueAt(i, row)
			// fmt.Printf("\t%e", v)
			rowsList = append(rowsList, v)
		}
		// fmt.Printf("\n")
		row++
		resStruct.ResValueList = append(resStruct.ResValueList, rowsList)
	}

	resStruct.YCount = len(fetchRes.DsNames)
	resStruct.XCount = row
	return
}
func (c *Csvhandler) FetchRRDXport(dbfile string, start, end time.Time, ratio float64) (resStruct *FetchRrd, err error) {
	// Xport

	resStruct = &FetchRrd{
		ResTimeList:  []time.Time{},
		ResValueList: [][]float64{},
	}

	e := rrd.NewExporter()

	for i := 0; i < len(c.ArrayDsName); i++ {
		def_name := c.ArrayDsName[i]
		cdef_ratio := fmt.Sprintf("%s,%f,*", def_name, ratio)
		cdef_name := fmt.Sprintf("%s_ratio", def_name)

		e.Def(def_name, dbfile, c.ArrayDsName[i], "AVERAGE") // "start="+strconv.FormatInt(sUnix, 10), "end="+strconv.FormatInt(eUnix, 10))
		e.CDef(cdef_name, cdef_ratio)                        // 设置等比大小
		e.XportDef(cdef_name, "ratio")
	}

	xportRes, err := e.Xport(start, end, step*time.Second)
	if err != nil {
		fmt.Println(err)
	}
	defer xportRes.FreeValues()
	fmt.Printf("XportResult:\n")
	fmt.Printf("Start: %s\n", xportRes.Start)
	fmt.Printf("End: %s\n", xportRes.End)
	fmt.Printf("Step: %s\n", xportRes.Step)
	for _, legend := range xportRes.Legends {
		fmt.Printf("\t%s", legend)
	}
	fmt.Printf("\n")

	row := 0
	for ti := xportRes.Start.Add(xportRes.Step); ti.Before(end) || ti.Equal(end); ti = ti.Add(xportRes.Step) {
		// fmt.Printf("%s / %d", ti, ti.Unix())
		resStruct.ResTimeList = append(resStruct.ResTimeList, ti)
		rowsList := []float64{}
		for i := 0; i < len(xportRes.Legends); i++ {
			v := xportRes.ValueAt(i, row)
			// fmt.Printf("\t%e", v)
			rowsList = append(rowsList, v)
		}
		// fmt.Printf("\n")
		row++
		resStruct.ResValueList = append(resStruct.ResValueList, rowsList)
	}
	resStruct.YCount = len(xportRes.Legends)
	resStruct.XCount = row
	return
}

func Sort95th(tempStruct *FetchRrd) {
	for i := 0; i < tempStruct.YCount; i++ {
		xRow := []float64{}
		for j := 0; j < tempStruct.XCount; j++ {
			xRow = append(xRow, tempStruct.ResValueList[j][i])
		}
		// log.Printf("count:x-%v,y-%v", tempStruct.XCount, i)
		sortFloat := MergeSort(xRow)
		rsult := sortFloat[int(float32(len(sortFloat))*0.95)+1]
		tempStruct.Rsult95th = append(tempStruct.Rsult95th, rsult)
	}
}

func Get95th(csvhandler *Csvhandler, ratio float64) (tempStruct *FetchRrd) {
	tempStruct = &FetchRrd{}
	var err error

	if ratio > 0 && ratio != 1 {
		tempStruct, err = csvhandler.FetchRRDXport(csvhandler.Dbfile, csvhandler.StartTime, csvhandler.EndTime, ratio)
		if err != nil {
			log.Println(err)
			return
		}
	} else if ratio == 1 {
		tempStruct, err = csvhandler.FetchRRD(csvhandler.Dbfile, csvhandler.StartTime, csvhandler.EndTime, ratio)
		if err != nil {
			log.Println(err)
			return
		}
		// tempStruct = temp
	}

	// 返回值多列, 求每列的百分比的第95个值
	Sort95th(tempStruct)
	return
}
func GoRRDtool(cli *cli.Context) {
	// func GoRRDtool() {
	log.Println("rrdtool")

	csvhandler := &Csvhandler{
		Headler:   []CsvHeadle{},
		TimeValue: make(map[int]CsvTimeValue),
	}
	csvhandler.Dbfile = cli.String("rrd_file")   // dbfile
	csvhandler.Pngfile = cli.String("png_file")  // pngfile
	csvhandler.SouCsv = cli.String("source_csv") // soucsv
	csvhandler.DstCsv = cli.String("dst_csv")    // dstcsv
	csvhandler.Ratio = cli.Float64("ratio")
	ratio := csvhandler.Ratio
	if ratio < 0 {
		log.Println("errors: ratio <=0 ")
		return
	}

	// data source file cvs
	err := csvhandler.GetFileCsv(csvhandler.SouCsv)
	if err != nil {
		log.Println(err)
		return
	}
	err = csvhandler.CreateRRD(csvhandler.Dbfile, csvhandler.StartTime, csvhandler.LenDS, step)
	if err != nil {
		log.Printf("createrrd:%v", err)
		return
	}
	err = csvhandler.UpdateRRD(csvhandler.Dbfile)
	if err != nil {
		log.Printf("UpdateRRD:%v", err)
		return
	}
	info, err := rrd.Info(csvhandler.Dbfile)
	if err != nil {
		log.Println(err)
		return
	}
	csvhandler.Rrd.Info = info

	// 从rrd中取95值
	typeStruct := Get95th(csvhandler, ratio)
	err = csvhandler.CreateGrapher(csvhandler.Dbfile, csvhandler.Pngfile, csvhandler.StartTime, csvhandler.EndTime, typeStruct, ratio)
	if err != nil {
		log.Printf("CreateGrapher:%v", err)
		return
	}

	log.Printf("start_unix:%v, end_unix:%v\n", csvhandler.StartUnix, csvhandler.EndUnix)
	log.Printf("time:%v,end:%v,count:%v", csvhandler.StartTime, csvhandler.EndTime, len(csvhandler.TimeValue))
}

func AddRRDtool(goCom []*cli.Command) []*cli.Command {
	// sliceFlag := &cli.StringSlice{} //&[]string{"kongbaiai2@qq.com"}
	Command := &cli.Command{
		Name:    "rrdtool",
		Aliases: []string{"rrd"},
		Usage:   "example:\trrdtool -s ./test.csv",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "png_file",
				Aliases: []string{"p"},
				Usage:   "specify create graph name case: *.png",
				Value:   "./test.png",
			},
			&cli.StringFlag{
				Name:    "rrd_file",
				Aliases: []string{"r"},
				Usage:   "specify create graph name case: *.rrd",
				Value:   "./test.rrd",
			},
			// &cli.StringSliceFlag{
			// 	Name:    "rrd_file",
			// 	Aliases: []string{"p"},
			// 	Usage:   "specify create rrd name case: *.rrd",
			// 	Value:   sliceFlag,
			// },
			&cli.StringFlag{
				Name:    "source_csv",
				Aliases: []string{"s"},
				Usage:   "requisite specify data csv file case: *.csv",
				Value:   "./test.csv",
			},
			&cli.StringFlag{
				Name:    "dst_csv",
				Aliases: []string{"d"},
				Usage:   "specify data csv save case: save_*.csv",
				Value:   "plase waite develop",
			},
			&cli.Float64Flag{
				Name:    "ratio",
				Aliases: []string{"ra"},
				Usage:   "specify data csv save case: save_*.csv",
				Value:   1,
			},
		},
		Action: func(cli *cli.Context) error {

			GoRRDtool(cli)
			return nil
		},
	}
	//
	goCom = append(goCom, Command)
	return goCom
}

//
