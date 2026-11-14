package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"math/bits"
)

func Gia_SimQualityOne(p *Gia_Man_t, pattern *Vec_Int_t, outputsOnly int32) *Vec_Int_t {
	inputs := Gia_ManCiNum(p)
	if vec.Vec_IntSize(pattern) != int(inputs) || inputs == 0 {
		panic("gia: invalid quality pattern")
	}
	words := util.Abc_Bit6WordNum(1 + inputs)
	simsIn := vec.Vec_WrdStart(int(inputs * words))
	data := vec.Vec_WrdArray(simsIn)
	for i, v := range vec.Vec_IntArray(pattern) {
		if v != 0 && v != 1 {
			panic("gia: invalid pattern value")
		}
		row := data[int32(i)*words : (int32(i)+1)*words]
		if v != 0 {
			for w := range row {
				row[w] = ^uint64(0)
			}
		}
		giaSimFlip(row, int32(i+1))
	}
	previous := p.VSimsPi
	p.VSimsPi = simsIn
	defer func() { p.VSimsPi = previous }()
	sims := Gia_ManSimPatSim(p)
	simData := vec.Vec_WrdArray(sims)
	count := func(id int32) int32 {
		row := simData[id*words : (id+1)*words]
		invert := row[0]&1 != 0
		var total int32
		for _, word := range row {
			if invert {
				word = ^word
			}
			total += int32(bits.OnesCount64(word))
		}
		return total
	}
	if outputsOnly != 0 {
		out := vec.Vec_IntStart(int(Gia_ManCoNum(p)))
		for i, id := range vec.Vec_IntArray(p.vCos) {
			vec.Vec_IntWriteEntry(out, i, count(id))
		}
		return out
	}
	out := vec.Vec_IntStart(int(p.nObjs))
	for _, obj := range p.pObjs {
		if Gia_ObjIsAnd(obj) != 0 {
			vec.Vec_IntWriteEntry(out, int(obj.id), count(obj.id))
		}
	}
	return out
}
func Gia_SimQualityTest(p *Gia_Man_t) {
	n := Gia_ManCiNum(p)
	if n > 10 {
		panic("gia: too many exhaustive quality inputs")
	}
	for mint := int32(0); mint < 1<<uint(n); mint++ {
		pat := vec.Vec_IntAlloc(int(n))
		for k := int32(0); k < n; k++ {
			vec.Vec_IntPush(pat, (mint>>uint(k))&1)
		}
		result := Gia_SimQualityOne(p, pat, 1)
		var total int32
		for _, v := range vec.Vec_IntArray(result) {
			total += v
		}
		fmt.Printf("%d : %0*b %d\n", mint, n, mint, total)
	}
}
func Gia_SimGenerateStats(p *Gia_Man_t) *Vec_Int_t {
	total := vec.Vec_IntStart(int(p.nObjs))
	util.Abc_Random(1)
	for i := 0; i < 1000; i++ {
		pat := vec.Vec_IntAlloc(int(Gia_ManCiNum(p)))
		for k := int32(0); k < Gia_ManCiNum(p); k++ {
			vec.Vec_IntPush(pat, int32(util.Abc_Random(0)&1))
		}
		result := Gia_SimQualityOne(p, pat, 0)
		for id, v := range vec.Vec_IntArray(result) {
			vec.Vec_IntAddToEntry(total, id, v)
		}
	}
	return total
}
func Gia_SimComputeScore(p *Gia_Man_t, total, current *Vec_Int_t) float64 {
	a, b := vec.Vec_IntArray(total), vec.Vec_IntArray(current)
	if len(a) != len(b) {
		panic("gia: mismatched simulation scores")
	}
	var score float64
	for i, v := range a {
		if v == 0 {
			v = 1
		}
		score += 1000 * float64(b[i]) / float64(v)
	}
	if score == 0 {
		return 1
	}
	if Gia_ManAndNum(p) == 0 {
		panic("gia: nonzero score without AND nodes")
	}
	return score / float64(Gia_ManAndNum(p))
}
func Gia_ManPatCollectOne(p *Gia_Man_t, patterns *Vec_Wrd_t, bit, words int32) *Vec_Int_t {
	out := vec.Vec_IntAlloc(int(Gia_ManCiNum(p)))
	data := vec.Vec_WrdArray(patterns)
	for id := int32(0); id < Gia_ManCiNum(p); id++ {
		vec.Vec_IntPush(out, int32(giaSimBit(data[id*words:(id+1)*words], bit)))
	}
	return out
}
func Gia_ManPatUpdateOne(p *Gia_Man_t, patterns *Vec_Wrd_t, bit, words int32, pattern *Vec_Int_t) {
	if vec.Vec_IntSize(pattern) != int(Gia_ManCiNum(p)) {
		panic("gia: invalid input pattern")
	}
	data := vec.Vec_WrdArray(patterns)
	for id, value := range vec.Vec_IntArray(pattern) {
		if value != 0 && value != 1 {
			panic("gia: invalid pattern bit")
		}
		row := data[int32(id)*words : (int32(id)+1)*words]
		if giaSimBit(row, bit) != uint64(value) {
			giaSimFlip(row, bit)
		}
	}
}
