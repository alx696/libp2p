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
	"github.com/libp2p/go-nat"
	//libp2pnat "github.com/libp2p/go-libp2p-nat"
	"github.com/multiformats/go-multiaddr"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"strconv"
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
		log.Println("读取错误:", e)
	} else {
		log.Println("读取内容:", txt)
	}

	_, e = stream.Write([]byte(strings.Join([]string{"远程地址:", clientAddr, "\n"}, "")))
	if e != nil {
		log.Println("写入错误:", e)
	}
	log.Println("结果已经返回")

	_ = stream.Close()

	////继续读取(用于检测客户断开)
	//for {
	//	txt, e := reader.ReadString('\n')
	//	txt = strings.Replace(txt, "\n", "", -1)
	//	if e != nil {
	//		log.Println("读取错误:", e)
	//		break
	//	} else {
	//		log.Println("读取内容:", txt)
	//	}
	//}
	//log.Println("客户断开:", clientAddr)
}

func main() {
	log.Println("DHT连接NAT穿越")

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

	if *bootstrap != "" {
		//如果设置了启发节点则连接
		ma, _ := multiaddr.NewMultiaddr(*bootstrap)
		a, _ := peer.AddrInfoFromP2pAddr(ma)
		e := node.Connect(ctx, *a)
		if e != nil {
			log.Fatalln(e)
		}
		log.Println("已经连接节点:", *bootstrap)

		//发送问候
		s, e := node.NewStream(ctx, a.ID, PROTOCOL_ID)
		if e != nil {
			log.Fatalln("建流错误:", a.ID, e)
		}
		_, e = s.Write([]byte("Hi! 启发节点\n"))
		if e != nil {
			log.Fatalln("发送错误:", a.ID, e)
		}
		_ = s.Close()
		log.Println("已经发送")
	} else {
		//启发节点每3秒显示一次DHT路由表中的节点
		var nodeMap = make(map[string]string)

		go func() {
			for {
				for _, v := range kadDHT.RoutingTable().ListPeers() {
					if _, exists := nodeMap[v.String()]; exists {
						continue
					}

					addr := kadDHT.FindLocal(v)
					log.Println("DHT发现节点:", addr)

					//发送问候
					s, e := node.NewStream(ctx, v, PROTOCOL_ID)
					if e != nil {
						log.Println("建流错误:", v, e)
						continue
					}
					_, e = s.Write([]byte("你好, DHT节点\n"))
					if e != nil {
						log.Println("发送错误:", v, e)
						continue
					}
					_ = s.Close()
					log.Println("已经发送")

					//缓存节点
					nodeMap[v.String()] = ""
				}

				log.Println("已连节点数量:", len(nodeMap))
				time.Sleep(time.Second * 3)
			}
		}()
	}

	//NAT穿越
	log.Println("准备NAT穿越")
	innerPort, _ := strconv.Atoi(*port)
	if innerPort == 0 {
		innerPort, e = strconv.Atoi(strings.Split(node.Addrs()[0].String(), "/")[4])
		if e != nil {
			log.Fatalln(e)
		}
	}
	var natIp string
	var natPort int
	natChan := nat.DiscoverNATs(ctx)
	select {
	case natGateway := <-natChan:
		natExternalAddress, e := natGateway.GetExternalAddress()
		if e != nil {
			log.Fatalln(e)
		}
		natIp = natExternalAddress.String()

		mappedExternalPort, e := natGateway.AddPortMapping("udp", innerPort, "P2P测试", time.Minute*3)
		if e != nil {
			log.Fatalln(e)
		}
		natPort = mappedExternalPort

		////不使用时记得移除
		//_ = natGateway.DeletePortMapping("udp", innerPort)
	}
	natProtoName := "ip4"
	if strings.ContainsAny(natIp, ":") {
		natProtoName = "ip6"
	}
	natP2pMa, _ := multiaddr.NewMultiaddr(strings.Join([]string{"/", natProtoName, "/", natIp, "/udp/", strconv.Itoa(natPort), "/quic/ipfs/", node.ID().String()}, ""))
	log.Println("NAT的P2P地址:", natP2pMa)

	// wait for a SIGINT or SIGTERM signal
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	log.Println("收到信号, 关闭...")

	_ = node.Close()
}
