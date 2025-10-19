package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"bypass/resolver"
	"bypass/firewall"
	"bypass/tools"
)

var (
	LogLVL = new(uint)
)

func getEnv(key string, fallback any) any {
	value, exists := os.LookupEnv(key)
	if !exists {
		return fallback
	}
	return value
}

func init() {
	LogLVL = flag.Uint("l", getEnv("BYPASS_LOG_LVL", uint(1)).(uint), "Log lvl")
	flag.Parse()
	switch *LogLVL {
	case 1:
		slog.SetDefault(slog.New(&CustomHandler{level: slog.LevelDebug}))
	case 2:
		slog.SetDefault(slog.New(&CustomHandler{level: slog.LevelInfo}))
	case 3:
		slog.SetDefault(slog.New(&CustomHandler{level: slog.LevelWarn}))
	case 4:
		slog.SetDefault(slog.New(&CustomHandler{level: slog.LevelError}))
	default:
		slog.SetDefault(slog.New(&CustomHandler{level: slog.LevelInfo}))
	}
}

func main() {
	if resolver.Args.Enable {
		resolver.Run()
	}
	if firewall.Args.Enable {
		firewall.Run()
	}
	if tools.Args.MetricsEnable {
		tools.RunMetrics()
	}
	select {}	
}

type CustomHandler struct {
	level slog.Level
}

func (h *CustomHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *CustomHandler) Handle(_ context.Context, record slog.Record) error {
	timestamp := record.Time.Format("2006/01/02 15:04:05")

	// Извлекаем тег (если есть)
	tag := ""
	record.Attrs(func(a slog.Attr) bool {
		if a.Key == "tag" {
			tag = fmt.Sprintf("[%s]", a.Value.String())
		}
		return true
	})

	// Определяем цвет и поток вывода
	var (
		color  string
		output *os.File
	)

	switch {
	case record.Level == slog.LevelDebug:
		color, output = colorPurple, os.Stdout
	case record.Level == slog.LevelInfo:
		color, output = colorBlue, os.Stdout
	case record.Level == slog.LevelWarn:
		color, output = colorYellow, os.Stderr
	case record.Level == slog.LevelError:
		color, output = colorRed, os.Stderr
	}

	// Форматированный вывод
	fmt.Fprintf(output, "%s %s%s%s %s %s\n",
		timestamp,
		color, record.Level.String(), colorReset,
		tag,
		record.Message,
	)

	return nil
}

func (h *CustomHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

func (h *CustomHandler) WithGroup(name string) slog.Handler {
	return h
}

const (
	colorReset  = "\033[0m"
	colorBlue   = "\033[1;34m"
	colorPurple = "\033[1;35m"
	colorYellow = "\033[1;33m"
	colorRed    = "\033[1;31m"
)