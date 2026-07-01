package main

import (
	"fmt"
	"log/slog"
	"net"
	"net/rpc"
	"strconv"
	"sync"

	sharedproto "github.com/navod-abay/mandelbrotset-go/shared_proto"
)

func serveRPC(ln net.Listener, wg *sync.WaitGroup) {
	defer wg.Done()
	keepConnection := true
	for keepConnection {
		conn, err := ln.Accept() // Accept a connection
		if err != nil {
			fmt.Printf("Error accepting TCP connection, err: %v ", err)
			continue
		}
		slog.Debug("Starting RPC server")
		rpc.ServeConn(conn)
		slog.Debug("Finished handling RPC")

	}
}

func rpcClientFlow(port int) {
	server := new(sharedproto.RpcServer)
	server.Port = port
	rpc.Register(server)

	ln, err := net.Listen("tcp", ":"+strconv.Itoa(port))

	if err != nil {
		fmt.Printf("Couldn't start listening on port :8080. Error: %v", err)
	}
	slog.Debug("Started Listening", "port", port)
	var wg = new(sync.WaitGroup)
	wg.Add(1)
	go serveRPC(ln, wg)
	wg.Wait()
}
