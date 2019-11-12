package main

import (
	"context"
	"flag"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/peer"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/multiformats/go-multiaddr"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	log.Println("DHT发现节点")

	//启发节点
	bootstrap := flag.String("bootstrap", "", "")
	flag.Parse()

	//创建上下文
	ctx := context.Background()

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
	log.Println("节点P2P地址:", p2pAddrs)

	//创建DHT
	kadDHT, e := dht.New(ctx, node)
	if e != nil {
		log.Fatalln(e)
	}
	//间隔显示DHT节点
	go func() {
		for {
			for _, v := range kadDHT.RoutingTable().ListPeers() {
				addr := kadDHT.FindLocal(v)
				log.Println("DHT节点:", addr)
			}

			log.Println("---")
			time.Sleep(time.Second * 6)
		}
	}()

	//如果设置了启发节点则连接
	if *bootstrap != "" {
		ma, _ := multiaddr.NewMultiaddr(*bootstrap)
		a, _ := peer.AddrInfoFromP2pAddr(ma)
		e := node.Connect(ctx, *a)
		if e != nil {
			log.Fatalln(e)
		}
		log.Println("已经连接节点:", *bootstrap)
	}

	// wait for a SIGINT or SIGTERM signal
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	log.Println("收到信号, 关闭...")

	e = node.Close()
	if e != nil {
		log.Fatalln(e)
	}
}
