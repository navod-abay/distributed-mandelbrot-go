package writers

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/navod-abay/mandelbrotset-go/core/models"
)

func WriteToCSVNoColor(pixelArray [][]bool) {

	fmt.Println("Writing output to a csv file(No Colors)")
	f, err := os.OpenFile("outputNoColor.csv", os.O_WRONLY|os.O_CREATE, 0644)
	writer := bufio.NewWriter(f)

	if err == nil {
		for i := range pixelArray {
			for j := range pixelArray[i] {
				if pixelArray[i][j] {
					writer.WriteString("1, ")
				} else {
					writer.WriteString("0,")
				}
			}
			writer.WriteString("\n")
		}
		writer.Flush()
	} else {
		log.Fatal(err)
	}
	defer f.Close()
}

func WriteToCSV(pixelArray [][]uint16, writeWaitGroup *sync.WaitGroup) {
	defer writeWaitGroup.Done()

	fmt.Println("Writing output to a csv file")
	f, err := os.Create("output.csv")
	writer := bufio.NewWriter(f)

	if err == nil {
		for i := range pixelArray {
			for j := range pixelArray[i] {
				writer.WriteString(strconv.Itoa(int(pixelArray[i][j])) + ",")
			}
			writer.WriteString("\n")
		}
		writer.Flush()
	} else {
		log.Fatal(err)
	}
	defer f.Close()
}

func SaveCsvSnapshot(pixelArray [][]uint16, imageDimensions models.ImageDimensions, skip int32, waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()
	currentTime := time.Now()
	fileName := currentTime.Format(time.RFC3339Nano) + ".csv"
	snapshotFilepath := filepath.Join("snapshots", fileName)
	// slog.Debug("Saving a csv snapshot", "skip", skip)
	f, err := os.Create(snapshotFilepath)
	writer := bufio.NewWriter(f)

	if err == nil {
		for i := int32(0); i < imageDimensions.X_size; i += skip {
			for j := int32(0); j < imageDimensions.X_size; j += skip {
				writer.WriteString(strconv.Itoa(int(pixelArray[i][j])) + ",")
			}
			writer.WriteString("\n")
		}
		writer.Flush()
	} else {
		log.Fatal(err)
	}
	defer f.Close()
}
