package gia

// A port of ABC's src/opt/dau/dauGia.c onto Gia_Man_t: DSD-based conversion
// of LUT truth tables into AIG logic (Dsm_ManTruthToGia) and LUT unmapping
// (Dsm_ManDeriveGia). ABC builds the result directly in a hashed GIA whose
// levels steer balancing of AND/XOR groups, so object order and structure
// follow from this exact construction; the opt/dau aig.Graph versions do not
// reproduce it.

import (
	"github.com/glycerine/abcgo/bool/kit"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/opt/dau"
)

const dauDsdMaxVar = 12 // DAU_DSD_MAX_VAR

// dauDsdParser walks a DSD expression like Dau_DsdToGia_rec's char pointer.
type dauDsdParser struct {
	p       *Gia_Man_t
	str     string
	matches []int32
	pos     int
	scratch []uint32
}

func dauLitNotCond(lit int32, c bool) int32 {
	if c {
		return lit ^ 1
	}
	return lit
}

// Dau_DsdAddToArray keeps pFans ordered by decreasing level.
func dauDsdAddToArray(p *Gia_Man_t, fans []int32, iFan int32) []int32 {
	fans = append(fans, iFan)
	for i := len(fans) - 1; i > 0; i-- {
		if Gia_ObjLevelId(p, fans[i]>>1) <= Gia_ObjLevelId(p, fans[i-1]>>1) {
			break
		}
		fans[i], fans[i-1] = fans[i-1], fans[i]
	}
	return fans
}

// Dau_DsdBalance combines the two lowest-level fanins until one remains.
func dauDsdBalance(p *Gia_Man_t, fans []int32, and bool) int32 {
	for len(fans) > 1 {
		n := len(fans)
		iFan0, iFan1 := fans[n-1], fans[n-2]
		fans = fans[:n-2]
		var iFan int32
		if vec.Vec_IntSize(&p.vHTable) == 0 {
			if and {
				iFan = Gia_ManAppendAnd2(p, iFan0, iFan1)
			} else if p.PMuxes != nil {
				compl := (iFan0 ^ iFan1) & 1
				iFan = Gia_ManAppendXorReal(p, iFan0&^1, iFan1&^1) ^ compl
			} else {
				iFan = Gia_ManAppendXor2(p, iFan0, iFan1)
			}
		} else {
			if and {
				iFan = Gia_ManHashAnd(p, iFan0, iFan1)
			} else if p.PMuxes != nil {
				iFan = Gia_ManHashXorReal(p, iFan0, iFan1)
			} else {
				iFan = Gia_ManHashXor(p, iFan0, iFan1)
			}
		}
		obj := Gia_ManObj(p, iFan>>1)
		if Gia_ObjIsAnd(obj) != 0 {
			if and {
				Gia_ObjSetAndLevel(p, obj)
			} else if p.PMuxes != nil {
				Gia_ObjSetXorLevel(p, obj)
			} else {
				if Gia_ObjIsAnd(Gia_ObjFanin0(obj)) != 0 {
					Gia_ObjSetAndLevel(p, Gia_ObjFanin0(obj))
				}
				if Gia_ObjIsAnd(Gia_ObjFanin1(obj)) != 0 {
					Gia_ObjSetAndLevel(p, Gia_ObjFanin1(obj))
				}
				Gia_ObjSetAndLevel(p, obj)
			}
		}
		fans = dauDsdAddToArray(p, fans, iFan)
	}
	return fans[0]
}

// muxParts reads a MUX "<...>" with optional "{...}" fanins (shared by both
// recursive builders). rec builds one subexpression with the given literals.
func (d *dauDsdParser) muxParts(lits []int32, rec func([]int32) int32) [3]int32 {
	start := d.pos
	q := int(d.matches[d.pos])
	lits2 := lits
	if q+1 < len(d.str) && d.str[q+1] == '{' {
		d.pos = q + 1
		q2 := int(d.matches[d.pos])
		var fanins []int32
		for d.pos++; d.pos < q2; d.pos++ {
			fanins = append(fanins, rec(lits))
		}
		lits2 = fanins
	}
	d.pos = start
	q = int(d.matches[d.pos])
	var temp [3]int32
	k := 0
	for d.pos++; d.pos < q; d.pos++ {
		temp[k] = rec(lits2)
		k++
	}
	if q+1 < len(d.str) && d.str[q+1] == '{' {
		d.pos++
		d.pos = int(d.matches[d.pos])
	}
	return temp
}

// hexFunction reads "<hex>{...}" and returns its truth table and fanins.
func (d *dauDsdParser) hexFunction(lits []int32, rec func([]int32) int32) ([]word, []int32) {
	fn := make([]word, 1<<max(0, dauDsdMaxVar-6))
	nVars := util.Abc_TtReadHex(fn, d.str[d.pos:])
	d.pos += int(util.Abc_TtHexDigitNum(nVars))
	q := int(d.matches[d.pos])
	var fanins []int32
	for d.pos++; d.pos < q; d.pos++ {
		fanins = append(fanins, rec(lits))
	}
	if len(fanins) != int(nVars) {
		panic("dau: DSD fanin count mismatch")
	}
	return fn, fanins
}

