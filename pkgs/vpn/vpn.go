package vpn

import (
	"fmt"
	"net/netip"
	"platform-vpn/pkgs/k3s"
	"platform-vpn/pkgs/log"
	"platform-vpn/pkgs/utils"
	"strconv"
	"strings"
	"time"

	"github.com/metacubex/mihomo/adapter"
	"github.com/metacubex/mihomo/adapter/outbound"
	c "github.com/metacubex/mihomo/config"
	C "github.com/metacubex/mihomo/constant"
	"github.com/metacubex/mihomo/hub/executor"
	mlog "github.com/metacubex/mihomo/log"
	"github.com/metacubex/mihomo/rules"
	"github.com/showa-93/go-mask"
	logrus "github.com/sirupsen/logrus"
)

var k3sClient *k3s.Client
var updateHostsTicker *time.Ticker

func StopVPN() error {
	if updateHostsTicker != nil {
		updateHostsTicker.Stop()
	}
	executor.Shutdown()
	err := utils.CleanPlatformHosts()
	times := 3
	for err != nil && times > 0 {
		times--
		time.Sleep(time.Second)
		err = utils.CleanPlatformHosts()
	}
	return err
}

func StartVPN(user string, password string, host string, port int, refreshInterval time.Duration, stackMode int) error {
	logrus.SetOutput(log.NewLogger())
	maskPassword, err := mask.String(mask.MaskTypeFilled, password)
	if err != nil {
		log.Error(fmt.Sprintf("密码掩码处理失败: %v", err))
		maskPassword = "********"
	}
	log.Info(fmt.Sprintf("连接 VPN 服务器[%s], 配置：%s, %d, %s", host, user, port, maskPassword))

	k3sClient, err = k3s.NewClient(host, port, user, password)
	if err != nil {
		return fmt.Errorf("创建 k3s 客户端失败: %v", err)
	}

	routeCidrs, err := k3sClient.GetCidrs()
	if err != nil {
		return fmt.Errorf("获取路由 CIDR 失败: %v", err)
	}
	for _, addr := range routeCidrs {
		log.Info(fmt.Sprintf("路由地址: %s", addr))
	}

	proxies := make(map[string]C.Proxy)

	iface, err := utils.DetectRouteInterface(host, strconv.Itoa(port))
	if err != nil {
		log.Error(fmt.Sprintf("检测路由接口失败: %v", err))
	} else {
		log.Info(fmt.Sprintf("通过接口 %s 连接到 %s:%d", iface, host, port))
	}

	sshProxy, err := outbound.NewSsh(outbound.SshOption{
		Server:   host,
		Port:     port,
		UserName: user,
		Password: password,
		BasicOption: outbound.BasicOption{
			Interface: iface,
		},
	})
	if err != nil {
		return fmt.Errorf("创建 SSH 代理失败: %v", err)
	}

	proxies["DIRECT"] = adapter.NewProxy(outbound.NewDirect())
	proxies["PROXY"] = adapter.NewProxy(sshProxy)

	clashRawConfig := c.DefaultRawConfig()
	clashConfig, err := c.ParseRawConfig(clashRawConfig)
	if err != nil {
		return fmt.Errorf("解析 Clash 配置失败: %v", err)
	}

	clashConfig.General.LogLevel = mlog.DEBUG
	clashConfig.General.SocksPort = 37890
	clashConfig.General.IPv6 = false
	clashConfig.General.Tun.Enable = true
	clashConfig.General.Tun.DNSHijack = []string{}
	clashConfig.General.Tun.Device = "utun8"
	clashConfig.General.Tun.StrictRoute = true
	clashConfig.General.Tun.MTU = 1400
	clashConfig.General.Tun.Inet4Address = []netip.Prefix{
		netip.PrefixFrom(netip.MustParseAddr("10.10.0.1"), 30),
	}
	clashConfig.General.Tun.Inet6Address = []netip.Prefix{}
	// 转换路由地址
	for _, addr := range routeCidrs {
		prefix := strings.Split(addr, "/")
		mask, _ := strconv.Atoi(prefix[1])
		clashConfig.General.Tun.RouteAddress = append(clashConfig.General.Tun.RouteAddress,
			netip.PrefixFrom(netip.MustParseAddr(prefix[0]), int(mask)))
	}
	clashConfig.DNS.Enable = false
	clashConfig.Rules = AddIPCIDRRule(routeCidrs)
	clashConfig.Proxies = proxies

	executor.ApplyConfig(clashConfig, true)

	if err := UpdateHosts(); err != nil {
		log.Error(fmt.Sprintf("首次更新 hosts 失败: %v", err))
	}

	updateHostsTicker = time.NewTicker(refreshInterval)
	go func() {
		for range updateHostsTicker.C {
			if err := UpdateHosts(); err != nil {
				log.Error(fmt.Sprintf("定时更新 hosts 失败: %v", err))
			}
		}
	}()
	return nil
}

func UpdateHosts() error {
	if err := utils.CleanPlatformHosts(); err != nil {
		return fmt.Errorf("清理 hosts 失败: %v", err)
	}

	hosts, err := k3sClient.GetServiceHosts()
	if err != nil {
		log.Error(fmt.Sprintf("更新 hosts 失败: %v", err))
		return err
	}

	if err := utils.UpdatePlatformHosts(hosts); err != nil {
		return fmt.Errorf("更新平台 hosts 失败: %v", err)
	}

	log.Info("已更新 hosts 列表")
	return nil
}

func AddIPCIDRRule(ipCidrs []string) []C.Rule {
	proxyRules := make([]C.Rule, 0, len(ipCidrs))
	for _, ipCidr := range ipCidrs {
		log.Info(fmt.Sprintf("添加 IPCIDR 规则: %s", ipCidr))
		proxyRule, parseErr := rules.ParseRule("IP-CIDR", ipCidr, "PROXY", []string{}, nil)
		if parseErr != nil {
			log.Error(fmt.Sprintf("添加 IPCIDR 规则失败: %s", parseErr.Error()))
			continue
		}
		proxyRules = append(proxyRules, proxyRule)
	}
	return proxyRules
}
