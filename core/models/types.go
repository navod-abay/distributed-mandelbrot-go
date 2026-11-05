package models

type ImageDimensions struct {
	X_high  float64 `json:"x_high"` // Highest number in the x axis
	X_low   float64 `json:"x_low"`  // Lowest number in the x axis
	Y_high  float64 `json:"y_high"` // Higest number in the y axis
	Y_low   float64 `json:"y_low"`  // Lowest number in the y axis
	X_start int32   `json:"-"`
	Y_start int32   `json:"-"`
	X_size  int32   `json:"-"`
	Y_size  int32   `json:"-"`

	Orig_x_low float64 `json:"-"`
	Orig_y_low float64 `json:"-"`
	Pixel_size float64 `json:"-"`

	Orig_x_start int32 `json:"-"`
	Orig_y_start int32 `json:"-"`
	Orig_y_size  int32 `json:"-"`
	Orig_x_size  int32 `json:"-"`

	// For merging after
	X_index int32 `json:"-"`
	Y_index int32 `json:"-"`

	// Color Information
	HueUpper int32 `json:"hue_upper"`
	HueLower int32 `json:"hue_lower"`
	Sat      int32 `json:"sat"`
	Value    int32 `json:"value"`
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

type Flags struct {
	OptimizationFlag      bool
	ColorFlag             bool
	CsvWriteFlag          bool
	BmpWriteFlag          bool
	SaveSnapShotsFlag     bool
	NoParellalizationFlag bool
}

type SolverResult struct {
	Color         bool
	NoColorMatrix *[][]bool
	ColorMatrix   *[][]uint16
}
