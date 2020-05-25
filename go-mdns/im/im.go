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
	"strconv"
	"strings"
	"time"
)

const ProtocolId = "/p2p/mdns/iim"

type discoveryNotifee struct {
	PeerChan chan peer.AddrInfo
}

// 消息
// 如果是文件, FileName和FileSize必须设置.
type Message struct {
	Text     string `json:"text"`      //文本
	FileName string `json:"file_name"` //文件名称
	FileSize int64  `json:"file_size"` //文件大小(字节数量)
	FromId   string `json:"from_id"`   //发送人ID(接收时设置真实来源)
	Ts       int64  `json:"ts"`        //接收时间
}

type Info struct {
	Id    string `json:"id"`
	Name  string `json:"name"`
	Photo string `json:"photo"` //Base64编码的字节,并非图片!
}

var fileDir string //文件目录
var myInfo Info
var ctx context.Context
var host host2.Host
var ser discovery.Service
var pn *discoveryNotifee
var messageChan chan Message //缓存接收到消息的信道
var ps map[string]peer.AddrInfo

// interface to be called when new  peer is found
// 注意:外部勿用!
func (n *discoveryNotifee) HandlePeerFound(pi peer.AddrInfo) {
	n.PeerChan <- pi
}

// 获取或生成密钥
func getOrGenerateKey(privatePath, publicPath string) (prKey crypto.PrivKey, puKey crypto.PubKey) {
	log.Println("密钥路径:", privatePath, publicPath)

	_, e := os.Stat(privatePath)
	if os.IsNotExist(e) {
		log.Println("生成密钥")
		//生成密钥
		rr := rand.Reader
		prKey, puKey, _ = crypto.GenerateKeyPairWithReader(crypto.Ed25519, -1, rr)

		//存储密钥
		privateKeyBytes, _ := crypto.MarshalPrivateKey(prKey)
		_ = ioutil.WriteFile(privatePath, privateKeyBytes, os.ModePerm)
		publicKeyBytes, _ := crypto.MarshalPublicKey(puKey)
		_ = ioutil.WriteFile(publicPath, publicKeyBytes, os.ModePerm)
	} else {
		log.Println("读取密钥")
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

// 从读写器中读取文本
func readText(rw *bufio.ReadWriter) (string, error) {
	str, err := rw.ReadString('\n')
	if err != nil {
		log.Println("读取文本出错:", err)
		return "", err
	}
	return strings.Replace(str, "\n", "", -1), nil
}

// 往读写器中写入文本
func writeText(rw *bufio.ReadWriter, text string) error {
	_, err := rw.Write([]byte(fmt.Sprint(text, "\n")))
	if err != nil {
		log.Println("写入文本失败:", err)
		return err
	}
	err = rw.Flush()
	if err != nil {
		log.Println("压出文本失败:", err)
		return err
	}
	return nil
}

// 从读写器中读取文件并保存
func readFileAndSave(rw *bufio.ReadWriter, path string, fileSize int64) error {
	//如果预期路径已经存在则重命名
	_, err := os.Stat(path)
	if err == nil {
		fileExt := filepath.Ext(path)
		path = fmt.Sprint(filepath.Dir(path), "/", strings.Replace(filepath.Base(path), fileExt, "", 1), "(", time.Now().Unix(), ")", fileExt)
	}

	f, err := os.Create(path)
	defer f.Close()
	if err != nil {
		log.Println("创建文件出错:", err)
		return err
	}

	var sizeSum int64
	buf := make([]byte, 1048576)
	for {
		n, err := rw.Read(buf)
		if err == io.EOF {
			//永远不会触发!!! 除非流关闭
			break
		} else if err != nil {
			log.Println("流中读取字节失败:", err)
			return err
		}
		//size := int64(binary.LittleEndian.Uint64(buf[0:8]))
		//log.Println("总共大小:", size)
		log.Println("收到:", n)
		_, err = f.Write(buf[0:n])
		if err != nil {
			log.Println("文件写入字节出错:", err)
			return err
		}
		sizeSum += int64(n)
		log.Println(sizeSum, fileSize)
		if sizeSum == fileSize {
			log.Println("文件读取完毕")
			break
		}
	}

	//for {
	//	someBytes, err := rw.ReadBytes('\n')
	//	if err == io.EOF {
	//		//永远不会触发?!
	//		log.Println("流中读取字节完毕(EOF)")
	//		break
	//	} else if err != nil {
	//		log.Println("流中读取字节失败:", err)
	//		return false
	//	}
	//	log.Println("读到数量:", len(someBytes))
	//	_, err = f.Write(someBytes)
	//	if err != nil {
	//		log.Println("文件写入字节出错:", err)
	//		return false
	//	}
	//}

	return nil
}

// 往读写器中写入文件
func writeFile(rw *bufio.ReadWriter, path string) error {
	f, err := os.Open(path)
	defer f.Close()
	if err != nil {
		log.Println("打开文件出错:", err)
		return err
	}

	buf := make([]byte, 1048576)
	for {
		n, err := f.Read(buf)
		if err == io.EOF {
			log.Println("文件中读取字节完毕(EOF)")
			break
		} else if err != nil {
			log.Println("文件中读取字节失败:", err)
			return err
		}
		wn, err := rw.Write(buf[0:n])
		if err != nil {
			log.Println("流中写入字节失败:", err)
			return err
		}
		log.Println("流中写入字节数量:", wn)
		err = rw.Flush()
		if err != nil {
			log.Println("流中压出字节失败:", err)
			return err
		}
		err = rw.Flush()
		if err != nil {
			log.Println("流中压出字节失败:", err)
			return err
		}
	}

	return nil
}

// 处理进来的流
func handleStream(s network.Stream) {
	remoteId := s.Conn().RemotePeer().String()
	log.Println("处理新流:", remoteId)

	rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))

	// 读取
	jsonText, err := readText(rw)
	if err != nil {
		_ = s.Close()
		return
	}
	var message Message
	err = json.Unmarshal([]byte(jsonText), &message)
	if err != nil {
		log.Println("消息格式错误:", err)
		_ = s.Close()
		return
	}
	message.FromId = remoteId
	message.Ts = time.Now().UnixNano() / 1e6

	if message.Text != "" {
		// 读取文本
		log.Println("收到文本消息:", message.Text)
		// 存入信道等待提取
		messageChan <- message
	} else if message.FileSize != 0 {
		log.Println("文件消息")
		// 读取文件并保存
		err = readFileAndSave(rw, fmt.Sprint(fileDir, "/", message.FileName), message.FileSize)
		if err != nil {
			_ = s.Close()
			return
		}
		// 存入信道等待提取
		messageChan <- message
	}

	// 回复
	if message.Text == "" && message.FileSize == 0 {
		//返回我的信息
		jsonBytes, _ := json.Marshal(myInfo)
		_ = writeText(rw, string(jsonBytes))
	} else {
		//返回时间
		_ = writeText(rw, strconv.FormatInt(time.Now().UnixNano()/1e6, 10))
	}

	_ = s.Close()
}

