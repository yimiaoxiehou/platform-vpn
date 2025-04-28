package utils

import (
	"context"
	"net"
	"net/netip"
	"syscall"

	"golang.org/x/sys/unix"
)

func bindIfaceToDialer(iface net.Interface, dialer *net.Dialer, _ net.IP) error {
	addControlToDialer(dialer, bindControl(iface))
	return nil
}

func bindControl(iface net.Interface) controlFn {
	return func(ctx context.Context, network, address string, c syscall.RawConn) (err error) {
		addrPort, err := netip.ParseAddrPort(address)
		if err == nil && !addrPort.Addr().IsGlobalUnicast() {
			return
		}

		var innerErr error
		err = c.Control(func(fd uintptr) {
			switch network {
			case "tcp4", "udp4":
				innerErr = unix.SetsockoptInt(int(fd), unix.IPPROTO_IP, unix.IP_BOUND_IF, iface.Index)
			case "tcp6", "udp6":
				innerErr = unix.SetsockoptInt(int(fd), unix.IPPROTO_IPV6, unix.IPV6_BOUND_IF, iface.Index)
			}
		})

		if innerErr != nil {
			err = innerErr
		}

		return
	}
}
