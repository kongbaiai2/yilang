package utils

import (
	"context"
	"net"
	"time"
)

// dnsServer = "114.114.114.114:53"
// func CustomLookupHost(dns_server, domain string) ([]string, error) {
// 	r := &net.Resolver{
// 		PreferGo: true,
// 		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
// 			d := net.Dialer{
// 				Timeout: time.Millisecond * time.Duration(10000),
// 			}
// 			return d.DialContext(ctx, network, dns_server)
// 		},
// 	}
// 	// ipadd, err := r.LookupIPAddr(context.Background(), domain)
// 	ipStrArr, err := r.LookupHost(context.Background(), domain)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return ipStrArr, nil
// }

// dnsServer = "114.114.114.114:53"
func NewResolverClient(dns_server string) *net.Resolver {
	return &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{
				Timeout: time.Millisecond * time.Duration(10000),
			}
			return d.DialContext(ctx, network, dns_server)
		},
	}
}