// 初始节点
// privateKeyPath: 密钥文件路径
// publicKeyPath: 公钥文件路径
// fileDirectory: 文件目录(末尾不带斜杠/)
// name: 我的名字
// photo: 头像图片字节的Base64字符
func Init(privateKeyPath, publicKeyPath, fileDirectory, name, photo string) error {
	log.Println("初始:", fileDirectory, name)

	fileDir = fileDirectory

	//准备消息信道
	messageChan = make(chan Message, 100)
	ps = make(map[string]peer.AddrInfo)

	//准备上下文
	var cancel context.CancelFunc
	ctx, cancel = context.WithCancel(context.Background())
	defer cancel()

	//获取密钥
	privateKey, _ := getOrGenerateKey(privateKeyPath, publicKeyPath)
	log.Println("密钥类型:", privateKey.Type().String())

	var err error
	host, err = libp2p.New(
		ctx,
		libp2p.Identity(privateKey),
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

	//缓存我的信息
	myInfo = Info{host.ID().String(), name, photo}

	// 别人向你创建创建流时进行处理
	host.SetStreamHandler(ProtocolId, handleStream)

	// 创建mDNS服务并注册发现列队
	ser, err = discovery.NewMdnsService(ctx, host, time.Second*6, "mdns-test")
	if err != nil {
		return err
	}
	pn = &discoveryNotifee{}
	pn.PeerChan = make(chan peer.AddrInfo)
	ser.RegisterNotifee(pn)

	// 循环发现节点
	for p := range pn.PeerChan {
		if _, ok := ps[p.ID.String()]; !ok {
			log.Println("发现节点:", p.ID.String())
			ps[p.ID.String()] = p
		}
	}
	log.Println("停止发现节点")

	return nil
}

// 销毁
func Destroy() {
	log.Println("销毁")
	close(pn.PeerChan)
	_ = ser.Close()
	_ = host.Close()
}

// 获取自己ID
func GetMyId() string {
	return host.ID().String()
}

// 获取节点
// 说明: 因为Java不支持,只能弄成轮训模式.
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
func SendText(id, text string) (string, error) {
	log.Println("发送文本:", id, text)

	// 创建读写器
	rw, err := newReadWriter(id)
	if err != nil {
		return "", err
	}

	// 发出
	message := Message{Text: text}
	jsonBytes, err := json.Marshal(message)
	if err != nil {
		return "", err
	}
	err = writeText(rw, string(jsonBytes))
	if err != nil {
		return "", err
	}

	// 读取结果
	str, err := readText(rw)
	if err != nil {
		return "", err
	}
	return str, nil
}

// 发送文件
func SendFile(id, path string) (string, error) {
	log.Println("发送文件:", id, path)

	f, err := os.Open(path)
	defer f.Close()
	if err != nil {
		log.Println("打开文件出错:", err)
		return "", err
	}
	fs, err := f.Stat()
	if err != nil {
		log.Println("获取文件信息出错:", err)
		return "", err
	}

	// 创建读写器
	rw, err := newReadWriter(id)
	if err != nil {
		return "", err
	}

	// 发出
	message := Message{FileName: filepath.Base(path), FileSize: fs.Size()}
	jsonBytes, err := json.Marshal(message)
	if err != nil {
		return "", err
	}
	err = writeText(rw, string(jsonBytes))
	if err != nil {
		return "", err
	}
	err = writeFile(rw, path)
	if err != nil {
		return "", err
	}

	// 读取结果
	str, err := readText(rw)
	if err != nil {
		return "", err
	}
	return str, nil
}

// 获取信息
func GetInfo(id string) (string, error) {
	log.Println("获取信息:", id)

	// 创建读写器
	rw, err := newReadWriter(id)
	if err != nil {
		return "", err
	}

	// 发出
	message := Message{}
	jsonBytes, err := json.Marshal(message)
	if err != nil {
		return "", err
	}
	err = writeText(rw, string(jsonBytes))
	if err != nil {
		return "", err
	}

	// 读取结果
	str, err := readText(rw)
	if err != nil {
		return "", err
	}
	return str, nil
}

// 提取消息
// 说明: 因为Java不支持,只能弄成轮训模式. 消息信道只有100容量, 入股并发量大应提升提取频率.
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
