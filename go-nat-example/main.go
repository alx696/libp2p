package main

import (
	"context"
	"flag"
	libp2pnat "github.com/libp2p/go-libp2p-nat"
	gonat "github.com/libp2p/go-nat"
	"log"
	"time"
)

func debugNat(natGateway gonat.NAT, port int) {
	log.Println("NAT网关类型:", natGateway.Type())

	natDeviceAddress, e := natGateway.GetDeviceAddress()
	if e != nil {
		log.Fatalln(e)
	}
	log.Println("NAT网关IP:", natDeviceAddress.String())

	natExternalAddress, e := natGateway.GetExternalAddress()
	if e != nil {
		log.Fatalln(e)
	}
	log.Println("NAT网关公网IP:", natExternalAddress.String())

	//进行端口映射
	mappedExternalPort, e := natGateway.AddPortMapping("udp", port, "P2P测试", time.Minute)
	if e != nil {
		log.Fatalln(e)
	}
	log.Println("NAT映射端口:", port, mappedExternalPort)

	//移除端口映射
	_ = natGateway.DeletePortMapping("udp", port)
}

// 目前有go-nat和go-libp2p-nat两个库可用,
// go-nat.DiscoverNATs(ctx)的网络环境适应性较好.
func main() {
	log.Println("NAT示例")

	natType := flag.String("type", "", "")
	flag.Parse()

	//内部端口
	port := 60000

	//创建上下文
	ctx := context.Background()

	if *natType == "libp2p" {
		//go-libp2p-nat库
		//H3C路由需要2分钟以上, 家里的路由5秒.
		log.Println("使用go-libp2p-nat")
		mNAT, e := libp2pnat.DiscoverNAT(ctx)
		if e != nil {
			log.Fatalln(e)
		}
		log.Println("已有NAT映射:", mNAT.Mappings())
		natMapping, e := mNAT.NewMapping("udp", port)
		if e != nil {
			log.Fatalln(e)
		}
		externalAddr, e := natMapping.ExternalAddr()
		if e != nil {
			log.Fatalln(e)
		}
		// 地址形式为IP加端口, 如 8.8.8.8:38288
		log.Println("NAT公网地址:", externalAddr.String())

		//移除端口映射
		_ = natMapping.Close()
	} else {
		//go-nat库
		if *natType == "gateway" {
			log.Println("使用go-nat DiscoverGateway()")
			//H3C路由需要2分钟以上, 家里的路由5秒.
			natGateway, e := gonat.DiscoverGateway()
			if e != nil {
				log.Fatalln(e)
			}
			log.Println("找到NAT网关:", natGateway)
			debugNat(natGateway, port)
		} else {
			log.Println("使用go-nat DiscoverNATs(ctx)")
			//企业路由器2秒, 家里路由器马上
			natChan := gonat.DiscoverNATs(ctx)
			select {
			case natGateway := <-natChan:
				if natGateway == nil {
					log.Println("没有找到网关")
					return
				}

				log.Println("找到NAT网关:", natGateway)
				debugNat(natGateway, port)
			}
		}
	}
}
