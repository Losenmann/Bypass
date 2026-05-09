package routing

const (
	constBirdSockPath = "/var/run/bird.ctl"
	constBirdConfPath = "/etc/bird/bird.conf"
)

const constBirdConf = `log stderr all;
ipv4 table bypass;
protocol device {
}

protocol kernel Kernel {
    learn;
    kernel table 85;
    ipv4 {
        table bypass;
        import all;
        export none;
    };
}

protocol rip RIPv2 {
    interface "eth0" {
        mode multicast;
        version 2;
        update time 5;
        timeout time 15;
        garbage time 10;
    };
    ipv4 {
        table bypass;
        import none;
        export all;
    };
}

protocol ospf v2 OSPFv2 {
    area 0.0.0.0 {
        interface "eth0" {
            type broadcast;
            cost 10;
        };
    };
    ipv4 {
        table bypass;
        import all;
        export all;
    };
}
`