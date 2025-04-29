package utils

import (
	"context"
	"fmt"
	"net"
	"platform-vpn/pkgs/log"
	"strings"
	"syscall"
)

type controlFn = func(ctx context.Context, network, address string, c syscall.RawConn) error

func addControlToDialer(d *net.Dialer, fn controlFn) {
	ld := *d
	d.ControlContext = func(ctx context.Context, network, address string, c syscall.RawConn) (err error) {
		switch {
		case ld.ControlContext != nil:
			if err = ld.ControlContext(ctx, network, address, c); err != nil {
				return
			}
		case ld.Control != nil:
			if err = ld.Control(network, address, c); err != nil {
				return
			}
		}
		return fn(ctx, network, address, c)
	}
}

func DetectRouteInterface(ip string, port string) (string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", fmt.Errorf("获取网络接口列表失败: %v", err)
	}
	for _, iface := range interfaces {
		log.Info(fmt.Sprintf("正在检查网络接口: %s", iface.Name))
		dialer := &net.Dialer{}

		dest := net.ParseIP(ip)
		err = bindIfaceToDialer(iface, dialer, dest)
		if err != nil {
			return "", fmt.Errorf("绑定网络接口失败: %v", err)
		}

		ifaceAddrs, err := iface.Addrs()
		if err != nil {
			log.Error(fmt.Sprintf("获取接口地址失败: %v", err))
			continue
		}

		found := false
		for _, ifaceAddr := range ifaceAddrs {
			addr := strings.Split(ifaceAddr.String(), "/")[0]
			addrIp := net.ParseIP(addr)
			if dest.To4() != nil && addrIp.To4() != nil {
				found = true
				break
			}
		}
		if !found {
			continue
		}

		address := net.JoinHostPort(ip, port)
		conn, err := dialer.Dial("tcp", address)
		if err == nil {
			defer conn.Close()
			log.Info(fmt.Sprintf("找到可用网络接口: %s", iface.Name))
			return iface.Name, nil
		} else {
			log.Error(fmt.Sprintf("连接失败: %v", err))
		}
	}
	return "", fmt.Errorf("未找到可连接到 %s:%s 的网络接口", ip, port)
}
