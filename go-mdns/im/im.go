package im

import (
	"bufio"
	"context"
	"fmt"
	"github.com/libp2p/go-libp2p"
	host2 "github.com/libp2p/go-libp2p-core/host"
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

	// 读取
	str, err := rw.ReadString('\n')
	str = strings.Replace(str, "\n", "", -1)
	if err != nil {
		log.Println("读取字符出错:", err)
		return
	}
	log.Println("读取字符:", str)

	// 回复2次
	for i := 0; i < 2; i++ {
		_, err := rw.Write([]byte(fmt.Sprint(time.Now().String(), "\n")))
		if err != nil {
			log.Println("回复失败:", err)
			return
		}
		err = rw.Flush()
		if err != nil {
			log.Println("压出失败:", err)
			return
		}
	}

	_ = s.Close()
}

func Init(handler func(s network.Stream),
	cb func(ctx context.Context, host host2.Host, p peer.AddrInfo) error) error {
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
	host.SetStreamHandler(ProtocolId, handler)

	// 创建mDNS服务并注册发现列队
	ser, err := discovery.NewMdnsService(ctx, host, time.Hour, "mdns-test")
	if err != nil {
		return err
	}
	n := &discoveryNotifee{}
	n.PeerChan = make(chan peer.AddrInfo)
	ser.RegisterNotifee(n)

	// 发现节点
	for p := range n.PeerChan {
		err := cb(ctx, host, p)
		if err != nil {
			log.Println("处理发现节点出错:", err)
		}
	}

	return nil
}

func Connect(ctx context.Context, host host2.Host, p peer.AddrInfo) (*bufio.ReadWriter, error) {
	log.Println("连接节点:", p.ID.String())

	// 连接并建流
	err := host.Connect(ctx, p)
	if err != nil {
		return nil, err
	}
	s, err := host.NewStream(ctx, p.ID, ProtocolId)
	if err != nil {
		return nil, err
	}

	// 通信
	rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))
	return rw, nil
}
