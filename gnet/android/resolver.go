package android

import (
	"context"
	"net"
	"net/http"
	"time"
)

var (
	Resolver *net.Resolver
)

// 解决 Android 上的 DNS 名称解析失败, https://github.com/golang/go/issues/8877
// 代码参考: https://czyt.tech/post/golang-http-use-custom-dns/
func init() {
	const (
		dnsResolverIP        = "8.8.8.8:53" // Google DNS resolver.
		dnsResolverProto     = "udp"        // Protocol to use for the DNS resolver
		dnsResolverTimeoutMs = 5000         // Timeout (ms) for the DNS resolver (optional)
	)

	Resolver = &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{
				Timeout: time.Duration(dnsResolverTimeoutMs) * time.Millisecond,
			}
			return d.DialContext(ctx, dnsResolverProto, dnsResolverIP)
		},
	}

	http.DefaultTransport.(*http.Transport).DialContext =
		func(ctx context.Context, network, addr string) (net.Conn, error) {
			d := &net.Dialer{Resolver: Resolver}
			return d.DialContext(ctx, network, addr)
		}
}
