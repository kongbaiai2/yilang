package sub_dns

import (
	"log"
	. "registration"

	"github.com/urfave/cli/v2"
)

type subDns struct{}

func init() {
	dnsRegister()
}

func dnsRegister() {
	var zTools Commander
	zTools = &subDns{}
	Register("sub_dns", zTools)
}

func (z *subDns) Appends(gCommadns []*cli.Command) []*cli.Command {
	return append(gCommadns, subCmdsubDns())
}

func subCmdsubDns() *cli.Command {
	return &cli.Command{
		Name:    "lookup",
		Aliases: []string{"l"},
		Usage:   "Specify to generate lookup dns host and ip",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "host",
				Usage:    "specify host name",
				Required: false,
			},
			&cli.StringFlag{
				Name:     "ip",
				Usage:    "specify ip name",
				Required: false,
			},
		},
		Action: func(c *cli.Context) error {
			err := deal_lookup(c)
			if err != nil {
				log.Println(err)
			}
			return err
		},
	}
}
