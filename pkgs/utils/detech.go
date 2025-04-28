package utils

import (
	"context"
	"net"
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

func DetechRouteInterface(ip string, port string) (net.Interface, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return net.Interface{}, err
	}
	for _, iface := range interfaces {
		dialer := &net.Dialer{}

		dest := net.ParseIP(ip)
		err = bindIfaceToDialer(iface, dialer, dest)
		if err != nil {
			return net.Interface{}, err
		}
		address := net.JoinHostPort(ip, port)
		conn, err := dialer.Dial("tcp", address)
		if err == nil {
			defer conn.Close()
			return iface, nil
		}
	}
	return net.Interface{}, nil

}
