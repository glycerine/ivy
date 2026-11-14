package gia

import (
	"github.com/glycerine/abcgo/misc/tim"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

// giaMfsNetworkData mirrors the network constructor inputs without coupling
// GIA to the still-incomplete SFM package.
type giaMfsNetworkData struct {
	fanins       [][]int
	truths       [][]uint64
	fixed, empty []bool
	pis, pos     int
}

func giaMfsExtract(p *Gia_Man_t) *giaMfsNetworkData {
	boxes := int(Gia_ManBoxNum(p))
	realPis, realPos := int(Gia_ManPiNum(p)), int(Gia_ManPoNum(p))
	var timing *tim.Tim_Man_t
	var blackIns, blackOuts int
	if p.PManTime != nil {
		timing = p.PManTime.(*tim.Tim_Man_t)
		tim.Tim_ManBlackBoxIoNum(timing, &blackIns, &blackOuts)
		if boxes > 0 {
			realPis, realPos = tim.Tim_ManPiNum(timing), tim.Tim_ManPoNum(timing)
		}
	}
	size := int(Gia_ManCiNum(p)+1+Gia_ManLutNum(p)+Gia_ManCoNum(p)) + blackIns + blackOuts
	d := &giaMfsNetworkData{fanins: make([][]int, size), truths: make([][]uint64, size), fixed: make([]bool, size), empty: make([]bool, size), pis: blackOuts + realPis, pos: blackIns + realPos}
	for i := range d.truths {
		d.truths[i] = []uint64{0}
	}
	count := blackOuts
	Gia_ManCleanCopyArray(p)
	for _, id := range vec.Vec_IntArray(p.vCis) {
		Gia_ObjSetCopyArray(p, id, int32(count))
		count++
	}
	d.fixed[count] = true
	Gia_ObjSetCopyArray(p, 0, int32(count))
	count++
	maxSize := max(Gia_ManLutSizeMax(p), 6)
	if maxSize >= 16 {
		panic("gia: MFS supports at most 15 LUT inputs")
	}
	leaves := vec.Vec_IntAlloc(int(maxSize))
	Gia_ObjComputeTruthTableStart(p, maxSize)
	for id := int32(1); id < p.nObjs; id++ {
		if Gia_ObjIsLut(p, id) == 0 {
			continue
		}
		vec.Vec_IntClear(leaves)
		var fanins []int32
		for _, fan := range Gia_ObjLutFanins(p, id) {
			copyID := Gia_ObjCopyArray(p, fan)
			if copyID < 0 {
				panic("gia: LUT fanin not mapped")
			}
			fanins = append(fanins, copyID)
			vec.Vec_IntPush(leaves, fan)
		}
		truth := append([]uint64(nil), Gia_ObjComputeTruthTableCut(p, Gia_ManObj(p, id), leaves)...)
		n := util.Abc_TtMinBase(truth, fanins, int32(len(fanins)), int32(len(fanins)))
		fanins = fanins[:n]
		for _, v := range fanins {
			d.fanins[count] = append(d.fanins[count], int(v))
		}
		words := 1
		if n > 6 {
			words = 1 << uint(n-6)
		}
		d.truths[count] = append([]uint64(nil), truth[:words]...)
		if Gia_ObjLutIsMux(p, id) != 0 {
			d.fixed[count], d.empty[count] = true, true
		}
		Gia_ObjSetCopyArray(p, id, int32(count))
		count++
	}
	Gia_ObjComputeTruthTableStop(p)
	for i, id := range vec.Vec_IntArray(p.vCos) {
		obj := Gia_ManObj(p, id)
		fan := Gia_ObjCopyArray(p, Gia_ObjFanin0(obj).id)
		if fan < 0 {
			panic("gia: output driver not mapped")
		}
		d.fanins[count] = []int{int(fan)}
		if i < int(Gia_ManCoNum(p))-realPos {
			d.fixed[count], d.empty[count] = true, true
			truth := uint64(0xaaaaaaaaaaaaaaaa)
			if Gia_ObjFaninC0(obj) != 0 {
				truth = ^truth
			}
			d.truths[count] = []uint64{truth}
		}
		Gia_ObjSetCopyArray(p, id, int32(count))
		count++
	}
	count += blackIns
	if count != size {
		panic("gia: inconsistent MFS node count")
	}
	if p.PAigExtra != nil {
		extra := p.PAigExtra
		if Gia_ManCiNum(extra) >= 16 {
			panic("gia: box truth exceeds 15 inputs")
		}
		Gia_ObjComputeTruthTableStart(extra, Gia_ManCiNum(extra))
		defer Gia_ObjComputeTruthTableStop(extra)
		ci, co, bbIn, bbOut := realPis, 0, 0, 0
		for box := 0; box < boxes; box++ {
			ins, outs := tim.Tim_ManBoxInputNum(timing, box), tim.Tim_ManBoxOutputNum(timing, box)
			if !tim.Tim_ManBoxIsBlack(timing, box) {
				vec.Vec_IntClear(leaves)
				for i := 0; i < ins; i++ {
					vec.Vec_IntPush(leaves, Gia_ManCi(extra, int32(i)).id)
				}
				for j := 0; j < outs; j++ {
					node := int(Gia_ObjCopyArray(p, Gia_ManCi(p, int32(ci+j)).id))
					fanins := make([]int32, ins)
					for i := 0; i < ins; i++ {
						fanins[i] = Gia_ObjCopyArray(p, Gia_ManCo(p, int32(co+i)).id)
					}
					d.fixed[node] = true
					output := Gia_ManCo(extra, int32(ci-realPis+j))
					driver := Gia_ObjFanin0(output)
					var truth []uint64
					if driver.id == 0 {
						truth = make([]uint64, max(1, 1<<uint(max(0, ins-6))))
					} else if Gia_ObjIsCi(driver) != 0 {
						truth = make([]uint64, max(1, 1<<uint(max(0, ins-6))))
						variable := uint(Gia_ObjCioId(driver))
						for k := range truth {
							for bit := uint(0); bit < 64; bit++ {
								if ((uint(k)*64+bit)>>variable)&1 != 0 {
									truth[k] |= 1 << bit
								}
							}
						}
					} else {
						truth = append([]uint64(nil), Gia_ObjComputeTruthTableCut(extra, driver, leaves)...)
					}
					if Gia_ObjFaninC0(output) != 0 {
						for i := range truth {
							truth[i] = ^truth[i]
						}
					}
					n := util.Abc_TtMinBase(truth, fanins, int32(ins), int32(ins))
					for _, v := range fanins[:n] {
						d.fanins[node] = append(d.fanins[node], int(v))
					}
					words := 1
					if n > 6 {
						words = 1 << uint(n-6)
					}
					d.truths[node] = append([]uint64(nil), truth[:words]...)
				}
			} else {
				for j := 0; j < outs; j++ {
					node := int(Gia_ObjCopyArray(p, Gia_ManCi(p, int32(ci+j)).id))
					d.fanins[node] = []int{bbOut}
					bbOut++
					d.fixed[node], d.empty[node] = true, true
					d.truths[node] = []uint64{0xaaaaaaaaaaaaaaaa}
				}
				for i := 0; i < ins; i++ {
					node := int(Gia_ObjCopyArray(p, Gia_ManCo(p, int32(co+i)).id))
					d.fanins[size-blackIns+bbIn] = []int{node}
					bbIn++
				}
			}
			for i := 0; i < ins; i++ {
				obj := Gia_ManCo(p, int32(co+i))
				if Gia_ObjIsCi(Gia_ObjFanin0(obj)) != 0 {
					d.empty[Gia_ObjCopyArray(p, Gia_ObjFanin0(obj).id)] = true
				}
			}
			co += ins
			ci += outs
		}
		if ci != int(Gia_ManCiNum(p)) || co+realPos != int(Gia_ManCoNum(p)) || ci-realPis != int(Gia_ManCoNum(extra)) || bbIn != blackIns || bbOut != blackOuts {
			panic("gia: inconsistent MFS box counts")
		}
	}
	return d
}
