package main

import (
	"bufio"
	"context"
	"crypto/rand"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p-quic-transport"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
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
	prKey, _ := rsaKey()

	ctx := context.Background()

	quicTransport, e := libp2pquic.NewTransport(prKey)
	if e != nil {
		log.Fatalln(e)
	}

	node, e := libp2p.New(ctx,
		libp2p.Transport(quicTransport),
		libp2p.Identity(prKey),
		libp2p.ListenAddrStrings("/ip4/0.0.0.0/udp/60000/quic"),
	)
	if e != nil {
		log.Fatalln(e)
	}
	log.Println("节点地址:", node.Addrs())

	//打印节点P2P地址
	p2pAddrs, e := peer.AddrInfoToP2pAddrs(&peer.AddrInfo{node.ID(), node.Addrs()})
	if e != nil {
		log.Fatalln(e)
	}
	log.Println("节点P2P地址:", p2pAddrs)

	//开启DHT
	katDHT, e := dht.New(ctx, node)
	if e != nil {
		log.Fatalln(e)
	}
	log.Println(katDHT.PeerID())

	//设置流处
	node.SetStreamHandler(protocolID, handleStream)

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
