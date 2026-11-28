package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/internal/testoracle"
	"slices"
	"strings"
	"testing"
)

func genReviewGraph(typed bool, signs int32, reorder bool) *Gia_Man_t {
	p := Gia_ManStart(64)
	if typed {
		p.EnableMuxes()
	}
	a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	var x int32
	if typed {
		x = Gia_ManAppendXorReal(p, a^(signs&1), b^(signs>>1&1))
	} else {
		x = Gia_ManAppendAnd(p, a^(signs&1), b^(signs>>1&1))
	}
	Gia_ManAppendCo(p, x^(signs>>2&1))
	c, d := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	var root int32
	if typed {
		root = Gia_ManAppendMuxReal(p, c^(signs>>3&1), x^(signs>>4&1), d^(signs>>5&1))
	} else {
		z := Gia_ManAppendAnd(p, x, c^(signs>>3&1))
		root = Gia_ManAppendAnd(p, z, d^(signs>>4&1))
	}
	Gia_ManAppendCo(p, root^(signs>>5&1))
	Gia_ManAppendCo(p, 1)
	buf := Gia_ManAppendBuf(p, d^(signs>>6&1))
	Gia_ManAppendCo(p, buf)
	Gia_ManSetRegNum(p, 1)
	if reorder {
		ci, co := Vec_IntArray(p.vCis), Vec_IntArray(p.vCos)
		ci[0], ci[3] = ci[3], ci[0]
		co[0], co[2] = co[2], co[0]
		for i, id := range ci {
			Gia_ObjSetCioId(p.pObjs[id], int32(i))
		}
		for i, id := range co {
			Gia_ObjSetCioId(p.pObjs[id], int32(i))
		}
	}
	for id, obj := range p.pObjs {
		obj.Value = uint32(17 + 3*id)
	}
	return p
}

func TestGenNativeWordSimulation(t *testing.T) {
	exe := compileGiaSourceOracle(t, "gen_words_review.c")
	native, err := testoracle.Command(t, exe).CombinedOutput()
	if err != nil {
		t.Fatalf("C word simulation: %v\n%s", err, native)
	}
	for _, mode := range []string{"random-zero", "explicit-zero"} {
		output, err := testoracle.Command(t, exe, mode).CombinedOutput()
		if err == nil || !strings.Contains(string(output), "Vec_WrdEntryP") {
			t.Fatalf("original C zero-width precondition changed: %v %s", err, output)
		}
	}
	var got strings.Builder
	for typed := 0; typed < 2; typed++ {
		nSigns := int32(64)
		if typed != 0 {
			nSigns = 128
		}
		for signs := int32(0); signs < nSigns; signs++ {
			for reorder := 0; reorder < 2; reorder++ {
				for _, words := range []int32{1, 2, 5} {
					for _, seed := range []int32{0, 7} {
						for explicit := 0; explicit < 2; explicit++ {
							p := genReviewGraph(typed != 0, signs, reorder != 0)
							inputs := Vec_WrdStart(int(4 * words))
							for i := 0; i < 4; i++ {
								for w := int32(0); w < words; w++ {
									Vec_WrdWriteEntry(inputs, int(int32(i)*words+w), uint64(i+1)*0x9e3779b97f4a7c15^uint64(w+1)*0xd1b54a32d192ed03^uint64(seed))
								}
							}
							snapshot := slices.Clone(Vec_WrdArray(inputs))
							Gia_ManRandomW(1)
							for i := int32(0); i < seed; i++ {
								Gia_ManRandomW(0)
							}
							if explicit != 0 {
								Gia_ManSimulateWordsInit(p, inputs)
							} else {
								Gia_ManSimulateWords(p, words)
							}
							fmt.Fprintf(&got, "typed%d signs%d reorder%d words%d seed%d explicit%d", typed, signs, reorder, words, seed, explicit)
							if typed == 0 {
								for id := int32(0); id < p.nObjs; id++ {
									fmt.Fprintf(&got, " obj%d", id)
									for _, v := range Gia_ManObjSim(p, id) {
										fmt.Fprintf(&got, " %016x", v)
									}
								}
							}
							for i := int32(0); i < 4; i++ {
								fmt.Fprintf(&got, " ci%d", i)
								for _, v := range Gia_ManObjSim(p, Gia_ManCi(p, i).id) {
									fmt.Fprintf(&got, " %016x", v)
								}
							}
							for i := int32(0); i < 4; i++ {
								fmt.Fprintf(&got, " co%d", i)
								for _, v := range Gia_ManObjSim(p, Gia_ManCo(p, i).id) {
									fmt.Fprintf(&got, " %016x", v)
								}
							}
							fmt.Fprintf(&got, " next%016x\n", Gia_ManRandomW(0))
							if !slices.Equal(snapshot, Vec_WrdArray(inputs)) {
								t.Fatal("simulation modified input words")
							}
							for id, obj := range p.pObjs {
								if obj.Value != uint32(17+3*id) {
									t.Fatal("simulation changed Value scratch")
								}
							}
							for w := int32(0); w < words; w++ {
								for bit := uint(0); bit < 64; bit++ {
									assignment := 0
									for i := int32(0); i < 4; i++ {
										if Gia_ManObjSim(p, Gia_ManCi(p, i).id)[w]>>bit&1 != 0 {
											assignment |= 1 << i
										}
									}
									values := cofReviewValues(p, assignment)
									for id := int32(0); id < p.nObjs; id++ {
										if (Gia_ManObjSim(p, id)[w]>>bit&1 != 0) != values[id] {
											t.Fatalf("typed%d signs%d reorder%d words%d seed%d explicit%d word%d bit%d obj%d", typed, signs, reorder, words, seed, explicit, w, bit, id)
										}
									}
								}
							}
							if explicit != 0 && words > 0 {
								before := Gia_ManObjSim(p, Gia_ManCi(p, 0).id)[0]
								Vec_WrdWriteEntry(inputs, 0, ^before)
								if Gia_ManObjSim(p, Gia_ManCi(p, 0).id)[0] != before {
									t.Fatal("input/result storage aliases")
								}
							}
							Vec_WrdFree(inputs)
							Gia_ManStop(p)
						}
					}
				}
			}
		}
	}
	esopReviewCompare(t, native, got.String())
}

// C asserts on zero-sized word vectors; retain the existing safe Go extension.
func TestGenWordZeroWidth(t *testing.T) {
	for _, explicit := range []bool{false, true} {
		p := genReviewGraph(true, 37, true)
		Gia_ManRandomW(1)
		want := Gia_ManRandomW(0)
		Gia_ManRandomW(1)
		if explicit {
			in := Vec_WrdStart(0)
			Gia_ManSimulateWordsInit(p, in)
			Vec_WrdFree(in)
		} else {
			Gia_ManSimulateWords(p, 0)
		}
		if p.NSimWords != 0 || Vec_WrdSize(p.vSims) != 0 {
			t.Fatal("zero-width storage changed")
		}
		for id := int32(0); id < p.nObjs; id++ {
			if len(Gia_ManObjSim(p, id)) != 0 {
				t.Fatal("nonempty zero-width object")
			}
		}
		if got := Gia_ManRandomW(0); got != want {
			t.Fatal("zero-width simulation consumed random data")
		}
		Gia_ManStop(p)
	}
}
