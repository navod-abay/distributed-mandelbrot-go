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

func (server *RpcServer) calculateSubImage(imageDimensions models.ImageDimensions, init_skip int32, c chan models.ImageFragment, outerWaitGroup *sync.WaitGroup, identityTime string) {
	defer outerWaitGroup.Done()
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
	var wg sync.WaitGroup
	wg.Add(1)
	result := solvers.SubImageOptimizedCalculation(imageDimensions, pixelArray, init_skip, &wg, true)
	filename := strconv.Itoa(server.Port) + "/" + identityTime + "/" + rand.Text() + ".bmp"
	wg.Wait()
	wg.Add(1)
	go writers.WriteToBmpFile(result, imageDimensions, filename, 1000, &wg)
	wg.Wait()
	slog.Debug("Going to write result to channel")

	c <- models.ImageFragment{Result: result, X_Index: imageDimensions.X_index, Y_Index: imageDimensions.Y_index}
	slog.Debug("Result written to channel from the goroutine")
}

func (server *RpcServer) StartWork(startWorkArgs StartWorkArgs, results *[]models.ImageFragment) error {
	slog.Debug("Starting Work", "len(startworkArgs.imageDimensions)", len(startWorkArgs.ImageDimensions))
	subdivision_level := startWorkArgs.Subdivision_level
	fmt.Print("Running with parallelization")
	var init_skip int32
	if subdivision_level == 0 {
		init_skip = 1
	} else {
		init_skip = int32(1) << (subdivision_level / 2)
	}
	subImageDimensionsArray := startWorkArgs.ImageDimensions
	for _, subImageDimension := range subImageDimensionsArray {
		slog.Debug("Checking subImage index values", "subImageDimension.X_Index", subImageDimension.X_index, "subImageDimension.Y_index", subImageDimension.Y_index)
	}
	c := make(chan models.ImageFragment, len(subImageDimensionsArray))
	timestamp := time.Now().GoString()
	var waitGroup sync.WaitGroup // Wait group to wait for parallelized sub images
	for _, subImageDimension := range subImageDimensionsArray {
		waitGroup.Add(1)
		go server.calculateSubImage(subImageDimension, init_skip, c, &waitGroup, timestamp)
	}
	slog.Debug("Waiting for calculating subimages")
	waitGroup.Wait()
	close(c)
	slog.Debug("Finished calculating subimages")
	resultArray := make([]models.ImageFragment, len(subImageDimensionsArray))
	index := 0
	for result := range c {
		resultArray[index] = result
		index++
	}
	for i := range resultArray {
		slog.Debug("Image fragment to be sent from the client", "imageFragment.X_Index", resultArray[i].X_Index, "imageFragment.Y_Index", resultArray[i].Y_Index)
	}

	*results = resultArray
	slog.Debug("Finished writing Calculating subimages")
	return nil
}
