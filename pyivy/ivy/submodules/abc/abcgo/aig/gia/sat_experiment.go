package gia

import (
	"fmt"
	"time"
)

func Gia_ManSatPartPrint(p *Gia_Man_t, obj *Gia_Obj_t, step int32) {
	if Gia_ObjIsAnd(obj) == 0 || obj.fMark0 != 0 {
		panic("gia: invalid SAT partition node")
	}
	id := Gia_ObjId(p, obj)
	for i, lit := range [2]int32{Gia_ObjFaninLit0(obj, id), Gia_ObjFaninLit1(obj, id)} {
		if i == 1 {
			if step&1 != 0 {
				fmt.Print(" + ")
			} else {
				fmt.Print("*")
			}
		}
		fan := Gia_ManObj(p, lit>>1)
		if fan.fMark0 != 0 {
			if lit&1 != 0 {
				fmt.Print("!")
			}
			fmt.Print(lit >> 1)
			continue
		}
		if lit&1 != 0 {
			fmt.Print("(")
		}
		Gia_ManSatPartPrint(p, fan, step+(lit&1))
		if lit&1 != 0 {
			fmt.Print(")")
		}
	}
}

// Gia_ManSatExperiment partitions an AIG and prints partition size statistics.
func Gia_ManSatExperiment(p *Gia_Man_t) {
	started := time.Now()
	manager := Gia_ManSatStart()
	defer Gia_ManSatStop(manager)
	defer Gia_ManCleanMark0(p)
	Gia_ManCreateValueRefs(p)
	for id := int32(0); id < p.nObjs; id++ {
		obj := Gia_ManObj(p, id)
		obj.fMark0 = 0
		switch {
		case Gia_ObjIsCo(obj) != 0:
			Gia_ObjFanin0(obj).fMark0 = 1
		case Gia_ObjIsCi(obj) != 0:
			obj.fMark0 = 1
		case Gia_ObjIsAnd(obj) != 0:
			if obj.Value > 1 || Gia_ManSatPartCountNodes(p, obj) >= GIA_LIMIT {
				obj.fMark0 = 1
			}
		}
		obj.Value = 0
	}
	var counts [2*GIA_LIMIT + 2]int32
	var all, words, packed int32
	var storage [1000]uint32
	for id := int32(0); id < p.nObjs; id++ {
		obj := Gia_ManObj(p, id)
		if Gia_ObjIsAnd(obj) == 0 || obj.fMark0 == 0 {
			continue
		}
		obj.fMark0 = 0
		all++
		n := Gia_ManSatPartCreate(p, obj, storage[:])
		packed += n
		manager.partitions = append(manager.partitions, append([]uint32(nil), storage[:n]...))
		var leaves, nodes int32
		Gia_ManSatPartCount(p, obj, &leaves, &nodes)
		words += leaves + nodes
		if nodes <= 2*GIA_LIMIT {
			counts[nodes]++
		} else {
			counts[2*GIA_LIMIT+1]++
		}
		obj.fMark0 = 1
	}
	fmt.Println()
	for i, count := range counts {
		var percentage, nodePercentage float64
		if all > 0 {
			percentage = 100 * float64(count) / float64(all)
		}
		if ands := Gia_ManAndNum(p); ands > 0 {
			nodePercentage = 100 * float64(i) * float64(count) / float64(ands)
		}
		fmt.Printf("%2d=%6d  %7.2f %%  %7.2f %%\n", i, count, percentage, nodePercentage)
	}
	bpn := float64(0)
	if p.nObjs > 0 {
		bpn = 4 * float64(packed) / float64(p.nObjs)
	}
	fmt.Printf("MemoryEst = %10.3f MB  MemoryReal = %10.3f MB  %5.2f bpn  Time = %.2f sec\n", 4*float64(words)/(1<<20), 4*float64(packed)/(1<<20), bpn, time.Since(started).Seconds())
}
