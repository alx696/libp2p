package main

import (
	"context"
	"crypto/rand"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/peer"
	libp2pquic "github.com/libp2p/go-libp2p-quic-transport"
	"github.com/libp2p/go-libp2p/p2p/protocol/ping"
	"github.com/multiformats/go-multiaddr"
	"log"
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
		libp2p.ListenAddrStrings("/ip4/127.0.0.1/udp/0/quic"),
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

	//连接到指定地址并发送ping
	ma, e := multiaddr.NewMultiaddr("/ip4/127.0.0.1/udp/60000/quic/ipfs/QmQk4ej59ncF23q4DY9h2r841y973fB69QHknVpFPc5F7j")
	if e != nil {
		log.Fatalln(e)
	}
	addr, e := peer.AddrInfoFromP2pAddr(ma)
	if e != nil {
		log.Fatalln(e)
	}
	e = node.Connect(ctx, *addr)
	if e != nil {
		log.Fatalln(e)
	}
	ch := pingService.Ping(ctx, addr.ID)
	res := <- ch
	log.Println("Pint RTT:", res.RTT)

	e = node.Close()
	if e != nil {
		log.Fatalln(e)
	}
}
