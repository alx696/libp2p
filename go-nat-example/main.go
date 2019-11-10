package main

import (
	"github.com/libp2p/go-nat"
	"log"
	"time"
)

func main() {
	log.Println("寻找NAT")
	natGateway, e := nat.DiscoverGateway()
	if e != nil {
		log.Fatalln("寻找NAT错误:", e)
	}

	for {
		natExternalAddress, e := natGateway.GetExternalAddress()
		if e != nil {
			log.Fatalln("获取NAT网关公网地址错误:", e)
		}
		natInternalAddress, e := natGateway.GetInternalAddress()
		if e != nil {
			log.Fatalln("获取节点内网地址错误:", e)
		}
		log.Println("NAT网关公网地址:", natExternalAddress.String(), "节点内网地址:", natInternalAddress)

		time.Sleep(time.Second)
	}
}
