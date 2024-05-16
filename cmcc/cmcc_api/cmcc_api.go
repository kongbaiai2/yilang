package cmcc_api

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"net/http"
	"strings"
	"time"
	"unsafe"
	"utils"

	"github.com/sirupsen/logrus"
)

func (t *TenantItem) GetToken(cmcc *Cmcc, url string) error {
	cmcc.Domain = url
	cmcc.ApiToken = t.Tenant.TokenApi
	cmcc.ApiStatistic = t.Tenant.ApiStatistic
	cmcc.TenantDomain = t.Tenant.Domain
	cmcc.tenant_id = t.Tenant.TenantId
	cmcc.tenant_key = t.Tenant.TenantKey
	err := cmcc.PostToken()
	if err != nil {
		return err
	}
	return nil
}

// tenantId + datetime + tenantKey
func (c *Cmcc) PostToken() error {
	timeformat := time.Now().Format(time.RFC3339)
	sign := Sha256(c.tenant_id + timeformat + c.tenant_key)

	postDate := &AuthRequestPostDate{}
	postDate.DateTime = timeformat
	postDate.Authorization.Tenant_id = c.tenant_id
	postDate.Authorization.Sign = sign
	structBody, err := json.Marshal(postDate)
	if err != nil {
		log.Println(err)
		return err
	}

	httpReq, err := http.NewRequest("POST", c.Domain+c.ApiToken, bytes.NewReader(structBody))
	if err != nil {
		log.Printf("NewRequest fail, url: %s, reqBody: %s, err: %v", c.Domain+c.ApiToken, structBody, err)
		return err
	}
	httpReq.Header.Add("Content-Type", "application/json")
	httpReq.Header.Add("Accept", "application/vnd.cmcdn+json")
	// httpReq.Header.Add("CMCDN-Auth-Token", "")
	// httpReq.Header.Add("HTTP-X-CMCDN-Signature", signature)
	// httpReq.Header.Add("X-CMCDN-Media-Type", "application/vnd.cmcdn.version; format=json")
	// httpReq.Header.Add("Access-Control-Allow-Methods", "POST")

	// caCert, _ := ioutil.ReadFile("/etc/ssl/cert.pem")
	// caCertPool := x509.NewCertPool()
	// caCertPool.AppendCertsFromPEM(caCert)
	client := &http.Client{Transport: &http.Transport{
		TLSClientConfig: &tls.Config{
			// RootCAs: caCertPool,
			InsecureSkipVerify: true,
		},
	},
		Timeout: time.Duration(timeout),
	}

	// requestDump, err := httputil.DumpRequest(httpReq, true)
	httpRsp, err := client.Do(httpReq)
	// fmt.Println(string(httpRsp.Proto))
	if err != nil {
		log.Printf("do http fail, url: %s, reqBody: %s, err:%v", c.Domain+c.ApiToken, structBody, err)
		return err
	}
	defer httpRsp.Body.Close()

	buf := new(bytes.Buffer)
	buf.ReadFrom(httpRsp.Body)
	b := buf.Bytes()
	s := *(*string)(unsafe.Pointer(&b))

	sMap := JsonToMap(s)
	if token, ok := sMap["token"]; ok {
		if tokenStr, ok := token.(string); ok {
			c.Token = tokenStr
			return nil
		}
		return fmt.Errorf("sMap['token'] type not string")
	}

	// log.Println(sMap["token"])
	return fmt.Errorf("sMap['token'] not exist")
}
func (c *Cmcc) GetStatisticArgs(end, start time.Duration) {
	now := time.Now()
	endTime := now.Add(-end).Format(time.RFC3339)
	startTime := now.Add(-start).Format(time.RFC3339)
	c.StatisticArgs.Domain = c.TenantDomain
	c.StatisticArgs.Detail = 1
	c.StatisticArgs.Start = startTime
	c.StatisticArgs.End = endTime
	c.StatisticArgs.IpProtocol = "0"
}

// GET：http://xxx.com/api/statistic/bw?domain=abc.com&domain=11.com.cn&detail=1&start=2011-12-03T10:15:30%2B08:00&end=2011-12- 03T10:15:35%2B08:00&ipProtocol=all
// /api/statistic/bw?domain=${domain}&detail=${detail}&start=${start}&end=${end}&ipProtocol=0
func (c *Cmcc) GetStatisticCDN(detailed string, getArgs StatisticArgs, result *[]ResponseStatistic) error {
	domain := c.Domain + c.ApiStatistic + detailed
	args := fmt.Sprintf("?domain=%s&detail=%d&start=%s&end=%s&ipProtocol=%s",
		getArgs.Domain, getArgs.Detail, getArgs.Start, getArgs.End, getArgs.IpProtocol)
	url := domain + args
	url = strings.Replace(url, "+", "%2B", -1)
	httpReq, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Printf("NewRequest fail, url: %s, reqBody: %s, err: %v", url, "", err)
		return err
	}

	// httpReq.Header.Add("Content-Type", "application/json")
	httpReq.Header.Add("Accept", "application/vnd.cmcdn+json")
	httpReq.Header.Add("CMCDN-Auth-Token", c.Token)
	sign := Sha256(c.Token)
	httpReq.Header.Add("HTTP-X-CMCDN-Signature", sign)

	client := &http.Client{Transport: &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	},
		Timeout: time.Duration(timeout),
	}

	httpRsp, err := client.Do(httpReq)
	if err != nil {
		log.Printf("do http fail, url: %s, reqBody: %s, err:%v", url, "", err)
		return err
	}
	defer httpRsp.Body.Close()

	if httpRsp.StatusCode != 200 {
		log.Printf("do http fail, url: %s, sign: %s, ", url, sign)
	}
	body, err := ioutil.ReadAll(httpRsp.Body)
	if err != nil {
		return err
	}
	log.Debug(string(body))
	if err := json.Unmarshal(body, result); err != nil {
		log.Println(string(body))
		return err
	}

	log.Printf("do http success, url: %s, sign: %s", url, sign)
	return nil
}

func (t *TenantItem) cmcc(end, start time.Duration) error {
	log.SetLevel(logrus.Level(cfg.LogLevel))
	err := utils.Licence(cfg.Licence, key)
	if err != nil {
		log.Fatalf("Licence errors: %v", err)
	}

	cmcc := &Cmcc{}

	err = t.GetToken(cmcc, t.Tenant.Url)
	if err != nil {
		return err
	}
	// 通用配置,获取多长时间段的数据
	cmcc.GetStatisticArgs(end, start)

	// 取指定bw数据
	var resultStruct []ResponseStatistic
	err = cmcc.GetStatisticCDN("/bw", cmcc.StatisticArgs, &resultStruct)
	if err != nil {
		return err
	}

	return writeBwToDb(resultStruct, t.GetData.Alert.MailWarn.Host)
}
