package tun

import (
	"log"

	"github.com/metacubex/mihomo/component/resolver"
	C "github.com/metacubex/mihomo/constant"
	"github.com/metacubex/mihomo/dns"
	"github.com/qdm12/dns/v2/pkg/nameserver"
)

func ReadSystemDNS() ([]dns.NameServer, error) {

	var nameServers []dns.NameServer

	dnss := nameserver.GetDNSServers()
	for _, v := range dnss {
		nameServers = append(nameServers, dns.NameServer{Addr: v.String()})
	}

	return nameServers, nil
}

func UpdateDNS() {

	sysDNS, err := ReadSystemDNS()
	for _, v := range sysDNS {
		log.Println("sysDNS:", v.Addr)
	}
	if err != nil {
		log.Println("ReadSystemDNS error:", err)
	}

	cfg := dns.Config{
		Main:         sysDNS,
		EnhancedMode: C.DNSNormal,
		Default:      sysDNS,
	}

	r := dns.NewResolver(cfg)
	m := dns.NewEnhancer(cfg)

	// reuse cache of old host mapper
	if old := resolver.DefaultHostMapper; old != nil {
		m.PatchFrom(old.(*dns.ResolverEnhancer))
	}

	resolver.DefaultResolver = r
	resolver.DefaultHostMapper = m
	resolver.DefaultLocalServer = dns.NewLocalServer(r.Resolver, m)
	resolver.UseSystemHosts = true

	if r.ProxyResolver.Invalid() {
		resolver.ProxyServerHostResolver = r.ProxyResolver
	} else {
		resolver.ProxyServerHostResolver = r.Resolver
	}

	if r.DirectResolver.Invalid() {
		resolver.DirectHostResolver = r.DirectResolver
	} else {
		resolver.DirectHostResolver = r.Resolver
	}

	dns.ReCreateServer("any:5399", r.Resolver, m)

}
