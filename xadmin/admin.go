package xadmin

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
)

var gServer *http.Server

func StartAdminAt(addr string) {
	if addr == "" {
		fmt.Printf("invalid admin addr(%s)\n", addr)
		return
	}

	go func(sAddr string) {
		if gServer != nil {
			fmt.Println("close default server at:", gServer.Addr)
			gServer.Close()
			gServer = nil
		}
		fmt.Printf("start admin server at %s\n", sAddr)
		gServer = &http.Server{Addr: sAddr, Handler: nil}
		gServer.ListenAndServe()
	}(addr)
}