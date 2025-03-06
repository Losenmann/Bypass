package main

import (
	"bypass/cache"
	"bypass/database"
	"fmt"
)



func main() {
	database.Check()
	cache.CacheAdd("test1", []byte("my value1"))
	cache.CacheAdd("test2", []byte("my value2"))
	cache.CacheAdd("test1", []byte("my value2"))
	fmt.Println(database.Select())
}
