// Package muxconn 封装网络多路复用能力，提供统一的 Muxer 接口。
//
// 经自测，smux 性能略高于 yamux 通道，但 smux 存在一个 BUG，虚拟子流关闭信号
// 不能及时感知，应该是 smux 项目本身的问题。
// 其它说明：
//   - smux 性能略高于 yamux （千兆网卡自测），但子流关闭信号双端无法感知（疑似 smux 实现缺陷）。
//   - yamux 实现更稳定（仅个人倾向），但 smux 项目更活跃。
//
// smux 和 yamux 两者在系统中可互为备选，也可用于交叉验证连接行为一致性。
package muxconn
