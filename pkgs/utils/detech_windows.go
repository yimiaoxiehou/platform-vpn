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

// Windows 系统网络接口相关常量
const (
	// IP_UNICAST_IF 用于设置 IPv4 单播接口
	IP_UNICAST_IF = 31

	// IPV6_UNICAST_IF 用于设置 IPv6 单播接口
	IPV6_UNICAST_IF = 31
)

// bindIfaceToDialer 将网络接口绑定到指定的 Dialer
//
// 参数:
//   - iface: 要绑定的网络接口
//   - dialer: 目标 Dialer
//   - dest: 目标 IP 地址
//
// 返回:
//   - error: 绑定过程中发生的错误，如果成功则返回 nil
func bindIfaceToDialer(iface net.Interface, dialer *net.Dialer, dest net.IP) error {
	addControlToDialer(dialer, bindControl(iface.Index, dest))
	return nil
}

// bind4 绑定 IPv4 接口到指定的系统套接字
//
// 参数:
//   - handle: 系统套接字句柄
//   - ifaceIdx: 接口索引
//
// 返回:
//   - error: 绑定过程中发生的错误，如果成功则返回 nil
func bind4(handle syscall.Handle, ifaceIdx int) error {
	var bytes [4]byte
	binary.BigEndian.PutUint32(bytes[:], uint32(ifaceIdx))
	idx := *(*uint32)(unsafe.Pointer(&bytes[0]))
	
	if err := syscall.SetsockoptInt(handle, syscall.IPPROTO_IP, IP_UNICAST_IF, int(idx)); err != nil {
		return fmt.Errorf("设置 IPv4 单播接口失败 (接口索引: %d): %w", ifaceIdx, err)
	}
	return nil
}

// bind6 绑定 IPv6 接口到指定的系统套接字
//
// 参数:
//   - handle: 系统套接字句柄
//   - ifaceIdx: 接口索引
//
// 返回:
//   - error: 绑定过程中发生的错误，如果成功则返回 nil
func bind6(handle syscall.Handle, ifaceIdx int) error {
	if err := syscall.SetsockoptInt(handle, syscall.IPPROTO_IPV6, IPV6_UNICAST_IF, ifaceIdx); err != nil {
		return fmt.Errorf("设置 IPv6 单播接口失败 (接口索引: %d): %w", ifaceIdx, err)
	}
	return nil
}

// bindControl 返回一个用于控制网络连接的函数
//
// 参数:
//   - ifaceIdx: 接口索引
//   - dest: 目标 IP 地址
//
// 返回:
//   - controlFn: 用于控制网络连接的函数
func bindControl(ifaceIdx int, dest net.IP) controlFn {
	return func(ctx context.Context, network, address string, c syscall.RawConn) (err error) {
		addrPort, err := netip.ParseAddrPort(address)
		if err == nil && !addrPort.Addr().IsGlobalUnicast() {
			return
		}

		var innerErr error
		err = c.Control(func(fd uintptr) {
			handle := syscall.Handle(fd)
			bind6err := bind6(handle, ifaceIdx)
			bind4err := bind4(handle, ifaceIdx)

			switch network {
			case "ip6", "tcp6":
				innerErr = bind6err
			case "ip4", "tcp4", "udp4":
				innerErr = bind4err
			case "udp6":
				if (!addrPort.Addr().IsValid() || addrPort.Addr().IsUnspecified()) && bind6err != nil && dest.To4() != nil {
					if bind4err != nil {
						innerErr = fmt.Errorf("IPv6 和 IPv4 绑定都失败: IPv6 错误: %w, IPv4 错误: %s", bind6err, bind4err)
					}
				} else {
					innerErr = bind6err
				}
			}
		})

		if innerErr != nil {
			err = fmt.Errorf("网络接口绑定失败 (网络类型: %s): %w", network, innerErr)
		}

		return
	}
}
