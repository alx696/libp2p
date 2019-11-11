# DHT连接 示例

Kademlia-based DHT，用于在互联网（局域网）中通过启发节点互相发现。连接启发节点或通过DHT发现节点时，通过创建流发送和接受数据。

示例在[go-dht-example](https://github.com/alx696/libp2p/tree/master/go-dht-example)的基础上增加了密钥生成、QUIC传输、连接节点、创建流发送数据、接受流读取数据的示例。

## 测试

构建程序, 然后局域网中运行两个程序, 观察控制台输出信息.

构建程序
```bash
$ go build -o 1
```

运行程序1
```bash
$ ./1
2019/11/11 15:17:02 节点P2P地址: [/ip4/127.0.0.1/tcp/38133/ipfs/QmS1BCwGa4yCTcaRsg6xopoL2vwso5ZHWP2QetAUEk9ohX /ip4/192.168.1.200/tcp/38133/ipfs/QmS1BCwGa4yCTcaRsg6xopoL2vwso5ZHWP2QetAUEk9ohX /ip4/192.168.122.1/tcp/38133/iQmS1BCwGa4yCTcaRsg6xopoL2vwso5ZHWP2QetAUEk9ohX /ip4/172.17.0.1/tcp/38133/ipfs/QmS1BCwGa4yCTcaRsg6xopoL2vwso5ZHWP2QetAUEk9ohX /ip6/::1/tcp/37939/ipfs/QmS1BCwGa4yCTcaRsg6xopoL2vwso5ZHWP2QetAUEk9ohX]
```

运行程序2
```bash
$ ./1 --bootstrap="/ip4/127.0.0.1/tcp/38133/ipfs/QmS1BCwGa4yCTcaRsg6xopoL2vwso5ZHWP2QetAUEk9ohX"
2019/11/11 17:23:56 已经发送
2019/11/11 17:23:58 处理流: /ip4/27.17.7.86/udp/22509/quic
2019/11/11 17:23:58 读取内容: 你好, DHT节点
2019/11/11 17:23:58 结果已经返回
```

> 如果需要需要在互联网上测试，请确认路由器上做了局域网IP和端口的映射，将启动参数中IP修改为对应互联网IP即可。