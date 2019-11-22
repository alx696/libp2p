# NAT示例

目前有两个库可用:
* https://github.com/libp2p/go-nat
* https://github.com/libp2p/go-libp2p-nat

经过测试**go-nat**中的`DiscoverNATs(ctx)`网络兼容性较好.

> 测试发现使用DHT后好像不需要自己处理NAT, 直接通过节点ID即可创建流进行相互通信.

## 启动参数

**type** 不设置时使用`go-nat.DiscoverNATs(ctx)`, 设为`gateway`时使用`go-nat.DiscoverGateway()`, 设为`libp2p`时使用`go-libp2p-nat.DiscoverNAT(ctx)`.
