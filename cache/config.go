package cache

const redisConf_c = `
bind 0.0.0.0
protected-mode no
port 6379
tcp-backlog 511
timeout 0
tcp-keepalive 300
daemonize no
pidfile /run/redis.pid
loglevel warning
databases 16
always-show-logo no
save 3600 1
stop-writes-on-bgsave-error no
rdbcompression yes
rdbchecksum yes
dir /tmp/redis
`