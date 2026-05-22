package muxconn_test

import (
	"context"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/xmx/muxconn"
)

func TestServer(t *testing.T) {
	wsu := &websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
	expose := &exposeHandler{wsu: wsu, virtual: new(serverVirtual)}
	route := http.NewServeMux()
	route.HandleFunc("/api/tunnel", expose.accept)

	http.ListenAndServe(":9999", route)
}

type exposeHandler struct {
	wsu     *websocket.Upgrader
	virtual http.Handler
}

func (e *exposeHandler) accept(w http.ResponseWriter, r *http.Request) {
	ws, err := e.wsu.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	query := r.URL.Query()
	conn := ws.NetConn()

	var mux muxconn.Muxer
	switch strings.ToLower(query.Get("protocol")) {
	case "smux":
		mux, err = muxconn.NewSMUX(nil, conn, nil, true)
	default:
		mux, err = muxconn.NewYaMUX(nil, conn, nil, true)
	}
	if err != nil {
		_ = conn.Close()
		return
	}

	vsrv := &http.Server{Handler: e.virtual}
	cli := &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return mux.Open(ctx)
			},
		},
	}
	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()

		for range ticker.C {
			resp, _ := cli.Get("http://victual.internal/api/ping")
			if resp != nil {
				io.Copy(os.Stdout, resp.Body)
				resp.Body.Close()
			}
		}

	}()

	vsrv.Serve(mux)
}

type serverVirtual struct{}

func (s *serverVirtual) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("HELLO, 我是 server。"))
}
