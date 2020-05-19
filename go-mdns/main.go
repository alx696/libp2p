package main

import (
	"context"
	"github.com/alx696/go-mdns/im"
	host2 "github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"log"
)

var ps map[string]peer.AddrInfo

func main() {
	log.Println("在内网用mDNS发现节点并通信")

	ps = make(map[string]peer.AddrInfo)

	im.Init(func(ctx context.Context, host host2.Host, p peer.AddrInfo) error {

		if _, ok := ps[p.ID.String()]; !ok {
			log.Println("发现节点:", p.ID.String())
			ps[p.ID.String()] = p

			im.Send(ctx, host, p, "你好")
		}

		return nil
	})
}
