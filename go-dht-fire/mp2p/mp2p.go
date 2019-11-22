package mp2p

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/json"
	"github.com/ipfs/go-cid"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/routing"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	libp2pquic "github.com/libp2p/go-libp2p-quic-transport"
	gonat "github.com/libp2p/go-nat"
	"github.com/multiformats/go-multiaddr"
	"github.com/multiformats/go-multihash"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

const (
	PROTOCOL_BOOTSTRAP = "/mp2p/bootstrap"
	PROTOCOL_HI        = "/mp2p/hi"
)

var natGateway gonat.NAT
var m sync.RWMutex
var ctx context.Context
var kadDHT *dht.IpfsDHT
var node host.Host
var peerIpfsMap = make(map[string]string)

func enableNAT(internalPort int) (net.IP, int, error) {
	natChan := gonat.DiscoverNATs(ctx)
	select {
	case natGateway = <-natChan:
		//获取公网IP
		netIp, e := natGateway.GetExternalAddress()
		if e != nil {
			return net.IP{}, 0, e
		}
		log.Println("NAT公网IP:", netIp.String())

		//映射端口
		externalPort, e := natGateway.AddPortMapping("udp", internalPort, "mp2p", time.Minute)
		if e != nil {
			return net.IP{}, 0, e
		}
		log.Println("NAT内部端口:", internalPort, "映射外部端口:", externalPort)

		return netIp, externalPort, nil

		////移除端口映射
		//_ = natGateway.DeletePortMapping("udp", internalPort)
	}
}

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
		prKey, puKey, _ = crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, rr)

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

// 流处
func handleStream(stream network.Stream) {
	streamPeerId := stream.Conn().RemotePeer().String()
	streamPeerMa := stream.Conn().RemoteMultiaddr().String()
	log.Println("流处:", stream.Protocol(), streamPeerId, streamPeerMa)

	//读取内容
	reader := bufio.NewReader(stream)
	txt, e := reader.ReadString('\n')
	if e != nil {
		log.Println(e)
		return
	}
	txt = strings.Replace(txt, "\n", "", -1)
	log.Println(txt)

	switch stream.Protocol() {

	case PROTOCOL_BOOTSTRAP:
		//缓存连接节点地址
		m.Lock()
		if txt != "" {
			peerIpfsMap[streamPeerId] = txt
		} else {
			peerIpfsMap[streamPeerId] = strings.Join([]string{streamPeerMa, "/ipfs/", streamPeerId}, "")
		}
		m.Unlock()

		//获取现有节点地址
		var idArray []string
		m.RLock()
		for k, v := range peerIpfsMap {
			if k == streamPeerId {
				continue
			}

			idArray = append(idArray, v)
		}
		m.RUnlock()

		//返回现有节点地址
		jsonText := "[]"
		if len(idArray) > 0 {
			jsonBytes, e := json.Marshal(idArray)
			if e != nil {
				log.Println(e)
				return
			}
			jsonText = string(jsonBytes)
		}
		_, e = stream.Write([]byte(strings.Join([]string{jsonText, "\n"}, "")))
		if e != nil {
			log.Println(e)
			return
		}

	case PROTOCOL_HI:
		_, e = stream.Write([]byte("Fine!\n"))
		if e != nil {
			log.Println(e)
			return
		}

	}
}

// 更新节点
func updatePeer(jsonText string) {
	var maTextArray []string
	e := json.Unmarshal([]byte(jsonText), &maTextArray)
	if e != nil {
		log.Println(e)
		return
	}

	for _, v := range maTextArray {
		addrInfo, e := textToAddrInfo(v)
		if e != nil {
			log.Println(e)
			continue
		}
		peerIpfsMap[addrInfo.ID.String()] = v
	}
}

// 引导
func bootstrap(addrText, natAddr string) error {
	ai, e := textToAddrInfo(addrText)
	if e != nil {
		return e
	}

	//连接
	e = node.Connect(ctx, *ai)
	if e != nil {
		return e
	}
	log.Println("已经连接启发节点:", addrText)

	//获取引导数据
	s, e := node.NewStream(ctx, ai.ID, PROTOCOL_BOOTSTRAP)
	if e != nil {
		return e
	}
	_, e = s.Write([]byte(strings.Join([]string{natAddr, "\n"}, "")))
	if e != nil {
		return e
	}
	reader := bufio.NewReader(s)
	txt, e := reader.ReadString('\n')
	if e != nil {
		return e
	}
	txt = strings.Replace(txt, "\n", "", -1)
	log.Println("拿到引导数据:", txt)
	e = s.Reset()
	if e != nil {
		return e
	}

	//更新节点
	updatePeer(txt)

	return nil
}

