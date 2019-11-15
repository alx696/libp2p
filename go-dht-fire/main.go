package main

import (
	"context"
	"crypto/rand"
	"flag"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/routing"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	mplex "github.com/libp2p/go-libp2p-mplex"
	libp2pquic "github.com/libp2p/go-libp2p-quic-transport"
	"github.com/libp2p/go-libp2p-secio"
	yamux "github.com/libp2p/go-libp2p-yamux"
	"github.com/multiformats/go-multiaddr"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

const (
	CONFIG_DIR            = "./config"
	RSA_FILE_PATH_PRIVATE = "./config/rsa-private"
	RSA_FILE_PATH_PUBLIC  = "./config/rsa-public"
)

func rsaKey() (prKey crypto.PrivKey, puKey crypto.PubKey) {
	_, e := os.Stat(CONFIG_DIR)
	if os.IsNotExist(e) {
		_ = os.Mkdir(CONFIG_DIR, 0755)

		//生成密钥
		rr := rand.Reader
		prKey, puKey, _ = crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, rr)

		//存储密钥
		privateKeyBytes, _ := crypto.MarshalPrivateKey(prKey)
		_ = ioutil.WriteFile(RSA_FILE_PATH_PRIVATE, privateKeyBytes, 0644)
		publicKeyBytes, _ := crypto.MarshalPublicKey(puKey)
		_ = ioutil.WriteFile(RSA_FILE_PATH_PUBLIC, publicKeyBytes, 0644)
	} else {
		//还原密钥
		privateKeyBytes, _ := ioutil.ReadFile(RSA_FILE_PATH_PRIVATE)
		publicKeyBytes, _ := ioutil.ReadFile(RSA_FILE_PATH_PUBLIC)
		prKey, _ = crypto.UnmarshalPrivateKey(privateKeyBytes)
		puKey, _ = crypto.UnmarshalPublicKey(publicKeyBytes)
	}

	return
}

// 参考 https://github.com/libp2p/go-libp2p-examples/blob/b7ac9e91865656b3ec13d18987a09779adad49dc/ipfs-camp-2019/06-Pubsub/main.go
func main() {
	log.Println("DHT星星之火")

	//指定端口,否则随机
	port := flag.String("port", "0", "")
	//启发节点
	//必须是P2P地址, 即 https://github.com/multiformats/multiaddr#protocols (含/ipfs/Qm...)
	bootstrap := flag.String("bootstrap", "", "")
	flag.Parse()

	//生成密钥
	prKey, _ := rsaKey()

	//创建传输层
	quicTransport, e := libp2pquic.NewTransport(prKey)
	if e != nil {
		log.Fatalln(e)
	}

	//创建上下文
	ctx := context.Background()

	//DHT定义
	var kadDHT *dht.IpfsDHT
	newDHT := func(h host.Host) (routing.PeerRouting, error) {
		var err error
		kadDHT, err = dht.New(ctx, h)
		return kadDHT, err
	}

	//创建节点
	node, e := libp2p.New(
		ctx,
		libp2p.Identity(prKey),               //保持私玥(节点ID)
		libp2p.Transport(quicTransport),      //使用QUIC传输
		libp2p.Security(secio.ID, secio.New), //使用secio加密
		libp2p.ListenAddrStrings(
			strings.Join([]string{"/ip4/0.0.0.0/udp/", *port, "/quic"}, ""), //监听IPv4
			strings.Join([]string{"/ip6/::/udp/", *port, "/quic"}, ""),      //监听IPv6
		),
		libp2p.Routing(newDHT), //路由DHT
		libp2p.ChainOptions(
			libp2p.Muxer("/yamux/1.0.0", yamux.DefaultTransport),
			libp2p.Muxer("/mplex/6.7.0", mplex.DefaultTransport),
		), //多路复用
	)
	if e != nil {
		log.Fatalln(e)
	}

	//节点地址转为P2P地址
	p2pAddrs, e := peer.AddrInfoToP2pAddrs(&peer.AddrInfo{node.ID(), node.Addrs()})
	if e != nil {
		log.Fatalln(e)
	}
	log.Println("节点:", p2pAddrs[0])

	//如果设置了启发节点则连接
	if *bootstrap != "" {
		bootstrapMa, e := multiaddr.NewMultiaddr(*bootstrap)
		if e != nil {
			log.Fatalln(e)
		}
		bootstrapAddrInfo, e := peer.AddrInfoFromP2pAddr(bootstrapMa)
		if e != nil {
			log.Fatalln(e)
		}

		//连接
		e = node.Connect(ctx, *bootstrapAddrInfo)
		if e != nil {
			log.Fatalln(e)
		}
		log.Println("已经连接启发节点:", *bootstrap)
	}

	//显示DHT节点
	go func() {
		for {
			kadDHT.RefreshRoutingTable()

			for _, peerId := range kadDHT.RoutingTable().ListPeers() {
				log.Println("DHT节点:", peerId)
			}

			time.Sleep(time.Second * 6)
		}
	}()

	// wait for a SIGINT or SIGTERM signal
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	log.Println("收到信号, 关闭...")

	_ = node.Close()
}
