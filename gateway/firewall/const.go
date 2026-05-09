package firewall
const (
	module                 		string = "IPTABLES"
	constFirewallTypeNftables	string = "nftables"
	constFirewallTypeIptables	string = "iptables"
	constRoutingTableName       string = "bypass"
	constRoutingTableId			byte   = 144
	constRoutingRuleMark    	uint32 = 144
	constRoutingRuleMask     	uint32 = 0xffffffff
	constRoutingRuleNextHop   	string = "tasks.vpn"
)

const constNFTablesConf = `#!/usr/sbin/nft -f

flush ruleset

table inet bypass_filter {
        set default_ipv4 {
                type ipv4_addr
                flags interval
                elements = { 8.8.8.8, 8.4.4.8 }
        }
        set default_ipv6 {
                type ipv6_addr
                flags interval
        }
        chain prerouting {
                type filter hook prerouting priority mangle; counter; policy accept;
                ip daddr @default_ipv4 meta mark set 0x85 counter
                ip6 daddr @default_ipv6 meta mark set 0x85 counter
        }
        chain output {
                type route hook output priority mangle; counter; policy accept;
                ip daddr @default_ipv4 meta mark set 0x85 counter
                ip6 daddr @default_ipv6 meta mark set 0x85 counter
        }
        chain postrouting {
                type nat hook postrouting priority srcnat; counter; policy accept;
                counter masquerade
        }
}
`
