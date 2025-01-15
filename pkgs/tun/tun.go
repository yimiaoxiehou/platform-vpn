package tun

import (
	"fmt"
	"net/netip"
	"platform-vpn/pkgs/log"
	"platform-vpn/pkgs/utils"
	"strconv"
	"strings"
	"sync"

	"github.com/metacubex/mihomo/adapter"
	"github.com/metacubex/mihomo/adapter/outbound"
	"github.com/metacubex/mihomo/common/observable"
	C "github.com/metacubex/mihomo/constant"
	LC "github.com/metacubex/mihomo/listener/config"
	"github.com/metacubex/mihomo/listener/sing_tun"
	mlog "github.com/metacubex/mihomo/log"
	"github.com/metacubex/mihomo/tunnel"
)

var (
	tunLister *sing_tun.Listener

	// lock for recreate function
	tunMux sync.Mutex

	LastTunConf  LC.Tun
	LastTuicConf LC.TuicServer
)

// 添加配置结构体
type TunConfig struct {
	Device      string
	Inet4Addr   string
	RouteAddrs  []string
	SSHServer   string
	SSHPort     int
	SSHUser     string
	SSHPassword string
	// Fallback_DNS string
}

var mlogSub observable.Subscription[mlog.Event]

func StopTun() error {
	tunnel.OnSuspend()
	mlog.UnSubscribe(mlogSub)
	closeTunListener()
	return utils.CleanPlatformHosts()
}

func StartTun(config TunConfig) error {
	mlogSub = mlog.Subscribe()
	go func() {
		for logM := range mlogSub {
			switch logM.LogLevel {
			case mlog.DEBUG:
				log.Debug(logM.Payload)
			case mlog.INFO:
				log.Info(logM.Payload)
			case mlog.WARNING:
				log.Warning(logM.Payload)
			case mlog.ERROR:
				log.Error(logM.Payload)
			default:
			}
		}
	}()

	tunConf := LC.Tun{
		Enable:      true,
		StrictRoute: true,
		Device:      config.Device,
		Inet4Address: []netip.Prefix{
			netip.PrefixFrom(netip.MustParseAddr(config.Inet4Addr), 16),
		},
		RouteAddress: make([]netip.Prefix, 0, len(config.RouteAddrs)),
		AutoRedirect: true,
		AutoRoute:    true,
		Stack:        C.TunMixed,
	}

	// 转换路由地址
	for _, addr := range config.RouteAddrs {
		prefix := strings.Split(addr, "/")
		mask, _ := strconv.Atoi(prefix[1])
		tunConf.RouteAddress = append(tunConf.RouteAddress,
			netip.PrefixFrom(netip.MustParseAddr(prefix[0]), int(mask)))
	}

	var Tunnel = tunnel.Tunnel

	proxies := make(map[string]C.Proxy)
	sshProxy, _ := outbound.NewSsh(outbound.SshOption{
		Server:   config.SSHServer,
		Port:     config.SSHPort,
		UserName: config.SSHUser,
		Password: config.SSHPassword,
	})
	proxies["DIRECT"] = adapter.NewProxy(sshProxy)
	tunnel.UpdateProxies(proxies, nil)
	err := ReCreateTun(tunConf, Tunnel)
	if err != nil {
		return err
	}
	tunnel.OnRunning()
	return nil
}

func ReCreateTun(tunConf LC.Tun, tunnel C.Tunnel) error {
	// Sort the TUN configuration.
	tunConf.Sort()

	// Lock the TUN mutex to prevent concurrent access.
	tunMux.Lock()
	defer func() {
		LastTunConf = tunConf
		tunMux.Unlock()
	}()

	var err error
	defer func() {
		if err != nil {
			log.Error(fmt.Sprintf("Start TUN listening error: %s", err.Error()))
		}
	}()

	// Close the current TUN listener.
	closeTunListener()
	UpdateDNS()
	// Create a new TUN listener with the provided configuration and tunnel.
	lister, err := sing_tun.New(tunConf, tunnel)
	if err != nil {
		return err
	}
	tunLister = lister

	// Log the address where the TUN adapter is listening.
	log.Info(fmt.Sprintf("[TUN] Tun adapter listening at: %s", tunLister.Address()))
	return nil
}

func closeTunListener() {
	if tunLister != nil {
		tunLister.Close()
		tunLister = nil
	}
}

func Cleanup() {
	closeTunListener()
}
