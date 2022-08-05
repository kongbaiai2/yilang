package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/robfig/cron"
	"github.com/tealeg/xlsx"
)

const (
	// reLinke = `[^\s]*.(?:com|cn|net|edu|gov|top|hk|org)$`
	// dnsServer = "114.114.114.114:53"

	retNu = 100
)

var (
	cfg      *AppCfg
	ipChan   chan map[string]string
	ipV6Chan chan map[string]string

	retMap map[string]map[string]string

	g_buildTimestamp   string
	g_buildGitRevision string
)

type AppCfg struct {
	RegExp         string   `json:"regexp"`
	CidrList       []string `json:"cidr_list"`
	Cidrv6List     []string `json:"cidrv6_list"`
	Dns_servers    []string `json:"dns_servers"`
	Domain_file    string   `json:"domain_file_excel"`
	TaskCronConfig string   `json:"task_cron_config"`
	Hostname       string   `json:"hostname"`
}

func AnalyzeExcel(path, regexpStr string) (map[string]string, error) {

	result := make(map[string]string)
	re := regexp.MustCompile(regexpStr)
	xlFile, err := xlsx.OpenFile(path) //打开文件
	if err != nil {
		return result, err
	}

	for _, sheet := range xlFile.Sheets { //遍历sheet层
		for rowIndex, row := range sheet.Rows { //遍历row层
			if rowIndex > 0 {
				if len(row.Cells) < 2 {
					break
				}
				for _, cell := range row.Cells { //遍历cell层
					text := cell.String() //把单元格的内容转成string
					if len(text) == 0 {
						break
					}

					if match := re.FindStringSubmatch(text); match != nil {
						// fmt.Println(match)
						for _, v := range match {
							result[v] = v
						}
					}
				}
			}
		}
	}

	return result, nil
}

// dnsServer = "114.114.114.114:53"
func CustomLookupHost(dns_server, host string) ([]string, error) {
	r := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{
				Timeout: time.Millisecond * time.Duration(10000),
			}
			return d.DialContext(ctx, network, dns_server)
		},
	}
	// ipadd, err := r.LookupIPAddr(context.Background(), host)
	ipStrArr, err := r.LookupHost(context.Background(), host)
	if err != nil {
		return nil, err
	}

	return ipStrArr, nil
}

func getIpArrayInNslookup(domains map[string]string, dns_server string) (ipv4Map, ipv6Map map[string]string) {
	number := len(domains)

	// ip去重
	ipv4Map = make(map[string]string, number)
	ipv6Map = make(map[string]string, number)
	i := 0

	for _, url := range domains {
		retIps, err := CustomLookupHost(dns_server, url)
		if err != nil {
			fmt.Printf("error: %v\n", err)
			retErr := make(map[string]string, 1)
			retErr[url] = fmt.Sprintf("error: %v\n", err)
			retMap[url] = retErr

			i++
			if i > retNu {
				return
			}
		}
		for _, ipStr := range retIps {
			ip := net.ParseIP(ipStr)
			if ip == nil {
				continue
			}
			if IsIPv4(ipStr) {
				if _, ok := ipv4Map[ipStr]; !ok {
					ipv4Map[ipStr] = url + " " + dns_server

					tmpMap := make(map[string]string)
					tmpMap[ipStr] = url + " " + dns_server
					ipChan <- tmpMap
				}
			} else if IsIPv6(ipStr) {
				if _, ok := ipv6Map[ipStr]; !ok {
					ipv6Map[ipStr] = url + " " + dns_server

					tmpMap := make(map[string]string)
					tmpMap[ipStr] = url + " " + dns_server
					ipV6Chan <- tmpMap
				}
			}
		}
	}
	defer fmt.Println("go end : ", time.Now())
	return
}
func IsIPv4(ipAddr string) bool {
	return strings.Contains(ipAddr, ".")
}
func IsIPv6(ipAddr string) bool {
	return strings.Contains(ipAddr, ":")
}

