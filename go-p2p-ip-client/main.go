package main

import (
	"bufio"
	"context"
	"crypto/rand"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/peer"
	libp2pquic "github.com/libp2p/go-libp2p-quic-transport"
	"github.com/multiformats/go-multiaddr"
	"log"
	"strings"
	"time"
)

const (
	protocolID = "/p2p/ip"
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
	)
	if e != nil {
		log.Fatalln(e)
	}
	log.Println("节点地址:", node.Addrs())

	//打印节点P2P地址
	p2pAddrs, e := peer.AddrInfoToP2pAddrs(&peer.AddrInfo{node.ID(), node.Addrs()})
	if e != nil {
		log.Fatalln(e)
	}
	log.Println("节点P2P地址:", p2pAddrs)

	//服务地址
	ma, e := multiaddr.NewMultiaddr("/ip4/127.0.0.1/udp/60000/quic/ipfs/QmV3W8GtRQ8JCv9JBTFiSjpufm4US94edAHDqKe1cdY3ye")
	if e != nil {
		log.Fatalln(e)
	}
	addr, e := peer.AddrInfoFromP2pAddr(ma)
	if e != nil {
		log.Fatalln(e)
	}

	for {
		//连接
		e = node.Connect(ctx, *addr)
		if e != nil {
			log.Println("连接错误:", e)
			//30秒后重试
			time.Sleep(time.Second * 30)
			continue
		}

		//建流
		s, e := node.NewStream(ctx, addr.ID, protocolID)
		if e != nil {
			log.Println("建流错误:", e)
			//30秒后重试
			time.Sleep(time.Second * 30)
			continue
		}
		_, e = s.Write([]byte("公网IP\n"))
		if e != nil {
			log.Println("写入错误:", e)
			//30秒后重试
			time.Sleep(time.Second * 30)
			continue
		}
		log.Println("请求已经发送")
		reader := bufio.NewReader(s)
		for {
			txt, e := reader.ReadString('\n')
			txt = strings.Replace(txt, "\n", "", -1)
			if e != nil {
				log.Println("读取错误:", e)
				break
			} else {
				log.Println("读取内容:", txt)
			}
		}

		//30秒后重试
		time.Sleep(time.Second * 30)
	}
}
