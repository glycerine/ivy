package gia

import (
	"encoding/binary"
	"fmt"
	"github.com/glycerine/abcgo/misc/vec"
	"os"
	"time"
)

func frcCrossCut(p *Frc_Man_t, o *Frc_Obj_t, reverse bool) int32 {
	if o.iFanout <= 0 {
		panic("gia: invalid FORCE fanout count")
	}
	first := o.iFanout == int32(len(o.fanouts))
	o.iFanout--
	if first {
		p.nCutCur++
		p.nCutMax = max(p.nCutMax, p.nCutCur)
		if Frc_ObjIsCi(o) == 0 {
			if reverse {
				for i := len(o.fanins) - 1; i >= 0; i-- {
					p.nCutCur -= frcCrossCut(p, o.fanins[i], reverse)
				}
			} else {
				for _, fanin := range o.fanins {
					p.nCutCur -= frcCrossCut(p, fanin, reverse)
				}
			}
		}
	}
	return giaBool(o.iFanout == 0)
}
func Frc_ManCrossCut_rec(p *Frc_Man_t, o *Frc_Obj_t) int32  { return frcCrossCut(p, o, false) }
func Frc_ManCrossCut2_rec(p *Frc_Man_t, o *Frc_Obj_t) int32 { return frcCrossCut(p, o, true) }
func Frc_ManCrossCut(p *Frc_Man_t, order *Vec_Int_t, reverse int32) int32 {
	if vec.Vec_IntSize(order) != int(Frc_ManCoNum(p)) {
		panic("gia: invalid FORCE output order")
	}
	p.nCutCur, p.nCutMax = 0, 0
	for _, o := range p.objects {
		o.iFanout = int32(len(o.fanouts))
	}
	for _, id := range vec.Vec_IntArray(order) {
		o := Frc_ManObj(p, id)
		if Frc_ObjIsCo(o) == 0 {
			panic("gia: FORCE order contains non-output")
		}
		p.nCutCur -= frcCrossCut(p, o.fanins[0], reverse != 0)
	}
	if p.nCutCur != 0 {
		panic("gia: FORCE cross cut did not close")
	}
	return p.nCutMax
}
func Frc_ManCollectCos(p *Frc_Man_t) *Vec_Int_t { return vec.Vec_IntDup(p.vCos) }
func Frc_ManCrossCutTest(p *Frc_Man_t, order *Vec_Int_t) {
	if order == nil {
		order = Frc_ManCollectCos(p)
	}
	fmt.Printf("CrossCut = %6d\nCrossCut = %6d\n", Frc_ManCrossCut(p, order, 0), Frc_ManCrossCut(p, order, 1))
	vec.Vec_IntReverseOrder(order)
	fmt.Printf("CrossCut = %6d\nCrossCut = %6d\n", Frc_ManCrossCut(p, order, 0), Frc_ManCrossCut(p, order, 1))
	vec.Vec_IntReverseOrder(order)
}
func Frc_ManPlaceRandom(p *Frc_Man_t) {
	order := make([]int32, p.nObjs)
	for i := range order {
		order[i] = int32(i)
	}
	for i := range order {
		j := int(Gia_ManRandom(0) % uint32(len(order)))
		order[i], order[j] = order[j], order[i]
	}
	for i, o := range p.objects {
		o.pPlace = order[i]
	}
}
func Frc_ManArrayShuffle(v *Vec_Int_t) {
	a := vec.Vec_IntArray(v)
	for i := range a {
		j := int(Gia_ManRandom(0) % uint32(len(a)))
		a[i], a[j] = a[j], a[i]
	}
}
func Frc_ManPlaceDfs_rec(p *Frc_Man_t, o *Frc_Obj_t, place *int32) {
	if o.iFanout <= 0 {
		panic("gia: invalid FORCE fanout count")
	}
	first := o.iFanout == int32(len(o.fanouts))
	o.iFanout--
	if first {
		if Frc_ObjIsCi(o) == 0 {
			for _, fanin := range o.fanins {
				Frc_ManPlaceDfs_rec(p, fanin, place)
			}
		}
		o.pPlace = *place
		*place++
	}
}
func Frc_ManPlaceDfs(p *Frc_Man_t, order *Vec_Int_t) {
	var n int32
	for _, o := range p.objects {
		o.iFanout = int32(len(o.fanouts))
		if len(o.fanouts) == 0 && Frc_ObjIsCo(o) == 0 {
			o.pPlace = n
			n++
		}
	}
	for _, id := range vec.Vec_IntArray(order) {
		o := Frc_ManObj(p, id)
		if Frc_ObjIsCo(o) == 0 {
			panic("gia: FORCE order contains non-output")
		}
		Frc_ManPlaceDfs_rec(p, o.fanins[0], &n)
		o.pPlace = n
		n++
	}
	if n != p.nObjs {
		panic("gia: incomplete FORCE placement")
	}
}
func Frc_ManPlaceDfsBoth(p *Frc_Man_t, order *Vec_Int_t, other *int32) int32 {
	a := Frc_ManCrossCut(p, order, 0)
	vec.Vec_IntReverseOrder(order)
	b := Frc_ManCrossCut(p, order, 0)
	if a <= b {
		vec.Vec_IntReverseOrder(order)
		Frc_ManPlaceDfs(p, order)
		*other = b
		return a
	}
	Frc_ManPlaceDfs(p, order)
	vec.Vec_IntReverseOrder(order)
	*other = a
	return b
}
func Frc_ManPlacementRefine(p *Frc_Man_t, iters, verbose int32) {
	start := time.Now()
	order := Frc_ManCollectCos(p)
	var other int32
	initial := Frc_ManPlaceDfsBoth(p, order, &other)
	positions := make([]float32, p.nObjs)
	for iter := int32(0); iter < iters; iter++ {
		var span float64
		for _, o := range p.objects {
			low, high := o.pPlace, o.pPlace
			for _, next := range o.fanouts {
				low = min(low, next.pPlace)
				high = max(high, next.pPlace)
			}
			o.fEdgeCenter = float32(low+high) * 0.5
			span += float64(high - low)
		}
		for i, o := range p.objects {
			center := o.fEdgeCenter
			for _, fanin := range o.fanins {
				center += fanin.fEdgeCenter
			}
			positions[i] = center / float32(len(o.fanins)+1)
		}
		permutation := embSortFloats(positions)
		vec.Vec_IntClear(order)
		for i, id := range permutation {
			obj := p.objects[id]
			obj.pPlace = int32(i)
			if Frc_ObjIsCo(obj) != 0 {
				vec.Vec_IntPush(order, obj.hHandle)
			}
		}
		cut := Frc_ManPlaceDfsBoth(p, order, &other)
		if verbose != 0 {
			percent := float64(0)
			if initial != 0 {
				percent = 100 * float64(initial-cut) / float64(initial)
			}
			fmt.Printf("%2d : Span = %e  Cut = %6d  (%5.2f %%)  CutR = %6d  Total = %.2f sec\n", iter+1, span, cut, percent, other, time.Since(start).Seconds())
		}
	}
}
func Frc_ObjFanoutsAreCos(o *Frc_Obj_t) int32 {
	for _, fanout := range o.fanouts {
		if Frc_ObjIsCo(fanout) == 0 {
			return 0
		}
	}
	return 1
}
func Frc_DumpGraphIntoFile(p *Frc_Man_t) {
	ids := make(map[*Frc_Obj_t]int32)
	var n int32
	for i, o := range p.objects {
		if i != 0 && (Frc_ObjIsCi(o) != 0 && Frc_ObjFanoutsAreCos(o) == 0 || Frc_ObjIsNode(o) != 0) {
			ids[o] = n
			n++
		}
	}
	f, err := os.Create("x\\large\\aig\\dg1.g")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()
	for _, o := range p.objects {
		id, ok := ids[o]
		if !ok {
			continue
		}
		for _, next := range o.fanouts {
			if target, ok := ids[next]; ok {
				fmt.Fprintf(f, "%d %d\n", id, target)
			}
		}
	}
}
func For_ManExperiment(g *Gia_Man_t, iters, cluster, verbose int32) {
	Gia_ManRandom(1)
	var p *Frc_Man_t
	if cluster != 0 {
		p = Frc_ManStart(g)
	} else {
		p = Frc_ManStartSimple(g)
	}
	if verbose != 0 {
		Frc_ManPrintStats(p)
	}
	Frc_ManPlacementRefine(p, iters, verbose)
	Frc_ManStop(p)
}
func For_ManFileExperiment() {
	const size = 1 << 25
	start := time.Now()
	data := make([]byte, 4*size)
	for i := 0; i < size; i++ {
		binary.LittleEndian.PutUint32(data[4*i:], uint32(i))
	}
	fmt.Printf("2^25 machine words (%d bytes).\nFillup = %.2f sec\n", len(data), time.Since(start).Seconds())
	start = time.Now()
	if f, err := os.Open("test.txt"); err == nil {
		_, _ = f.Read(data)
		f.Close()
	}
	fmt.Printf("Read = %.2f sec\n", time.Since(start).Seconds())
	start = time.Now()
	if err := os.WriteFile("test.txt", data, 0666); err != nil {
		fmt.Println(err)
	}
	fmt.Printf("Write = %.2f sec\n", time.Since(start).Seconds())
}
