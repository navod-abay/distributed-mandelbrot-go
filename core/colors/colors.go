package colors

import (
	"encoding/binary"
	"math"
)

func MapIterationsToUint16Colors(iterations uint16, hueUpper int32, hueLower int32, saturation int32, value int32) []byte {
	buf := make([]byte, 2)
	hue := hueLower + (hueUpper-hueLower)/100*int32(iterations)
	R, G, B := hsvTo2ByteRGB(hue, float64(saturation), float64(value))
	// fmt.Printf("R: %v, B: %v, G: %v", R, G, B)
	var colorNum uint16
	colorNum |= (B & 0x003F)
	colorNum |= (G & 0x007F) << 5
	colorNum |= (R & 0x003F) << 11
	binary.LittleEndian.PutUint16(buf, colorNum)
	return buf
}
func hsvTo2ByteRGB(hue int32, sat float64, val float64) (uint16, uint16, uint16) {
	// fmt.Printf("H : %v, S: %v, V: %v", hue, sat, val)
	val /= 100
	sat /= 100
	C := val * sat
	X := C * (1 - math.Abs(float64((hue/60)%2-1)))
	m := val - C
	var R_, G_, B_ float64
	switch {
	case hue < 60:
		R_, G_, B_ = C, X, 0
	case hue < 120:
		R_, G_, B_ = X, C, 0
	case hue < 180:
		R_, G_, B_ = 0, C, X
	case hue < 240:
		R_, G_, B_ = 0, X, C
	case hue < 300:
		R_, G_, B_ = X, 0, C
	case hue < 360:
		R_, G_, B_ = C, 0, X
	}
	return uint16((R_ + m) * 32), uint16((G_ + m) * 64), uint16((B_ + m) * 32)
}
