package firewall
//	ip rule add fwmark 0x1 lookup 144
//	ip route add default dev wg0 table 144
//	ip route add default via 192.168.1.1 table 144
// nft add element inet bypass main_ipv4 {  }
const NFTablesConf_c = `#!/usr/sbin/nft -f

flush ruleset

table inet bypass {
        set main_ipv4 {
                type ipv4_addr
                flags interval
                comment "Bypass: IPv4 main address list"
        }

        set main_ipv6 {
                type ipv6_addr
                flags interval
                comment "Bypass: IPv6 main address list"
        }

        chain prerouting {
                type filter hook prerouting priority mangle; counter; policy accept; comment "Bypass: marking a packet before routing"
                ip daddr @main_ipv4 meta mark set 0x00000001 counter comment "Bypass: mark IPv4 traffic destined to main_ipv4"
                ip6 daddr @main_ipv6 meta mark set 0x00000001 counter comment "Bypass: mark IPv6 traffic destined to main_ipv6"
        }

        chain postrouting {
                type nat hook postrouting priority srcnat; counter; policy accept; comment "Bypass: NAT for outgoing packets with the 0x1 mark"
                meta mark 0x00000001 counter masquerade comment "Bypass: masquerade marked traffic via route table bypass_main"
        }
}
`
