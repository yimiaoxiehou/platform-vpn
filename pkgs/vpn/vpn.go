package vpn

import (
	"fmt"
	"platform-vpn/pkgs/k3s"
	"platform-vpn/pkgs/log"
	"platform-vpn/pkgs/tun"
	"platform-vpn/pkgs/utils"
	"time"

	"github.com/showa-93/go-mask"
)

var k3sClinet *k3s.Client

func StopVPN() error {
	if updateHostsTicker != nil {
		updateHostsTicker.Stop()
	}
	return tun.StopTun()
}

var updateHostsTicker *time.Ticker

func StartVPN(user string, password string, host string, port int, RefreshInterval time.Duration) error {
	maskPassword, _ := mask.String(mask.MaskTypeFilled, password)
	log.Info(fmt.Sprintf("连接 VPN 服务器[%s], 配置：%s, %d, %s", host, user, port, maskPassword))
	var err error
	k3sClinet, err = k3s.NewClient(host, port, user, password)
	if err != nil {
		return err
	}

	config, err := k3sClinet.GetK3sConfig()
	if err != nil {
		return err
	}

	err = tun.StartTun(tun.TunConfig{
		Device:      "demo",
		Inet4Addr:   "10.10.0.1",
		RouteAddrs:  []string{config.ClusterCIDR, config.ServiceCIDR},
		SSHServer:   host,
		SSHPort:     port,
		SSHUser:     user,
		SSHPassword: password,
	})
	if err == nil {
		// 首次立即执行一次
		UpdateHosts()

		// 创建一个定时器，每5分钟触发一次
		updateHostsTicker = time.NewTicker(RefreshInterval)
		// 在后台循环执行
		go func() {
			for range updateHostsTicker.C {
				UpdateHosts()
			}
		}()
	}
	return err
}

func UpdateHosts() error {
	utils.CleanPlatformHosts()
	if hosts, err := k3sClinet.GetServiceHosts(); err != nil {
		log.Error(fmt.Sprintf("更新hosts失败: %v", err))
		return err
	} else {
		utils.UpdatePlatformHosts(hosts)
		log.Info("已更新hosts列表。")
		return err
	}
}
