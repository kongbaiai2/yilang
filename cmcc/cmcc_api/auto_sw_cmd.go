package cmcc_api

import (
	"net/http"
	"strings"

	"utils"

	"github.com/gin-gonic/gin"
	// logs "github.com/sirupsen/logrus"
	"github.com/satori/uuid"
)

func Auto_sw_cmds(c *gin.Context) {
	undo := c.DefaultQuery("undo", "false")
	uuidstr := c.DefaultQuery("uuid", "glb_uuid")
	if uuidstr != glb_uuid {
		c.String(http.StatusUnauthorized, "Status Unauthorized")
		return
	}
	if undo == "true" {
		if err := Auto_cfg(true); err != nil {
			c.String(http.StatusNotModified, "failed: %s", err)
			return
		}
	} else if undo == "false" {
		if err := Auto_cfg(false); err != nil {
			c.String(http.StatusNotModified, "failed: %s", err)
			return
		}
	}
	GenerateUuid()
	c.String(200, "success %s", "auto limit")
}
func Auto_cfg(isUndoCmd bool) error {
	// log.SetLevel(logs.WarnLevel)

	// 获取配置命令，执行的条件，description = auto*,交换机ip。
	for _, swconf := range cfg.AutoLimits {
		// 连接 ssh执行命令
		if swconf.IsDisable {
			log.Info("skip: ", swconf.SwIpList)
			continue
		}

		ssh_ctl := utils.NewSshCfg()
		ssh_ctl.User = swconf.User
		pwd, err := utils.DecryptDes(swconf.Encrypt, key)
		if err != nil {
			return err
		}

		ssh_ctl.Pwd = pwd
		for _, ip := range swconf.SwIpList {
			var inChannel chan string = make(chan string, 200)
			// 通过description取接口名
			ssh_ctl.Addr = ip
			err := ssh_ctl.RunShell([]string{swconf.Description}, inChannel)
			if err != nil && !strings.Contains(err.Error(), "remote command exited without exit status or exit signal") {
				log.Errorf("ssh: %s, %v", ip, err)
				return err
			}

			interArr := []string{}
			for results := range inChannel {
				if !strings.Contains(results, "XGE") {
					continue
				}

				// resultSpt := strings.Split(results, "\n")
				resultArr := strings.Fields(results)

				interArr = append(interArr, "interface XGigabitEthernet"+strings.Split(resultArr[0], "XGE")[1])
				if isUndoCmd {
					interArr = append(interArr, swconf.UndoCmd)
				} else {
					interArr = append(interArr, swconf.Cmd)
				}

			}

			log.Infof("cmds: [%s]", interArr)
			inChannel = make(chan string, 200)
			err = ssh_ctl.RunShell(interArr, inChannel)
			if err != nil && !strings.Contains(err.Error(), "remote command exited without exit status or exit signal") {
				log.Errorf("ssh: %s, %v", ip, err)
				return err
			}
			var b strings.Builder
			for results := range inChannel {
				b.WriteString(results)
				b.WriteString("\n")
			}
			log.Warn(b.String())
		}
	}
	return nil
}

func GenerateUuid() {
	glb_uuid = uuid.NewV4().String()

	url := "http://zabbix.yipeng888.com:7001/auto?uuid=" + glb_uuid + "&undo=false"
	url += "\n\nhttp://zabbix.yipeng888.com:7001/auto?uuid=" + glb_uuid + "&undo=true"
	log.Warn(url)
	warnerChan <- NewAlarmDb("uuid", "uuid", "yunchen.guangdongyunchen.com", url, "mail", 1)
}

// var log *logs.Logger

// func main() {
// 	log = logs.New()
// 	t := time.Now()
// 	defer func() {
// 		log.Warn(time.Since(t))
// 	}()
// 	cfg := cmcc_api.GetCfg()
// 	// read config
// 	cmcc_api.GetConfig(cfg, "limit_sw_intface", "json", ".", "./config", "../config")

// 	// init log
// 	cmcc_api.InitLog(log, "limit.log")

// 	Auto_cfg(false)

// 	// runGin()
// }
