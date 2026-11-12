// Port of ABC/MVSIS FXU (MVSIS Group, UC Berkeley, 2003).
// Original algorithm and licensing: src/opt/fxu; see repository copyright notice.
package fxu

import (
	"encoding/binary"
	"sort"
)

func pairParts(a, b []int32) (first, second, base []int32) {
	i, j := 0, 0
	for i < len(a) && j < len(b) {
		switch {
		case a[i] == b[j]:
			base = append(base, a[i])
			i++
			j++
		case a[i] < b[j]:
			first = append(first, a[i])
			i++
		default:
			second = append(second, b[j])
			j++
		}
	}
	first = append(first, a[i:]...)
	second = append(second, b[j:]...)
	if len(first) > 0 && len(second) > 0 && first[0] > second[0] {
		first, second = second, first
	}
	return
}
func pairKey(a, b []int32) string {
	data := make([]byte, 4+4*(len(a)+len(b)))
	binary.LittleEndian.PutUint32(data, uint32(len(a)))
	for i, v := range a {
		binary.LittleEndian.PutUint32(data[4+4*i:], uint32(v))
	}
	for i, v := range b {
		binary.LittleEndian.PutUint32(data[4+4*len(a)+4*i:], uint32(v))
	}
	return string(data)
}
func sameColumns(a, b []int32) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}
func Fxu_PairCanonicize(a, b **Fxu_Cube) {
	ca, cb := (*a).columns(), (*b).columns()
	i, j := 0, 0
	for i < len(ca) && j < len(cb) && ca[i] == cb[j] {
		i++
		j++
	}
	if i == len(ca) || j == len(cb) {
		return
	}
	if ca[i] > cb[j] {
		*a, *b = *b, *a
	}
}
func Fxu_PairCanonicize2(a, b **Fxu_Cube) {
	if (*a).Index > (*b).Index {
		*a, *b = *b, *a
	}
}
func Fxu_PairCompare(a, b *Fxu_Pair) int32 {
	return boolInt(sameColumns(a.First, b.First) && sameColumns(a.Second, b.Second))
}
func Fxu_PairHashKeyArray(p *Fxu_Matrix, a, b []int32, na, nb int32) uint32 {
	h := uint32(2166136261)
	for _, v := range []byte(pairKey(a[:na], b[:nb])) {
		h = (h ^ uint32(v)) * 16777619
	}
	return h
}
func Fxu_PairHashKey(p *Fxu_Matrix, a, b *Fxu_Cube, base, na, nb *int32) uint32 {
	first, second, common := pairParts(a.columns(), b.columns())
	*base = int32(len(common))
	*na = int32(len(first))
	*nb = int32(len(second))
	return Fxu_PairHashKeyArray(p, first, second, *na, *nb)
}
func Fxu_PairAlloc(p *Fxu_Matrix, a, b *Fxu_Cube) *Fxu_Pair {
	first, second, base := pairParts(a.columns(), b.columns())
	return &Fxu_Pair{Cube1: a, Cube2: b, First: first, Second: second, Base: base}
}
func Fxu_PairAdd(pair *Fxu_Pair) {
	if pair.Divisor != nil {
		Fxu_ListDoubleAddPairLast(pair.Divisor, pair)
	}
}
func Fxu_PairMinCube(pair *Fxu_Pair) *Fxu_Cube {
	if pair.Cube1.Index < pair.Cube2.Index {
		return pair.Cube1
	}
	return pair.Cube2
}
func Fxu_PairMaxCube(pair *Fxu_Pair) *Fxu_Cube {
	if pair.Cube1.Index > pair.Cube2.Index {
		return pair.Cube1
	}
	return pair.Cube2
}
func Fxu_PairMinCubeInt(pair *Fxu_Pair) int32 { return Fxu_PairMinCube(pair).Index }
func sortedColumns(columns []int32) []int32 {
	result := append([]int32(nil), columns...)
	sort.Slice(result, func(i, j int) bool { return result[i] < result[j] })
	return result
}
func boolInt(v bool) int32 {
	if v {
		return 1
	}
	return 0
}
