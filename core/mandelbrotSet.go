package core

import (
	"bufio"
	"encoding/binary"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"github.com/navod-abay/mandelbrotset-go/core/models"
	"github.com/navod-abay/mandelbrotset-go/core/solvers"
	"github.com/navod-abay/mandelbrotset-go/core/writers"
)

const (
	maximum_iteration_depth = 100
	base_resolution         = 4
)

func getIntWithDefaultValue(reader *bufio.Reader, prompt string, _default int) int {
	isSuccess := false
	var value int
	for !isSuccess {
		fmt.Printf("%s [%d]:", prompt, _default)
		str, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error Reading input.")
			fmt.Println(err)
			continue
		}
		str = strings.TrimSpace(str)
		if str == "" {
			value = _default
			isSuccess = true
		} else {
			var err error
			value, err = strconv.Atoi(str)
			if err != nil {
				fmt.Println("Error scanning the input")
				continue
			}
			isSuccess = true
		}
	}

	return value
}

func getFloatWithDefaultValue(reader *bufio.Reader, prompt string, _default float64) float64 {
	isSuccess := false
	var value float64
	for !isSuccess {

		fmt.Printf("%s [%f]:", prompt, _default)
		str, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Error Reading input.%v", err)
			continue
		}
		str = strings.TrimSpace(str)
		if str == "" {
			value = _default
			isSuccess = true
		} else {
			var err error
			value, err = strconv.ParseFloat(str, 64)
			if err != nil {
				fmt.Println("Error scanning the input")
				continue
			}
			isSuccess = true
		}
	}

	return value
}

func getInputWithCondition(prompt string, condition func(int) bool) int {
	input := getOneIntInput(prompt)
	for !condition(input) {
		fmt.Println("The input doesn't satisify the condition")
		getOneIntInput(prompt)
	}
	return input
}

func getOneIntInput(prompt string) int {
	isCorrect := false
	var i int
	for !isCorrect {
		fmt.Println(prompt)
		numScanned, err := fmt.Scan(&i)
		if numScanned > 1 {
			fmt.Println("Only 1 value is expected")
		} else if err != nil {
			fmt.Println("Error Scanning input")
		} else {
			return i
		}
	}
	return i
}

func calculatePixelSize(imageDimensions models.ImageDimensions, subdivision_level int) models.ImageDimensions {
	X_range := imageDimensions.X_high - imageDimensions.X_low
	Y_range := imageDimensions.Y_high - imageDimensions.Y_low
	var pixel_size float64
	if X_range < Y_range {
		pixel_size = float64(X_range) / float64(base_resolution*(int(1)<<subdivision_level))
	} else {
		pixel_size = float64(Y_range) / float64(base_resolution*(int(1)<<subdivision_level))
	}
	X_size := int32(float64(X_range) / pixel_size)
	Y_size := int32(float64(Y_range) / pixel_size)
	updateImageDimensions := models.ImageDimensions{
		X_low:      imageDimensions.X_low,
		X_high:     imageDimensions.X_low + pixel_size*float64(X_size),
		Y_low:      imageDimensions.Y_low,
		Y_high:     imageDimensions.Y_high,
		Pixel_size: pixel_size,
		X_size:     X_size,
		Y_size:     Y_size,
		X_start:    0,
		Y_start:    0,
		HueLower:   imageDimensions.HueLower,
		HueUpper:   imageDimensions.HueUpper,
		Sat:        imageDimensions.Sat,
		Value:      imageDimensions.Value,
	}
	return updateImageDimensions
}

func GetImageDimensions() (models.ImageDimensions, int) {
	reader := bufio.NewReader(os.Stdin)

	var imageDimensions models.ImageDimensions
	imageDimensions.X_low = getFloatWithDefaultValue(reader, "Enter the x axis lower limit", -2)
	imageDimensions.X_high = getFloatWithDefaultValue(reader, "Enter the x axis upper limit", 2)
	imageDimensions.Y_low = getFloatWithDefaultValue(reader, "Enter y axis lower limit", -2)
	imageDimensions.Y_high = getFloatWithDefaultValue(reader, "Enter y axis upper limit", 2)
	subdivision_level := getIntWithDefaultValue(reader, "Enter the subdivision level", 4)
	imageDimensions.HueLower = int32(getIntWithDefaultValue(reader, "Enter the Hue lower bound", 0))
	imageDimensions.HueUpper = int32(getIntWithDefaultValue(reader, "Enter the Hue upper bound", 360))
	imageDimensions.Sat = int32(getFloatWithDefaultValue(reader, "Enter Saturation", 30))
	imageDimensions.Value = int32(getFloatWithDefaultValue(reader, "Enter value", 40))
	var pixel_size float64
	var X_size, Y_size int
	imageDimensions = calculatePixelSize(imageDimensions, subdivision_level)
	fmt.Printf("HU: %v,HL: %v, S: %v, V: %v\n", imageDimensions.HueUpper, imageDimensions.HueLower, imageDimensions.Sat, imageDimensions.Value)
	fmt.Println("Calculated array size")
	fmt.Printf("X axis size: %v\n", X_size)
	fmt.Printf("Y axis size: %v\n", Y_size)
	fmt.Printf("Pixel size: %v\n", pixel_size)
	return imageDimensions, subdivision_level
}

