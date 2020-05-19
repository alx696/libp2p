package main

import (
	"bufio"
	"context"
	"fmt"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p/p2p/discovery"
	"log"
	"os"
	"time"
)

type discoveryNotifee struct {
	PeerChan chan peer.AddrInfo
}

//interface to be called when new  peer is found
func (n *discoveryNotifee) HandlePeerFound(pi peer.AddrInfo) {
	n.PeerChan <- pi
}

func handleStream(stream network.Stream) {
	log.Println("处理新的流!")

	// Create a buffer stream for non blocking read and write.
	rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))

	go readData(rw)
	go writeData(rw)

	// 'stream' will stay open until you close it (or the other side closes it).
}

func readData(rw *bufio.ReadWriter) {
	for {
		str, err := rw.ReadString('\n')
		if err != nil {
			log.Println("Error reading from buffer")
			log.Fatalln(err)
		}

		if str == "" {
			return
		}
		if str != "\n" {
			// Green console colour: 	\x1b[32m
			// Reset console colour: 	\x1b[0m
			log.Printf("\x1b[32m%s\x1b[0m> ", str)
		}

	}
}

func writeData(rw *bufio.ReadWriter) {
	stdReader := bufio.NewReader(os.Stdin)

	for {
		log.Print("> ")
		sendData, err := stdReader.ReadString('\n')
		if err != nil {
			log.Println("Error reading from stdin")
			log.Fatalln(err)
		}

		_, err = rw.WriteString(fmt.Sprintf("%s\n", sendData))
		if err != nil {
			log.Println("Error writing to buffer")
			log.Fatalln(err)
		}
		err = rw.Flush()
		if err != nil {
			log.Println("Error flushing buffer")
			log.Fatalln(err)
		}
	}
}

func main() {
	log.Println("在内网用mDNS发现节点并通信")

	//创建上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 创建节点
	node, e := libp2p.New(ctx)
	if e != nil {
		log.Fatalln(e)
	}
	// 节点地址转为P2P地址
	p2pAddrs, e := peer.AddrInfoToP2pAddrs(&peer.AddrInfo{node.ID(), node.Addrs()})
	if e != nil {
		log.Fatalln(e)
	}
	log.Println("我的节点地址:", p2pAddrs)

	// Set a function as stream handler.
	// This function is called when a peer initiates a connection and starts a stream with this peer.
	node.SetStreamHandler("/p2p/mdns", handleStream)

	// 创建服务
	// interval最小时间小于5秒时强制变成5秒
	ser, e := discovery.NewMdnsService(ctx, node, time.Hour, "mdns-psp")
	if e != nil {
		log.Fatalln(e)
	}
	// register with service so that we get notified about peer discovery
	n := &discoveryNotifee{}
	n.PeerChan = make(chan peer.AddrInfo)
	ser.RegisterNotifee(n)

	// 等待节点连接
	p := <-n.PeerChan
	log.Println("发现节点:", p.ID.String())

	// 通信
	if e := node.Connect(ctx, p); e != nil {
		log.Println("连接失败:", e)
	}

	// open a stream, this stream will be handled by handleStream other end
	stream, e := node.NewStream(ctx, p.ID, "/p2p/mdns")

	if e != nil {
		log.Println("创建流失败:", e)
	} else {
		rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))

		go writeData(rw)
		go readData(rw)
		log.Println("已经连接:", p)
	}

	select {} //wait here
}
