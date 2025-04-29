package utils

import (
	"context"
	"fmt"
	"net"
	"net/netip"
	"syscall"

	"golang.org/x/sys/unix"
)

// bindIfaceToDialer 将网络接口绑定到指定的 Dialer
//
// 参数:
//   - iface: 要绑定的网络接口
//   - dialer: 目标 Dialer
//   - _: 目标 IP 地址（在 Darwin 系统下未使用）
//
// 返回:
//   - error: 绑定过程中发生的错误，如果成功则返回 nil
func bindIfaceToDialer(iface net.Interface, dialer *net.Dialer, _ net.IP) error {
	addControlToDialer(dialer, bindControl(iface))
	return nil
}

// bindControl 返回一个用于控制网络连接的函数
//
// 参数:
//   - iface: 要绑定的网络接口
//
// 返回:
//   - controlFn: 用于控制网络连接的函数
func bindControl(iface net.Interface) controlFn {
	return func(ctx context.Context, network, address string, c syscall.RawConn) (err error) {
		// 解析地址和端口
		addrPort, err := netip.ParseAddrPort(address)
		// 如果地址不是全局单播地址，则直接返回
		if err == nil && !addrPort.Addr().IsGlobalUnicast() {
			return
		}

		var innerErr error
		err = c.Control(func(fd uintptr) {
			switch network {
			case "tcp4", "udp4", "ip4":
				innerErr = unix.SetsockoptInt(int(fd), unix.IPPROTO_IP, unix.IP_BOUND_IF, iface.Index)
				if innerErr != nil {
					innerErr = fmt.Errorf("绑定 IPv4 接口失败 (接口: %s, 索引: %d): %w", 
						iface.Name, iface.Index, innerErr)
				}
			case "tcp6", "udp6", "ip6":
				innerErr = unix.SetsockoptInt(int(fd), unix.IPPROTO_IPV6, unix.IPV6_BOUND_IF, iface.Index)
				if innerErr != nil {
					innerErr = fmt.Errorf("绑定 IPv6 接口失败 (接口: %s, 索引: %d): %w", 
						iface.Name, iface.Index, innerErr)
				}
			default:
				innerErr = fmt.Errorf("不支持的网络类型: %s", network)
			}
		})

		if innerErr != nil {
			err = innerErr
		}

		return
	}
}