func Run() {
	opts := &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}
	handler := slog.NewTextHandler(os.Stdout, opts)
	logger := slog.New(handler)
	slog.SetDefault(logger)
	imageDimensions, subdivision_level := GetImageDimensions()

	optimizationFlag := flag.Bool("optimization", true, "Set to true to enable optimization")
	colorFlag := flag.Bool("not-colorized", true, "Produces a two colored image when set to true")
	csvWriteFlag := flag.Bool("write-csv", true, "Set to true to write the end result to a csv")
	bmpWriteFlag := flag.Bool("write-bmp", true, "Set to true to create a bmp image file")
	saveSnapShotsFlag := flag.Bool("save-snapshots", false, "Set to save intermediate results in the optimization process")
	noParellalizationFlag := flag.Bool("no-parellalization", true, "Set to true to stop parallelization and run everything in one thread")

	flag.Parse()
	fmt.Println(*noParellalizationFlag)
	// Creating the image array
	var writeWaitGroup sync.WaitGroup
	if !*optimizationFlag {
		if *colorFlag {
			pixelArray := solvers.ConstructAndCalculatePixelArray(imageDimensions)
			if *csvWriteFlag {
				writeWaitGroup.Add(1)
				go writers.WriteToCSV(pixelArray, &writeWaitGroup)
			}
			if *bmpWriteFlag {
				writeWaitGroup.Add(1)
				includedColor := make([]byte, 2)
				binary.LittleEndian.PutUint16(includedColor, 0)
				excludedColor := make([]byte, 2)
				binary.LittleEndian.PutUint16(excludedColor, 255)
				go writers.WriteToBmpFile(pixelArray, imageDimensions, "output.bmp", maximum_iteration_depth, &writeWaitGroup)

			}
		} else {
			pixelArray := solvers.ConstructAndCalculateNoColorPixelArray(imageDimensions)
			if *csvWriteFlag {
				writers.WriteToCSVNoColor(pixelArray)
			}
			if *bmpWriteFlag {
				includedColor := make([]byte, 2)
				binary.LittleEndian.PutUint16(includedColor, 0)
				excludedColor := make([]byte, 2)
				binary.LittleEndian.PutUint16(excludedColor, 255)
				writers.WriteToBmpFileNoColor(pixelArray, imageDimensions, includedColor, excludedColor)

			}
		}
	} else {

		if !*noParellalizationFlag {
			fmt.Print("Running with no parallelization")
			pixelArray := solvers.OptimizedCalculation(imageDimensions, subdivision_level, *saveSnapShotsFlag)

			if *csvWriteFlag {
				writeWaitGroup.Add(1)
				go writers.WriteToCSV(pixelArray, &writeWaitGroup)
			}
			if *bmpWriteFlag {
				writeWaitGroup.Add(1)
				includedColor := make([]byte, 2)
				binary.LittleEndian.PutUint16(includedColor, 0)
				excludedColor := make([]byte, 2)
				binary.LittleEndian.PutUint16(excludedColor, 255)
				go writers.WriteToBmpFile(pixelArray, imageDimensions, "output.bmp", maximum_iteration_depth, &writeWaitGroup)

			}
		} else {
			fmt.Print("Running with parallelization")
			var waitGroup sync.WaitGroup // Wait group to wait for parallelized sub images
			var init_skip int32
			if subdivision_level == 0 {
				init_skip = 1
			} else {
				init_skip = int32(1) << (subdivision_level / 2)
			}
			processorGroupSize := runtime.NumCPU()

			subImageDimensionsArray := solvers.GetSubImageDimensionsArrays(imageDimensions, processorGroupSize)
			pixelArray := make([][]uint16, imageDimensions.X_size)
			for i := range imageDimensions.X_size {
				pixelArray[i] = make([]uint16, imageDimensions.Y_size)
			}
			for _, subImageDimension := range subImageDimensionsArray {
				waitGroup.Add(1)
				go solvers.SubImageOptimizedCalculation(subImageDimension, pixelArray, init_skip, &waitGroup, *saveSnapShotsFlag)
			}
			waitGroup.Wait()
			if *csvWriteFlag {
				writeWaitGroup.Add(1)
				go writers.WriteToCSV(pixelArray, &writeWaitGroup)
			}
			if *bmpWriteFlag {
				writeWaitGroup.Add(1)
				includedColor := make([]byte, 2)
				binary.LittleEndian.PutUint16(includedColor, 0)
				excludedColor := make([]byte, 2)
				binary.LittleEndian.PutUint16(excludedColor, 255)
				go writers.WriteToBmpFile(pixelArray, imageDimensions, "output.bmp", maximum_iteration_depth, &writeWaitGroup)

			}
		}
	}
	writeWaitGroup.Wait()
	fmt.Println("Calculation is over")

}
