// Package utils 提供了网络接口绑定相关的实用工具函数
package utils

import (
	"context"
	"encoding/binary"
	"fmt"
	"net"
	"net/netip"
	"syscall"
	"unsafe"
)

// Windows 系统下用于设置单播接口的常量
const (
	IP_UNICAST_IF   = 31 // IPv4 单播接口选项
	IPV6_UNICAST_IF = 31 // IPv6 单播接口选项
)

// bindIfaceToDialer 将网络接口绑定到指定的 Dialer
// iface: 要绑定的网络接口
// dialer: 目标 Dialer
// destAddress: 目标地址
func bindIfaceToDialer(iface net.Interface, dialer *net.Dialer, dest net.IP) error {
	addControlToDialer(dialer, bindControl(iface.Index, dest))
	return nil
}

// bind4 绑定 IPv4 接口
// handle: 系统套接字句柄
// ifaceIdx: 接口索引
func bind4(handle syscall.Handle, ifaceIdx int) error {
	var bytes [4]byte
	// 将接口索引转换为大端字节序
	binary.BigEndian.PutUint32(bytes[:], uint32(ifaceIdx))
	idx := *(*uint32)(unsafe.Pointer(&bytes[0]))
	// 设置 IPv4 单播接口选项
	err := syscall.SetsockoptInt(handle, syscall.IPPROTO_IP, IP_UNICAST_IF, int(idx))
	if err != nil {
		err = fmt.Errorf("bind4: %w", err)
	}
	return err
}

// bind6 绑定 IPv6 接口
// handle: 系统套接字句柄
// ifaceIdx: 接口索引
func bind6(handle syscall.Handle, ifaceIdx int) error {
	// 设置 IPv6 单播接口选项
	err := syscall.SetsockoptInt(handle, syscall.IPPROTO_IPV6, IPV6_UNICAST_IF, ifaceIdx)
	if err != nil {
		err = fmt.Errorf("bind6: %w", err)
	}
	return err
}

// bindControl 返回一个用于控制网络连接的函数
// ifaceIdx: 接口索引
// dest: 目标地址
func bindControl(ifaceIdx int, dest net.IP) controlFn {
	return func(ctx context.Context, network, address string, c syscall.RawConn) (err error) {
		// 解析地址和端口
		addrPort, err := netip.ParseAddrPort(address)
		// 如果地址不是全局单播地址，则直接返回
		if err == nil && !addrPort.Addr().IsGlobalUnicast() {
			return
		}

		var innerErr error
		err = c.Control(func(fd uintptr) {
			handle := syscall.Handle(fd)
			bind6err := bind6(handle, ifaceIdx)
			bind4err := bind4(handle, ifaceIdx)
			// 根据网络类型选择不同的绑定方式
			switch network {
			case "ip6", "tcp6":
				innerErr = bind6err
			case "ip4", "tcp4", "udp4":
				innerErr = bind4err
			case "udp6":
				// 当在通配符 IP 上监听 UDP 时，Go 会将网络设置为 udp6
				if (!addrPort.Addr().IsValid() || addrPort.Addr().IsUnspecified()) && bind6err != nil && dest.To4() != nil {
					// 尝试绑定 IPv6，如果失败则忽略。这是 Windows 禁用接口 IPv6 的解决方案
					if bind4err != nil {
						innerErr = fmt.Errorf("%w (%s)", bind6err, bind4err)
					} else {
						innerErr = nil
					}
				} else {
					innerErr = bind6err
				}
			}
		})

		if innerErr != nil {
			err = innerErr
		}

		return
	}
}
