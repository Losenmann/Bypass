/system/script/add name="funcBypassctl" source="\
:global funcBypassctl do={
    :local command \$1
    :local action \$2
    :local gateway \$3
    :local version \"1.0.0\"
    :local vdnscachesize 25
    :local vaddresslist \"bypass\"
    :local vroutetable \"rtab-bypass\"
    :local vroutemark \"rtab-bypass\"

    # Setup Bypass
    :if (\$command = \"conf\" && \$action = \"setup\") do={
        :if ([:len \$gateway] = 0) do={
            :put \"[ERROR] Set gateway or interface\"
            :return 1
        }

        # Resize DNS cache
        :do {
            /ip/dns/set cache-size=((([/system/resource/get free-memory]/1024)/100)*\$vdnscachesize)
            :put \"[INFO] DNS cache resize successfully\"
        } on-error={
            :put \"[WARN] Error DNS cache resize\"
        }

        # Add route table
        :do {
            :if ([/routing/table/find where name=\"\$vroutetable\"]) do={
                :put \"[INFO] Exists route table\"
            } else {
                /routing/table/add name=\"\$vroutetable\" fib comment=\"custom: table of bypass routes\"
                :put \"[INFO] Added route table successfully\"
            }
        } on-error={
            :put \"[WARN] Error adding route table\"
        }

        # Add routing rule
        :do {
            :if ([/routing/rule/find where routing-mark=\$vroutemark action=lookup-only-in-table table=\$vroutetable]) do={
                :put \"[INFO] Exists routing rule\"
            } else {
                /routing/rule/add routing-mark=\$vroutemark action=lookup-only-in-table table=\$vroutetable place-before=0 comment=\"custom: rule bypass route\"
                :put \"[INFO] Added routing rule successfully\"
            }
        } on-error={
            :put \"[WARN] Error adding routing rule\"
        }

        # Add route
        :do {
            :if ([/ip/route/find where dst-address=0.0.0.0/0 routing-table=\$vroutetable]) do={
                :put \"[INFO] Exists route\"
                /ip/route/set [/ip/route/find where dst-address=0.0.0.0/0 routing-table=\$vroutetable] gateway=\$gateway
            } else {
                /ip/route/add dst-address=0.0.0.0/0 gateway=\$gateway distance=100 routing-table=\$vroutetable comment=\"custom: route bypass\"
                :put \"[INFO] Added route successfully\"
            }
        } on-error={
            :put \"[WARN] Error adding route\"
        }

        # Add mangle rule
        :do {
            :if ([/ip/firewall/mangle/find where chain=prerouting dst-address-list=\$vaddresslist action=mark-routing new-routing-mark=\$vroutemark]) do={
                :put \"[INFO] Exists mangle rule\"
            } else {
                /ip/firewall/mangle/add chain=prerouting dst-address-list=\$vaddresslist action=mark-routing new-routing-mark=\$vroutemark passthrough=yes place-before=0 comment=\"custom: route marking bypass\"
                :put \"[INFO] Added mangle rule successfully\"
            }
        } on-error={
            :put \"[WARN] Error mangle rule\"
        }
        :return 0
    }

    # Remove Bypass
    :if (\$command = \"conf\" && \$action = \"rm\") do={
        # Remove mangle rule
        :do {
            /ip/firewall/mangle/remove [/ip/firewall/mangle/find where chain=prerouting dst-address-list=\$vaddresslist action=mark-routing new-routing-mark=\$vroutemark]
            :put \"[INFO] Removed mangle rule successfully\"
        } on-error={
            :put \"[WARN] Error remove mangle rule\"
        }

        # Remove route
        :do {
            /ip/route/remove [/ip/route/find where dst-address=0.0.0.0/0 routing-table=\$vroutetable]
            :put \"[INFO] Removed route successfully\"
        } on-error={
            :put \"[WARN] Error remove route\"
        }

        # Remove routing rule
        :do {
            /routing/rule/remove [/routing/rule/find where routing-mark=\$vroutemark action=lookup-only-in-table table=\$vroutetable]
            :put \"[INFO] Removed routing rule successfully\"
        } on-error={
            :put \"[WARN] Error remove routing rule\"
        }

        # Remove route table
        :do {
            /routing/table/remove [/routing/table/find name=\$vroutetable] 
            :put \"[INFO] Removed route table successfully\"
        } on-error={
            :put \"[WARN] Error remove route table\"
        }

        # Set default size DNS cache 
        :do {
            /ip/dns/set cache-size=2048
            :put \"[INFO] DNS cache set size default successfully\"
        } on-error={
            :put \"[WARN] Error DNS cache set size default\"
        }
        :return 0
    }

    # Service control
    :if (\$command = \"svc\") do={
        :if ($action = true) do={
            :do {
                /ip/firewall/mangle/enable [/ip/firewall/mangle/find where chain=\"prerouting\" dst-address-list=\"\$vaddresslist\" action=\"mark-routing\" new-routing-mark=\"\$vroutemark\"]
                :return 0
            } on-error={
                :put \"[ERROR] Not found mangle rule. Please run the setup process «\$funcBypassctl conf 1 \\\"VPN_interface/address\\\"»\"
                :return 1
            }
        }
        :if (\$action = false) do={
            :do {
                /ip/firewall/mangle/disable [/ip/firewall/mangle/find where chain=\"prerouting\" dst-address-list=\"\$vaddresslist\" action=\"mark-routing\" new-routing-mark=\"\$vroutemark\"]
                :return 0
            } on-error={
                :put \"[ERROR] Not found mangle rule. Please run the setup process «\$funcBypassctl conf 1 \\\"VPN_interface/address\\\"»\"
                :return 1
            }
        }
    }

    # Version
    :if (\$command = \"version\") do={
        :put \$version
        :return 0
    }

    # Scheduler
    :if (\$command = \"startup\") do={
        :if (\$action = true) do={
            :do {
                /system/scheduler/add name=\"funcBypassctl\" on-event=\"/system/script/run funcBypassctl\" start-time=\"startup\"
                :put \"[INFO] Added scheduler task\"
            } on-error={
                :put \"[INFO] Exists scheduler task\"
            }
            :return 0
        }
        :if (\$action = false) do={
            :do {
                /system/scheduler/remove \"funcBypassctl\"
                :put \"[INFO] Removed scheduler task\"
            } on-error={
                :put \"[INFO] Not exists scheduler task\"
            }
            :return 0
        }
    }

    # Help
    :if (\$command = \"help\") do={
        :put \"help    - Reference information\"
        :put \"version - The version of the currently running software\"
        :put \"conf    - Service setup (setup/rm)\"
        :put \"svc     - Service control (true/false)\"
        :put \"startup - Start at power on (true/false)\"
        :return 0
    }
    :return 0
}
"
/system/script/run funcBypassctl
$funcBypassctl startup true