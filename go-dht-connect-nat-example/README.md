# DHT连接NAT穿越 示例

Kademlia-based DHT，用于在互联网（局域网）中通过启发节点互相发现。连接启发节点或通过DHT发现节点时，通过创建流发送和接受数据。

示例在[go-dht-connect-example](https://github.com/alx696/libp2p/tree/master/go-dht-connect-example)的基础上增加了NAT穿越的示例。

## 测试

构建程序, 然后局域网中运行两个程序, 观察控制台输出信息.

构建程序
```bash
$ go build -o 1
```

运行程序1
```bash
2019/11/11 17:48:47 HT连接NAT穿越
2019/11/11 17:48:47 节点P2P地址: [/ip4/127.0.0.1/udp/60000/quic/ipfs/QmbHiUwXw9SDNys9aZ486sD4Qauca63XcGY9ES3GwBWNBU /ip4/192.168.1.200/udp/60000/quic/ipfs/QmbHiUwXw9SDNys9aZ486sD4Qauca63XcGY9ES3GwBWNBU /ip4/192.168.122.1/udp/60000/quic/ipfs/QmbHiUwXw9SDNys9aZ486sD4Qauca63XcGY9ES3GwBWNBU /ip4/172.17.0.1/udp/60000/quic/ipfs/QmbHiUwXw9SDNys9aZ486sD4Qauca63XcGY9ES3GwBWNBU /ip6/::1/udp/60000/quic/ipfs/QmbHiUwXw9SDNys9aZ486sD4Qauca63XcGY9ES3GwBWNBU]
2019/11/11 17:48:47 准备NAT穿越
2019/11/11 17:48:47 已连节点数量: 0
2019/11/11 17:48:49 NAT网关公网IP: 8.8.8.8
2019/11/11 17:48:49 NAT已经映射内网端口: 60000 为: 42676
```
> 程序1运行在局域网中！

远端运行程序2
```bash
$ ./1 --bootstrap="/ip4/8.8.8.8/udp/42676/quic/ipfs/QmbHiUwXw9SDNys9aZ486sD4Qauca63XcGY9ES3GwBWNBU"
2019/11/11 17:49:06 DHT发现节点
2019/11/11 17:49:06 节点P2P地址: [/ip4/127.0.0.1/udp/40933/quic/ipfs/QmXDunpuNNS93eCEv66UnAzuMBgdENZY7MSE3TuhXv /ip4/192.168.0.3/udp/40933/quic/ipfs/QmXDunpuNNS93eCEv66UnAzuMBgdENZY7MSE3TuhXNtEjv /ip4/172.18.0.1/udp/40933/quic/ipfs/QmXDunpuNNS93eCEv66UnAzuMBgdENZY7MSE3TuhXNtEjv /ip4/172.17.0.1/udp/40933/quic/ipfs/QmXDunpuNNS93eCEv66UnAzuMBgdENZY7MSE3TuhXNtEjv /ip6/::1/udp/34411/quic/ipfs/QmXDunpuNNS93eCEv66UnAzuMBgdENZY7MSE3TuhXNtEjv]
2019/11/11 17:49:06 已经连接节点: /ip4/8.8.8.8/udp/42676/quic/ipfs/QmbHiUwXw9SDNys9aZ486sD4Qauca63XcGY9ES3GwBWNBU
2019/11/11 17:49:06 已经发送
2019/11/11 17:49:08 处理流: /ip4/8.8.8.8/udp/42676/quic
2019/11/11 17:49:08 读取内容: 你好, DHT节点
2019/11/11 17:49:08 结果已经返回
```
> 程序2运行在其它局域网中，通过互联网用程序1的`NAT已经映射内网端口`连接。
