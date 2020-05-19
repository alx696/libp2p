# mdns示例

https://github.com/libp2p/go-libp2p-examples/tree/master/chat-with-mdns

多播DNS，用于在局域网中通过广播申明自己的存在。

## 测试

分别构建两个程序, 然后局域网中运行两个程序, 连接成功后即可在控制台互发消息(enter键发送).

构建两个程序
```bash
$ go build -o 1
$ go build -o 2
```

运行程序1
```bash
$ ./1
2020/05/19 17:10:33 在内网用mDNS发现节点
2020/05/19 17:10:33 我的节点地址: [/ip6/::1/tcp/36143/p2p/QmTgPSCerkMhc5nPjjEZzQ7dy4sgKT8JN16qm4gbY3skx5 /ip4/127.0.0.1/tcp/43775/p2p/QmTgPSCerkMhc5nPjjEZzQ7dy4sgKT8JN16qm4gbY3skx5 /ip4/192.168.1.200/tcp/43775/p2p/QmTgPSCerkMhc5nPjjEZzQ7dy4sgKT8JN16qm4gbY3skx5 /ip4/192.168.122.1/tcp/43775/p2p/QmTgPSCerkMhc5nPjjEZzQ7dy4sgKT8JN16qm4gbY3skx5 /ip4/172.17.0.1/tcp/43775/p2p/QmTgPSCerkMhc5nPjjEZzQ7dy4sgKT8JN16qm4gbY3skx5]
2020/05/19 17:10:33 发现节点地址: [/ip4/172.17.0.1/tcp/36143/p2p/QmbtqzBBds2gdCH29HudWfFFyWKG859TGpmuepmicJc45g]
2020/05/19 17:10:33 发现节点地址: [/ip4/172.17.0.1/tcp/36143/p2p/QmbtqzBBds2gdCH29HudWfFFyWKG859TGpmuepmicJc45g]
```

运行程序2
```bash
$ ./2
2020/05/19 17:10:52 在内网用mDNS发现节点
2020/05/19 17:10:52 我的节点地址: [/ip4/127.0.0.1/tcp/33461/p2p/QmZXZ6h8cY1owXFcXRG981pVPY7z3PySd2KUHPubV5yyCH /ip4/192.168.1.200/tcp/33461/p2p/QmZXZ6h8cY1owXFcXRG981pVPY7z3PySd2KUHPubV5yyCH /ip4/192.168.122.1/tcp/33461/p2p/QmZXZ6h8cY1owXFcXRG981pVPY7z3PySd2KUHPubV5yyCH /ip4/172.17.0.1/tcp/33461/p2p/QmZXZ6h8cY1owXFcXRG981pVPY7z3PySd2KUHPubV5yyCH /ip6/::1/tcp/43427/p2p/QmZXZ6h8cY1owXFcXRG981pVPY7z3PySd2KUHPubV5yyCH]
2020/05/19 17:10:52 发现节点地址: [/ip4/172.17.0.1/tcp/36143/p2p/QmTgPSCerkMhc5nPjjEZzQ7dy4sgKT8JN16qm4gbY3skx5]
2020/05/19 17:10:52 发现节点地址: [/ip4/172.17.0.1/tcp/36143/p2p/QmTgPSCerkMhc5nPjjEZzQ7dy4sgKT8JN16qm4gbY3skx5]
```