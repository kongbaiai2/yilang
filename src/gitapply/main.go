package main

import (
	"flag"
	_ "init_sub"
	"log"
	"os"
	"registration"
	"time"
)

var (
	buildGitRevison, buildTimestamp string
)

func initLog() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.SetOutput(os.Stdout)
}
func main() {
	pAppVersion := flag.Bool("version", false, "show version")
	if *pAppVersion {
		log.Fatalf("Git Revision:%s\nBuild Time: %s\n", buildGitRevison, buildTimestamp)
		os.Exit(0)
	}

	initLog()

	t := time.Now()
	defer func() { log.Println(time.Since(t)) }()

	registration.Run(os.Args)

}
