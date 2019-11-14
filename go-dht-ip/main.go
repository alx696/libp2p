package main

import (
	"bytes"
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
	"html/template"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

const (
	CONFIG_DIR            = "./config"
	RSA_FILE_PATH_PRIVATE = "./config/rsa-private"
	RSA_FILE_PATH_PUBLIC  = "./config/rsa-public"
)

var kadDHT *dht.IpfsDHT

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

// 检测IP是否为内网IP4
func isInnerIp4(ip string) bool {
	// 检测IP是否为172内网段
	// 参考 https://stackoverflow.com/questions/19882961/go-golang-check-ip-address-in-range?answertab=votes#tab-top
	is172Inner := func(ip net.IP) bool {
		ipStart := net.ParseIP("172.16.0.0")
		ipEnd := net.ParseIP("172.31.255.255")

		if bytes.Compare(ip, ipStart) >= 0 && bytes.Compare(ip, ipEnd) <= 0 {
			return true
		}
		return false
	}

	// 内网IP段:
	// 10.0.0.0-10.255.255.255
	// 172.16.0.0-172.31.255.255
	// 192.168.0.0-192.168.255.255
	if !strings.HasPrefix(ip, "127.") && !strings.HasPrefix(ip, "10.") &&
		!strings.HasPrefix(ip, "192.168.") && !is172Inner(net.ParseIP(ip)) {
		return false
	}
	return true
}

// 检测IP是否为内网IP6
func isInnerIp6(ipText string) bool {
	ip := net.ParseIP(ipText)
	_, ipv6Net, e := net.ParseCIDR("fe80::/10")
	if e != nil {
		log.Fatalln(e)
	}
	if ipv6Net.Contains(ip) {
		return true
	}

	return false
}

func webServer(port string) {
	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		//获取DHT节点ID
		peerIds := kadDHT.RoutingTable().ListPeers()
		//获取请求ID
		id := request.URL.Query().Get("id")

		if id == "" {
			//未设置ID参数时返回首页
			t, e := template.ParseFiles("./template/index.html")
			if e != nil {
				writer.WriteHeader(http.StatusInternalServerError)
				_, _ = writer.Write([]byte(e.Error()))
				return
			}
			_ = t.Execute(writer, len(peerIds))
		} else {
			//设置ID时返回ID地址信息
			ip := "离线状态"
			for _, peerId := range peerIds {
				if peerId.String() == id {
					addrInfo := kadDHT.FindLocal(peerId)
					log.Println(addrInfo)

					//提取IP
					for _, ma := range addrInfo.Addrs {
						maSplit := strings.Split(ma.String(), "/")
						//ip4
						if maSplit[1] == "ip4" && !isInnerIp4(maSplit[2]) {
							ip = maSplit[2]
						} else if maSplit[1] == "ip6" && !isInnerIp6(maSplit[2]) {
							//ip6
							ip = maSplit[2]
						}
					}

					break
				}
			}

			t, e := template.ParseFiles("./template/ip.html")
			if e != nil {
				writer.WriteHeader(http.StatusInternalServerError)
				_, _ = writer.Write([]byte(e.Error()))
				return
			}
			_ = t.Execute(writer, ip)
		}
	})

	log.Println("Web服务端口:", port)
	log.Fatalln(http.ListenAndServe(strings.Join([]string{"", port}, ":"), nil))
}

// 参考 https://github.com/libp2p/go-libp2p-examples/blob/b7ac9e91865656b3ec13d18987a09779adad49dc/ipfs-camp-2019/06-Pubsub/main.go
func main() {
	log.Println("DHT IP")

	//指定端口,否则随机
	port := flag.String("port", "0", "")
	//指定web端口
	webPort := flag.String("web-port", "60000", "")
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
		bootstrapA, e := peer.AddrInfoFromP2pAddr(bootstrapMa)
		if e != nil {
			log.Fatalln(e)
		}

		//连接
		e = node.Connect(ctx, *bootstrapA)
		if e != nil {
			log.Fatalln(e)
		}
		log.Println("已经连接启发节点:", *bootstrap)
	}

	//启动Web服务
	go webServer(*webPort)

	// wait for a SIGINT or SIGTERM signal
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	log.Println("收到信号, 关闭...")

	_ = node.Close()
}
