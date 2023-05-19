package utils

import (
	"encoding/csv"
	"fmt"
	. "glb_config"
	"log"
	"os"
	"time"
)

func UnixToTimeString(unix int64) string {
	// timelayout := time.RFC3339
	timeLayout := "2006-01-02 15:04:05"
	return time.Unix(unix, 0).Format(timeLayout)
}

func SaveCsv(csvfilename string, Datas []DataMapForHistory) error {
	file, err := os.Create(csvfilename)
	if err != nil {
		log.Println(err)
		return err
	}
	defer file.Close()
	file.WriteString("\xEF\xBB\xBF")
	writer := csv.NewWriter(file)
	writer.Comma = ','

	allRow := [][]string{}
	for _, v := range Datas {
		xRow := []string{}
		xRow = append(xRow, UnixToTimeString(v.Clock), fmt.Sprint(v.Value))
		allRow = append(allRow, xRow)
	}
	writer.WriteAll(allRow)

	writer.Flush()
	log.Printf("write to %s success", csvfilename)
	return nil
}
