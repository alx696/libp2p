# NAT示例

目前有两个库可用:
* https://github.com/libp2p/go-nat
* https://github.com/libp2p/go-libp2p-nat

经过测试**go-nat**中的`DiscoverNATs(ctx)`网络兼容性较好.

## 启动参数

**type** 不设置时使用`go-nat.DiscoverNATs(ctx)`, 设为`libp2p`时使用`go-libp2p-nat.DiscoverNAT(ctx)`, 设为`gateway`时使用`go-nat.DiscoverGateway()`.