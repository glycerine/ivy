package gia

import "math"

// Emb_ManMatrAlloc creates the square, contiguous matrix used by the embedding.
func Emb_ManMatrAlloc(nDims int32) [][]float32 {
	if nDims < 0 {
		panic("gia: negative embedding dimension")
	}
	rows := make([][]float32, int(nDims))
	data := make([]float32, int(nDims)*int(nDims))
	for i := range rows {
		rows[i] = data[i*int(nDims) : (i+1)*int(nDims) : (i+1)*int(nDims)]
	}
	return rows
}
func Emb_ManVecRandom(v []float32, nDims int32) {
	for i := int32(0); i < nDims; i++ {
		v[i] = float32(Gia_ManRandom(0))
	}
}
func Emb_ManVecNormal(v []float32, nDims int32) {
	var norm float64
	wide := false
	for _, x := range v[:nDims] {
		square := x * x
		norm += float64(square)
		wide = wide || x != 0 && square < 0x1p-126
	}
	// Keep the source's float32 products for ordinary vectors. Recompute in
	// float64 if squaring overflowed or lost precision in the subnormal range.
	if wide || norm == 0 || math.IsInf(norm, 0) {
		norm = 0
		for _, x := range v[:nDims] {
			y := float64(x)
			norm += y * y
		}
		// Zero covariance can yield an iterate with no direction to normalize.
		if norm == 0 {
			return
		}
	}
	norm = math.Sqrt(norm)
	for i := int32(0); i < nDims; i++ {
		v[i] = float32(float64(v[i]) / norm)
	}
}
func Emb_ManVecMultiplyOne(a, b []float32, nDims int32) float32 {
	var sum float32
	for i := int32(0); i < nDims; i++ {
		sum += a[i] * b[i]
	}
	return sum
}
func Emb_ManVecCopyOne(dst, src []float32, nDims int32) { copy(dst[:nDims], src[:nDims]) }
func Emb_ManVecMultiply(matrix [][]float32, v []float32, nDims int32, result []float32) {
	for i := int32(0); i < nDims; i++ {
		result[i] = Emb_ManVecMultiplyOne(matrix[i], v, nDims)
	}
}
func Emb_ManVecOrthogonolizeOne(eigen, v []float32, nDims int32, result []float32) {
	// Preserve the source's sequential evaluation when result aliases v.
	for i := int32(0); i < nDims; i++ {
		result[i] = v[i] - eigen[i]*Emb_ManVecMultiplyOne(v, eigen, nDims)
	}
}
