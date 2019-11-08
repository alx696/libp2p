package main

import (
	"bufio"
	"context"
	"crypto/rand"
	"flag"
	"fmt"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/protocol"
	"github.com/libp2p/go-libp2p/p2p/discovery"
	"github.com/multiformats/go-multiaddr"
	"log"
	"os"
	"strings"
	"time"
)

func readData(rw *bufio.ReadWriter) {
	for {
		str, err := rw.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading from buffer")
			panic(err)
		}

		if str == "" {
			return
		}
		if str != "\n" {
			// Green console colour: 	\x1b[32m
			// Reset console colour: 	\x1b[0m
			fmt.Printf("\x1b[32m%s\x1b[0m> ", str)
		}

	}
}

func writeData(rw *bufio.ReadWriter) {
	stdReader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("> ")
		sendData, err := stdReader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading from stdin")
			panic(err)
		}

		_, err = rw.WriteString(fmt.Sprintf("%s\n", sendData))
		if err != nil {
			fmt.Println("Error writing to buffer")
			panic(err)
		}
		err = rw.Flush()
		if err != nil {
			fmt.Println("Error flushing buffer")
			panic(err)
		}
	}
}

func handleStream(stream network.Stream) {
	fmt.Println("Got a new stream!")

	// Create a buffer stream for non blocking read and write.
	rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))

	go readData(rw)
	go writeData(rw)

	// 'stream' will stay open until you close it (or the other side closes it).
}

//接口需要的结构
type discoveryNotifee struct {
	PeerChan chan peer.AddrInfo
}

//interface to be called when new  peer is found
func (n *discoveryNotifee) HandlePeerFound(pi peer.AddrInfo) {
	n.PeerChan <- pi
}

// 参考https://github.com/libp2p/go-libp2p-examples/tree/master/chat-with-mdns
func main() {
	port := *flag.String("port", "60000", "listen port")
	flag.Parse()
	log.Println("P2P:", port)

	protocolID := protocol.ID("/p2p/_testing")
	ctx := context.Background()

	//生成密钥
	rr := rand.Reader
	prKey, _, e := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, rr)
	if e != nil {
		log.Fatalln(e)
	}

	//创建多址
	ma, _ := multiaddr.NewMultiaddr(strings.Join([]string{"/ip4/0.0.0.0/tcp/", port}, ""))
	host, e := libp2p.New(ctx, libp2p.ListenAddrs(ma), libp2p.Identity(prKey))
	if e != nil {
		log.Fatalln(e)
	}

	//设置流处
	host.SetStreamHandler(protocolID, handleStream)

	//创建mDNS服务并注册监听
	s, e := discovery.NewMdnsService(ctx, host, time.Second * 3, "juliao-mdns")
	if e != nil {
		log.Fatalln(e)
	}
	n := &discoveryNotifee{}
	n.PeerChan = make(chan peer.AddrInfo)
	s.RegisterNotifee(n)

	for {
		select {
		case p := <-n.PeerChan:
			log.Println("发现节点:", p)
		}
	}

	//p := <- n.PeerChan
	//log.Println("发现节点:", p)

	////连接节点
	//e = host.Connect(ctx, p)
	//if e != nil {
	//	log.Fatalln(e)
	//}
	//
	////创建流
	//stream, e := host.NewStream(ctx, p.ID, protocolID)
	//if e != nil {
	//	log.Fatalln(e)
	//}
	//handleStream(stream)
	//log.Println("完成")
	//
	//select {}
}
