package main

import (
	"cmcc_api"
	"flag"
	"log"
	"utils"
)

var (
	buildGitRevision string
	buildTimestamp   string
	key              []byte = []byte("t7h3he2d")
)

func main() {
	pAppVersion := flag.Bool("version", false, "show version")
	pEncrypt := flag.String("encrypt", "", "use encrypt")
	flag.Parse()
	if *pAppVersion {
		log.Fatalf("Git Revision: %s\ntime: %s", buildGitRevision, buildTimestamp)
	}
	if *pEncrypt != "" {
		log.Fatal(utils.EncryptDes(*pEncrypt, key))
	}
	cmcc_api.NewRun(key)
}
