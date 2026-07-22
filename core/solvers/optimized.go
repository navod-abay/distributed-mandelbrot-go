package solvers

import (
	"fmt"
	"log/slog"
	"sync"

	"github.com/navod-abay/mandelbrotset-go/core/models"
	"github.com/navod-abay/mandelbrotset-go/core/writers"
)

const (
	maximum_iteration_depth = 1000
)

func getHorizontalMiddleVerticalEdgePixel(pixelArray [][]uint16, x int32, y int32, skip int32) int {
	num := 0
	if pixelArray[x-skip][y] == maximum_iteration_depth {
		num++
	}
	if pixelArray[x][y] == maximum_iteration_depth {
		num++
	}
	if pixelArray[x+skip][y] == maximum_iteration_depth {
		num++
	}
	return num
}

func getHorizontalEdgeVerticalMiddlePixel(pixelArray [][]uint16, x int32, y int32, skip int32) int {
	num := 0
	if pixelArray[x-skip/2][y] == maximum_iteration_depth {
		num++
	}
	if pixelArray[x+skip/2][y] == maximum_iteration_depth {
		num++
	}
	return num
}

func getHorizontalMiddleVerticalMiddlePixelMemberNeighbors(pixelArray [][]uint16, x int32, y int32, skip int32, rightColumn int, leftCol int) int {
	num := rightColumn + leftCol
	if pixelArray[x][y-skip/2] == maximum_iteration_depth {
		num++
	}
	if pixelArray[x][y+skip/2] == maximum_iteration_depth {
		num++
	}
	return num
}

func GetSubImageDimensionsArrays(imageDimensions models.ImageDimensions, processorGroupSize int) []models.ImageDimensions {
	x_length := imageDimensions.X_size
	y_length := imageDimensions.Y_size
	slog.Debug("Starting subImageDimensions calculation", "x_length", x_length, "y_length", y_length)
	x_subdivisions := int32(1)
	y_subdivisions := int32(1)
	fmt.Printf("Number of processes: %v\n", processorGroupSize)
	for processorGroupSize > 1 {
		var longerAxisSubdivisions *int32
		var longerAxisLength *int32
		if x_length > y_length {
			longerAxisSubdivisions = &x_subdivisions
			longerAxisLength = &x_length
		} else {
			longerAxisSubdivisions = &y_subdivisions
			longerAxisLength = &y_length
		}
		if processorGroupSize%2 == 0 {
			*longerAxisSubdivisions *= 2
			processorGroupSize /= 2
			*longerAxisLength /= 2
		} else {
			break
		}
	}
	slog.Debug("Finished calculating subImage subdivisions", "x_subdivision", x_subdivisions, "y_subdivisions", y_subdivisions, "x_length", x_length, "y_length", y_length)
	x_pos := int32(0)
	X_val := imageDimensions.X_low
	Y_val := imageDimensions.Y_low
	var newImageDimensions []models.ImageDimensions
	num := 0
	for i := int32(0); i < x_subdivisions; i++ {
		y_pos := int32(0)
		for j := int32(0); j < y_subdivisions; j++ {
			newImageDimension := imageDimensions
			newImageDimension.X_start = x_pos
			newImageDimension.X_low = X_val
			newImageDimension.Y_start = y_pos
			newImageDimension.Y_low = Y_val
			y_pos += y_length
			Y_val += float64(y_length) * imageDimensions.Pixel_size
			newImageDimension.X_size = x_pos + x_length
			newImageDimension.X_high = X_val + float64(x_length)*imageDimensions.Pixel_size
			newImageDimension.Y_size = y_pos
			newImageDimension.Y_high = Y_val
			newImageDimension.Orig_x_low = imageDimensions.X_low
			newImageDimension.Orig_y_low = imageDimensions.Y_low
			newImageDimension.X_index = i
			newImageDimension.Y_index = j
			newImageDimensions = append(newImageDimensions, newImageDimension)
			num++
		}
		Y_val = imageDimensions.Y_low
		X_val += float64(x_length) * imageDimensions.Pixel_size
		x_pos += x_length
	}
	fmt.Printf(`Number of sub images: %v\n`, len(newImageDimensions))
	return newImageDimensions
}

