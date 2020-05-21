package im

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	host2 "github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	secio "github.com/libp2p/go-libp2p-secio"
	libp2ptls "github.com/libp2p/go-libp2p-tls"
	"github.com/libp2p/go-libp2p/p2p/discovery"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const ProtocolId = "/p2p/mdns"

type discoveryNotifee struct {
	PeerChan chan peer.AddrInfo
}

type Message struct {
	Type    string `json:"type"`    //文本或文件
	Content string `json:"content"` //Type为文本时是文本内容, Type为文件时是文件名称.
}

var fileDirectory string //文件目录
var ctx context.Context
var host host2.Host
var messageChan chan Message //缓存接收到消息的信道
var ps map[string]peer.AddrInfo

// interface to be called when new  peer is found
// 注意:外部勿用!
func (n *discoveryNotifee) HandlePeerFound(pi peer.AddrInfo) {
	n.PeerChan <- pi
}

// 生成或读取密钥
// 注意: Android可用"/sdcard/rsa"定位到存储中rsa文件夹, 但记得在应用权限中申请写外部存储权限.
func rsaKey(dir string) (prKey crypto.PrivKey, puKey crypto.PubKey) {
	log.Println("密钥文件夹路径:", dir)
	privatePath := strings.Join([]string{dir, "private"}, "/")
	publicPath := strings.Join([]string{dir, "public"}, "/")

	_, e := os.Stat(dir)
	if os.IsNotExist(e) {
		e = os.MkdirAll(dir, os.ModePerm)
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

// 创建读写器
func newReadWriter(id string) (*bufio.ReadWriter, error) {
	p, ok := ps[id]
	if !ok {
		return nil, errors.New("ID错误")
	}

	// 连接并建流
	err := host.Connect(ctx, p)
	if err != nil {
		delete(ps, id)
		return nil, err
	}
	s, err := host.NewStream(ctx, p.ID, ProtocolId)
	if err != nil {
		delete(ps, id)
		return nil, err
	}

	// 创建读写器
	rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))
	return rw, nil
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

// 读取文件并保存
func ReadFileAndSave(rw *bufio.ReadWriter, path string) bool {
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

// 处理进来的流
func handleStream(s network.Stream) {
	log.Println("处理新流")

	rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))

	// 读取
	messageJson := ReadText(rw)
	var message Message
	err := json.Unmarshal([]byte(messageJson), &message)
	if err != nil {
		log.Println("消息格式错误:", err)
		_ = s.Close()
		return
	}

	if message.Type == "文本" {
		// 读取文本
		log.Println("文本消息:", message.Content)
		// 存入信道等待提取
		messageChan <- message
	} else if message.Type == "文件" {
		log.Println("文件消息")
		// 读取文件并保存
		filename := fmt.Sprint(time.Now().Unix(), "-", message.Content)
		ReadFileAndSave(rw, fmt.Sprint(fileDirectory, "/", filename))
		message.Content = filename
		// 存入信道等待提取
		messageChan <- message
	}

	// 回复2次
	for i := 0; i < 2; i++ {
		WriteText(rw, time.Now().String())
	}

	_ = s.Close()
}

// 初始化节点
// dir: 末尾不带斜杠/
func Init(dir string) error {
	log.Println("初始化:", dir)

	//准备文件文件夹
	fileDirectory = fmt.Sprint(dir, "/file")
	err := os.MkdirAll(fileDirectory, os.ModePerm)
	if err != nil {
		return err
	}
	log.Println("文件文件夹路径:", fileDirectory)

	//准备缓存
	messageChan = make(chan Message, 100)
	ps = make(map[string]peer.AddrInfo)

	//准备上下文
	var cancel context.CancelFunc
	ctx, cancel = context.WithCancel(context.Background())
	defer cancel()

	//生成密钥
	prKey, _ := rsaKey(fmt.Sprint(dir, "/rsa"))

	host, err = libp2p.New(
		ctx,
		libp2p.Identity(prKey),
		// support TLS connections
		libp2p.Security(libp2ptls.ID, libp2ptls.New),
		// support secio connections
		libp2p.Security(secio.ID, secio.New),
	)
	if err != nil {
		return err
	}
	addrs, err := peer.AddrInfoToP2pAddrs(&peer.AddrInfo{host.ID(), host.Addrs()})
	if err != nil {
		return err
	}
	log.Println("我的地址:", addrs)

	// 别人向你创建创建流时进行处理
	host.SetStreamHandler(ProtocolId, handleStream)

	// 创建mDNS服务并注册发现列队
	ser, err := discovery.NewMdnsService(ctx, host, time.Second*6, "mdns-test")
	if err != nil {
		return err
	}
	n := &discoveryNotifee{}
	n.PeerChan = make(chan peer.AddrInfo)
	ser.RegisterNotifee(n)

	// 发现节点
	for p := range n.PeerChan {
		if _, ok := ps[p.ID.String()]; !ok {
			log.Println("发现节点:", p.ID.String())
			ps[p.ID.String()] = p
		}
	}

	return nil
}

// 获取自己ID
func GetSelfId() string {
	return host.ID().String()
}

// 获取节点
func FindPeer() string {
	var ids []string
	for k, _ := range ps {
		ids = append(ids, k)
	}
	jsonBytes, _ := json.Marshal(ids)
	txt := string(jsonBytes)
	if txt == "null" {
		txt = "[]"
	}
	return txt
}

// 发送文本
func SendText(id, text string) error {
	log.Println("发送文本:", id, text)

	// 创建读写器
	rw, err := newReadWriter(id)
	if err != nil {
		return err
	}

	// 发出
	message := Message{"文本", text}
	jsonBytes, err := json.Marshal(message)
	if err != nil {
		return err
	}
	WriteText(rw, string(jsonBytes))

	// 读取结果
	for {
		str, err := rw.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				log.Println("读取文本出错:", err)
				return err
			} else {
				log.Println("读取文本完毕")
				break
			}
		}
		str = strings.Replace(str, "\n", "", -1)
		log.Println("收到回复:", str)
	}

	return nil
}

// 发送文件
func SendFile(id, path string) error {
	log.Println("发送文件:", id, path)

	// 创建读写器
	rw, err := newReadWriter(id)
	if err != nil {
		return err
	}

	// 发出
	message := Message{"文件", filepath.Base(path)}
	jsonBytes, err := json.Marshal(message)
	if err != nil {
		return err
	}
	WriteText(rw, string(jsonBytes))
	WriteFile(rw, path)

	// 读取结果
	for {
		str, err := rw.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				log.Println("读取文本出错:", err)
				return err
			} else {
				log.Println("读取文本完毕")
				break
			}
		}
		str = strings.Replace(str, "\n", "", -1)
		log.Println("收到回复:", str)
	}

	return nil
}

// 提取消息
// 说明: 因为Java不支持,只能弄成轮训模式.
func ExtractMessage() string {
	var ms []Message

	select {
	case m := <-messageChan:
		ms = append(ms, m)
	default:
		log.Println("信道缓存消息提取完毕")
		break
	}

	jsonBytes, _ := json.Marshal(ms)
	txt := string(jsonBytes)
	if txt == "null" {
		txt = "[]"
	}
	return txt
}
