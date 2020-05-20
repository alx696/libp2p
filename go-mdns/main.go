package main

import (
	"bufio"
	"context"
	"fmt"
	"github.com/alx696/go-mdns/im"
	host2 "github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"io"
	"io/ioutil"
	"log"
	"strings"
	"time"
)

var ps map[string]peer.AddrInfo

func readText(rw *bufio.ReadWriter) string {
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

func writeText(rw *bufio.ReadWriter, text string) bool {
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

func readFile(rw *bufio.ReadWriter, path string) bool {
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

func writeFile(rw *bufio.ReadWriter, path string) bool {
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

func handleStream(s network.Stream) {
	log.Println("处理新流")

	rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))

	// 读取
	messageType := readText(rw)
	if messageType == "文本" {
		log.Println("文本消息")
		// 读取文本
		messageText := readText(rw)
		log.Println(messageText)
	} else if messageType == "文件" {
		log.Println("文件消息")
		// 读取文件
		readFile(rw, "/home/km/下载/r.txt")
	}

	// 回复2次
	for i := 0; i < 2; i++ {
		writeText(rw, time.Now().String())
	}

	_ = s.Close()
}

func main() {
	log.Println("在内网用mDNS发现节点并通信")

	ps = make(map[string]peer.AddrInfo)

	im.Init(handleStream,
		func(ctx context.Context, host host2.Host, p peer.AddrInfo) error {
			if _, ok := ps[p.ID.String()]; !ok {
				log.Println("发现节点:", p.ID.String())
				ps[p.ID.String()] = p

				//连接
				rw, err := im.Connect(ctx, host, p)
				if err != nil {
					log.Println("连接失败:", err)
					return err
				}

				//发送
				//writeText(rw, "文本")
				//writeText(rw, "你好")
				writeText(rw, "文件")
				writeFile(rw, "/home/km/下载/s.txt")
				//等待回复
				for {
					str, err := rw.ReadString('\n')
					if err != nil {
						if err != io.EOF {
							log.Println("读取文本出错:", err)
						} else {
							log.Println("读取文本完毕")
						}
						break
					}
					str = strings.Replace(str, "\n", "", -1)
					log.Println("收到回复:", str)
				}
			}

			return nil
		})
}
