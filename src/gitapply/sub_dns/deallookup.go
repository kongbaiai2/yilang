package sub_dns

import (
	"context"
	"log"
	. "utils"

	"github.com/urfave/cli/v2"
)

func deal_lookup(c *cli.Context) error {
	domain := "mn-cdn.shzcdata.com.c8c10c27.d.cdn.10086.cn"

	r := NewResolverClient("112.4.0.55:53")
	ipadd, err := r.LookupIP(context.Background(), "ip", domain)
	if err != nil {
		log.Panic(err)
		return err
	}
	for _, v := range ipadd {
		log.Printf("%#v", v.To4().String())
	}

	// addr := "120.240.82.255"
	// addr := "39.136.117.197"
	// domains, err := r.LookupCNAME(context.Background(), domain)
	// if err != nil {
	// 	return err
	// }
	// for _, v := range domains {
	// log.Printf("%#v", ipadd)
	// }
	return nil
}
