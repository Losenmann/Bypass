package tools

import "flag"

type Flags_t struct {
	MetricsEnable bool
}

var (
	Args Flags_t
)

func init() {
	flag.BoolVar(&Args.MetricsEnable, "M", false, "Enable metrics")
}