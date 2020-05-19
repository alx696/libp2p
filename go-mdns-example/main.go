package main

import (
	"context"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/peer"
	"log"
)

// mDNS接待的结构
type mDnsNotifee struct {
	PeerChan chan peer.AddrInfo
}

// mDNS接待的方法
// 发现节点时会调用此方法(测试发现会被执行两次?)
func (n *mDnsNotifee) HandlePeerFound(i peer.AddrInfo) {
	n.PeerChan <- i
}

func main() {
	log.Println("内网用mDNS发现节点")

	//创建上下文
	ctx := context.Background()

	//创建节点
	node, e := libp2p.New(ctx)
	if e != nil {
		log.Fatalln(e)
	}
	log.Println(node.ID().String())

	//创建mDNS
	//interval最小时间小于5秒时强制变成5秒
	//mdnsService, e := discovery.NewMdnsService(ctx, node, time.Second*6, "p2p-mdns")
	//if e != nil {
	//	log.Fatalln(e)
	//}
	//log.Println(mdnsService.Close())
	////监听mDNS
	//mdnsNotifee := &mDnsNotifee{}
	//mdnsNotifee.PeerChan = make(chan peer.AddrInfo)
	//mdnsService.RegisterNotifee(mdnsNotifee)
	//for {
	//	select {
	//	case p := <-mdnsNotifee.PeerChan:
	//		log.Println("发现节点:", p)
	//	}
	//}
}
