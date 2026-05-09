package firewall

import (
)

func Firewall() {
	go RunIPRoute()
	RunNFTables()
}