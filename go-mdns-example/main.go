package main

import (
	"context"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p/p2p/discovery"
	"log"
	"time"
)

type discoveryNotifee struct {
	PeerChan chan peer.AddrInfo
}

//interface to be called when new  peer is found
func (n *discoveryNotifee) HandlePeerFound(pi peer.AddrInfo) {
	n.PeerChan <- pi
}

func main() {
	log.Println("在内网用mDNS发现节点")

	//创建上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//创建节点
	node, e := libp2p.New(ctx)
	if e != nil {
		log.Fatalln(e)
	}
	//节点地址转为P2P地址
	p2pAddrs, e := peer.AddrInfoToP2pAddrs(&peer.AddrInfo{node.ID(), node.Addrs()})
	if e != nil {
		log.Fatalln(e)
	}
	log.Println("我的节点地址:", p2pAddrs)

	// 创建服务
	// interval最小时间小于5秒时强制变成5秒
	ser, err := discovery.NewMdnsService(ctx, node, time.Second*6, "mdns-test")
	if err != nil {
		panic(err)
	}
	// register with service so that we get notified about peer discovery
	n := &discoveryNotifee{}
	n.PeerChan = make(chan peer.AddrInfo)
	ser.RegisterNotifee(n)

	//等待并显示节点
	for {
		select {
		case p := <-n.PeerChan:
			// 测试发现结果每次会是2条?
			// 节点地址转为P2P地址
			p2pAddrs, e := peer.AddrInfoToP2pAddrs(&p)
			if e != nil {
				log.Fatalln(e)
			}
			log.Println("发现节点地址:", p2pAddrs)
		}
	}
}
