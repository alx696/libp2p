package mp2p

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/json"
	"github.com/libp2p/go-libp2p"
	autonat "github.com/libp2p/go-libp2p-autonat-svc"
	connmgr "github.com/libp2p/go-libp2p-connmgr"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/routing"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	libp2pquic "github.com/libp2p/go-libp2p-quic-transport"
	secio "github.com/libp2p/go-libp2p-secio"
	libp2ptls "github.com/libp2p/go-libp2p-tls"
	gonat "github.com/libp2p/go-nat"
	"github.com/multiformats/go-multiaddr"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

const (
	PROTOCOL_BOOTSTRAP = "/p2p/bootstrap"
)

var ctx context.Context
var mDHT *dht.IpfsDHT
var node host.Host
var sm sync.RWMutex
var peerMap = make(map[string]string)
var natGateway gonat.NAT

// 生成或读取密钥
// 注意: Android可用"/sdcard/rsa"定位到存储中rsa文件夹, 但记得在应用权限中申请写外部存储权限.
func rsaKey(dir string) (prKey crypto.PrivKey, puKey crypto.PubKey) {
	log.Println("密钥文件夹路径:", dir)
	privatePath := strings.Join([]string{dir, "private"}, "/")
	publicPath := strings.Join([]string{dir, "public"}, "/")

	_, e := os.Stat(dir)
	if os.IsNotExist(e) {
		e = os.MkdirAll(dir, 0755)
		if e != nil {
			log.Println("创建密钥文件夹出错:", e)
			return
		}

		//生成密钥
		rr := rand.Reader
		prKey, puKey, _ = crypto.GenerateKeyPairWithReader(crypto.Ed25519, -1, rr)

		//存储密钥
		privateKeyBytes, _ := crypto.MarshalPrivateKey(prKey)
		_ = ioutil.WriteFile(privatePath, privateKeyBytes, 0644)
		publicKeyBytes, _ := crypto.MarshalPublicKey(puKey)
		_ = ioutil.WriteFile(publicPath, publicKeyBytes, 0644)
	} else {
		//还原密钥
		privateKeyBytes, _ := ioutil.ReadFile(privatePath)
		publicKeyBytes, _ := ioutil.ReadFile(publicPath)
		prKey, _ = crypto.UnmarshalPrivateKey(privateKeyBytes)
		puKey, _ = crypto.UnmarshalPublicKey(publicKeyBytes)
	}

	return
}

// P2P地址转地址信息
func textToAddrInfo(text string) (*peer.AddrInfo, error) {
	ai := &peer.AddrInfo{}

	ma, e := multiaddr.NewMultiaddr(text)
	if e != nil {
		return ai, e
	}
	ai, e = peer.AddrInfoFromP2pAddr(ma)
	if e != nil {
		return ai, e
	}

	return ai, nil
}

//从流中读取文本
func readTextFormStream(s network.Stream) (string, error) {
	reader := bufio.NewReader(s)
	text, e := reader.ReadString('\n')
	if e != nil {
		return "", e
	}
	text = strings.Replace(text, "\n", "", -1)
	return text, nil
}

func handleBootstrapStream(s network.Stream) {
	peerId := s.Conn().RemotePeer().String()
	peerMa := s.Conn().RemoteMultiaddr().String()
	log.Println("流处:", peerId, peerMa)

	//读取流
	text, e := readTextFormStream(s)
	if e != nil {
		log.Println(e)
		return
	}
	log.Println("流处收到数据:", text)

	//缓存连接节点地址
	sm.Lock()
	if text != "" {
		peerMap[peerId] = text
	} else {
		peerMap[peerId] = strings.Join([]string{peerMa, "/ipfs/", peerId}, "")
	}
	sm.Unlock()

	//获取现有节点地址
	var maArray []string
	sm.RLock()
	for k, v := range peerMap {
		if k == peerId {
			continue
		}

		maArray = append(maArray, v)
	}
	sm.RUnlock()

	//返回现有节点地址
	jsonText := "[]"
	if len(maArray) > 0 {
		jsonBytes, e := json.Marshal(maArray)
		if e != nil {
			log.Println(e)
			return
		}
		jsonText = string(jsonBytes)
	}
	_, e = s.Write([]byte(strings.Join([]string{jsonText, "\n"}, "")))
	if e != nil {
		log.Println(e)
		return
	}

	log.Println("流处完毕")
}

// 引导
func bootstrap(internalPort int, addrText string) error {
	//NAT穿越
	natAddr := ""
	natChan := gonat.DiscoverNATs(ctx)
	select {
	case natGateway = <-natChan:
		if natGateway == nil {
			log.Println("没有找到NAT网关!")
			break
		}
		log.Println("NAT网关类型:", natGateway.Type())

		//获取公网IP
		netIp, e := natGateway.GetExternalAddress()
		if e != nil {
			log.Fatalln(e)
		}
		log.Println("NAT公网IP:", netIp.String())

		log.Println("内部端口:", internalPort)
		//映射端口
		externalPort, e := natGateway.AddPortMapping("udp", internalPort, "mp2p", time.Second*3)
		if e != nil {
			log.Fatalln(e)
		}
		log.Println("NAT内部端口:", internalPort, "映射外部端口:", externalPort)
		natAddr = strings.Join([]string{"/ip4/", netIp.String(), "/udp/", strconv.Itoa(externalPort), "/quic/ipfs/", node.ID().String()}, "")
	}
	log.Println("节点NAT地址:", natAddr)

	//转换地址
	ai, e := textToAddrInfo(addrText)
	if e != nil {
		return e
	}

	//连接节点
	e = node.Connect(ctx, *ai)
	if e != nil {
		return e
	}
	log.Println("已连启发节点")

	//请给节点
	s, e := node.NewStream(ctx, ai.ID, PROTOCOL_BOOTSTRAP)
	if e != nil {
		return e
	}
	_, e = s.Write([]byte(strings.Join([]string{natAddr, "\n"}, "")))
	if e != nil {
		return e
	}
	text, e := readTextFormStream(s)
	if e != nil {
		return e
	}
	log.Println("启发收到数据:", text)
	e = s.Reset()
	if e != nil {
		return e
	}

	//逐个连接
	var maArray []string
	e = json.Unmarshal([]byte(text), &maArray)
	if e != nil {
		return e
	}
	sm.Lock()
	for _, v := range maArray {
		addrInfo, e := textToAddrInfo(v)
		if e != nil {
			log.Println(e)
			continue
		}

		//连接节点, 触发DHT路由刷新
		e = node.Connect(ctx, *addrInfo)
		if e != nil {
			log.Println(e)
			continue
		}
		log.Println("已连节点:", v)

		//缓存节点
		peerMap[addrInfo.ID.String()] = v
	}
	sm.Unlock()

	return nil
}

