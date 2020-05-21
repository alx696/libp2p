package main

import (
	"encoding/json"
	"github.com/alx696/go-mdns/im"
	"log"
	"time"
)

func main() {
	log.Println("在内网用mDNS发现节点并通信")

	// /sdcard/android/data/red.lilu.red.iim/cache
	go im.Init("./config", im.Info{Name: "电脑", Photo: "[base64]"})

	go func() {
		for {
			msg := im.ExtractMessage()
			log.Println("提取消息:", msg)

			time.Sleep(time.Second * 3)
		}
	}()

	go func() {
		for {
			ps := im.FindPeer()
			log.Println("现有节点:", ps)

			var ids []string
			err := json.Unmarshal([]byte(ps), &ids)
			if err != nil {
				log.Println(err)
				continue
			}
			for _, v := range ids {
				//err := im.SendText(v, "你好")
				//err := im.SendFile(v, "/home/km/下载/s.txt")
				info, err := im.GetInfo(v)
				if err != nil {
					log.Println(err)
					continue
				}
				log.Println("信息:", *info)
			}

			time.Sleep(time.Second * 6)
		}
	}()

	select {}
}
