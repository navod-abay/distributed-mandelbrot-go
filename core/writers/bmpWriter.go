package writers

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/navod-abay/mandelbrotset-go/core/colors"
	"github.com/navod-abay/mandelbrotset-go/core/models"
)

const (
	max_iteration int = 1000 // TODO: Use a command line argument with default values for max_iteration value
)

type BmpHeaderDetails struct {
	fileSize       uint32
	reserved       uint32
	infoHeaderSize uint32
	dataOffset     uint32
	width          int32
	height         int32
	planes         uint16
	bitCount       uint16
	compression    int32
	imageSize      int32
	endInfoHeader  []int32
}

func WriteFragmentsToBmp(fragments []models.ImageFragment, filename string, origImageDimension models.ImageDimensions, writeWaitGroup *sync.WaitGroup) {
	defer writeWaitGroup.Done()
	slog.Debug("Total fragments", "total Fragments", len(fragments))
	fragmentPointers := make([][]int, len(fragments)) // dynamic sizing because there are not much fragments
	for i := range len(fragmentPointers) {            // an n x n matrix even though have a total of n fragments. Won't be a memory issue because it is insignificant
		fragmentPointers[i] = make([]int, len(fragments))
	}
	var maxXIndex int = 0
	var maxYIndex int = 0
	for index, fragment := range fragments {
		if fragment.X_Index > int32(maxXIndex) {
			maxXIndex = int(fragment.X_Index)
		}
		if fragment.Y_Index > int32(maxYIndex) {
			maxYIndex = int(fragment.Y_Index)
		}
		fragmentPointers[fragment.X_Index][fragment.Y_Index] = index
	}
	slog.Debug("Finished organizgin fragments", "maxXIndex", maxXIndex, "maxYIndex", maxXIndex, "fragmentPointers", fragmentPointers)
	fmt.Println("Writing output to bmp file")
	dir := filepath.Dir("output/" + filename)
	dir_err := os.MkdirAll(dir, os.ModePerm)
	if dir_err != nil {
		slog.Error("Failed to make directory to store files")
		return
	}
	bmp_f, err := os.OpenFile("output/"+filename, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		fmt.Println("Failed to opena  writer for the bmp file")
		slog.Error(err.Error())
	} else {
		WriteBmpHeader(bmp_f, calculateBMPHeaderDetailsFromFragments(origImageDimension))
		writer := bufio.NewWriter(bmp_f)
		for x_index := range maxXIndex {
			for y_index := range maxYIndex {
				pixelArray := fragments[fragmentPointers[x_index][y_index]].Result
				slog.Debug("Writing Image Fragmet", "X_index", x_index, "y_index", y_index, "len(pixelArray)", len(pixelArray))
				for i := range pixelArray[0] {
					for j := range pixelArray {
						writer.Write(colors.MapIterationsToUint16Colors(pixelArray[j][i], origImageDimension.HueUpper, origImageDimension.HueLower, origImageDimension.Sat, origImageDimension.Value))
					}
				}
			}
		}

		writer.Flush()
	}
	defer bmp_f.Close()
}

func calculateBMPHeaderDetailsFromFragments(origImageDimension models.ImageDimensions) BmpHeaderDetails {
	var details BmpHeaderDetails
	details.infoHeaderSize = 40
	details.width = int32(origImageDimension.X_size)
	details.height = int32(origImageDimension.Y_size)
	details.planes = 1
	details.compression = 0
	details.imageSize = 0
	details.bitCount = 16
	details.dataOffset = 54
	details.reserved = 0
	details.fileSize = 2*uint32(details.width)*uint32(details.height) + 54
	fmt.Println("FileSize: ", details.fileSize)
	details.endInfoHeader = []int32{0, 0, 0, 0, 0, 0}
	return details
}
func WriteBmpHeader(file *os.File, headerDetails BmpHeaderDetails) {
	slog.Debug("Writing tp BMP file", "headerDetails", headerDetails)
	bufferedWriter := bufio.NewWriter(file)
	bufferedWriter.WriteString("BM")
	binary.Write(bufferedWriter, binary.LittleEndian, headerDetails.fileSize)
	binary.Write(bufferedWriter, binary.LittleEndian, headerDetails.reserved)
	binary.Write(bufferedWriter, binary.LittleEndian, headerDetails.dataOffset)
	binary.Write(bufferedWriter, binary.LittleEndian, headerDetails.infoHeaderSize)
	binary.Write(bufferedWriter, binary.LittleEndian, []int32{headerDetails.width, headerDetails.height})

	buf := make([]byte, 2)
	binary.LittleEndian.PutUint16(buf, headerDetails.planes)
	bufferedWriter.Write(buf)

	binary.LittleEndian.PutUint16(buf, headerDetails.bitCount)
	bufferedWriter.Write(buf)

	binary.Write(bufferedWriter, binary.LittleEndian, headerDetails.endInfoHeader)
	bufferedWriter.Flush()
}

