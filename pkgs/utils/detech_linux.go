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
			innerErr = unix.BindToDevice(int(fd), iface.Name)
		})

		if innerErr != nil {
			err = innerErr
		}

		return
	}
}
