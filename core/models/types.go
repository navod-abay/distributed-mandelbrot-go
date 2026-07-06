package models

type ImageDimensions struct {
	X_high  float64 // Highest number in the x axis
	X_low   float64 // Lowest number in the x axis
	Y_high  float64 // Higest number in the y axis
	Y_low   float64 // Lowest number in the y axis
	X_start int32
	Y_start int32
	X_size  int32
	Y_size  int32

	Orig_x_low float64
	Orig_y_low float64
	Pixel_size float64

	Orig_x_start int32
	Orig_y_start int32
	Orig_y_size  int32
	Orig_x_size  int32

	// For merging after
	X_index int32
	Y_index int32

	// Color Information
	HueUpper int32
	HueLower int32
	Sat      int32
	Value    int32
}

type ImageFragment struct {
	Result  [][]uint16
	X_Index int32
	Y_Index int32
}

type Index struct {
	X int
	Y int
}
