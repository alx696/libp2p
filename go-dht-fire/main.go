package main

import (
	"flag"
	"github.com/alx696/libp2p/go-dht-fire/mp2p"
	"log"
)

// 参考 https://github.com/libp2p/go-libp2p-examples/blob/b7ac9e91865656b3ec13d18987a09779adad49dc/ipfs-camp-2019/06-Pubsub/main.go
func main() {
	log.Println("DHT星星之火")

	//指定端口,否则随机
	portFlag := flag.String("port", "0", "")
	//启发节点
	//必须是P2P地址, 即 https://github.com/multiformats/multiaddr#protocols (含/ipfs/Qm...)
	bootstrapFlag := flag.String("bootstrap", "", "")
	flag.Parse()

	mp2p.Init(*portFlag, *bootstrapFlag)
}
