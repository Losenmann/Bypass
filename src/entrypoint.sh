#!/bin/bash
awg-quick up wg0
sleep 3
nft -f /etc/nftables.d/bypass.nft
redis-server /etc/redis/redis.conf
awg-exporter
