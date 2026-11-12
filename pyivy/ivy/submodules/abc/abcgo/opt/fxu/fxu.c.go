package fxu

import (
	"fmt"
	"os"
)

func Extract(data *Fxu_Data_t) (int32, error) {
	data.LastError = nil
	data.NewSOPs = nil
	data.NewFanins = nil
	p, err := createMatrix(data)
	if err != nil {
		data.LastError = err
		return -1, err
	}
	log := data.Log
	if log == nil {
		log = os.Stdout
	}
	count := int32(0)
	for count < data.NodesExt {
		var single *Fxu_Single
		var double *Fxu_Double
		weight := int32(-1)
		switch {
		case data.OnlyS:
			single = Fxu_HeapSingleReadMax(p.singleHeap)
			weight = Fxu_HeapSingleReadMaxWeight(p.singleHeap)
		case data.OnlyD:
			double = Fxu_HeapDoubleReadMax(p.doubleHeap)
			weight = Fxu_HeapDoubleReadMaxWeight(p.doubleHeap)
		case data.UseCompl:
			weight = Fxu_Select(p, &single, &double)
		default:
			ws, wd := Fxu_HeapSingleReadMaxWeight(p.singleHeap), Fxu_HeapDoubleReadMaxWeight(p.doubleHeap)
			if ws >= wd {
				single = Fxu_HeapSingleReadMax(p.singleHeap)
				weight = ws
			} else {
				double = Fxu_HeapDoubleReadMax(p.doubleHeap)
				weight = wd
			}
		}
		if single == nil && double == nil || !(weight > data.WeightMin || weight == 0 && data.Use0) {
			break
		}
		if data.Verbose {
			fmt.Fprintf(log, "Div %5d : selected weight = %5d\n", count, weight)
		}
		Fxu_Update(p, single, double)
		count++
	}
	if count > 0 {
		Fxu_CreateCovers(p, data)
	} else {
		data.NewSOPs = append([]string(nil), data.SOPs...)
		data.NewFanins = make([][]int32, len(data.Fanins))
		for i, f := range data.Fanins {
			data.NewFanins[i] = append([]int32(nil), f...)
		}
	}
	if data.Verbose {
		fmt.Fprintf(log, "Total single = %3d. Total double = %3d. Total compl = %3d.\n", p.SinglesTaken, p.DoublesTaken, p.ComplementsTaken)
	}
	return count, nil
}
func Fxu_FastExtract(data *Fxu_Data_t) int32 {
	count, err := Extract(data)
	data.LastError = err
	if err != nil {
		return -1
	}
	return count
}
func Fxu_MatrixRingCubesUnmark(p *Fxu_Matrix) { p.cubeRing = nil }
func Fxu_MatrixRingVarsUnmark(p *Fxu_Matrix)  { p.varRing = nil }

// Matrix memory is owned by Go. Explicit byte buffers remain available for
// clients of the old allocator API; recycling releases the caller's reference.
func Fxu_MemFetch(p *Fxu_Matrix, n int32) []byte {
	if n < 0 {
		panic("negative FXU allocation")
	}
	return make([]byte, n)
}
func Fxu_MemRecycle(p *Fxu_Matrix, item *[]byte, n int32) {
	if n < 0 || int(n) > len(*item) {
		panic("invalid FXU recycle size")
	}
	*item = nil
}
