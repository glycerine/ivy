package gia

import (
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"math/bits"
)

func giaSimWithPatterns(p *Gia_Man_t, patterns *Vec_Wrd_t) *Vec_Wrd_t {
	old := p.VSimsPi
	p.VSimsPi = patterns
	defer func() { p.VSimsPi = old }()
	return Gia_ManSimPatSim(p)
}
func Gia_SimCollectRare(p *Gia_Man_t, patterns *Vec_Wrd_t, limit int32) *Vec_Int_t {
	ci := Gia_ManCiNum(p)
	if ci <= 0 || vec.Vec_WrdSize(patterns)%int(ci) != 0 {
		panic("gia: invalid rare-pattern dimensions")
	}
	words := int32(vec.Vec_WrdSize(patterns)) / ci
	total := 64 * words
	sims := giaSimWithPatterns(p, patterns)
	data := vec.Vec_WrdArray(sims)
	result := vec.Vec_IntAlloc(100)
	for _, obj := range p.pObjs {
		if Gia_ObjIsAnd(obj) == 0 {
			continue
		}
		var count int32
		for _, word := range data[obj.id*words : (obj.id+1)*words] {
			count += int32(bits.OnesCount64(word))
		}
		rareOne := count < total/2
		rare := total - count
		phase := int32(0)
		if rareOne {
			rare = count
			phase = 1
		}
		if rare <= limit {
			vec.Vec_IntPushTwo(result, 2*obj.id+phase, rare)
		}
	}
	return result
}
func Gia_SimQualityImpact(p *Gia_Man_t, pattern, rare *Vec_Int_t) *Vec_Flt_t {
	ci := Gia_ManCiNum(p)
	if vec.Vec_IntSize(pattern) != int(ci) {
		panic("gia: invalid impact pattern")
	}
	words := util.Abc_Bit6WordNum(1 + ci)
	inputs := vec.Vec_WrdStart(int(ci * words))
	data := vec.Vec_WrdArray(inputs)
	for i, v := range vec.Vec_IntArray(pattern) {
		row := data[int32(i)*words : (int32(i)+1)*words]
		if v != 0 {
			for j := range row {
				row[j] = ^uint64(0)
			}
		}
		giaSimFlip(row, int32(i+1))
	}
	sims := giaSimWithPatterns(p, inputs)
	values := vec.Vec_WrdArray(sims)
	out := vec.Vec_FltStart(int(ci))
	entries := vec.Vec_IntArray(rare)
	if len(entries)%2 != 0 {
		panic("gia: invalid rare counts")
	}
	for i := 0; i < len(entries); i += 2 {
		lit, count := entries[i], entries[i+1]
		row := values[(lit>>1)*words : ((lit>>1)+1)*words]
		original := giaSimBit(row, 0)
		increment := float32(1) / float32(count+1)
		if original == uint64(lit&1) {
			increment = -increment
		}
		for k := int32(0); k < ci; k++ {
			if giaSimBit(row, k+1) != original {
				vec.Vec_FltAddToEntry(out, int(k), increment)
			}
		}
	}
	return out
}
func Gia_SimCollectBest(scores *Vec_Flt_t) *Vec_Int_t {
	values := vec.Vec_FltArray(scores)
	var best float32
	for _, v := range values {
		if v > best {
			best = v
		}
	}
	if best <= 0 {
		return nil
	}
	out := vec.Vec_IntAlloc(100)
	for i, v := range values {
		if v == best {
			vec.Vec_IntPush(out, int32(i))
		}
	}
	return out
}
func Gia_ManPatGetQuo(p *Gia_Man_t, rare *Vec_Int_t, sims *Vec_Wrd_t, pattern, words int32) float32 {
	if vec.Vec_WrdSize(sims) != int(p.nObjs*words) {
		panic("gia: invalid simulation quality dimensions")
	}
	var quality float32
	data := vec.Vec_WrdArray(sims)
	entries := vec.Vec_IntArray(rare)
	for i := 0; i < len(entries); i += 2 {
		lit, count := entries[i], entries[i+1]
		row := data[(lit>>1)*words : ((lit>>1)+1)*words]
		if giaSimBit(row, pattern) == uint64(lit&1) {
			quality += 1 / float32(count+1)
		}
	}
	return quality
}
func Gia_ManPatGetTotalQuo(p *Gia_Man_t, limit int32, patterns *Vec_Wrd_t, words int32) float32 {
	rare := Gia_SimCollectRare(p, patterns, limit)
	sims := giaSimWithPatterns(p, patterns)
	var total float32
	for i := int32(0); i < 64*words; i++ {
		total += Gia_ManPatGetQuo(p, rare, sims, i, words)
	}
	return total
}
func Gia_ManPatGetOneQuo(p *Gia_Man_t, limit int32, patterns *Vec_Wrd_t, words, pattern int32) float32 {
	rare := Gia_SimCollectRare(p, patterns, limit)
	return Gia_ManPatGetQuo(p, rare, giaSimWithPatterns(p, patterns), pattern, words)
}
