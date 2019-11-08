package main

import (
	"context"
	"crypto/rand"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p/p2p/protocol/ping"
	"github.com/libp2p/go-libp2p-quic-transport"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx := context.Background()

	//生成密钥
	rr := rand.Reader
	prKey, _, e := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, rr)
	if e != nil {
		log.Fatalln(e)
	}

	quicTransport, e := libp2pquic.NewTransport(prKey)
	if e != nil {
		log.Fatalln(e)
	}

	node, e := libp2p.New(ctx,
		libp2p.Transport(quicTransport),
		libp2p.Identity(prKey),
		libp2p.ListenAddrStrings("/ip4/127.0.0.1/udp/60000/quic"),
		libp2p.Ping(false),
	)
	if e != nil {
		log.Fatalln(e)
	}
	log.Println("节点地址:", node.Addrs())

	//定义ping协议
	pingService := &ping.PingService{Host: node}
	node.SetStreamHandler(ping.ID, pingService.PingHandler)

	//打印节点P2P地址
	p2pAddrs, e := peer.AddrInfoToP2pAddrs(&peer.AddrInfo{node.ID(), node.Addrs()})
	if e != nil {
		log.Fatalln(e)
	}
	log.Println("节点P2P地址:", p2pAddrs)

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
