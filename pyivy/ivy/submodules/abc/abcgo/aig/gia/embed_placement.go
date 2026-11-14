package gia

import (
	"fmt"
	"time"
)

const GIA_PLACE_SIZE = 0x7fff

// embSortFloats follows Gia_SortFloats, including its permutation of ties.
func embSortFloats(a []float32) []int {
	perm := make([]int, len(a))
	for i := range perm {
		perm[i] = i
	}
	var sortRange func(int, int)
	sortRange = func(start, end int) {
		n := end - start
		if n <= 15 {
			for i := start; i < end-1; i++ {
				best := i
				for j := i + 1; j < end; j++ {
					if a[j] < a[best] {
						best = j
					}
				}
				a[i], a[best] = a[best], a[i]
				perm[i], perm[best] = perm[best], perm[i]
			}
			return
		}
		pivot := a[start+n/2]
		i, j := start-1, end
		for {
			for {
				i++
				if !(a[i] < pivot) {
					break
				}
			}
			for {
				j--
				if !(pivot < a[j]) {
					break
				}
			}
			if i >= j {
				break
			}
			a[i], a[j] = a[j], a[i]
			perm[i], perm[j] = perm[j], perm[i]
		}
		sortRange(start, i)
		sortRange(i, end)
	}
	sortRange(0, len(a))
	return perm
}
func Emb_ManDerivePlacement(p *Emb_Man_t, nSols int32) {
	if nSols != 2 {
		return
	}
	var perms [2][]int
	for d := int32(0); d < 2; d++ {
		values := Emb_ManSol(p, d)
		lo, hi := float32(embInfinity), float32(-embInfinity)
		for _, v := range values {
			lo = min(lo, v)
			hi = max(hi, v)
		}
		var scale float32
		if hi > lo {
			scale = float32(GIA_PLACE_SIZE) / (hi - lo)
		}
		for i, v := range values {
			values[i] = (v - lo) * scale
		}
		perms[d] = embSortFloats(values)
	}
	p.pPlacement = make([]uint16, 2*p.nObjs)
	for i := int32(0); i < p.nObjs; i++ {
		coordinate := uint16(int64(i) * GIA_PLACE_SIZE / int64(p.nObjs))
		p.pPlacement[2*perms[0][i]] = coordinate
		p.pPlacement[2*perms[1][i]+1] = coordinate
	}
}
func embBounds(p *Emb_Man_t, o *Emb_Obj_t) (uint16, uint16, uint16, uint16) {
	minX, maxX := p.pPlacement[2*o.Value], p.pPlacement[2*o.Value]
	minY, maxY := p.pPlacement[2*o.Value+1], p.pPlacement[2*o.Value+1]
	for _, next := range o.fanouts {
		x, y := p.pPlacement[2*next.Value], p.pPlacement[2*next.Value+1]
		minX = min(minX, x)
		maxX = max(maxX, x)
		minY = min(minY, y)
		maxY = max(maxY, y)
	}
	return minX, maxX, minY, maxY
}
func Emb_ManComputeHPWL(p *Emb_Man_t) float64 {
	if p.pPlacement == nil {
		return 0
	}
	var result float64
	for _, o := range p.objects {
		x0, x1, y0, y1 := embBounds(p, o)
		result += float64(int(x1) - int(x0) + int(y1) - int(y0))
	}
	return result
}
func Emb_ManPlacementRefine(p *Emb_Man_t, nIters, verbose int32) {
	if p.pPlacement == nil {
		return
	}
	edgeX, edgeY := make([]float32, p.nObjs), make([]float32, p.nObjs)
	vertX, vertY := make([]float32, p.nObjs), make([]float32, p.nObjs)
	start := time.Now()
	for iter := int32(0); iter < nIters; iter++ {
		var cost float64
		for _, o := range p.objects {
			x0, x1, y0, y1 := embBounds(p, o)
			edgeX[o.Value] = float32(int(x0)+int(x1)) * 0.5
			edgeY[o.Value] = float32(int(y0)+int(y1)) * 0.5
			cost += float64(int(x1) - int(x0) + int(y1) - int(y0))
		}
		for _, o := range p.objects {
			x, y := edgeX[o.Value], edgeY[o.Value]
			for _, fanin := range o.fanins {
				x += edgeX[fanin.Value]
				y += edgeY[fanin.Value]
			}
			vertX[o.Value] = x / float32(len(o.fanins)+1)
			vertY[o.Value] = y / float32(len(o.fanins)+1)
		}
		px, py := embSortFloats(vertX), embSortFloats(vertY)
		for k := int32(0); k < p.nObjs; k++ {
			coordinate := uint16(int64(k) * GIA_PLACE_SIZE / int64(p.nObjs))
			p.pPlacement[2*px[k]] = coordinate
			p.pPlacement[2*py[k]+1] = coordinate
		}
		if verbose != 0 {
			fmt.Printf("%2d : HPWL = %e  Time = %.2f sec\n", iter+1, cost, time.Since(start).Seconds())
		}
	}
}
