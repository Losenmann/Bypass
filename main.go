package main

import (
	"bypass/gateway"
)

func main() {
	go gateway.Gateway()

	select {}	
}