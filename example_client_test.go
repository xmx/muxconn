package muxconn_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/xmx/muxconn"
)

func TestClient(t *testing.T) {
	parent := context.Background()
	addrs := []string{"http://localhost:9999/api/tunnel?protocol=yamux"}
	mux, err := muxconn.DialContext(parent, addrs, nil)

	if err != nil {
		t.Fatal(err)
	}
	defer mux.Close()

	route := http.NewServeMux()
	route.HandleFunc("/api/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("HELLO 我是客户端 Server\n"))
	})

	vitual := &http.Server{Handler: route}
	vitual.Serve(mux)
}
