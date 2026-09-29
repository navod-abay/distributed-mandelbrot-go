package solvers

func complexMultiplicationSomponents(num complex128) (float64, float64, float64) {
	r := real(num)
	i := imag(num)
	return r * r, r * i, i * i
}

func checkMandelbrotSetInclusionNoColor(z_0 complex128, max_iteration int) bool {
	z_i := z_0
	r := real(z_0)
	i := imag(z_0)
	num := 0
	for ; num < max_iteration; num++ {
		rr, ri, ii := complexMultiplicationSomponents(z_i)
		if rr+ii > 4 {
			return false
		} else {
			z_i = complex(rr-ii+r, 2*ri+i)
		}
	}
	return true
}

func CheckMandelbrotSetInclusion(z_0 complex128, max_iteration int) uint16 {
	return checkMandelbrotSetInclusion(z_0, max_iteration)
}

func checkMandelbrotSetInclusion(z_0 complex128, max_iteration int) uint16 {
	z_i := z_0
	r := real(z_0)
	i := imag(z_0)
	var num uint16 = 0
	for ; num < uint16(max_iteration); num++ {
		rr, ri, ii := complexMultiplicationSomponents(z_i)
		if rr+ii > 4 {
			return num
		} else {
			z_i = complex(rr-ii+r, 2*ri+i)
		}
	}
	return maximum_iteration_depth
}
