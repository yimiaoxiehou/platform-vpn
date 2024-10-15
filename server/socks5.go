package main

import (
	"context"
	"log"
	"net"

	"github.com/armon/go-socks5"
)

// CustomResolver 实现 socks5.NameResolver 接口
type CustomResolver struct {
	// 你可以在这里添加自定义的 DNS 服务器地址
	DNSServer string
}

// Resolve 实现自定义的域名解析逻辑
func (r *CustomResolver) Resolve(ctx context.Context, name string) (context.Context, net.IP, error) {
	log.Printf("Resolving domain: %s", name)

	// 使用自定义 DNS 服务器解析域名
	resolver := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{}
			return d.DialContext(ctx, "udp", r.DNSServer)
		},
	}

	ips, err := resolver.LookupIP(ctx, "ip4", name)
	if err != nil {
		return ctx, nil, err
	}

	if len(ips) == 0 {
		return ctx, nil, net.ErrClosed
	}

	log.Printf("Resolved %s to %s", name, ips[0].String())
	return ctx, ips[0], nil
}

func handleSocks5(conn net.Conn) {
	// 创建一个自定义解析器
	resolver := &CustomResolver{
		DNSServer: "10.96.0.10:53", // 使用 Google 的公共 DNS 服务器，你可以替换成你想要的 DNS 服务器
	}

	// 创建一个 SOCKS5 配置
	conf := &socks5.Config{
		Resolver: resolver,
	}

	// 创建一个 SOCKS5 服务器
	server, err := socks5.New(conf)
	if err != nil {
		log.Printf("Failed to create SOCKS5 server: %v", err)
		return
	}

	// 开始服务
	if err := server.ServeConn(conn); err != nil {
		log.Printf("SOCKS5 server failed to serve connection: %v", err)
	}
}
