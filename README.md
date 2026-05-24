# muxconn

Go 语言流式多路复用库，将 [smux](https://github.com/xtaci/smux) 和 [yamux](https://github.com/hashicorp/yamux) 统一在同一个 `Muxer` 接口之下，可互换使用。

## 核心接口

- **`muxconn.Muxer`** — 多路复用会话，嵌入 `net.Listener`，提供 `Open()` / `Accept()` / `Close()` 等操作
- **`muxconn.Streamer`** — 虚拟子流，嵌入 `net.Conn`，额外提供 `Stats()` 获取流级别统计

## 功能特性

- **双后端支持** — smux 与 yamux 通过同一接口切换，可互为备选或交叉验证
- **速率限制** — 每个 Muxer 粒度的令牌桶限速，所有子流共享配额
- **流量统计** — 按 Muxer 聚合、按 Stream 细分的 RX/TX 字节计数
- **WebSocket 拨号** — 内置 `DialContext`，支持 `ws://` / `wss://`，自动协商协议
- **Context 生命周期** — 基于 context 的取消传播，支持优雅关闭

## 使用建议

smux 吞吐略高，但存在子流关闭信号不能及时感知的 bug；yamux 实现更稳定。两者可在系统中互为备选，也可用于交叉验证行为一致性。

## 使用示例

[github.com/xmx/muxconn-example](https://github.com/xmx/muxconn-example)