func RunOneSkipPass(pixelArray [][]uint16, imageDimensions models.ImageDimensions, skip int32, saveSnapShotsFlag bool, waitGroup *sync.WaitGroup) {
	fmt.Printf("\n\nskip: %v\n", skip)
	slog.Debug("Starting calculating edges", "imageDimensions", imageDimensions)
	// Handling edges of the array without changing the maximum iterations according to the neighbors inclusion
	x_val_end := float64(imageDimensions.X_size-1) * (imageDimensions.Pixel_size)
	y_val := imageDimensions.Y_low + float64(skip/2)*imageDimensions.Pixel_size + imageDimensions.Orig_y_low
	skip_pixel_size := float64(skip) * imageDimensions.Pixel_size
	for j := imageDimensions.Y_start + skip/2; j < imageDimensions.Y_size; j += skip {
		pixelArray[imageDimensions.X_start][j] = checkMandelbrotSetInclusion(complex(imageDimensions.X_low, y_val), maximum_iteration_depth)
		pixelArray[imageDimensions.X_size-1][j] = checkMandelbrotSetInclusion(complex(x_val_end, y_val), maximum_iteration_depth)
		y_val += skip_pixel_size
	}
	x_val := float64(imageDimensions.X_start+skip/2)*imageDimensions.Pixel_size + imageDimensions.Orig_x_low
	y_val_end := float64(imageDimensions.Y_size-1)*imageDimensions.Pixel_size + imageDimensions.Orig_y_low
	for i := imageDimensions.X_start + skip/2; i < imageDimensions.X_size; i += skip {
		//		pixelArray[i][imageDimensions.Y_start] = checkMandelbrotSetInclusion(pixelArray[i][imageDimensions.Y_start].Number, maximum_iteration_depth)
		pixelArray[i][imageDimensions.Y_start] = checkMandelbrotSetInclusion(complex(x_val, imageDimensions.Y_low), maximum_iteration_depth)
		// pixelArray[i][imageDimensions.Y_size-1] = checkMandelbrotSetInclusion(pixelArray[i][imageDimensions.X_size-1].Number, maximum_iteration_depth)
		pixelArray[i][imageDimensions.Y_size-1] = checkMandelbrotSetInclusion(complex(x_val, y_val_end), maximum_iteration_depth)

		x_val += skip_pixel_size
	}

	slog.Debug("Finished handling edges for iteration, staring horizontal middle vertical edge", "skip", skip)
	// Handling horizontal middle vertical left edge pixels (8s in a 3 by 3 grid)
	var leftCol int
	var rightCol int
	var middleCol int
	// Handling the center of the arrays
	x_val = imageDimensions.X_low + skip_pixel_size
	for i := imageDimensions.X_start + skip; i < imageDimensions.X_size-skip; i += skip {
		y_val = imageDimensions.Y_low + skip_pixel_size/2
		leftCol = getHorizontalMiddleVerticalEdgePixel(pixelArray, i, imageDimensions.Y_start, skip)
		rightCol = getHorizontalMiddleVerticalEdgePixel(pixelArray, i, imageDimensions.Y_start+skip, skip)
		for j := imageDimensions.Y_start + skip/2; j < imageDimensions.Y_size-2*skip; j += skip {
			pixelArray[i][j] = checkMandelbrotSetInclusion(complex(x_val, y_val), (7-leftCol-rightCol)*100)
			leftCol = rightCol
			rightCol = getHorizontalMiddleVerticalEdgePixel(pixelArray, i, j+(3*skip/2), skip)
			y_val += skip_pixel_size
		}
		x_val += skip_pixel_size
	}

	// slog.Debug("Starting Horizontal Edge Vertical Middle")
	// Handling Horizontal Edge Vertical Middle pixels (4/6s in a grid)
	x_val = imageDimensions.X_low + skip_pixel_size/2
	for i := imageDimensions.X_start + skip/2; i < imageDimensions.X_size-skip; i += skip {
		y_val = imageDimensions.Y_low + skip_pixel_size
		leftCol = getHorizontalEdgeVerticalMiddlePixel(pixelArray, i, imageDimensions.Y_start, skip)
		middleCol = getHorizontalEdgeVerticalMiddlePixel(pixelArray, i, imageDimensions.Y_start+skip, skip)
		rightCol = getHorizontalEdgeVerticalMiddlePixel(pixelArray, i, imageDimensions.Y_start+2*skip, skip)
		for j := imageDimensions.Y_start + skip; j < imageDimensions.Y_size-skip; j += skip {
			pixelArray[i][j] = checkMandelbrotSetInclusion(complex(x_val, y_val), (7-leftCol-rightCol-middleCol)*100)
			leftCol = middleCol
			middleCol = rightCol
			rightCol = getHorizontalEdgeVerticalMiddlePixel(pixelArray, i, j+skip, skip)
			y_val += skip_pixel_size
		}
		x_val += skip_pixel_size
	}

	// Handling Horizontal middle and Verticle Middle pixels
	// slog.Debug("Starting Horizontal Middle Vertical Middle")

	x_val = imageDimensions.X_low + skip_pixel_size/2
	for i := imageDimensions.X_start + skip/2; i < imageDimensions.X_size-skip; i += skip {
		y_val = imageDimensions.Y_low + skip_pixel_size/2
		leftCol = getHorizontalMiddleVerticalEdgePixel(pixelArray, imageDimensions.X_start+skip/2, imageDimensions.Y_start, skip/2)
		for j := imageDimensions.Y_start + skip/2; j < imageDimensions.Y_size-skip; j += skip {
			rightCol = getHorizontalMiddleVerticalEdgePixel(pixelArray, i, j+skip/2, skip/2)
			pixelArray[i][j] = checkMandelbrotSetInclusion(complex(x_val, y_val), (9-getHorizontalMiddleVerticalMiddlePixelMemberNeighbors(pixelArray, i, j, skip, rightCol, leftCol))*100)
			leftCol = rightCol
			y_val += skip_pixel_size
		}
		x_val += skip_pixel_size
	}

	slog.Debug("Finished iteration with skip", "skip", skip)
	if saveSnapShotsFlag {
		waitGroup.Add(1)
		go writers.SaveCsvSnapshot(pixelArray, imageDimensions, skip, waitGroup)
		waitGroup.Add(1)
		go writers.SaveSnapShotBMP(pixelArray, imageDimensions, int(skip), waitGroup)
	}
}

