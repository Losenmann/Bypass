package setup

import (
	"log/slog"
	"strconv"
)

var (
	compMemcache = "[MEMCACHED]"
	compDatabase = "[DATABASE]"
	compDefault = "[DEFAULT]"
)

func LoggingHendler(lvl int, msg interface{}, comp int) {
	switch v := msg.(type) {
	case int64:
		msg = strconv.FormatInt(v, 10)
	default:
	}
		component := compDefault
		switch comp {
		case 1:
			component = compDatabase
		case 2:
			component = compMemcache
		}
		
		switch lvl {
		case 1:
			if *LogLVL <= lvl {
				slog.Info(component + " " + msg.(string))
			}
		case 2:
			if *LogLVL <= lvl {
				slog.Warn(component + " " + msg.(string))
			}
		case 3:
			if *LogLVL <= lvl {
				slog.Error(component + " " + msg.(string))
			}
		}
}
