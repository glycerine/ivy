package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/opt/dau"
	"time"
)

func Gia_ManResubTest(p *Gia_Man_t) {
	start := time.Now()
	q := Gia_ManDupMuxes(p, 2)
	mffcs := Gia_ManComputeMffcs(q, 4, 100, 8, 100)
	Gia_ManAddDivisors(q, mffcs)
	fmt.Printf("Time = %s\n", time.Since(start).Round(time.Millisecond))
	Gia_ManStop(q)
}
func Gia_ManResubTest3() {
	vars := []word{0xaaaaaaaaaaaaaaaa, 0xcccccccccccccccc, 0xf0f0f0f0f0f0f0f0, 0xff00ff00ff00ff00}
	Abc_ResubPrepareManager(1)
	defer Abc_ResubPrepareManager(0)
	for i := 0; i <= 1000; i++ {
		truth := util.Abc_Tt6Stretch(word(i), 4)
		fmt.Printf("%3d : %04x ", i, i)
		dau.Dau_DsdPrintFromTruth2([]word{truth}, 4)
		divs := [][]word{{^truth}, {truth}, {vars[0]}, {vars[1]}, {vars[2]}, {vars[3]}}
		Abc_ResubComputeFunction(divs, 1, 16, 50, 0, 0, 1, 1)
		fmt.Println()
	}
}
func Gia_ManResubTest3_() {
	a, b, c, d := word(0xaaaaaaaaaaaaaaaa), word(0xcccccccccccccccc), word(0xf0f0f0f0f0f0f0f0), word(0xff00ff00ff00ff00)
	truth := (a | b) & (c & d)
	divs := vec.Vec_PtrAlloc(6)
	for _, v := range []word{^truth, truth, a, b, c, d} {
		vec.Vec_PtrPush(divs, []word{v})
	}
	fmt.Printf("%016x ", truth)
	dau.Dau_DsdPrintFromTruth2([]word{truth}, 6)
	p := Gia_ResbAlloc(1)
	Gia_ManResubPerform(p, divs, 1, 100, 0, 50, 1, 1, 0)
	Gia_ResbFree(p)
}

// resubSimulate produces complete truth tables in object-ID order.
func resubSimulate(p *Gia_Man_t) ([][]word, int) {
	n := int(Gia_ManCiNum(p))
	if n >= 16 {
		panic("gia: resubstitution supports fewer than 16 inputs")
	}
	inputs := vec.Vec_WrdStartTruthTables(n)
	words := 1
	if n > 6 {
		words = 1 << (n - 6)
	}
	p.VSimsPi = inputs
	data := make([][]word, p.nObjs)
	for i := range data {
		data[i] = make([]word, words)
	}
	for i, id := range vec.Vec_IntArray(p.vCis) {
		copy(data[id], vec.Vec_WrdArray(inputs)[i*words:(i+1)*words])
	}
	read := func(lit int32, i int) word {
		v := data[lit>>1][i]
		if lit&1 != 0 {
			v = ^v
		}
		return v
	}
	for id := int32(1); id < p.nObjs; id++ {
		a, b, c, kind := p.Gate(id)
		if kind == 1 {
			continue
		}
		for i := 0; i < words; i++ {
			switch kind {
			case 2:
				data[id][i] = read(a, i) & read(b, i)
			case 3:
				data[id][i] = read(a, i) ^ read(b, i)
			case 4:
				control := read(a, i)
				data[id][i] = (control & read(b, i)) | (^control & read(c, i))
			case 5, 6:
				data[id][i] = read(a, i)
			default:
				panic("gia: invalid resubstitution simulation gate")
			}
		}
	}
	return data, words
}
func Gia_ManTryResub(p *Gia_Man_t) {
	start := time.Now()
	sims, nWords := resubSimulate(p)
	Gia_ManLevelNum(p)
	Gia_ManCreateRefs(p)
	Gia_ManStaticFanoutStart(p)
	defer Gia_ManStaticFanoutStop(p)
	Abc_ResubPrepareManager(int32(nWords))
	defer Abc_ResubPrepareManager(0)
	total, unrealizable := int32(0), int32(0)
	var resubTime time.Duration
	for _, obj := range p.pObjs {
		if Gia_ObjIsAnd(obj) == 0 {
			continue
		}
		id := obj.id
		level := Gia_ObjLevel(p, obj)
		size := Gia_NodeMffcSizeMark(p, obj)
		off := make([]word, nWords)
		for i, v := range sims[id] {
			off[i] = ^v
		}
		divs := [][]word{off, sims[id]}
		Gia_ManUnivTfo(p, []int32{id}, nil, nil)
		for _, ci := range vec.Vec_IntArray(p.vCis) {
			divs = append(divs, sims[ci])
		}
		for _, candidate := range p.pObjs {
			if Gia_ObjIsAnd(candidate) != 0 && Gia_ObjIsTravIdCurrent(p, candidate) == 0 && Gia_ObjIsTravIdPrevious(p, candidate) == 0 && Gia_ObjLevel(p, candidate) <= level {
				divs = append(divs, sims[candidate.id])
			}
		}
		t := time.Now()
		gates := Abc_ResubComputeFunction(divs, int32(nWords), min(size-1, 20), 200, 0, 1, 1, 0)
		resubTime += time.Since(t)
		if len(gates) > 0 {
			total += size - int32(len(gates)/2)
		} else {
			unrealizable++
		}
	}
	fmt.Printf("Total nodes = %5d.  Non-realizable = %5d.  Gain = %6d.  Time = %s  Pure resub time = %s\n", Gia_ManAndNum(p), unrealizable, total, time.Since(start).Round(time.Millisecond), resubTime.Round(time.Millisecond))
}
