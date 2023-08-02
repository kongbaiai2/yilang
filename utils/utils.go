package utils

import (
	"errors"
	"log"
	"strconv"
	"time"
)

func UtilsTemp() {
	log.Println("utilstemp.go")
}

func Licence(licence string, key []byte) error {
	t := time.Now()
	time_str, err := DecryptDes(licence, key)
	if err != nil {
		return err
	}
	// licen_time, err := time.ParseInLocation("2006-01-02 15:04:05", time_str, time.Local)
	licen_int, err := strconv.ParseInt(time_str, 10, 64)
	if err != nil {
		return err
	}
	// 300 day
	if t.Unix()-licen_int > 60*60*24*300 {
		err_new := "licence errors: " + licence
		return errors.New(err_new)
	}
	return nil
}
