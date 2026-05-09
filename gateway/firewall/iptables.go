package firewall

import (
	"github.com/coreos/go-iptables/iptables"
	"log/slog"
	"github.com/lrh3321/ipset-go"
)



func SetupIPtables() {
	ipt, _ := iptables.New()
	ruleSpec := []string{"-m", "set", "--match-set", "to_marked", "dst", "-j", "MARK", "--set-mark", "0x1"}
	linkSpec := []string{"-j", "bypass"}

	if exists, err := ipt.ChainExists("mangle", "bypass"); err != nil {
		slog.Error(err.Error(), "tag", module)
	} else if !exists {
		if err := ipt.NewChain("mangle", "bypass"); err != nil {
			slog.Error(err.Error(), "tag", module)
		}
	}

	if hasRule, err := ipt.Exists("mangle", "bypass", ruleSpec...); err != nil {
		slog.Error(err.Error(), "tag", module)
	} else if !hasRule {
		if err := ipt.AppendUnique("mangle", "bypass", ruleSpec...); err != nil {
			slog.Error(err.Error(), "tag", module)
		}
	}

	if hasLink, err := ipt.Exists("mangle", "PREROUTING", linkSpec...); err != nil {
		slog.Error(err.Error(), "tag", module)
	} else if !hasLink {
		if err := ipt.AppendUnique("mangle", "PREROUTING", linkSpec...); err != nil {
			slog.Error(err.Error(), "tag", module)
		}
	}
	
	if err := ipset.Create("bypass", ipset.TypeHashIP, ipset.CreateOptions{}); err != nil {
		slog.Error(err.Error(), "tag", module)
	}
}

