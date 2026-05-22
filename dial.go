package muxconn

import (
	"context"
	"crypto/tls"
	"errors"
	"log/slog"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

type DialOptions struct {
	// Dialer websocket 拨号器。
	Dialer *websocket.Dialer

	// PerTimeout 每次连接的超时时间。
	PerTimeout time.Duration

	// Logger 日志输出。
	Logger *slog.Logger
}

func DialContext(parent context.Context, addrs []string, opts *DialOptions) (Muxer, error) {
	if opts == nil {
		opts = new(DialOptions)
	}

	return opts.dialContext(parent, addrs)
}

func (d DialOptions) dialContext(parent context.Context, addrs []string) (Muxer, error) {
	if len(addrs) == 0 {
		return nil, errors.New("连接地址不能为空")
	}

	var errs []error
	for _, addr := range addrs {
		pu, err := url.Parse(addr)
		if err != nil {
			errs = append(errs, err)
			d.log().Warn("地址格式错误", "addr", addr, "err", err)
			continue
		}
		switch pu.Scheme {
		case "http":
			pu.Scheme = "ws"
		case "https":
			pu.Scheme = "wss"
		}

		mux, err2 := d.protocolDials(parent, pu)
		if len(err2) != 0 {
			errs = append(errs, err2...)
			d.log().Warn("连接失败", "addr", addr, "err", err2)
			continue
		}

		proto, module := mux.Library()
		d.log().Info("连接成功", "addr", addr, "protocol", proto, "module", module)

		return mux, nil
	}

	err := errors.Join(errs...)
	d.log().Error("全部连接失败", "addr", addrs, "err", err)

	return nil, err
}

func (d DialOptions) protocolDials(parent context.Context, dest *url.URL) (Muxer, []error) {
	query := dest.Query()
	proto := strings.ToLower(query.Get("protocol"))
	var protocols []string
	switch proto {
	case "smux", "yamux": // 目前仅支持这些协议
		protocols = append(protocols, proto)
	default:
		// 如果用户指定了错误的协议或未指定协议，则将协议都尝试一遍。
		protocols = []string{"smux", "yamux"}
	}

	var errs []error
	for _, protocol := range protocols {
		query.Set("protocol", protocol)
		dest.RawQuery = query.Encode()

		mux, err := d.dialOne(parent, dest, protocol)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		return mux, nil
	}

	return nil, errs
}

func (d DialOptions) dialOne(parent context.Context, dest *url.URL, proto string) (Muxer, error) {
	strURL := dest.String()
	dialer := d.websocketDialer()

	ctx, cancel := context.WithTimeout(parent, d.perTimeout())
	defer cancel()

	ws, _, err := dialer.DialContext(ctx, strURL, nil)
	if err != nil {
		return nil, err
	}

	var mux Muxer
	conn := ws.NetConn()

	switch proto {
	case "smux":
		mux, err = NewSMUX(parent, conn, nil, false)
	default:
		mux, err = NewYaMUX(parent, conn, nil, false)
	}

	if err != nil {
		_ = ws.Close()
		return nil, err
	}

	return mux, nil
}

func (d DialOptions) log() *slog.Logger {
	if d.Logger != nil {
		return d.Logger
	}

	return slog.Default()
}

func (d DialOptions) websocketDialer() *websocket.Dialer {
	if d.Dialer != nil {
		return d.Dialer
	}

	return &websocket.Dialer{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
}

func (d DialOptions) perTimeout() time.Duration {
	if d.PerTimeout > 0 {
		return d.PerTimeout
	}

	return 10 * time.Second
}
