package utils

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
)

func UtilsTemp() {
	log.Println("utilstemp.go")
}

func Licence(licence string, key []byte) error {
	t := time.Now()
	str := "gophersoul"
	decrypt, err := DecryptDes(licence, key)
	if err != nil {
		return err
	}
	strSpl := strings.Split(decrypt, "-")
	if strSpl[0] != str {
		return fmt.Errorf("licence is error: %s", licence)
	}
	// licen_time, err := time.ParseInLocation("2006-01-02 15:04:05", time_str, time.Local)
	licen_int, err := strconv.ParseInt(strSpl[1], 10, 64)
	if err != nil {
		return fmt.Errorf("licence format error: %s", licence)
	}

	endLicen := licen_int + 60*60*24*300
	// 300 day
	if t.Unix() > endLicen || t.Unix() < licen_int {
		return fmt.Errorf("licence expiration: %v", time.Unix(endLicen, 0).Format(time.RFC3339))
	}
	return nil
}
