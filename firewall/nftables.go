package firewall

import (
	//	"log"
	"fmt"
	"net"
	"time"

	"github.com/google/nftables"
	"github.com/google/nftables/expr"
	// "golang.org/x/sys/unix"
)

func SetupNFT() {
	conn := &nftables.Conn{}
	table := &nftables.Table{
		Family: nftables.TableFamilyINet,
		Name:   "bypass",
	}
	set_main_ipv4 := &nftables.Set{
		Table: table,
		Name:  "main_ipv4",
		KeyType: nftables.SetDatatype{
			Name: "ipv4_addr",
		},
		Interval: true,
	}
	set_main_ipv6 := &nftables.Set{
		Table: table,
		Name:  "main_ipv6",
		KeyType: nftables.SetDatatype{
			Name: "ipv6_addr",
		},
		Interval: true,
	}
	prerouting := &nftables.Chain{
		Name:     "prerouting",
		Table:    table,
		Type:     nftables.ChainTypeFilter,
		Hooknum:  nftables.ChainHookPrerouting,
		Priority: nftables.ChainPriorityMangle,
	}
	postrouting := &nftables.Chain{
		Name:     "postrouting",
		Table:    table,
		Type:     nftables.ChainTypeNAT,
		Hooknum:  nftables.ChainHookPostrouting,
		Priority: nftables.ChainPriorityNATSource,
	}
	rule_nat_main := &nftables.Rule{
		Table: table,
		Chain: postrouting,
		Exprs: []expr.Any{
			&expr.Meta{Key: expr.MetaKeyMARK, Register: 1},
			&expr.Counter{},
			&expr.Masq{},
		},
	}
	rule_mangle_main := &nftables.Rule{
		Table: table,
		Chain: prerouting,
		Exprs: []expr.Any{
			&expr.Lookup{SourceRegister: 1, SetID: set_main_ipv4.ID, SetName: set_main_ipv4.Name},
			&expr.Meta{Key: expr.MetaKeyMARK, Register: 1},
			&expr.Counter{},
		},
	}

	/*
		rule := &nftables.Rule{
			Table: table,
			Chain: prerouting,
			Exprs: []expr.Any{
				// Загружаем IPv4-адрес назначения в регистр 1
				&expr.Payload{
					DestRegister: 1,
					Base:         expr.PayloadBaseNetworkHeader,
					Offset:       16, // IPv4 dst offset
					Len:          4,
				},
				// Проверяем, что ip daddr ∈ @to_marked
				&expr.Lookup{
					SourceRegister: 1,
					SetName:        set_main_ipv4.Name,
					SetID:          set_main_ipv4.ID,
				},
				// Если совпало — записываем mark = 0x1
				&expr.Meta{
					Key:      expr.MetaKeyMARK,
					Register: 1,
				},
				&expr.Immediate{
					Register: 1,
					Data:     []byte{0x01, 0x00, 0x00, 0x00}, // 0x1 (little-endian)
				},
				&expr.Meta{
					Key:      expr.MetaKeyMARK,
					Register: 1,
					SourceRegister:   false, // записать (set)
				},
			},
		}
	*/

	conn.AddTable(table)
	conn.AddSet(set_main_ipv4, nil)
	conn.AddSet(set_main_ipv6, nil)
	conn.AddChain(prerouting)
	conn.AddChain(postrouting)
	conn.AddRule(rule_nat_main)
	conn.AddRule(rule_mangle_main)
	conn.Flush()

}

func addIPSet(conn *nftables.Conn, table *nftables.Table, name string, ipv4 bool, ips []string) *nftables.Set {
	var elements []nftables.SetElement
	set := &nftables.Set{
		Table:    table,
		Name:     name,
		KeyType:  nftables.TypeIPAddr,
		Interval: true,
	}

	if !ipv4 {
		set.KeyType = nftables.TypeIP6Addr
	}

	for _, ipStr := range ips {
		ip := net.ParseIP(ipStr)
		if ipv4 {
			ip = ip.To4()
		}
		elements = append(elements, nftables.SetElement{Key: ip})
	}

	conn.AddSet(set, elements)
	return set
}

