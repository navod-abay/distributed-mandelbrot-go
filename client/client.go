package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"os"
	"runtime"
	"strconv"
	"sync"
	"time"

	sharedproto "github.com/navod-abay/mandelbrotset-go/shared_proto"
)

const port string = ":8080"

func handleConnection(conn net.Conn, lock *sync.Mutex) error {
	lock.Lock() // only one master can push work into the client
	defer lock.Unlock()
	fmt.Println("Lock Acquired")
	buffReader := bufio.NewReader(conn)
	conn.SetReadDeadline(time.Now().Add(20 * time.Second))
	err := readHandshakeRequest(buffReader)
	if err != nil {
		fmt.Printf("%v", err.Error())
		return err
	}
	buffWriter := bufio.NewWriter(conn)
	sendClientHandshakeResponse(buffWriter)
	return nil
}

func readHandshakeRequest(buff *bufio.Reader) error {
	slog.Debug("Waiting for the handshake request")
	requestType, contentBytes, err := sharedproto.ReadMessage(buff)
	if err != nil || requestType != sharedproto.HandShake {
		return sharedproto.ErrIncorrectMessage
	}
	slog.Debug("Handshake request successfully parsed.", "Extra Bytes", contentBytes.String())
	return nil
}

func sendClientHandshakeResponse(buffWriter *bufio.Writer) error {
	fmt.Println("Sending Handshake Res")
	var buff bytes.Buffer
	cpus := runtime.NumCPU()
	buff.WriteString("\nNumCPU: " + strconv.Itoa(cpus))
	err := sharedproto.SendMessage(buffWriter, buff.Bytes(), "handshakeResponse")
	return err
}

func KeepListening(ln net.Listener, wg *sync.WaitGroup) {
	defer wg.Done()
	var lock sync.Mutex
	keepConnection := true
	for keepConnection {
		conn, err := ln.Accept() // Accept a connection
		if err != nil {
			fmt.Printf("Error accepting TCP connection, err: %v ", err)
			continue
		}
		go handleConnection(conn, &lock)

	}
}

func main() {
	var opts = &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}
	var handler = slog.NewTextHandler(os.Stdout, opts)
	var logger = slog.New(handler)
	slog.SetDefault(logger)

	isCustomProto := true
	flag.Func("protocolType", "Spcify whether to use RPC or the custom protocol for messaging", func(val string) error {
		slog.Debug("Parsing flags", "isCustomProto", val)
		if val != "Custom" {
			if val != "RPC" {
				return fmt.Errorf("Invalid Protocol Name. Only accepts 'RPC' or 'Custom'. Defaulting to custom protocol")
			}
			isCustomProto = false
		}
		return nil
	})
	port := flag.Int("port", 8080, "Port to run on")
	flag.Parse()
	if !isCustomProto {
		rpcClientFlow(*port)

	} else {
		ln, err := net.Listen("tcp", ":"+strconv.Itoa(*port))

		if err != nil {
			fmt.Printf("Couldn't start listening on port :8080. Error: %v", err)
		}
		var wg = new(sync.WaitGroup)
		wg.Add(1)
		go KeepListening(ln, wg)
		wg.Wait()
	}

}
