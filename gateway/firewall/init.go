package firewall

import (
	"github.com/prometheus/client_golang/prometheus"
	"flag"
	"bypass/tools"
)

type typeFlags struct {
	Enable bool
	Type flag.Value
	DSL bool
}

var (
	Args typeFlags
)

func init() {
	var DescType string
	Args.Type, DescType = tools.NewEnumStringFlag(constFirewallTypeNftables, []string{constFirewallTypeNftables, constFirewallTypeIptables}, "Firewall run mode")
	flag.BoolVar(&Args.DSL, "n", false, "Firewall show config")
	flag.BoolVar(&Args.Enable, "F", false, "Enable firewall")
	flag.Var(Args.Type, "f", DescType)
	prometheus.MustRegister(Metrics.Bytes, Metrics.Packets)
}