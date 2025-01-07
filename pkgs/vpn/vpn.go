package vpn

import (
	"context"
	"fmt"
	"platform-vpn/pkgs/k3s"
	"platform-vpn/pkgs/tun"
	"platform-vpn/pkgs/utils"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

var k3sClinet *k3s.Client

func StopVPN(ctx context.Context) error {
	if updateHostsTicker != nil {
		updateHostsTicker.Stop()
	}
	return tun.StopTun(ctx)
}

var updateHostsTicker *time.Ticker

func StartVPN(ctx context.Context, user string, password string, host string, port int, RefreshInterval time.Duration) error {
	var err error
	k3sClinet, err = k3s.NewClient(host, port, user, password)
	if err != nil {
		return err
	}

	config, err := k3sClinet.GetK3sConfig()
	if err != nil {
		return err
	}

	err = tun.StartTun(ctx, tun.TunConfig{
		Device:      "demo",
		Inet4Addr:   "10.10.0.1",
		RouteAddrs:  []string{config.ClusterCIDR, config.ServiceCIDR},
		SSHServer:   host,
		SSHPort:     port,
		SSHUser:     user,
		SSHPassword: password,
	})
	if err != nil {
		// 首次立即执行一次
		UpdateHosts(ctx)

		// 创建一个定时器，每5分钟触发一次
		updateHostsTicker = time.NewTicker(RefreshInterval)
		// 在后台循环执行
		go func() {
			for range updateHostsTicker.C {
				UpdateHosts(ctx)
			}
		}()
	}
	return err
}

func UpdateHosts(ctx context.Context) error {
	utils.CleanPlatformHosts(ctx)
	if hosts, err := k3sClinet.GetServiceHosts(ctx); err != nil {
		runtime.LogError(ctx, fmt.Sprintf("更新hosts失败: %v", err))
		return err
	} else {
		utils.UpdatePlatformHosts(hosts)
		runtime.LogPrint(ctx, "已更新hosts列表。")
		return err
	}
}
