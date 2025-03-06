package setup

import (
	"flag"
	"os"
	"fmt"
)

var (
	LogLVL       = new(int)
	MemcacheSock = new(string)
	DatabasePath = new(string)
)

func init() {
	fmt.Println(logo)
	LogLVL = flag.Int("l", getEnv("BYPASS_LOG_LVL", 1).(int), "Log lvl")
	MemcacheSock = flag.String("s", getEnv("BYPASS_PATH_MEMSOCK", "./memcached.sock").(string), "Path to memcached socket file")
	DatabasePath = flag.String("d", getEnv("BYPASS_PATH_DB", "./database.db").(string), "Path to database file")
	flag.Parse()
}

func getEnv(key string, fallback interface{}) interface{} {
	value, exists := os.LookupEnv(key)
	if !exists {
		return fallback
	}
	return value
}
