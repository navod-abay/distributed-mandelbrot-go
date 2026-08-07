package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"net/rpc"
	"os"
	"strconv"
	"strings"
	"sync"

	generate "github.com/navod-abay/mandelbrotset-go/core"
	"github.com/navod-abay/mandelbrotset-go/core/models"
	"github.com/navod-abay/mandelbrotset-go/core/solvers"
	sharedproto "github.com/navod-abay/mandelbrotset-go/shared_proto"
)

type ClientIdentifier struct {
	id           string
	numProcesses int
	conn         net.Conn
}

type RPCClientIdentifier struct {
	id           int
	numProcesses int16
	client       *rpc.Client
}

func sendHandshakeRequest(conn *bufio.Writer, id int) error {
	slog.Debug("Sending Handshake Request", "id:", id)
	err := sharedproto.SendMessage(conn, []byte{}, "handshake")
	slog.Debug("Handshake Request Sent")
	return err
}

func readHandshakeResponse(conn *bufio.Reader) (int, error) {
	requestType, bytes, err := sharedproto.ReadMessage(conn)
	if err != nil || requestType != sharedproto.HandshakeResponse {
		return 0, sharedproto.ErrIncorrectMessage
	}
	slog.Debug("Recieved Handshake Response", "received Extra content", bytes.String())
	// conn.Reset(conn)
	contentMap, err := sharedproto.ContentDeserialization(bytes)
	numProcesses, err := strconv.Atoi(contentMap["NumCPU"])
	return numProcesses, err
}

func handleConnection(conn net.Conn, id int, wg *sync.WaitGroup, c chan ClientIdentifier) {
	defer wg.Done()
	fmt.Printf("Connected with Id: %v\n", id)
	buffConn := bufio.NewWriter(conn)
	sendHandshakeRequest(buffConn, id)

	bufferedReader := bufio.NewReader(conn)
	slog.Debug("Gonna wait for the Handshake response")
	numProcesses, err := readHandshakeResponse(bufferedReader)
	if err != nil {
		slog.Error(err.Error())
		return
	}
	clientIdentity := ClientIdentifier{id: strconv.Itoa(id), numProcesses: numProcesses, conn: conn}
	slog.Debug("Going to write the client identity to the channel", "clientIdentity", clientIdentity)
	c <- clientIdentity
	slog.Debug("Finished initiating connection", "clientIdentity", clientIdentity)
}

func sendWorkRequest(id string, conn net.Conn, subImages []models.ImageDimensions, subdivision_levels int, wg *sync.WaitGroup) {
	defer wg.Done()
	fmt.Println("Sending work request to: " + id + "\n")
	buffWriter := bufio.NewWriter(conn)
	content := map[string]string{
		"id":                 id,
		"subdivision_levels": strconv.Itoa(subdivision_levels),
	}
	byteBuffer := new(bytes.Buffer)
	sharedproto.ContentSerialization(byteBuffer, content)
	for i, subimage := range subImages {
		key := "subimage_" + strconv.Itoa(i) + ":"
		byteBuffer.WriteString(key)
		err := binary.Write(byteBuffer, binary.LittleEndian, subimage)
		byteBuffer.WriteString("\n")
		if err != nil {
			fmt.Println(err)
			continue
		}
	}
	err := sharedproto.SendMessage(buffWriter, byteBuffer.Bytes(), "delegateWork")
	if err != nil {
		fmt.Println(err)
	}

}

func delegate(subImages []models.ImageDimensions, subdivision_levels int, clients map[string]ClientIdentifier) {
	index := 0
	var wg sync.WaitGroup
	for id, client := range clients {
		wg.Add(1)
		go sendWorkRequest(id, client.conn, subImages[index:index+client.numProcesses], subdivision_levels, &wg)
	}
	wg.Wait()
	fmt.Println("Finished delegating work")
}

func main() {
	var opts = &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}

	var handler = slog.NewTextHandler(os.Stdin, opts)
	var logger = slog.New(handler)

	isCustomProto := true
	flag.Func("protocolType", "Spcify whether to use RPC or the custom protocol for messaging", func(val string) error {
		val = strings.ToLower(val)
		if val != "custom" {
			if val != "rpc" {
				return fmt.Errorf("Invalid Protocol Name. Only accepts 'RPC' or 'Custom'. Defaulting to custom protocol")
			}
			isCustomProto = false
		}
		return nil
	})

	flag.Parse()

	slog.SetDefault(logger)
	fmt.Printf("Running master node\n")

	fileBytes, err := os.ReadFile("terraform_out.json")
	var IPs []string
	if err != nil {
		slog.Debug("Can't find the terraform output file. Defaulting to hardcoded values")
		IPs = []string{"127.0.0.1:8080", "127.0.0.1:8081"}
	} else if json.Unmarshal(fileBytes, &IPs) != nil {
		slog.Debug("Error unmarshaling JSON values. Defaulting to hardcoded IP values")
		IPs = []string{"127.0.0.1:8080", "127.0.0.1:8081"}
	} else if len(IPs) == 0 {
		slog.Debug("No IP addresses found in json file")
		IPs = []string{"127.0.0.1:8080", "127.0.0.1:8081"}
	} else {
		for i, v := range IPs {
			IPs[i] = v + ":8080"
		}
	}

	var wg sync.WaitGroup
	c := make(chan ClientIdentifier, 5)
	slog.Debug("Flags Parsed", "isCustomProto", isCustomProto)
	if isCustomProto {
		for id, ip := range IPs {
			fmt.Printf("Trying to connect to IP: %v\n", ip)
			conn, err := net.Dial("tcp", ip)
			if err != nil {
				fmt.Printf("Couldn't connect to IP: %v, err: %v\n", ip, err)
				continue
			}
			wg.Add(1)
			fmt.Printf("Successfully connected to ip: %v\n", ip)
			go handleConnection(conn, id, &wg, c)
		}
		wg.Wait()
		close(c)
		identities := map[string]ClientIdentifier{}
		for identity := range c {
			identities[identity.id] = identity
		}
		fmt.Println("Ready slaves:", identities)
		imageDimensions, subdivision_levels := generate.GetImageDimensions()
		totalProcessors := 0
		for _, client := range identities {
			totalProcessors += client.numProcesses
		}
		subImageDimensionsArray := solvers.GetSubImageDimensionsArrays(imageDimensions, totalProcessors)
		delegate(subImageDimensionsArray, subdivision_levels, identities)
	} else {
		RpcFlow(IPs)
	}
	fmt.Printf("Exiting master node\n")
}
