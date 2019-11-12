# DHT连接 示例

Kademlia-based DHT，用于在互联网（局域网）中通过启发节点互相发现。连接启发节点或通过DHT发现节点时，通过创建流发送和接受数据。

示例在[go-dht-example](https://github.com/alx696/libp2p/tree/master/go-dht-example)的基础上增加了密钥生成、QUIC传输、连接节点、创建流发送数据、接受流读取数据的示例。

## 测试

构建程序, 然后局域网中运行两个程序, 观察控制台输出信息.

构建程序
```bash
$ go build -o dht
```

运行程序1
```bash
$ ./dht
2019/11/11 15:17:02 节点P2P地址: [/ip4/127.0.0.1/tcp/38133/ipfs/QmS1BCwGa4yCTcaRsg6xopoL2vwso5ZHWP2QetAUEk9ohX /ip4/192.168.1.200/tcp/38133/ipfs/QmS1BCwGa4yCTcaRsg6xopoL2vwso5ZHWP2QetAUEk9ohX /ip4/192.168.122.1/tcp/38133/iQmS1BCwGa4yCTcaRsg6xopoL2vwso5ZHWP2QetAUEk9ohX /ip4/172.17.0.1/tcp/38133/ipfs/QmS1BCwGa4yCTcaRsg6xopoL2vwso5ZHWP2QetAUEk9ohX /ip6/::1/tcp/37939/ipfs/QmS1BCwGa4yCTcaRsg6xopoL2vwso5ZHWP2QetAUEk9ohX]
```

运行程序2
```bash
$ ./dht --bootstrap="/ip4/127.0.0.1/tcp/38133/ipfs/QmS1BCwGa4yCTcaRsg6xopoL2vwso5ZHWP2QetAUEk9ohX"
2019/11/11 17:23:56 已经发送
2019/11/11 17:23:58 处理流: /ip4/27.17.7.86/udp/22509/quic
2019/11/11 17:23:58 读取内容: 你好, DHT节点
2019/11/11 17:23:58 结果已经返回
```

> 注意: 互联网中必须出现至少一次交叉连接, 节点才能相互发现. 比如节点a连启发节点, 节点b连启发节点, 则节点a和b无法相互发现. 但如果节点a连启发节点, 节点b连a, 则3个节点可以相互发现.

如果需要互联网中节点能够稳定相互发现, 至少需要2个启发节点. 启发节点a首先启动, 让启发节点b连接启发节点a, 让其它节点连接启发节点b.