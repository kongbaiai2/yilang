package zabbix_tools

import (
	"log"
	. "registration"

	"github.com/urfave/cli/v2"
)

type zabbixTools struct{}

func init() {
	zabbixRegister()
}

func zabbixRegister() {
	var zTools Commander
	zTools = &zabbixTools{}
	Register("zabbix_tools", zTools)
}

func (z *zabbixTools) Appends(gCommadns []*cli.Command) []*cli.Command {
	return append(gCommadns, subCmdZabbixTools())
}

func subCmdZabbixTools() *cli.Command {
	return &cli.Command{
		Name:    "zabbix",
		Aliases: []string{"z"},
		Usage:   "Specify to generate an average of all data from zabbix db within 5 minutes\ncase: zabbix --host='S6880-48S4Q' --intname='10GE1/0/8' --inout='ifHCInOctets' --start='2023/05/01 00:00:00' --end='2023/05/01 00:01:00'",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "host",
				Usage:    "Required specify host name",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "intname",
				Usage:    "Required specify interface name",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "inout",
				Usage:    "Required inbound or outbound: ifHCInOctets or ifHCOutOctets",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "start",
				Usage:    "Required start time: 2023/04/30 00:00:00",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "end",
				Usage:    "Required end time: 2023/05/01 00:00:00",
				Required: true,
			},
		},
		Action: func(c *cli.Context) error {
			err := TurnMinutes(c)
			if err != nil {
				log.Println(err)
			}
			return err
		},
	}
}
