package firewall

import (
	"github.com/prometheus/client_golang/prometheus"
	"log/slog"
	"encoding/binary"
	"github.com/google/nftables"
	"github.com/google/nftables/expr"
	"strconv"
)


type typeMetrics struct {
	Packets *prometheus.GaugeVec
	Bytes   *prometheus.GaugeVec
}

var (
	Test interface{}
	Metrics       = typeMetrics{
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