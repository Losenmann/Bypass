package resolver

import (
	"bypass/cache"
	"flag"
	"log/slog"
	"strings"
	"sync"
	"time"
	"github.com/miekg/dns"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	module = "RESOLVER"
)

type Flags_t struct {
	Enable  	 	bool
}

type Metrics_t struct {
	LastResolve	prometheus.Gauge
}

type ResolveOptions struct {
	Timeout   		time.Duration
	IPv4Only  		bool
	IPv6Only  		bool
	MaxWorker 		int
	DNSServer 		string
}

type Resolved struct {
	Host string
	IPv4s  []string
	IPv6s  []string
	TTL  uint32
	Err  error
}

var (
	Args Flags_t
	opts ResolveOptions
	Metrics       = Metrics_t{
		LastResolve: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "bypass_resolver_last_resolve",
				Help: "Statistics processed packages.",
			},
		),
	}
)


func init() {
	flag.BoolVar(&Args.Enable, "R", false, "Enable resolver")
	prometheus.MustRegister(Metrics.LastResolve)
	opts = ResolveOptions{
		Timeout:   2 * time.Second,
		IPv4Only:  false,
		IPv6Only:  false,
		MaxWorker: 500,
		DNSServer: "8.8.8.8:53",
	}
}

func Run() {
	cache.Run()
	cache.CreateCache()
	
	for res := range ResolveLargeDNS(*cache.GetCacheExpiry(), opts) {
		if res.Err == nil {
			if res.IPv4s != nil {
				go cache.AddIPs(res.IPv4s, false)
			}
			if res.IPv6s != nil {
				go cache.AddIPs(res.IPv6s, true)
			}
			cache.UpdateCacheExpiry(res.Host, res.TTL)
		}
	}
	if resp, err := cache.GetIPs(); err != nil {
		slog.Warn(err.Error(), "tag", module)
	} else {
		cache.RecreateCIDRs(GenCIDR(resp))
	}
}

// queryDNS делает DNS-запрос типа A или AAAA через miekg/dns
func queryDNS(host string, opts ResolveOptions) ([]string, []string, uint32, error) {
	client := new(dns.Client)
	client.Timeout = opts.Timeout

	msg := new(dns.Msg)
	if opts.IPv4Only {
		msg.SetQuestion(dns.Fqdn(host), dns.TypeA)
	} else if opts.IPv6Only {
		msg.SetQuestion(dns.Fqdn(host), dns.TypeAAAA)
	} else {
		msg.SetQuestion(dns.Fqdn(host), dns.TypeA)
	}

	r, _, err := client.Exchange(msg, opts.DNSServer)
	if err != nil {
		return nil, nil, 0, err
	}

	var results_ipv4, results_ipv6 []string
	ttl := uint32(0)
	for _, ans := range r.Answer {
		switch t := ans.(type) {
		case *dns.A:
			if !opts.IPv6Only {
				results_ipv4 = append(results_ipv4, t.A.String())
				if t.Hdr.Ttl > ttl {
					ttl = t.Hdr.Ttl
				}
			}
		case *dns.AAAA:
			if !opts.IPv4Only {
				results_ipv6 = append(results_ipv6, t.AAAA.String())
				if t.Hdr.Ttl > ttl {
					ttl = t.Hdr.Ttl
				}
			}
		}
	}
	return results_ipv4, results_ipv6, ttl, nil
}

func worker(jobs <-chan string, results chan<- Resolved, opts ResolveOptions, wg *sync.WaitGroup) {
	defer wg.Done()
	for host := range jobs {
		ipv4s, ipv6s, ttl, err := queryDNS(host, opts)
		if err != nil {
			slog.Warn("nslookup " + host + ": " + err.Error(), "tag", module)
		} else {
			slog.Debug("resolved " + host + ": "+ strings.Join(ipv4s, ","), "tag", module)
			slog.Debug("resolved " + host + ": "+ strings.Join(ipv6s, ","), "tag", module)
		}
		results <- Resolved{Host: host, IPv4s: ipv4s, IPv6s: ipv6s, TTL: ttl, Err: err}
	}
}

func ResolveLargeDNS(hosts []string, opts ResolveOptions) <-chan Resolved {
	jobs := make(chan string, opts.MaxWorker)
	results := make(chan Resolved, opts.MaxWorker)

	var wg sync.WaitGroup
	for i := 0; i < opts.MaxWorker; i++ {
		wg.Add(1)
		go worker(jobs, results, opts, &wg)
	}

	go func() {
		for _, host := range hosts {
			jobs <- host
		}
		close(jobs)
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	return results
}

func Resolve(hosts []string) (<-chan Resolved) {
	opts := ResolveOptions{
		Timeout:   2 * time.Second,
		IPv4Only:  false,
		IPv6Only:  false,
		MaxWorker: 500,
		DNSServer: "8.8.8.8:53",
	}

	return ResolveLargeDNS(hosts, opts)
}