func cidrContainsIp(cidr, ip string) bool {
	_, ipnetCidr, _ := net.ParseCIDR(cidr)
	ipA := net.ParseIP(ip)

	return ipnetCidr.Contains(ipA)
}

// Check that the IP in the channel is included by the CIDR
func goCidrContainIp(cidrs, cidrV6s []string) {
	fmt.Println("goCidrContainIp")
	defer fmt.Println("goCidrContainIp go end : ", time.Now())
	for {
		select {
		case ip_map := <-ipChan:
			for ip, url := range ip_map {
				for _, cidr := range cidrs {

					errMap := make(map[string]string)
					if cidrContainsIp(cidr, ip) {
						fmt.Printf("\ncidr (%s) contain: ip (%s) for (%v)\n", cidr, ip, url)
						errMap[cidr] = url
						retMap[ip] = errMap
					}
				}
			}

		case ipV6_map := <-ipV6Chan:
			for ipV6, url := range ipV6_map {
				for _, cidrV6 := range cidrV6s {

					errMap := make(map[string]string)
					if cidrContainsIp(cidrV6, ipV6) {
						fmt.Printf("\ncidrv6 (%s) contain: ip (%s) for (%v)\n", cidrV6, ipV6, url)
						errMap[cidrV6] = url
						retMap[ipV6] = errMap
					}
				}
			}
		}
	}
}
func setCron() {

	go goCidrContainIp(cfg.CidrList, cfg.Cidrv6List)

	c := cron.New()
	c.AddFunc(cfg.TaskCronConfig, doNslookup)

	c.Start()

}

func doNslookup() {
	// now := time.Now()
	retMap = make(map[string]map[string]string)

	// get domain array for excel
	domains, _ := AnalyzeExcel(cfg.Domain_file, cfg.RegExp)

	// get ip list of dns server to goroutinue
	for _, dnsServer := range cfg.Dns_servers {
		go getIpArrayInNslookup(domains, dnsServer)
	}

}

func listenPort() {
	router := gin.Default()

	router.GET("/", func(c *gin.Context) {

		// c.String(http.StatusOK, fmt.Sprintf("%v", retMap))
		c.JSON(http.StatusOK, retMap)
	})

	router.Run(":3333")
}
func readConfig(configFile string) error {
	c := newConfig(configFile)
	if c == nil {
		return fmt.Errorf("load config file failed")
	}
	cfg = c.cache
	return nil
}

func initConfig() {
	pConfigFile := flag.String("f", "./config.json", "config file path")
	pAppVersion := flag.Bool("V", false, "show version")
	flag.Parse()

	if *pAppVersion {
		fmt.Printf("Git Revision: %s\nBuild Time: %s\n", g_buildGitRevision, g_buildTimestamp)
		os.Exit(0)
	}

	err := readConfig(*pConfigFile)
	if err != nil {
		log.Fatalf("read config file %s error: %s\n", *pConfigFile, err.Error())
	}
}
func main() {
	ipChan = make(chan map[string]string, 100)
	ipV6Chan = make(chan map[string]string, 100)
	defer close(ipChan)
	defer close(ipV6Chan)

	initConfig()

	setCron()

	listenPort()

}

/*

> www.baidu.com
Server:         114.114.114.114
Address:        114.114.114.114#53

Non-authoritative answer:
www.baidu.com   canonical name = www.a.shifen.com.
Name:   www.a.shifen.com
Address: 180.101.49.12
Name:   www.a.shifen.com
Address: 180.101.49.11


> www.baidu.com
Server:         30.30.30.30
Address:        30.30.30.30#53

Non-authoritative answer:
www.baidu.com   canonical name = www.a.shifen.com.
Name:   www.a.shifen.com
Address: 110.242.68.4
Name:   www.a.shifen.com
Address: 110.242.68.3

*/
