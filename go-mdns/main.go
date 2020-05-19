package main

import (
	"github.com/alx696/go-mdns/im"
	"log"
)

func main() {
	log.Println("在内网用mDNS发现节点并通信")

	im.Init()
}
