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
	"io"
	"io/ioutil"
	"log"
	"strings"
	"time"
)

const ProtocolId = "/p2p/mdns"

type discoveryNotifee struct {
	PeerChan chan peer.AddrInfo
}

// interface to be called when new  peer is found
// 注意:外部勿用!
func (n *discoveryNotifee) HandlePeerFound(pi peer.AddrInfo) {
	n.PeerChan <- pi
}

// 读取文本
// 注意:不能阻塞线程,直接用里面的东西则可以!
func ReadText(rw *bufio.ReadWriter) string {
	str, err := rw.ReadString('\n')
	if err != nil {
		if err != io.EOF {
			log.Println("读取文本出错:", err)
		} else {
			log.Println("读取文本完毕")
		}
		return ""
	}
	return strings.Replace(str, "\n", "", -1)
}

// 写入文本
func WriteText(rw *bufio.ReadWriter, text string) bool {
	_, err := rw.Write([]byte(fmt.Sprint(text, "\n")))
	if err != nil {
		log.Println("写入文本失败:", err)
		return false
	}
	err = rw.Flush()
	if err != nil {
		log.Println("压出文本失败:", err)
		return false
	}
	return true
}

// 读取文件
func ReadFile(rw *bufio.ReadWriter, path string) bool {
	fileBytes, err := rw.ReadBytes('\n')
	if err != nil {
		log.Println("读取文件字节出错:", err)
		return false
	}
	err = ioutil.WriteFile(path, fileBytes, 0644)
	if err != nil {
		log.Println("保存文件出错:", err)
		return false
	}
	return true
}

// 写入文件
func WriteFile(rw *bufio.ReadWriter, path string) bool {
	fileBytes, err := ioutil.ReadFile(path)
	if err != nil {
		log.Println("读取文件失败:", err)
		return false
	}
	_, err = rw.Write(fileBytes)
	if err != nil {
		log.Println("写入文件字节失败:", err)
		return false
	}
	err = rw.Flush()
	if err != nil {
		log.Println("压出文件字节失败:", err)
		return false
	}
	return true
}

// 初始化节点
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

// 连接节点
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
