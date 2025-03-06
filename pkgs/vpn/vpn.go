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
	"github.com/metacubex/mihomo/rules"
	"github.com/showa-93/go-mask"
)

var k3sClinet *k3s.Client
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
	maskPassword, _ := mask.String(mask.MaskTypeFilled, password)
	log.Info(fmt.Sprintf("连接 VPN 服务器[%s], 配置：%s, %d, %s", host, user, port, maskPassword))
	var err error
	k3sClinet, err = k3s.NewClient(host, port, user, password)
	if err != nil {
		return err
	}

	routeCidrs, err := k3sClinet.GetCidrs()
	if err != nil {
		return err
	}

	proxies := make(map[string]C.Proxy)
	sshProxy, _ := outbound.NewSsh(outbound.SshOption{
		Server:   host,
		Port:     port,
		UserName: user,
		Password: password,
	})
	proxies["DIRECT"] = adapter.NewProxy(outbound.NewDirect())
	proxies["PROXY"] = adapter.NewProxy(sshProxy)

	clashRawConfig := c.DefaultRawConfig()
	clashConfig, _ := c.ParseRawConfig(clashRawConfig)
	clashConfig.General.SocksPort = 37890
	clashConfig.General.IPv6 = false
	clashConfig.General.Tun.Enable = true
	clashConfig.General.Tun.DNSHijack = []string{}
	clashConfig.General.Tun.Device = "platform-vpn"
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
	// 首次立即执行一次
	UpdateHosts()
	// 创建一个定时器，每5分钟触发一次
	updateHostsTicker = time.NewTicker(refreshInterval)
	// 在后台循环执行
	go func() {
		for range updateHostsTicker.C {
			UpdateHosts()
		}
	}()
	return nil
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

func AddIPCIDRRule(ipCidrs []string) []C.Rule {
	proxyRules := make([]C.Rule, 0, len(ipCidrs))
	for _, ipCidr := range ipCidrs {
		log.Info(fmt.Sprintf("Add IPCIDR rule: %s", ipCidr))
		proxyRule, parseErr := rules.ParseRule("IP-CIDR", ipCidr, "PROXY", []string{}, nil)
		if parseErr != nil {
			log.Error(fmt.Sprintf("Add IPCIDR rule error: %s", parseErr.Error()))
		}
		proxyRules = append(proxyRules, proxyRule)
	}
	return proxyRules
}
