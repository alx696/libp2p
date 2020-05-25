package main

import (
	"encoding/json"
	"github.com/alx696/go-mdns/im"
	"log"
	"time"
)

func main() {
	log.Println("在内网用mDNS发现节点并通信")

	//txt := dns.DigShort("iim.app.lilu.red", 16)
	//log.Println(txt)

	// /sdcard/android/data/red.lilu.red.iim/cache
	go im.Init("./config/private", "./config/public", "./config/file",
		"电脑", "[bytes base64]")

	//go func() {
	//	time.Sleep(time.Second * 6)
	//	im.Destroy()
	//}()

	//go func() {
	//	for {
	//		msg := im.ExtractMessage()
	//		log.Println("提取消息:", msg)
	//
	//		time.Sleep(time.Second * 3)
	//	}
	//}()

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
				//result, err := im.SendText(v, "你好")
				result, err := im.SendFile(v, "/home/m/下载/a.ttf")
				//infoStr, err := im.GetInfo(v)
				if err != nil {
					log.Println(err)
					continue
				}
				log.Println("通信结果:", result)
			}

			time.Sleep(time.Second * 6)
		}
	}()

	select {}
}
