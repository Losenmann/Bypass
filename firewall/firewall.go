package firewall

import (
	"bypass/tools"
	"encoding/binary"
	"errors"
	"flag"
	"log/slog"
	"strconv"
	"net"
	"github.com/google/nftables"
	"github.com/google/nftables/expr"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/vishvananda/netlink"
)

const (
	module                 = "IPTABLES"
	firewallType1_c string = "nftables"
	firewallType2_c string = "iptables"
	routeTable             = 144
)

type Flags_t struct {
	Enable bool
	Type flag.Value
	DSL bool
}

type Metrics_t struct {
	Packets *prometheus.GaugeVec
	Bytes   *prometheus.GaugeVec
}

var (
	Args Flags_t
	Test interface{}
	Metrics       = Metrics_t{
		Packets: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "bypass_stats_packets",
				Help: "Statistics processed packages.",
			},
			[]string{"chain", "mark", "sets"},
		),
		Bytes: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "bypass_stats_bytes",
				Help: "Statistics processed bytes.",
			},
			[]string{"chain", "mark", "sets"},
		),
	}
)

func init() {
	var DescType string
	Args.Type, DescType = tools.NewEnumStringFlag(firewallType1_c, []string{firewallType1_c, firewallType2_c}, "Firewall run mode")
	flag.BoolVar(&Args.DSL, "n", false, "Firewall show config")
	flag.BoolVar(&Args.Enable, "F", false, "Enable firewall")
	flag.Var(Args.Type, "f", DescType)
	prometheus.MustRegister(Metrics.Bytes, Metrics.Packets)
}

func Run() {
	//NewNFTManager()
	
	// Доработать функцию получения ip адреса шлюза
	//SetupIPRoute2("127.0.0.1")
	//RecreateIPv4Set(false)
}


func SetupIPRoute2(gateway string) error {
	rule := &netlink.Rule{Mark: 1, Table: routeTable}
	route := &netlink.Route{Gw: nil, Scope: netlink.SCOPE_UNIVERSE, Table: routeTable}

	if gw := net.ParseIP(gateway); gw == nil {
		err := errors.New("incorrect gateway ip address")
		slog.Error(err.Error(), "tag", module)
		return err
	} else {
		route.Gw = gw
	}

	if rules, err := netlink.RuleList(netlink.FAMILY_ALL); err != nil {
		slog.Error(err.Error(), "tag", module)
		return err
	} else {
		for _, r := range rules {
			if r.Mark == rule.Mark && r.Table == rule.Table {
				slog.Info("rule already exist", "tag", module)
				return nil
			}
		}
		if err := netlink.RuleAdd(rule); err != nil {
			slog.Error(err.Error(), "tag", module)
			return err
		} else {
			slog.Info("rule added success", "tag", module)
			if err := netlink.RouteAdd(route); err != nil {
				slog.Error(err.Error(), "tag", module)
				return err
			} else {
				slog.Info("route added success", "tag", module)
				return nil
			}
		}
	}
}

func GetMetrics() {
	conn := &nftables.Conn{}
	table := &nftables.Table{Family: nftables.TableFamilyINet, Name: "bypass"}
	if chains, err := conn.ListChains(); err != nil {
		slog.Warn(err.Error(), "tag", module)
	} else {
		slog.Debug("get chains list success", "tag", module)
		for _, chain := range chains {
			if rules, err := conn.GetRules(table, chain); err != nil {
				slog.Warn(err.Error(), "tag", module)
			} else {
				slog.Debug("get rules list success", "tag", module)
				for _, rule := range rules {
					var packets, bytes float64
					var sets, mark string
					for _, e := range rule.Exprs {
						switch c := e.(type) {
						case *expr.Counter:
							packets = float64(c.Packets)
							bytes = float64(c.Bytes)
						case *expr.Immediate:
							mark = strconv.Itoa(int(binary.LittleEndian.Uint32(c.Data)))
						case *expr.Cmp:
							mark = strconv.Itoa(int(binary.LittleEndian.Uint32(c.Data)))
						case *expr.Lookup:
							sets = c.SetName
						}
					}
					Metrics.Packets.WithLabelValues(rule.Chain.Name, mark, sets).Set(packets)
					Metrics.Bytes.WithLabelValues(rule.Chain.Name, mark, sets).Set(bytes)
					slog.Debug("set metrics success", "tag", module)
				}
			}
		}
	}
}



/*
		link, err := netlink.LinkByName("eth0")
	    if err != nil {
	        panic(err)
	    }
		route := &netlink.Route{
	        LinkIndex: link.Attrs().Index,
	        Scope:     netlink.SCOPE_UNIVERSE,
	        Table:     100, // таблица 100
	    }

	    // добавляем default (0.0.0.0/0)
	    dst, _ := netlink.ParseIPNet("0.0.0.0/0")
	    route.Dst = dst

	    if err := netlink.RouteAdd(route); err != nil {
	        panic(err)
	    }
*/

/*
func GetType() string {
	conn := &nftables.Conn{}
	_, err := conn.ListTables()
	if err != nil {
		slog.Error(err.Error(), "tag", module)
		return "iptables"
	} else {
		return "nftables"
	}
}

*/
