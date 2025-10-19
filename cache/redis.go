package cache

import (
	"bufio"
	"bypass/database"
	"context"
	"flag"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"
)

type Flags_t struct {
	Enable bool
	ConnetionStr string
}

type Metrics_t struct {
	IPs 	prometheus.Gauge
	CIDRs   prometheus.Gauge
	Domains prometheus.Gauge
}

const (
	module = "REDIS CLIENT"
	key_expiry_c = "bypass:domains:expiry"
	key_ip4s_c = "bypass:resolved:ip4s"
	key_cidr4s_c = "bypass:resolved:cidr4s"
	key_ip6s_c = "bypass:resolved:ip6s"
	key_cidr6s_c = "bypass:resolved:cidr6s"
)

var (
	Args 		  Flags_t
	Ctx           = context.Background()
	Rdb           *redis.Client
	EvalTtlExpiry string
	Metrics       = Metrics_t{
		IPs: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "bypass_cached_ips",
				Help: "Statistics processed packages.",
			},
		),
		CIDRs: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "bypass_cached_cidrs",
				Help: "Statistics processed bytes.",
			},
		),
		Domains: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "bypass_cached_domains",
				Help: "Statistics processed bytes.",
			},
		),
	}
)

func init() {
	flag.BoolVar(&Args.Enable, "C", false, "Enable local redis cache")
	flag.StringVar(&Args.ConnetionStr, "r", "redis://localhost:6379/0", "Redis server connection string")
	prometheus.MustRegister(Metrics.IPs, Metrics.CIDRs, Metrics.Domains)
}

func Run() () {
	database.Run()
	if Args.Enable {
		go runCacheLocal()
		time.Sleep(3 * time.Second)
		connectCache()
		return
	} else {
		connectCache()
		return
	}
}

func runCacheLocal() () {
    if conf, err := os.Create("/etc/redis.conf"); err != nil {
        slog.Error(err.Error(), "tag", module)
		os.Exit(1)
    } else {
		defer conf.Close()
		_, err = conf.WriteString(redisConf_c)
		if err != nil {
			slog.Error(err.Error(), "tag", module)
			os.Exit(1)
		}
	}
    
	cmd := exec.Command("redis-server", "/etc/redis.conf")

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
					slog.Error(err.Error(), "tag", module)
				}
				break
			}
		//processLine(line)
		slog.Debug(line, "tag", module)
		}
	}(pipe)
	if err := cmd.Wait(); err != nil {
		// handle error
	}
}

func ConnectCheck() (error) {
	if pong, err := Rdb.Ping(Ctx).Result(); err != nil {
		return err
	} else {
		slog.Debug("Redis ping: "+pong, "tag", module)
		return nil
	}
}

func connectCache() () {
	if opt, err := redis.ParseURL(Args.ConnetionStr); err != nil {
		slog.Error(err.Error(), "tag", module)
		os.Exit(1)
	} else {
		Rdb = redis.NewClient(opt)
		go func() {
			for {
				if err := ConnectCheck(); err != nil {
					slog.Error(err.Error(), "tag", module)
					os.Exit(1)
				}
				time.Sleep(60 * time.Second)
			}
		}()
	}
}

// Создание кеша из базы данных
func CreateCache() () {
	var address string
	rows := database.Query(database.PGQuerySelectAll_c)
	defer rows.Close()
	for rows.Next() {
		if err := rows.Scan(&address); err != nil {
			slog.Warn(err.Error(), "tag", module)
			continue
		}
		if err := Rdb.ZAdd(Ctx, key_expiry_c, redis.Z{Member: address, Score: float64(0)}).Err(); err != nil {
			slog.Warn(err.Error(), "tag", module)
		} else {
			//slog.Debug("cached domain: "+address, "tag", module)
		}
	}
	slog.Info("cache created successfully", "tag", module)
}

func GetCacheExpiry() (*[]string) {
	if resp, err := Rdb.ZRange(Ctx, key_expiry_c, 0, time.Now().Unix()).Result(); err != nil {
		slog.Warn(err.Error(), "tag", module)
	} else {
		return &resp
	}
	return nil
}

func UpdateCacheExpiry(dns string, t uint32) () {
	ttl := float64(time.Now().Unix() + int64(t))
	if err := Rdb.ZAdd(Ctx, key_expiry_c, redis.Z{Member: dns, Score: ttl}).Err(); err != nil {
		slog.Warn(err.Error(), "tag", module)
	} else {
		slog.Debug("update ttl for the domain "+dns + ": TTL " + strconv.Itoa(int(t)), "tag", module)
	}
}

func AddIPs(ips []string, ipv6 bool) () {
	if ipv6 {
		if err := Rdb.SAdd(Ctx, key_ip6s_c, ips).Err(); err != nil {
			slog.Warn(err.Error(), "tag", module)
		} else {
			slog.Debug("cached ips: "+strings.Join(ips, " "), "tag", module)
		}
	} else {
		if err := Rdb.SAdd(Ctx, key_ip4s_c, ips).Err(); err != nil {
			slog.Warn(err.Error(), "tag", module)
		} else {
			slog.Debug("cached ips: "+strings.Join(ips, " "), "tag", module)
		}
	}
}
	