func dauTruthWords(fn []word) []uint32 {
	words := make([]uint32, 2*len(fn))
	for i, w := range fn {
		words[2*i], words[2*i+1] = uint32(w), uint32(w>>32)
	}
	return words
}

// Dau_DsdToGia_rec: level-aware construction.
func (d *dauDsdParser) toGiaRec(lits []int32) int32 {
	p := d.p
	compl := false
	if d.str[d.pos] == '!' {
		d.pos++
		compl = true
	}
	c := d.str[d.pos]
	switch {
	case c >= 'a' && c < 'a'+dauDsdMaxVar:
		return dauLitNotCond(lits[c-'a'], compl)
	case c == '(' || c == '[':
		q := int(d.matches[d.pos])
		var fans []int32
		for d.pos++; d.pos < q; d.pos++ {
			fans = dauDsdAddToArray(p, fans, d.toGiaRec(lits))
		}
		return dauLitNotCond(dauDsdBalance(p, fans, c == '('), compl)
	case c == '<':
		temp := d.muxParts(lits, d.toGiaRec)
		var res int32
		if vec.Vec_IntSize(&p.vHTable) == 0 {
			if p.PMuxes != nil {
				res = Gia_ManAppendMux(p, temp[0], temp[1], temp[2])
			} else {
				res = Gia_ManAppendMux2(p, temp[0], temp[1], temp[2])
			}
		} else if p.PMuxes != nil {
			res = Gia_ManHashMuxReal(p, temp[0], temp[1], temp[2])
		} else {
			res = Gia_ManHashMux(p, temp[0], temp[1], temp[2])
		}
		obj := Gia_ManObj(p, res>>1)
		if Gia_ObjIsAnd(obj) != 0 {
			if p.PMuxes != nil && vec.Vec_IntSize(&p.vHTable) != 0 {
				Gia_ObjSetMuxLevel(p, obj)
			} else {
				if Gia_ObjIsAnd(Gia_ObjFanin0(obj)) != 0 {
					Gia_ObjSetAndLevel(p, Gia_ObjFanin0(obj))
				}
				if Gia_ObjIsAnd(Gia_ObjFanin1(obj)) != 0 {
					Gia_ObjSetAndLevel(p, Gia_ObjFanin1(obj))
				}
				Gia_ObjSetAndLevel(p, obj)
			}
		}
		return dauLitNotCond(res, compl)
	case c >= 'A' && c <= 'F' || c >= '0' && c <= '9':
		fn, fanins := d.hexFunction(lits, d.toGiaRec)
		nObjOld := Gia_ManObjNum(p)
		hash := int32(0)
		if vec.Vec_IntSize(&p.vHTable) != 0 {
			hash = 1
		}
		res := kit.Kit_TruthToGia(p, dauTruthWords(fn), int32(len(fanins)), &d.scratch, fanins, hash)
		for i := nObjOld; i < Gia_ManObjNum(p); i++ {
			Gia_ObjSetGateLevel(p, Gia_ManObj(p, i))
		}
		return dauLitNotCond(res, compl)
	}
	panic("dau: invalid DSD expression")
}

// Dau_DsdToGia2_rec: construction without levels.
func (d *dauDsdParser) toGia2Rec(lits []int32) int32 {
	p := d.p
	compl := false
	if d.str[d.pos] == '!' {
		d.pos++
		compl = true
	}
	c := d.str[d.pos]
	switch {
	case c >= 'a' && c < 'a'+dauDsdMaxVar:
		return dauLitNotCond(lits[c-'a'], compl)
	case c == '(':
		q := int(d.matches[d.pos])
		res := int32(1)
		for d.pos++; d.pos < q; d.pos++ {
			res = Gia_ManHashAnd(p, res, d.toGia2Rec(lits))
		}
		return dauLitNotCond(res, compl)
	case c == '[':
		q := int(d.matches[d.pos])
		res := int32(0)
		for d.pos++; d.pos < q; d.pos++ {
			lit := d.toGia2Rec(lits)
			if p.PMuxes != nil {
				res = Gia_ManHashXorReal(p, res, lit)
			} else {
				res = Gia_ManHashXor(p, res, lit)
			}
		}
		return dauLitNotCond(res, compl)
	case c == '<':
		temp := d.muxParts(lits, d.toGia2Rec)
		var res int32
		if p.PMuxes != nil {
			res = Gia_ManHashMuxReal(p, temp[0], temp[1], temp[2])
		} else {
			res = Gia_ManHashMux(p, temp[0], temp[1], temp[2])
		}
		return dauLitNotCond(res, compl)
	case c >= 'A' && c <= 'F' || c >= '0' && c <= '9':
		fn, fanins := d.hexFunction(lits, d.toGia2Rec)
		res := kit.Kit_TruthToGia(p, dauTruthWords(fn), int32(len(fanins)), &d.scratch, fanins, 1)
		return dauLitNotCond(res, compl)
	}
	panic("dau: invalid DSD expression")
}

