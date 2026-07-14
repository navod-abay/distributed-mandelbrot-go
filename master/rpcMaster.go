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
	"github.com/navod-abay/mandelbrotset-go/core/writers"
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

func sendRPCWorkRequest(id int, client *rpc.Client, subImages []models.ImageDimensions, subdivision_levels int, wg *sync.WaitGroup, c chan models.ImageFragment) {
	defer wg.Done()
	fmt.Println("Sending work request to: " + strconv.Itoa(id) + "\n")
	startWorkArgs := sharedproto.StartWorkArgs{ImageDimensions: subImages, Subdivision_level: subdivision_levels}
	var results []models.ImageFragment
	err := client.Call("RpcServer.StartWork", startWorkArgs, &results)
	slog.Debug("RCP call returned")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	slog.Debug("Gathered completed models for client", "clientID: ", id, "subimage count", len(results))
	for i := range results {
		slog.Debug("Returned image fragment from the client", "imageFragment.X_Index", results[i].X_Index, "imageFragment.Y_Index", results[i].Y_Index)
		c <- results[i]
	}

}

func delegateRPC(subImages []models.ImageDimensions, subdivision_levels int, clients map[int]RPCClientIdentifier, totalProcessors int) []models.ImageFragment {
	index := 0
	var wg sync.WaitGroup
	c := make(chan models.ImageFragment, totalProcessors)
	for id, client := range clients {
		wg.Add(1)
		clientImageSet := subImages[index : index+int(client.numProcesses)]
		slog.Debug("Going to send RPC request to client", "client", strconv.Itoa(id), "numSubImages", strconv.Itoa(len(clientImageSet)))
		go sendRPCWorkRequest(id, client.client, clientImageSet, subdivision_levels, &wg, c)
		index += int(client.numProcesses)
	}
	wg.Wait()
	close(c)
	var fragments []models.ImageFragment
	for fragment := range c {
		fragments = append(fragments, fragment)
	}
	fmt.Println("Finished delegating work")
	return fragments
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
	fragments := delegateRPC(subImageDimensionsArray, subdivision_levels, identities, totalProcessors)
	var writeWaitGroup sync.WaitGroup
	writeWaitGroup.Add(1)
	writers.WriteFragmentsToBmp(fragments, "output.bmp", imageDimensions, &writeWaitGroup)
	writeWaitGroup.Wait()
	fmt.Printf("Exiting master node\n")
}