// 参考 https://github.com/libp2p/go-libp2p-examples/blob/master/libp2p-host/host.go
func Init(port, bootstrapAddr string) {
	log.Println("启动节点:", port, bootstrapAddr)

	internalPort, e := strconv.Atoi(port)
	if e != nil {
		log.Fatalln(e)
	}

	//生成密钥
	prKey, _ := rsaKey("./config/rsa")

	// The context governs the lifetime of the libp2p node.
	// Cancelling it will stop the the host.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//创建节点
	node, e = libp2p.New(
		ctx,
		libp2p.Identity(prKey), //保持节点ID
		libp2p.ListenAddrStrings(
			strings.Join([]string{"/ip4/0.0.0.0/tcp/", port}, ""),          //监听IPv4
			strings.Join([]string{"/ip4/0.0.0.0/udp/", port, "/quic"}, ""), //监听IPv4
			//strings.Join([]string{"/ip6/::/udp/", port, "/quic"}, ""),      //监听IPv6
		),
		// support TLS connections
		libp2p.Security(libp2ptls.ID, libp2ptls.New),
		// support secio connections
		libp2p.Security(secio.ID, secio.New),
		// support QUIC
		libp2p.Transport(libp2pquic.NewTransport),
		// support any other default transports (TCP)
		libp2p.DefaultTransports,
		// Let's prevent our peer from having too many
		// connections by attaching a connection manager.
		libp2p.ConnectionManager(connmgr.NewConnManager(
			100,         // Lowwater
			300,         // HighWater,
			time.Minute, // GracePeriod
		)),
		// Attempt to open ports using uPNP for NATed hosts.
		libp2p.NATPortMap(),
		// Let this host use the DHT to find other hosts
		libp2p.Routing(func(h host.Host) (routing.PeerRouting, error) {
			mDHT, e = dht.New(ctx, h)
			return mDHT, e
		}),
		// Let this host use relays and advertise itself on relays if
		// it finds it is behind NAT. Use libp2p.Relay(options...) to
		// enable active relays and more.
		libp2p.EnableAutoRelay(),
	)
	if e != nil {
		log.Fatalln(e)
	}

	// If you want to help other peers to figure out if they are behind
	// NATs, you can launch the server-side of AutoNAT too (AutoRelay
	// already runs the client)
	_, e = autonat.NewAutoNATService(ctx, node,
		// Support same non default security and transport options as
		// original host.
		libp2p.Security(libp2ptls.ID, libp2ptls.New),
		libp2p.Security(secio.ID, secio.New),
		libp2p.Transport(libp2pquic.NewTransport),
		libp2p.DefaultTransports,
	)

	//节点地址转为P2P地址
	p2pAddrs, e := peer.AddrInfoToP2pAddrs(&peer.AddrInfo{node.ID(), node.Addrs()})
	if e != nil {
		log.Fatalln(e)
	}
	log.Println("节点地址:", p2pAddrs)

	//设置引导流处
	node.SetStreamHandler(PROTOCOL_BOOTSTRAP, handleBootstrapStream)

	//如果设置了引导节点则连接
	if bootstrapAddr != "" {
		e = bootstrap(internalPort, bootstrapAddr)
		if e != nil {
			log.Println(e)
		}
	}

	//显示DHT节点
	go func() {
		for {
			mDHT.RefreshRoutingTable()

			for _, peerId := range mDHT.RoutingTable().ListPeers() {
				log.Println("DHT节点:", peerId.String())
			}

			//var idMap = make(map[string]int)
			//sm.Lock()
			//for _, peerId := range kadDHT.RoutingTable().ListPeers() {
			//	idMap[peerId.String()] = 0
			//
			//	_, exists := peerMap[peerId.String()]
			//	if exists {
			//		continue
			//	}
			//
			//	log.Println("发现节点:", peerId.String())
			//	peerMap[peerId.String()] = ""
			//}
			//for k, _ := range peerMap {
			//	_, exists := idMap[k]
			//	if exists {
			//		continue
			//	}
			//
			//	log.Println("失去节点:", k)
			//	delete(peerMap, k)
			//}
			//log.Println("DHT节点数量:", len(peerMap))
			//sm.Unlock()

			time.Sleep(time.Second * 6)
		}
	}()

	// wait for a SIGINT or SIGTERM signal
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	log.Println("收到信号, 关闭...")

	//移除端口映射
	if natGateway != nil {
		_ = natGateway.DeletePortMapping("udp", internalPort)
	}

	_ = node.Close()
}