// dauDsdToGia is Dau_DsdToGia (levels) or Dau_DsdToGia2.
func dauDsdToGia(p *Gia_Man_t, dsd string, lits []int32, levels bool) int32 {
	if dsd == "0" {
		return 0
	}
	if dsd == "1" {
		return 1
	}
	d := &dauDsdParser{p: p, str: dsd, matches: dau.Dau_DsdComputeMatches(dsd)}
	var res int32
	if levels {
		res = d.toGiaRec(lits)
	} else {
		res = d.toGia2Rec(lits)
	}
	if d.pos+1 != len(dsd) {
		panic("dau: unparsed DSD expression")
	}
	return res
}

// Dsm_ManTruthToGia converts a LUT function over leaf literals into logic in p.
func Dsm_ManTruthToGia(p *Gia_Man_t, truth []word, leaves []int32) int32 {
	n := len(leaves)
	if n > dauDsdMaxVar {
		panic("dau: LUT exceeds twelve leaves")
	}
	if n == 0 {
		return int32(truth[0] & 1)
	}
	if n == 1 {
		return dauLitNotCond(leaves[0], truth[0]&1 != 0)
	}
	copyWords := append([]uint64(nil), truth[:util.Abc_TtWordNum(int32(n))]...)
	var dsd []byte
	dau.Dau_DsdDecompose(copyWords, int32(n), 0, 1, &dsd)
	text := string(dsd)
	if i := indexByte0(text); i >= 0 {
		text = text[:i]
	}
	return dauDsdToGia(p, text, leaves, p.vLevels != nil)
}

func indexByte0(s string) int {
	for i := 0; i < len(s); i++ {
		if s[i] == 0 {
			return i
		}
	}
	return -1
}

// Dsm_ManDeriveGia unmaps a LUT-mapped GIA into a hashed AIG (dauGia.c).
func Dsm_ManDeriveGia(p *Gia_Man_t, useMuxes int32) *Gia_Man_t {
	if Gia_ManHasMapping(p) == 0 {
		panic("gia: Dsm_ManDeriveGia requires a mapping")
	}
	pNew := Gia_ManStart(6*Gia_ManObjNum(p)/5 + 100)
	pNew.pName, pNew.PSpec = p.pName, p.PSpec
	pNew.vLevels = vec.Vec_IntStart(int(6*Gia_ManObjNum(p)/5 + 100))
	if useMuxes != 0 {
		pNew.PMuxes = make([]uint32, pNew.nObjsAlloc)
	}
	Gia_ManFillValue(p)
	Gia_ManConst0(p).Value = 0
	for i := int32(0); i < Gia_ManCiNum(p); i++ {
		Gia_ManCi(p, i).Value = uint32(Gia_ManAppendCi(pNew))
	}
	leaves := vec.Vec_IntAlloc(16)
	Gia_ManHashStart(pNew)
	Gia_ObjComputeTruthTableStart(p, Gia_ManLutSizeMax(p))
	for id := int32(1); id < Gia_ManObjNum(p); id++ {
		obj := Gia_ManObj(p, id)
		if Gia_ObjIsAnd(obj) == 0 {
			continue
		}
		if Gia_ObjIsBuf(obj) != 0 {
			obj.Value = uint32(Gia_ManAppendBuf(pNew, Gia_ObjFanin0Copy(obj)))
			continue
		}
		if Gia_ObjIsLut(p, id) == 0 {
			continue
		}
		vec.Vec_IntClear(leaves)
		for _, fanin := range Gia_ObjLutFanins(p, id) {
			vec.Vec_IntPush(leaves, fanin)
		}
		truth := Gia_ObjComputeTruthTableCut(p, obj, leaves)
		lits := make([]int32, 0, vec.Vec_IntSize(leaves))
		for _, fanin := range Gia_ObjLutFanins(p, id) {
			lits = append(lits, int32(Gia_ManObj(p, fanin).Value))
		}
		obj.Value = uint32(Dsm_ManTruthToGia(pNew, truth, lits))
	}
	Gia_ObjComputeTruthTableStop(p)
	for i := int32(0); i < Gia_ManCoNum(p); i++ {
		co := Gia_ManCo(p, i)
		co.Value = uint32(Gia_ManAppendCo(pNew, Gia_ObjFanin0Copy(co)))
	}
	Gia_ManHashStop(pNew)
	Gia_ManSetRegNum(pNew, Gia_ManRegNum(p))
	result := Gia_ManCleanup(pNew)
	Gia_ManStop(pNew)
	return result
}
