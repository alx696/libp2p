package im

import (
	"bufio"
	"context"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p/p2p/discovery"
	"log"
	"strings"
	"time"
)

const ProtocolId = "/p2p/mdns"

type discoveryNotifee struct {
	PeerChan chan peer.AddrInfo
}

//interface to be called when new  peer is found
func (n *discoveryNotifee) HandlePeerFound(pi peer.AddrInfo) {
	n.PeerChan <- pi
}

func readData(rw *bufio.ReadWriter) {
	for {
		str, err := rw.ReadString('\n')
		str = strings.Replace(str, "\n", "", -1)
		if err != nil {
			log.Println("读取字符出错:", err)
			return
		}
		log.Println("读取字符:", str)
	}
}

func writeData(rw *bufio.ReadWriter) {
	for {
		_, err := rw.Write([]byte("Hi\n"))
		if err != nil {
			log.Println("回复失败:", err)
			return
		}
		err = rw.Flush()
		if err != nil {
			log.Println("压出失败:", err)
			return
		}

		time.Sleep(time.Second * 6)
	}
}

func handleStream(s network.Stream) {
	log.Println("处理新流")

	rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))
	// 持续读取
	go readData(rw)
	// 持续回复
	go writeData(rw)
}

func Init() error {
	log.Println("初始化")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	host, err := libp2p.New(ctx)
	if err != nil {
		return err
	}
	addrs, err := peer.AddrInfoToP2pAddrs(&peer.AddrInfo{host.ID(), host.Addrs()})
	if err != nil {
		return err
	}
	log.Println("我的地址:", addrs)

	// 别人向你创建创建流时进行处理
	host.SetStreamHandler(ProtocolId, handleStream)

	// 创建mDNS服务并注册发现列队
	ser, err := discovery.NewMdnsService(ctx, host, time.Hour, "mdns-test")
	if err != nil {
		return err
	}
	n := &discoveryNotifee{}
	n.PeerChan = make(chan peer.AddrInfo)
	ser.RegisterNotifee(n)

	// 发现节点
	p := <-n.PeerChan
	discoveryAddrs, err := peer.AddrInfoToP2pAddrs(&peer.AddrInfo{p.ID, p.Addrs})
	if err != nil {
		return err
	}
	log.Println("发现地址:", discoveryAddrs)

	// 连接并建流
	err = host.Connect(ctx, p)
	if err != nil {
		log.Println("连接失败:", err)
		return err
	}
	log.Println("已经连接")
	s, err := host.NewStream(ctx, p.ID, ProtocolId)
	if err != nil {
		log.Println("建流失败:", err)
		return err
	}
	log.Println("已经建流")

	// 通信
	rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))
	// 持续读取
	go readData(rw)
	// 持续回复
	go writeData(rw)

	select {}
}