func addMarkRule(conn *nftables.Conn, table *nftables.Table, chain *nftables.Chain, set *nftables.Set) {
	conn.AddRule(&nftables.Rule{
		Table: table,
		Chain: chain,
		Exprs: []expr.Any{
			&expr.Lookup{
				SourceRegister: 1,
				SetName:        set.Name,
				SetID:          set.ID,
			},
			&expr.Meta{Key: expr.MetaKeyMARK, Register: 1},
			&expr.Immediate{Register: 1, Data: []byte{0x01, 0x00, 0x00, 0x00}},
			&expr.Meta{Key: expr.MetaKeyMARK, Register: 1},
			&expr.Counter{},
		},
	})
}

// helper: создаёт правило masquerade по метке
func addMasqRule(conn *nftables.Conn, table *nftables.Table, chain *nftables.Chain, mark uint32) {
	data := []byte{byte(mark), 0x00, 0x00, 0x00}
	conn.AddRule(&nftables.Rule{
		Table: table,
		Chain: chain,
		Exprs: []expr.Any{
			&expr.Meta{Key: expr.MetaKeyMARK, Register: 1},
			&expr.Cmp{Register: 1, Op: expr.CmpOpEq, Data: data},
			&expr.Counter{},
			&expr.Masq{},
		},
	})
}

func SetupNFtables() {
	var fwmark []byte = []byte{0x01, 0x00, 0x00, 0x00}
	conn := &nftables.Conn{}
	conn2 := &nftables.Conn{}
	table := &nftables.Table{
		Family: nftables.TableFamilyINet,
		Name:   "bypass",
	}
	set_main4 := &nftables.Set{
		Table:    table,
		Name:     "main_ipv4",
		KeyType:  nftables.TypeIPAddr,
		Interval: true,
	}
	set_main6 := &nftables.Set{
		Table:    table,
		Name:     "main_ipv6",
		KeyType:  nftables.TypeIP6Addr,
		Interval: true,
	}
	prerouting := &nftables.Chain{
		Name:     "prerouting",
		Table:    table,
		Type:     nftables.ChainTypeFilter,
		Hooknum:  nftables.ChainHookPrerouting,
		Priority: nftables.ChainPriorityMangle,
	}
	postrouting := &nftables.Chain{
		Name:     "postrouting",
		Table:    table,
		Type:     nftables.ChainTypeNAT,
		Hooknum:  nftables.ChainHookPostrouting,
		Priority: nftables.ChainPriorityNATSource,
	}
	rule_mangle_main4 := &nftables.Rule{
		Table: table,
		Chain: prerouting,
		Exprs: []expr.Any{
			&expr.Payload{OperationType: expr.PayloadLoad, DestRegister: 1, Base: expr.PayloadBaseNetworkHeader, Offset: 16, Len: 4},
			&expr.Lookup{SourceRegister: 1, SetID: set_main4.ID, SetName: set_main4.Name},
			&expr.Immediate{Register: 1, Data: fwmark},
			&expr.Meta{Key: expr.MetaKeyMARK, Register: 1, SourceRegister: true},
			&expr.Counter{},
		},
	}
	rule_mangle_main6 := &nftables.Rule{
		Table: table,
		Chain: prerouting,
		Exprs: []expr.Any{
			&expr.Payload{OperationType: expr.PayloadLoad, DestRegister: 1, Base: expr.PayloadBaseNetworkHeader, Offset: 24, Len: 16},
			&expr.Lookup{SourceRegister: 1, SetID: set_main6.ID, SetName: set_main6.Name},
			&expr.Immediate{Register: 1, Data: fwmark},
			&expr.Meta{Key: expr.MetaKeyMARK, Register: 1, SourceRegister: true},
			&expr.Counter{},
		},
	}
	rule_nat_main := &nftables.Rule{
		Table: table,
		Chain: postrouting,
		Exprs: []expr.Any{
			&expr.Meta{Key: expr.MetaKeyMARK, Register: 1},
			&expr.Cmp{Op: expr.CmpOpEq, Register: 1, Data: fwmark},
			&expr.Counter{},
			&expr.Masq{},
		},
	}
	conn.AddTable(table)
	if err := conn.AddSet(set_main4, nil); err != nil {
		fmt.Println(err.Error())
	}
	if err := conn.AddSet(set_main6, nil); err != nil {
		fmt.Println(err.Error())
	}
	conn2.FlushTable(table)
	conn2.DelTable(table)
	conn2.Flush()
	time.Sleep(1 * time.Second)
	conn.AddTable(table)
	conn.AddChain(prerouting)
	conn.AddChain(postrouting)
	conn.AddRule(rule_mangle_main4)
	conn.AddRule(rule_mangle_main6)
	conn.AddRule(rule_nat_main)
	conn.Flush()
}

