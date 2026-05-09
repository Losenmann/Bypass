package firewall

import (
	"net"
	"time"

	"github.com/vishvananda/netlink"
)

type IPRouteManager struct {
	Route *netlink.Route
	Rule  *netlink.Rule
}

var (
	iproute = NewIPRouteManager()
)

func NewIPRouteManager() *IPRouteManager {
	manager := &IPRouteManager{
		Route: &netlink.Route{
			Dst:   nil,
			Gw:    net.ParseIP(resolv(constRoutingRuleNextHop)),
			Table: int(constRoutingTableId),
		},
		Rule: &netlink.Rule{
			Mark:  constRoutingRuleMark,
			Mask:  new(constRoutingRuleMask),
			Table: int(constRoutingTableId),
		},
	}
	return manager
}

func (n *IPRouteManager) Setup() {
	if err := netlink.RouteAdd(n.Route); err != nil {
		panic(err)
	}
	if err := netlink.RuleAdd(n.Rule); err != nil {
		panic(err)
	}
}

func RunIPRoute() {
	iproute.Setup()
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		if newTTT := resolv("tasks.vpn"); newTTT != iproute.Route.Gw.String() {
			iproute.Route.Gw = net.ParseIP(newTTT)
			if err := netlink.RouteReplace(iproute.Route); err != nil {
				panic(err)
			}
		}
	}
}

func resolv(dns string) string {
	if ips, err := net.LookupIP(dns); err != nil {
		panic(err)
	} else {
		for _, ip := range ips {
			return ip.String()
		}
	}
	return ""
}