func RecreateCIDRs(cidrs []string) () {
	if err := Rdb.SAdd(Ctx, key_cidr4s_c, cidrs).Err(); err != nil {
		slog.Warn(err.Error(), "tag", module)
		return
	} else {
		slog.Debug("resteated cahche CIDR", "tag", module)
	}
}

func GetIPs() (*[]string, error) {
	if resp, err := Rdb.SMembers(Ctx, key_ip4s_c).Result(); err != nil {
		slog.Warn(err.Error(), "tag", module)
		return nil, err
	} else {
		slog.Debug("cache get ips: " + strconv.Itoa(len(resp)), "tag", module)
		return &resp, nil
	}
}

func GetCIDRs() (*[]string) {
	if resp, err := Rdb.SMembers(Ctx, key_ip4s_c).Result(); err != nil {
		slog.Warn(err.Error(), "tag", module)
		return nil
	} else {
		slog.Debug("cache get cidrs: " + strconv.Itoa(len(resp)), "tag", module)
		return &resp
	}
}





func CreateScripts() () {
	var err error
	if EvalTtlExpiry, err = Rdb.ScriptLoad(Ctx, evalRegenCIDRs).Result(); err != nil {
		slog.Warn(err.Error(), "tag", module)
	} else {
		slog.Info("cached eval TTL Expiry", "tag", module)
	}
}
/*
func UpdateDomainsTtl(dns string, t uint32) () {
	ttl := float64(time.Now().Unix() + int64(t))
	if err := Rdb.ZAdd(Ctx, "bypass:resolve:ttl", redis.Z{Member: dns, Score: ttl}).Err(); err != nil {
		slog.Warn(err.Error(), "tag", module)
	} else {
		slog.Debug("update ttl for the domain "+dns + ": TTL " + strconv.FormatUint(uint64(t), 10), "tag", module)
	}
}

func GetDomainsTtlExpiry() () {
	ttl := time.Now().Unix()
	if resp, err := Rdb.ZRange(Ctx, "bypass:resolve:ttl", 0, ttl).Result(); err != nil {
		slog.Warn(err.Error(), "tag", module)
	} else {
		for res := range resolver.Resolve(resp) {
			if res.Err == nil {
				if res.IPs != nil {
					AddIPs(res.IPs)
					UpdateDomainsTtl(res.Host, res.TTL)
				}
			}
		}
	}
}

func AddIPs(ips []string) () {
	if err := Rdb.SAdd(Ctx, "bypass:resolved:ips", ips).Err(); err != nil {
		slog.Warn(err.Error(), "tag", module)
	}
	slog.Debug("cached ips: "+strings.Join(ips, " "), "tag", module)
}

func RecreateCIDRs(cidrs []string) () {
	if err := Rdb.SAdd(Ctx, "bypass:resolved:cidrs", cidrs).Err(); err != nil {
		slog.Warn(err.Error(), "tag", module)
	}
	slog.Debug("resteated cahche CIDR", "tag", module)
	//health.Metrics.CIDRs.Set(float64(len(cidrs)))
}

func GetIPs() ([]string, error) {
	if resp, err := Rdb.SMembers(Ctx, key_ips_t).Result(); err != nil {
		slog.Warn(err.Error(), "tag", module)
		return nil, err
	} else {
		slog.Debug("cache get ips: " + strconv.Itoa(len(resp)), "tag", module)
		//health.Metrics.IPs.Set(float64(len(resp)))
		return resp, nil
	}
}

func GetCIDRs(def float64) (float64) {
	if resp, err := Rdb.SMembers(Ctx, key_cidrs_t).Result(); err != nil {
		slog.Warn(err.Error(), "tag", module)
		return def
	} else {
		slog.Debug("cache get cidrs: " + strconv.Itoa(len(resp)), "tag", module)
		return float64(len(resp))
	}
}

func CreateDomains() () {
	rows := database.Query(database.Querys.GetDomains)
	var address string
	defer rows.Close()
	for rows.Next() {
		if err := rows.Scan(&address); err != nil {
			slog.Warn(err.Error(), "tag", module)
			continue
		}
		UpdateDomainsTtl(address, 0)
		if err := Rdb.SAdd(Ctx, "bypass:resolve:domains", address).Err(); err != nil {
			slog.Warn(err.Error(), "tag", module)
		}
		slog.Debug("cached domain: "+address, "tag", module)
	}
}
*/
func GetMetrics() {
	if resp, err := Rdb.SCard(Ctx, "bypass:resolved:cidrs").Result(); err != nil {
		slog.Warn("failure get metrics CIDRs: "+err.Error(), "tag", module)
	} else {
		slog.Warn("get metrics CIDRs success", "tag", module)
		Metrics.CIDRs.Set(float64(resp))
	}
	if resp, err := Rdb.SCard(Ctx, "bypass:resolved:ips").Result(); err != nil {
		slog.Warn("failure get metrics IPs: "+err.Error(), "tag", module)
	} else {
		slog.Warn("get metrics IPs success", "tag", module)
		Metrics.IPs.Set(float64(resp))
	}
}