// 问候
func sayHi(maText string) {
	ai, e := textToAddrInfo(maText)
	if e != nil {
		return
	}
	log.Println("问候:", ai)

	//连接
	e = node.Connect(ctx, *ai)
	if e != nil {
		return
	}
	log.Println("已经连接启发节点:", maText)

	s, e := node.NewStream(ctx, ai.ID, PROTOCOL_HI)
	if e != nil {
		log.Println(e)
		return
	}
	_, e = s.Write([]byte("你好\n"))
	if e != nil {
		log.Println(e)
		return
	}
	reader := bufio.NewReader(s)
	txt, e := reader.ReadString('\n')
	if e != nil {
		log.Println(e)
		return
	}
	txt = strings.Replace(txt, "\n", "", -1)
	log.Println("Hi收到回复:", txt)
	e = s.Reset()
	if e != nil {
		log.Println(e)
		return
	}
}

// 参考 https://github.com/libp2p/go-libp2p-examples/blob/b7ac9e91865656b3ec13d18987a09779adad49dc/ipfs-camp-2019/06-Pubsub/main.go
func Init(port, bootstrapAddr string) {
	log.Println("启动P2P节点:", port, bootstrapAddr)

	if port == "0" {
		log.Fatalln("必须明确端口")
	}

	//生成密钥
	prKey, _ := rsaKey("./config/rsa")

	//创建传输层
	quicTransport, e := libp2pquic.NewTransport(prKey)
	if e != nil {
		log.Fatalln(e)
	}

	//创建上下文
	ctx = context.Background()

	//DHT定义
	newDHT := func(h host.Host) (routing.PeerRouting, error) {
		var err error
		kadDHT, err = dht.New(ctx, h)
		return kadDHT, err
	}

	//创建节点
	node, e = libp2p.New(
		ctx,
		libp2p.Identity(prKey),          //保持节点ID
		libp2p.Transport(quicTransport), //使用QUIC传输
		libp2p.ListenAddrStrings(
			strings.Join([]string{"/ip4/0.0.0.0/udp/", port, "/quic"}, ""), //监听IPv4
			strings.Join([]string{"/ip6/::/udp/", port, "/quic"}, ""),      //监听IPv6
		),
		libp2p.Routing(newDHT), //路由DHT
	)
	if e != nil {
		log.Fatalln(e)
	}

	//节点地址转为P2P地址
	p2pAddrs, e := peer.AddrInfoToP2pAddrs(&peer.AddrInfo{node.ID(), node.Addrs()})
	if e != nil {
		log.Fatalln(e)
	}
	log.Println("节点:", p2pAddrs)

	node.SetStreamHandler(PROTOCOL_BOOTSTRAP, handleStream)
	node.SetStreamHandler(PROTOCOL_HI, handleStream)

	//NAT穿越
	natAddr := ""
	internalPort, e := strconv.Atoi(port)
	if e != nil {
		log.Println(e)
	} else {
		natIp, natPort, e := enableNAT(internalPort)
		if e != nil {
			log.Println(e)
		} else {
			natAddr = strings.Join([]string{"/ip4/", natIp.String(), "/udp/", strconv.Itoa(natPort), "/quic/ipfs/", node.ID().String()}, "")
		}
	}
	log.Println("节点NAT:", natAddr)

	//如果设置了引导节点则连接
	if bootstrapAddr != "" {
		e := bootstrap(bootstrapAddr, natAddr)
		if e != nil {
			log.Println(e)
		}

		//进行一次连接和通信, 触发DHT列表刷新
		for _, v := range peerIpfsMap {
			sayHi(v)
		}
	}

	//CID前缀
	prefix := cid.Prefix{
		Version:  1,
		Codec:    cid.Raw,
		MhType:   multihash.SHA2_256,
		MhLength: -1,
	}
	c, e := prefix.Sum([]byte("你好"))
	if e != nil {
		log.Println(e)
	} else {
		log.Println(c)
	}

	//显示DHT节点
	go func() {
		for {
			kadDHT.RefreshRoutingTable()

			for _, peerId := range kadDHT.RoutingTable().ListPeers() {
				log.Println("DHT节点:", peerId)
			}
			log.Println("---")

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
