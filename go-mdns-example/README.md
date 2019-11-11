# mdns-discovery 示例

https://github.com/libp2p/go-libp2p/tree/master/p2p/discovery

多播DNS，用于在局域网中通过广播申明自己的存在。

## 测试

分别构建两个程序, 然后局域网中运行两个程序, 观察控制台输出信息.

构建两个程序
```bash
$ go build -o 1
$ go build -o 2
```

运行程序1
```bash
$ ./1
2019/11/11 14:02:39 内网用mDNS发现节点
2019/11/11 14:02:45 发现节点: {QmSAfxecscAx1nnjMzLPtcaRPhYUxXAqnm2NuAbU2K9t9M: [/ip4/172.17.0.1/tcp/40875]}
2019/11/11 14:02:45 发现节点: {QmSAfxecscAx1nnjMzLPtcaRPhYUxXAqnm2NuAbU2K9t9M: [/ip4/172.17.0.1/tcp/40875]}
```

运行程序2
```bash
$ ./2
2019/11/11 14:02:44 内网用mDNS发现节点
2019/11/11 14:02:44 发现节点: {QmXAy7X2mpzM3FbPw9XXKLUWMWGgkFPevpwuUs1wrxyd1q: [/ip4/172.17.0.1/tcp/36203]}
2019/11/11 14:02:44 发现节点: {QmXAy7X2mpzM3FbPw9XXKLUWMWGgkFPevpwuUs1wrxyd1q: [/ip4/172.17.0.1/tcp/36203]}
```