# Kademlia-based DHT启发节点

本身不提供任何功能，只用于发现节点。

## 开发

### 编译Android aar包

编译aar方法见[文档](https://github.com/alx696/share/wiki/Go-mobile-(Android)#%E5%AE%89%E8%A3%85%E5%92%8C%E5%88%9D%E5%A7%8B)

注意:
* gomobile不支持module, 文档中已有提到
* Android 9需要申请写外部存储权限才能写文件(之前无需申请)

## 用法

### 启动启发节点A

```bash
$ ./dht --port=60000
```

记录节点地址，例如 `/ip4/127.0.0.1/udp/60000/quic/ipfs/QmXDunpuNNS93eCEv66UnAzuMBgdENZY7MSE3TuhXNtEjv` ，将 `127.0.0.1` 替换成互联网IP。

### 启动启发节点B

```bash
./dht --port=60000 --bootstrap=/ip4/启发节点A的IP/udp/60000/quic/ipfs/QmXDunpuNNS93eCEv66UnAzuMBgdENZY7MSE3TuhXNtEjv
```

让启发节点B启动后立即连接启发节点A，这样启发节点A和B的组网就完成了。

### 将启发节点B作为引导节点

此时其它节点启动时以启发节点B作为引导节点，这样所有节点就能互相发现彼此。

## 注意

经过测试，互联网中发现节点需要至少2个启发节点。启发节点a首先启动，让启发节点b连接启发节点a，让其它节点连接启发节点b。

如果只有一个启发节点，节点a和b都连接启发节点，那么a和b之间是无法相互发现的。