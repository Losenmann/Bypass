package cache

import (
	"bufio"
	"bypass/setup"
	"fmt"
	"io"
	"os/exec"
	"runtime"
	"strconv"
	"time"
	"github.com/bradfitz/gomemcache/memcache"
)

const (
	constMCTimeout = 500000000
	constComp      = 2
)

var (
	mc       = new(memcache.Client)
	mcRunArg = []string{"-vv", "-s", *setup.MemcacheSock, "-t", strconv.Itoa(runtime.NumCPU())}
)

func init() {
	go CacheRun(mcRunArg)
	time.Sleep(constMCTimeout)
	mc = memcache.New(*setup.MemcacheSock)
	for mc.Ping() != nil {
		time.Sleep(constMCTimeout)
	}
}

func CacheAdd(key string, value []byte) {
	mc.Set(&memcache.Item{Key: key, Value: value})

	it, err := mc.Get(key)
	if err != nil {
		fmt.Println(err)
	}
	setup.LoggingHendler(2, string(it.Value), constComp)
}

func CacheRun(arg []string) {
	cmd := exec.Command("/usr/bin/memcached", arg...)

	pipe, _ := cmd.StderrPipe()
	if err := cmd.Start(); err != nil {
		// handle error
	}
	go func(p io.ReadCloser) {
		reader := bufio.NewReader(pipe)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				if err.Error() != "EOF" {
					fmt.Printf("Error read stdout: %v\n", err)
				}
				break
			}
			processLine(line)
		}
	}(pipe)

	if err := cmd.Wait(); err != nil {
		// handle error
	}
}

func processLine(line string) {
	setup.LoggingHendler(1, line, constComp)
	//fmt.Print(line) // Здесь можно добавить любую обработку строки
}