func CalculateBMPHeaderDetails(imageDimensions models.ImageDimensions) BmpHeaderDetails {
	var details BmpHeaderDetails
	details.infoHeaderSize = 40
	details.width = int32(imageDimensions.X_size)
	details.height = int32(imageDimensions.Y_size)
	details.planes = 1
	details.compression = 0
	details.imageSize = 0
	details.bitCount = 16
	details.dataOffset = 54
	details.reserved = 0
	details.fileSize = 2*uint32(details.width)*uint32(details.height) + 54
	fmt.Println("FileSize: ", details.fileSize)
	details.endInfoHeader = []int32{0, 0, 0, 0, 0, 0}
	return details
}

func WriteToBmpFileNoColor(pixelArray [][]bool, imageDimensions models.ImageDimensions, includedColor []byte, excludedColor []byte) {
	fmt.Println("Writing output to bmp file (No Color)")
	bmp_f, err := os.OpenFile("outputNoColor.bmp", os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		fmt.Println("Failed to opena  writer for the bmp file")
	} else {
		WriteBmpHeader(bmp_f, CalculateBMPHeaderDetails(imageDimensions))
		writer := bufio.NewWriter(bmp_f)
		for i := range pixelArray[0] {
			for j := range pixelArray {
				if pixelArray[j][i] {
					writer.Write(includedColor)
				} else {
					writer.Write(excludedColor)
				}
			}
		}
		writer.Flush()
	}

	defer bmp_f.Close()
}

func WriteToBmpFile(pixelArray [][]uint16, imageDimensions models.ImageDimensions, filename string, iterationThreshold int, writeWaitgroup *sync.WaitGroup) {
	defer writeWaitgroup.Done()
	fmt.Println("Writing output to bmp file")
	dir := filepath.Dir("output/" + filename)
	dir_err := os.MkdirAll(dir, os.ModePerm)
	if dir_err != nil {
		slog.Error("Failed to make directory to store files")
		return
	}
	bmp_f, err := os.OpenFile("output/"+filename, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		fmt.Println("Failed to opena  writer for the bmp file")
		slog.Error(err.Error())
	} else {
		WriteBmpHeader(bmp_f, CalculateBMPHeaderDetails(imageDimensions))
		writer := bufio.NewWriter(bmp_f)
		for i := range pixelArray[0] {
			for j := range pixelArray {
				writer.Write(colors.MapIterationsToUint16Colors(pixelArray[j][i], imageDimensions.HueUpper, imageDimensions.HueLower, imageDimensions.Sat, imageDimensions.Value))
			}
		}
		writer.Flush()
	}
	defer bmp_f.Close()
}

func SaveSnapShotBMP(pixelArray [][]uint16, imageDimensions models.ImageDimensions, skip int, waitgroup *sync.WaitGroup) error {
	defer waitgroup.Done()
	currentTime := time.Now()
	fileName := currentTime.Format(time.RFC3339Nano) + ".bmp"
	snapshotFilepath := filepath.Join("snapshots", fileName)
	slog.Debug("Saving a snapshot", "skip", skip)
	bmp_f, err := os.Create(snapshotFilepath)
	if err != nil {
		fmt.Println("Failed to opena  writer for the bmp file")
	} else {
		WriteBmpHeader(bmp_f, CalculateBMPHeaderDetails(imageDimensions))
		writer := bufio.NewWriter(bmp_f)
		for i := range pixelArray[0] {
			for j := range pixelArray {
				writer.Write(colors.MapIterationsToUint16Colors(pixelArray[j][i], imageDimensions.HueUpper, imageDimensions.HueLower, imageDimensions.Sat, imageDimensions.Value))
			}
		}
		writer.Flush()
	}

	defer bmp_f.Close()
	return err
}
