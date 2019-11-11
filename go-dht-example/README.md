# DHT 示例

Kademlia-based DHT，用于在互联网（局域网）中通过启发节点互相发现。

## 测试

构建程序, 然后局域网中运行三个程序, 观察控制台输出信息.

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
2019/11/11 15:17:48 DHT发现节点
2019/11/11 15:17:48 节点P2P地址: [/ip6/::1/tcp/33151/ipfs/QmXfzBJazCZ1kaJmmHJdExGiJ6xM8m63XdRr2eCmdsdWYq /ip4/127.0.0.1/tcp/33831/ipfs/QmXfzBJazCZ1kaJmmHJdExGiJ6xM8m63XdRr2eCmdsdWYq /ip4/192.168.1.200/tcp/33831/ipfs/QmXfzBZ1kaJmmHJdExGiJ6xM8m63XdRr2eCmdsdWYq /ip4/192.168.122.1/tcp/33831/ipfs/QmXfzBJazCZ1kaJmmHJdExGiJ6xM8m63XdRr2eCmdsdWYq /ip4/172.17.0.1/tcp/33831/ipfs/QmXfzBJazCZ1kaJmmHJdExGiJ6xM8m63XdRr2eCmdsdWYq]
2019/11/11 15:17:48 已经连接节点: /ip4/127.0.0.1/tcp/38133/ipfs/QmS1BCwGa4yCTcaRsg6xopoL2vwso5ZHWP2QetAUEk9ohX
```

运行程序3
```bash
$ ./1 --bootstrap="/ip6/::1/tcp/33151/ipfs/QmXfzBJazCZ1kaJmmHJdExGiJ6xM8m63XdRr2eCmdsdWYq"
2019/11/11 15:18:34 DHT发现节点
2019/11/11 15:18:35 节点P2P地址: [/ip4/127.0.0.1/tcp/42141/ipfs/QmY8SKqdd4HpqtUvwCKmgsH9XnYLq6Sspyfz3WcHxDVPRB /ip4/192.168.1.200/tcp/42141/ipfs/QmY8SKqdd4HpqtUvwCKmgsH9XnYLq6Sspyfz3WcHxDVPRB /ip4/192.168.122.1/tcp/42141/iQmY8SKqdd4HpqtUvwCKmgsH9XnYLq6Sspyfz3WcHxDVPRB /ip4/172.17.0.1/tcp/42141/ipfs/QmY8SKqdd4HpqtUvwCKmgsH9XnYLq6Sspyfz3WcHxDVPRB /ip6/::1/tcp/40309/ipfs/QmY8SKqdd4HpqtUvwCKmgsH9XnYLq6Sspyfz3WcHxDVPRB]
2019/11/11 15:18:35 已经连接节点: /ip6/::1/tcp/33151/ipfs/QmXfzBJazCZ1kaJmmHJdExGiJ6xM8m63XdRr2eCmdsdWYq
```

观察程序1的控制台输出:
```bash
2019/11/11 15:19:29 DHT发现节点: {QmY8SKqdd4HpqtUvwCKmgsH9XnYLq6Sspyfz3WcHxDVPRB: [/ip4/127.0.0.1/tcp/42141 /ip4/192.168.1.200/tcp/42141 /ip4/192.168.122.1/tcp/42141 /ip4/172.17.0.1/tcp/42141 /ip6/::1/tcp/40309]}
2019/11/11 15:19:29 DHT发现节点: {QmXfzBJazCZ1kaJmmHJdExGiJ6xM8m63XdRr2eCmdsdWYq: [/ip4/192.168.1.200/tcp/33831 /ip4/192.168.122.1/tcp/33831 /ip4/172.17.0.1/tcp/33831 /ip6/::1/tcp/33151 /ip4/127.0.0.1/tcp/33831]}
2019/11/11 15:19:29 ---
```

程序2只连接了程序1，程序3只连接了程序2，但程序1能通过程序2发现程序3。节点之间通过DHT进行了相互传播，最终所有相互连接的节点都能知道彼此的存在。
> 如果需要需要在互联网上测试，请确认路由器上做了局域网IP和端口的映射，将启动参数中IP修改为对应互联网IP即可。