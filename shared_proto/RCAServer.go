package sharedproto

import (
	"crypto/rand"
	"fmt"
	"log/slog"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/navod-abay/mandelbrotset-go/core/models"
	"github.com/navod-abay/mandelbrotset-go/core/solvers"
	"github.com/navod-abay/mandelbrotset-go/core/writers"
)

type RpcServer struct {
	Port int
}

type EmptyInput struct{}

type StartWorkArgs struct {
	ImageDimensions   []models.ImageDimensions
	Subdivision_level int
}

func (server *RpcServer) HelloWorld(name string, out *string) error {
	fmt.Printf("Hello %v", name)
	*out = "Great to meet you"
	return nil
}

func (server *RpcServer) GetNumProcessor(nothing EmptyInput, numCPU *int) error {
	*numCPU = runtime.NumCPU()
	return nil
}

func (servre *RpcServer) transformImageDimensions(imageDimensions models.ImageDimensions) models.ImageDimensions {
	imageDimensions.Orig_x_start = imageDimensions.X_start
	imageDimensions.Orig_y_start = imageDimensions.Y_start
	imageDimensions.X_size = imageDimensions.X_size - imageDimensions.X_start
	imageDimensions.Y_size = imageDimensions.Y_size - imageDimensions.Y_start
	imageDimensions.X_start = 0
	imageDimensions.Y_start = 0
	return imageDimensions
}

func (server *RpcServer) calculateSubImage(imageDimensions models.ImageDimensions, init_skip int32, c chan [][]uint16, outerWaitGroup *sync.WaitGroup, identityTime string) {
	imageDimensions = server.transformImageDimensions(imageDimensions)
	slog.Debug("X_low: " + strconv.FormatFloat(imageDimensions.X_low, 'f', -1, 64))
	slog.Debug("Y_low: " + strconv.FormatFloat(imageDimensions.Y_low, 'f', -1, 64))
	pixelArray := make([][]uint16, imageDimensions.X_size)
	slog.Debug("Sanity checking ", "imageDimensions", imageDimensions)
	for i := range imageDimensions.X_size {
		pixelArray[i] = make([]uint16, imageDimensions.Y_size)
	}
	for i := range imageDimensions.X_size {
		pixelArray[i] = make([]uint16, imageDimensions.Y_size)
	}
	result := solvers.SubImageOptimizedCalculation(imageDimensions, pixelArray, init_skip, outerWaitGroup, true)
	var wg sync.WaitGroup
	wg.Add(1)
	filename := strconv.Itoa(server.Port) + "/" + identityTime + "/" + rand.Text() + ".bmp"
	writers.WriteToBmpFile(result, imageDimensions, filename, 1000, &wg)
	wg.Wait()
	c <- result
}

func (server *RpcServer) StartWork(startWorkArgs StartWorkArgs, result *int) error {
	subdivision_level := startWorkArgs.Subdivision_level
	fmt.Print("Running with parallelization")
	var waitGroup sync.WaitGroup // Wait group to wait for parallelized sub images
	var init_skip int32
	if subdivision_level == 0 {
		init_skip = 1
	} else {
		init_skip = int32(1) << (subdivision_level / 2)
	}
	subImageDimensionsArray := startWorkArgs.ImageDimensions
	c := make(chan [][]uint16)
	timestamp := time.Now().GoString()
	for _, subImageDimension := range subImageDimensionsArray {
		waitGroup.Add(1)
		go server.calculateSubImage(subImageDimension, init_skip, c, &waitGroup, timestamp)
	}
	waitGroup.Wait()
	resultArray := make([][][]uint16, len(subImageDimensionsArray))
	for result := range c {
		resultArray[0] = result
	}
	return nil
}
