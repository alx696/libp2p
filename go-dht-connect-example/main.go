package main

import (
	"bufio"
	"context"
	"crypto/rand"
	"flag"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	libp2pquic "github.com/libp2p/go-libp2p-quic-transport"
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
	PROTOCOL_ID           = "/p2p/dht"
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

func handleStream(stream network.Stream) {
	clientAddr := stream.Conn().RemoteMultiaddr().String()
	log.Println("处理流:", clientAddr)

	reader := bufio.NewReader(stream)
	txt, e := reader.ReadString('\n')
	txt = strings.Replace(txt, "\n", "", -1)
	if e != nil {
		log.Println("处理流读取错误:", e)
	} else {
		log.Println("处理流读取内容:", txt)
	}

	_ = stream.Close()
}

func main() {
	log.Println("DHT节点连接")

	//指定端口,否则随机
	port := flag.String("port", "0", "")
	//启发节点
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

	//创建节点
	node, e := libp2p.New(
		ctx,
		libp2p.Identity(prKey),          //保持私玥(节点ID)
		libp2p.Transport(quicTransport), //使用QUIC传输
		libp2p.ListenAddrStrings(
			strings.Join([]string{"/ip4/0.0.0.0/udp/", *port, "/quic"}, ""), //监听IPv4
			strings.Join([]string{"/ip6/::/udp/", *port, "/quic"}, ""),      //监听IPv6
		),
	)
	if e != nil {
		log.Fatalln(e)
	}

	//节点地址转为P2P地址
	p2pAddrs, e := peer.AddrInfoToP2pAddrs(&peer.AddrInfo{node.ID(), node.Addrs()})
	if e != nil {
		log.Fatalln(e)
	}
	log.Println("节点P2P地址:", p2pAddrs)

	//设置流处
	node.SetStreamHandler(PROTOCOL_ID, handleStream)

	//创建DHT
	kadDHT, e := dht.New(ctx, node)
	if e != nil {
		log.Fatalln(e)
	}

	go func() {
		for {
			for _, v := range kadDHT.RoutingTable().ListPeers() {
				addr := kadDHT.FindLocal(v)
				log.Println("DHT节点:", addr)

				//发送问候
				s, e := node.NewStream(ctx, v, PROTOCOL_ID)
				if e != nil {
					log.Println("DHT建流错误:", v, e)
					continue
				}
				_, e = s.Write([]byte("你好, DHT节点\n"))
				if e != nil {
					log.Println("DHT发送错误:", v, e)
					continue
				}
				_ = s.Close()
				log.Println("DHT发送完毕")
			}

			time.Sleep(time.Second * 6)
		}
	}()

	//如果设置了启发节点则连接
	if *bootstrap != "" {
		ma, _ := multiaddr.NewMultiaddr(*bootstrap)
		a, _ := peer.AddrInfoFromP2pAddr(ma)

		//连接
		e := node.Connect(ctx, *a)
		if e != nil {
			log.Fatalln(e)
		}
		log.Println("已经连接启发节点:", *bootstrap)

		//创建流, 发送
		s, e := node.NewStream(ctx, a.ID, PROTOCOL_ID)
		if e != nil {
			log.Fatalln("启发节点建流错误:", e)
		}
		_, e = s.Write([]byte("Hi! 启发节点\n"))
		if e != nil {
			log.Fatalln("启发节点发送错误:", e)
		}
		_ = s.Close()
		log.Println("启发节点发送完毕")
	}

	// wait for a SIGINT or SIGTERM signal
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	log.Println("收到信号, 关闭...")

	_ = node.Close()
}
