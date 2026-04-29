package main

import (
	"net"
	_ "sync"
	"time"
	"github.com/vishvananda/netlink"
	"github.com/google/nftables"
	"github.com/google/nftables/expr"
)

type NFTManager struct {
	Conn   nftables.Conn
	fwmark []byte
	Table  *nftables.Table
	Sets   []*nftables.Set
	Chain  []*nftables.Chain
	Rule   []*nftables.Rule
}

var (
	nft = NewNFTManager()
)

func main() {

	route := &netlink.Route{
		Dst:   nil,
		Gw:    net.ParseIP(resolv("tasks.vpn")),
		Table: 85,
	}

	if err := netlink.RouteAdd(route); err != nil {
		panic(err)
	}

	rule := netlink.NewRule()
	rule.Mark = 0x85
	rule.Mask = new(uint32(0xffffffff))
	rule.Table = 85
	if err := netlink.RuleAdd(rule); err != nil {
		panic(err)
	}

	RunNFTables()

	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		if newTTT := resolv("tasks.vpn"); newTTT != route.Gw.String() {
			route.Gw = net.ParseIP(newTTT)
			if err := netlink.RouteReplace(route); err != nil {
				panic(err)
			}
		}
	}
}

func NewNFTManager() *NFTManager {
	manager := &NFTManager{
		fwmark: []byte{0x85, 0x00, 0x00, 0x00},
		Conn:   nftables.Conn{},
		Table: &nftables.Table{
			Family: nftables.TableFamilyINet,
			Name:   "bypass",
		},
	}

	manager.Sets = []*nftables.Set{
		{
			Table:    manager.Table,
			Name:     "default_ipv4",
			KeyType:  nftables.TypeIPAddr,
			Interval: true,
			Comment:  "default IPv4 address list",
		},
		{
			Table:    manager.Table,
			Name:     "default_ipv6",
			KeyType:  nftables.TypeIP6Addr,
			Interval: true,
			Comment:  "default IPv6 address list",
		},
	}

	manager.Chain = []*nftables.Chain{
		{
			Table:    manager.Table,
			Name:     "prerouting",
			Type:     nftables.ChainTypeFilter,
			Hooknum:  nftables.ChainHookPrerouting,
			Priority: nftables.ChainPriorityMangle,
		},
		{
			Table:    manager.Table,
			Name:     "output",
			Type:     nftables.ChainTypeRoute,
			Hooknum:  nftables.ChainHookOutput,
			Priority: nftables.ChainPriorityMangle,
		},
		{
			Table:    manager.Table,
			Name:     "postrouting",
			Type:     nftables.ChainTypeNAT,
			Hooknum:  nftables.ChainHookPostrouting,
			Priority: nftables.ChainPriorityNATSource,
		},
	}

	manager.Rule = []*nftables.Rule{
		{
			Table: manager.Table,
			Chain: manager.Chain[0],
			Exprs: []expr.Any{
				&expr.Payload{OperationType: expr.PayloadLoad, DestRegister: 1, Base: expr.PayloadBaseNetworkHeader, Offset: 16, Len: 4},
				&expr.Lookup{SourceRegister: 1, SetID: manager.Sets[0].ID, SetName: manager.Sets[0].Name},
				&expr.Immediate{Register: 1, Data: manager.fwmark},
				&expr.Meta{Key: expr.MetaKeyMARK, Register: 1, SourceRegister: true},
				&expr.Counter{},
			},
		},
		{
			Table: manager.Table,
			Chain: manager.Chain[0],
			Exprs: []expr.Any{
				&expr.Payload{OperationType: expr.PayloadLoad, DestRegister: 1, Base: expr.PayloadBaseNetworkHeader, Offset: 24, Len: 16},
				&expr.Lookup{SourceRegister: 1, SetID: manager.Sets[1].ID, SetName: manager.Sets[1].Name},
				&expr.Immediate{Register: 1, Data: manager.fwmark},
				&expr.Meta{Key: expr.MetaKeyMARK, Register: 1, SourceRegister: true},
				&expr.Counter{},
			},
		},
		{
			Table: manager.Table,
			Chain: manager.Chain[1],
			Exprs: []expr.Any{
				&expr.Payload{OperationType: expr.PayloadLoad, DestRegister: 1, Base: expr.PayloadBaseNetworkHeader, Offset: 16, Len: 4},
				&expr.Lookup{SourceRegister: 1, SetID: manager.Sets[0].ID, SetName: manager.Sets[0].Name},
				&expr.Immediate{Register: 1, Data: manager.fwmark},
				&expr.Meta{Key: expr.MetaKeyMARK, Register: 1, SourceRegister: true},
				&expr.Counter{},
			},
		},
		{
			Table: manager.Table,
			Chain: manager.Chain[1],
			Exprs: []expr.Any{
				&expr.Payload{OperationType: expr.PayloadLoad, DestRegister: 1, Base: expr.PayloadBaseNetworkHeader, Offset: 24, Len: 16},
				&expr.Lookup{SourceRegister: 1, SetID: manager.Sets[1].ID, SetName: manager.Sets[1].Name},
				&expr.Immediate{Register: 1, Data: manager.fwmark},
				&expr.Meta{Key: expr.MetaKeyMARK, Register: 1, SourceRegister: true},
				&expr.Counter{},
			},
		},
		{
			Table: manager.Table,
			Chain: manager.Chain[2],
			Exprs: []expr.Any{
				&expr.Meta{Key: expr.MetaKeyMARK, Register: 1},
				&expr.Cmp{Op: expr.CmpOpEq, Register: 1, Data: manager.fwmark},
				&expr.Counter{},
				&expr.Masq{},
			},
		},
	}
	return manager
}

