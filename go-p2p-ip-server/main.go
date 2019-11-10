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
	"github.com/libp2p/go-libp2p-quic-transport"
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
	protocolID = "/p2p/ip"
)

func rsaKey() (prKey crypto.PrivKey, puKey crypto.PubKey) {
	privateKeyBytes, e := ioutil.ReadFile("./rsa-private.txt")
	if e != nil {
		//生成密钥
		rr := rand.Reader
		prKey, puKey, _ = crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, rr)

		//存储密钥
		privateKeyBytes, _ := crypto.MarshalPrivateKey(prKey)
		_ = ioutil.WriteFile("./rsa-private.txt", privateKeyBytes, 0644)
		publicKeyBytes, _ := crypto.MarshalPublicKey(puKey)
		_ = ioutil.WriteFile("./rsa-public.txt", publicKeyBytes, 0644)
	} else {
		//还原密钥
		publicKeyBytes, _ := ioutil.ReadFile("./rsa-public.txt")

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
		log.Println("读取错误:", e)
	} else {
		log.Println("读取内容:", txt)
	}

	_, e = stream.Write([]byte("你的IP\n"))
	if e != nil {
		log.Println("写入错误:", e)
	}
	log.Println("结果已经返回")

	//不关掉流
	//_ = stream.Close()

	//继续读取(用于检测客户断开)
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
	log.Println("客户断开:", clientAddr)
}

func main() {
	//指定端口,否则随机
	port := flag.String("port", "0", "")
	//指定目标
	target := flag.String("target", "", "")
	flag.Parse()

	prKey, _ := rsaKey()
	quicTransport, e := libp2pquic.NewTransport(prKey)
	if e != nil {
		log.Fatalln("创建QUIC传输层出错:", e)
	}
	ctx := context.Background()
	node, e := libp2p.New(ctx,
		libp2p.Transport(quicTransport),
		libp2p.Identity(prKey),
		libp2p.ListenAddrStrings(strings.Join([]string{"/ip4/0.0.0.0/udp/", *port, "/quic"}, "")),
	)
	if e != nil {
		log.Fatalln("创建libp2p出错:", e)
	}
	log.Println("节点ID:", node.ID(), "节点地址:", node.Addrs())

	//打印节点P2P地址
	p2pAddrs, e := peer.AddrInfoToP2pAddrs(&peer.AddrInfo{node.ID(), node.Addrs()})
	if e != nil {
		log.Fatalln(e)
	}
	log.Println("节点P2P地址:", p2pAddrs)

	//设置流处
	node.SetStreamHandler(protocolID, handleStream)

	//开启DHT
	//注意:必须至少连接一个节点,然后才能通过相互发现!
	kadDHT, e := dht.New(ctx, node)
	if e != nil {
		log.Fatalln("创建DHT出错:", e)
	}
	go func() {
		for {
			for _, v := range kadDHT.RoutingTable().ListPeers() {
				localAddr := kadDHT.FindLocal(v)
				log.Println(localAddr)
			}
			time.Sleep(time.Second * 6)
		}
	}()

	//设置种点
	//TODO 成品应该多设置几个种点, 每次只连一个, 不行才换另外一个. 一旦连到一些节点后, 可缓存下来下次先连缓存节点, 种点保底.
	devP2pAddr, _ := multiaddr.NewMultiaddr("/dns4/test.dev.lilu.red/udp/10002/quic/ipfs/QmfFf8UjpnNtyqVSdx5GCcaEafq4s5vy1mUQdaQvZ4SSRd")
	devAddr, _ := peer.AddrInfoFromP2pAddr(devP2pAddr)
	for {
		e = node.Connect(ctx, *devAddr)
		if e == nil {
			break
		}
		log.Println("连接种子出错(1分钟后重试):", e)
		time.Sleep(time.Minute)
	}

	if *target != "" {
		targetP2pAddr, _ := multiaddr.NewMultiaddr(*target)
		targetAddr, _ := peer.AddrInfoFromP2pAddr(targetP2pAddr)

		for {
			//连接
			e = node.Connect(ctx, *targetAddr)
			if e != nil {
				log.Println("连接错误:", e)
				//30秒后重试
				time.Sleep(time.Second * 30)
				continue
			}

			//建流
			s, e := node.NewStream(ctx, targetAddr.ID, protocolID)
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
