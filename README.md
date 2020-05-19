# libp2p

https://libp2p.io

P2P网络的形成需要经过以下步骤:
1. 运行
2. 发现
3. 传输

## 发现

发现其它节点是P2P网络的开端，节点能够自行发现才能摆脱中央服务器。

局域网可以使用广播mDNS，互联网可以使用DHT。互联网上的发现显然比局域网麻烦，至少需要一个启发节点。

### mDNS

示例 [go-mdns-example]

### DHT

示例 [go-dht-fire]

## 传输

传输是P2P网络的最终目的，也是最复杂的部分。

### NAT穿越

特别是在中国，很少有设备能够分到互联网IP，需要使用[NAT穿越](https://docs.libp2p.io/reference/glossary/#nat-traversal)能够让不同局域网中的设备建立通信。

### 短路中继

很多设备位于防火墙之后或无法进行NAT穿越，对于这些设备需要使用[短路中继](https://docs.libp2p.io/reference/glossary/#circuit-relay)。