func (n *NFTManager) Setup() {
	n.Conn.FlushTable(n.Table)
	n.Conn.DelTable(n.Table)
	n.Conn.Flush()
	n.Conn.AddTable(n.Table)
	n.Conn.AddChain(n.Chain[0])
	n.Conn.AddChain(n.Chain[1])
	n.Conn.AddChain(n.Chain[2])
	n.Conn.AddSet(n.Sets[0], nil)
	n.Conn.AddSet(n.Sets[1], nil)
	n.Conn.AddRule(n.Rule[0])
	n.Conn.AddRule(n.Rule[1])
	n.Conn.AddRule(n.Rule[2])
	n.Conn.AddRule(n.Rule[3])
	n.Conn.AddRule(n.Rule[4])
	n.Conn.Flush()
}

func RunNFTables() {
	nft.Setup()
	//      resp, _ := consumer.GetCIDRs("", false)
	nft.AddSetElem("default", new([]string{"0.0.0.0/0"}))

}

func (n *NFTManager) AddSetElem(name string, cidrs *[]string) (err error) {
	var sets4, sets6 *nftables.Set
	var elem4, elem6 []nftables.SetElement

	if resp, err := n.Conn.GetSets(n.Table); err != nil {
		return err
	} else {
		for _, v := range resp {
			switch v.Name {
			case name + "_ipv4":
				sets4 = v
			case name + "_ipv6":
				sets6 = v
			}
		}

		for _, v := range *cidrs {
			first, end, _ := SetElemGenStartEnd(v)
			if first.To4() != nil {
				elem4 = append(elem4, []nftables.SetElement{{Key: first.To4(), IntervalEnd: false}, {Key: end.To4(), IntervalEnd: true}}...)
				continue
			}
			if first.To16() != nil {
				elem6 = append(elem6, []nftables.SetElement{{Key: first.To16(), IntervalEnd: false}, {Key: end.To16(), IntervalEnd: true}}...)
				continue
			}

		}

		if n.Conn.SetAddElements(sets4, elem4) != nil {
			return err
		}
		if n.Conn.SetAddElements(sets6, elem6) != nil {
			return err
		}
		n.Conn.Flush()
		return nil
	}
}

func (n *NFTManager) AddSet(name, desc string, flush bool) {
	if flush {
		resp, _ := n.Conn.GetSets(n.Table)
		for _, v := range resp {
			n.Conn.DelSet(v)
		}
		n.Conn.Flush()
		n.Sets = []*nftables.Set{
			{
				Table:    n.Table,
				Name:     name + "_ipv4",
				KeyType:  nftables.TypeIPAddr,
				Interval: true,
				Comment:  desc,
			},
			{
				Table:    n.Table,
				Name:     name + "_ipv6",
				KeyType:  nftables.TypeIP6Addr,
				Interval: true,
				Comment:  desc,
			},
		}
	} else {
		n.Sets = append(n.Sets, []*nftables.Set{
			{
				Table:    n.Table,
				Name:     name + "_ipv4",
				KeyType:  nftables.TypeIPAddr,
				Interval: true,
				Comment:  desc,
			},
			{
				Table:    n.Table,
				Name:     name + "_ipv6",
				KeyType:  nftables.TypeIP6Addr,
				Interval: true,
				Comment:  desc,
			},
		}...)
	}
	for _, v := range n.Sets {
		n.Conn.AddSet(v, nil)
	}
	n.Conn.Flush()
}

func SetElemGenStartEnd(cidr string) (first, end net.IP, err error) {
	_, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, nil, err
	}

	ip := ipnet.IP
	mask := ipnet.Mask
	ipLen := len(ip)

	// Быстрое создание IP с правильной длиной
	first = make(net.IP, ipLen)
	end = make(net.IP, ipLen)

	// Один проход по байтам
	for i := range ip {
		first[i] = ip[i] & mask[i]
		end[i] = ip[i] | ^mask[i]

	}

	// Быстрый инкремент
	for i := ipLen - 1; i >= 0; i-- {
		end[i]++
		if end[i] > 0 {
			break
		}
	}
	return first, end, nil
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