func OptimizedCalculation(imageDimensions models.ImageDimensions, subdivision_level int, saveSnapShotsFlag bool) [][]uint16 {
	var waitGroup sync.WaitGroup
	pixelArray := make([][]uint16, imageDimensions.X_size)
	var init_skip int32
	if subdivision_level == 0 {
		init_skip = 1
	} else {
		init_skip = int32(1) << (subdivision_level / 2)
	}
	X_val := imageDimensions.X_low
	Y_val := imageDimensions.Y_low
	slog.Debug("InitSkip", "init_skip", init_skip)
	for i := imageDimensions.X_start; i < imageDimensions.X_size; i++ {
		Y_val = imageDimensions.Y_low
		pixelArray[i] = make([]uint16, imageDimensions.Y_size)
		for j := imageDimensions.Y_start; j < imageDimensions.Y_size; j++ {
			cur_num := complex(X_val, Y_val)
			if i%init_skip == 0 && j%init_skip == 0 {
				pixelArray[i][j] = checkMandelbrotSetInclusion(cur_num, maximum_iteration_depth)
				// slog.Debug("Checked complex num", "num", pixelArray[i][j].Number)
			}
			Y_val += imageDimensions.Pixel_size
		}
		X_val += imageDimensions.Pixel_size
	}
	println("Finished initializing the array and calculating the initial pass ")
	if saveSnapShotsFlag {
		waitGroup.Add(1)
		go writers.SaveSnapShotBMP(pixelArray, imageDimensions, int(init_skip), &waitGroup)
		waitGroup.Add(1)
		go writers.SaveCsvSnapshot(pixelArray, imageDimensions, init_skip, &waitGroup)
	}

	for skip := init_skip; skip > 1; skip /= 2 {
		RunOneSkipPass(pixelArray, imageDimensions, skip, saveSnapShotsFlag, &waitGroup)
	}
	waitGroup.Wait()
	fmt.Println("Pixel Array initialization is over")

	return pixelArray
}

func SubImageOptimizedCalculation(imageDimensions models.ImageDimensions, pixelArray [][]uint16, init_skip int32, outerWaitGroup *sync.WaitGroup, saveSnapShotsFlag bool) [][]uint16 {
	defer outerWaitGroup.Done()
	var waitGroup sync.WaitGroup
	X_val := imageDimensions.X_low
	Y_val := imageDimensions.Y_low
	slog.Debug("InitSkip", "init_skip", init_skip)
	for i := int32(0); i < imageDimensions.X_size-imageDimensions.X_start; i++ {
		Y_val = imageDimensions.Y_low
		for j := int32(0); j < imageDimensions.Y_size-imageDimensions.Y_start; j++ {
			cur_num := complex(X_val, Y_val)
			if i%init_skip == 0 && j%init_skip == 0 {
				pixelArray[i+imageDimensions.X_start][j+imageDimensions.Y_start] = checkMandelbrotSetInclusion(cur_num, maximum_iteration_depth)
				// slog.Debug("Checked complex num", "num", pixelArray[i][j].Number)
			}
			Y_val += imageDimensions.Pixel_size
		}
		X_val += imageDimensions.Pixel_size
	}
	fmt.Printf("Finished initializing the array and calculating the initial pass. saveSnapshotFlag %v ", saveSnapShotsFlag)
	if saveSnapShotsFlag {
		waitGroup.Add(1)
		go writers.SaveSnapShotBMP(pixelArray, imageDimensions, int(init_skip), &waitGroup)
		waitGroup.Add(1)
		go writers.SaveCsvSnapshot(pixelArray, imageDimensions, init_skip, &waitGroup)
	}
	for skip := init_skip; skip > 1; skip /= 2 {
		RunOneSkipPass(pixelArray, imageDimensions, skip, saveSnapShotsFlag, &waitGroup)
	}

	waitGroup.Wait()
	return pixelArray
}
