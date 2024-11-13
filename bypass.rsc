/routing/table/add name="bypass" fib comment="custom: bypass"
/routing/rule/add routing-mark="route-baypass" action="lookup-only-in-table" table="bypass" comment="custom: bypass"
/ip/dns/set cache-size=((([/system/resource/get total-memory]/100)*8)/1024)
/ip/firewall/mangle/add chain="prerouting" dst-address-list="bypass" connection-state="new,related,established" action="mark-connection" new-connection-mark="conn-bypass" passthrough="yes" place-before="0" comment="custom: bypass"
/ip/firewall/mangle/add chain="prerouting" connection-mark="conn-bypass" action="mark-routing" new-routing-mark="route-baypass" passthrough="yes" place-before="1" comment="custom: bypass"
/ip/route/add dst-address="0.0.0.0/0" gateway="vpn-out1" distance="100" routing-table="bypass" comment="custom: bypass"
