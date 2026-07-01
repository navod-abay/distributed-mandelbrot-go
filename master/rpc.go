package main

import (
	"fmt"
	"log/slog"
	"net/rpc"
	"strconv"
	"sync"

	generate "github.com/navod-abay/mandelbrotset-go/core"
	"github.com/navod-abay/mandelbrotset-go/core/models"
	"github.com/navod-abay/mandelbrotset-go/core/solvers"
	sharedproto "github.com/navod-abay/mandelbrotset-go/shared_proto"
)

func RpcHandshake(client *rpc.Client, wg *sync.WaitGroup, c chan RPCClientIdentifier, id int) error {
	defer wg.Done()
	var numProcesses int16
	err := client.Call("RpcServer.GetNumProcessor", new(sharedproto.EmptyInput), &numProcesses)
	if err != nil {
		slog.Error(err.Error())
		return err
	}
	clientIdentity := RPCClientIdentifier{id: id, numProcesses: numProcesses, client: client}
	slog.Debug("NumProcess retrieval done", "numProcesses", numProcesses)
	c <- clientIdentity
	return nil
}

func sendRPCWorkRequest(id int, client *rpc.Client, subImages []models.ImageDimensions, subdivision_levels int, wg *sync.WaitGroup) {
	defer wg.Done()
	fmt.Println("Sending work request to: " + strconv.Itoa(id) + "\n")
	startWorkArgs := sharedproto.StartWorkArgs{ImageDimensions: subImages, Subdivision_level: subdivision_levels}
	var result *int
	err := client.Call("RpcServer.StartWork", startWorkArgs, result)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

}

func delegateRPC(subImages []models.ImageDimensions, subdivision_levels int, clients map[int]RPCClientIdentifier) {
	index := 0
	var wg sync.WaitGroup
	for id, client := range clients {
		wg.Add(1)
		go sendRPCWorkRequest(id, client.client, subImages[index:index+int(client.numProcesses)], subdivision_levels, &wg)
	}
	wg.Wait()
	fmt.Println("Finished delegating work")
}

func RpcFlow(IPs []string) {
	slog.Debug("Running in RPC mode")
	var wg sync.WaitGroup
	c := make(chan RPCClientIdentifier, 5)
	for index, ip := range IPs {
		fmt.Printf("Trying to connect(RPC) to IP: %v\n", ip)
		client, err := rpc.Dial("tcp", ip)
		if err != nil {
			fmt.Printf("Couldn't connect to IP: %v, err: %v\n", ip, err)
			continue
		}
		wg.Add(1)
		fmt.Printf("Successfully dialed ip: %v\n", ip)
		go RpcHandshake(client, &wg, c, index)
	}
	slog.Debug("waiting for waitgroups")
	wg.Wait()
	slog.Debug("Finished waiting for waitgroups")
	close(c)
	slog.Debug("Closed channel")
	identities := map[int]RPCClientIdentifier{}
	for identity := range c {
		identities[identity.id] = identity
	}
	fmt.Println("Ready slaves:", identities)
	imageDimensions, subdivision_levels := generate.GetImageDimensions()
	totalProcessors := 0
	for _, client := range identities {
		totalProcessors += int(client.numProcesses)
	}
	subImageDimensionsArray := solvers.GetSubImageDimensionsArrays(imageDimensions, totalProcessors)
	delegateRPC(subImageDimensionsArray, subdivision_levels, identities)
	fmt.Printf("Exiting master node\n")
}
