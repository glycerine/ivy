package gia

import (
	"bytes"
	"fmt"
	"github.com/glycerine/abcgo/internal/testoracle"
	"slices"
	"strings"
	"testing"

	"github.com/glycerine/abcgo/misc/tim"
	"github.com/glycerine/abcgo/misc/util"
)

func TestDuplicateRetainsInterleavedPortOrder(t *testing.T) {
	for _, attributes := range []bool{false, true} {
		t.Run(map[bool]string{false: "basic", true: "attributes"}[attributes], func(t *testing.T) {
			src := Gia_ManStart(1)
			a, b, r := Gia_ManAppendCi(src), Gia_ManAppendCi(src), Gia_ManAppendCi(src)
			root := Gia_ManAppendAnd(src, Gia_ManAppendAnd(src, a, b), r^1)
			Gia_ManAppendCo(src, root)
			Gia_ManAppendCo(src, a^1)
			Gia_ManAppendCo(src, b)
			Gia_ManSetRegNum(src, 1)
			src.nConstrs = 1
			src = Gia_ManDupOrderDfsReverse(src)
			var out *Gia_Man_t
			if attributes {
				out = Gia_ManDupWithAttributes(src)
			} else {
				out = Gia_ManDup(src)
			}
			if out.nRegs != 1 || out.nConstrs != 1 {
				t.Fatal("state/constraint roles changed")
			}
			for mask := 0; mask < 8; mask++ {
				a, b, r := mask&1 != 0, mask&2 != 0, mask&4 != 0
				checkQuantificationOutputs(t, out, []bool{a, b, r}, []bool{a && b && !r, !a, b})
			}
		})
	}
}

func attributesReviewGraph(signs int32, typed, layout bool, regs int32) *Gia_Man_t {
	src := Gia_ManStart(1)
	a, b, c, r0, r1 := Gia_ManAppendCi(src), Gia_ManAppendCi(src), Gia_ManAppendCi(src), Gia_ManAppendCi(src), Gia_ManAppendCi(src)
	var x, y int32
	if typed {
		x = Gia_ManAppendXorReal(src, a^(signs&1), b^((signs>>1)&1))
		src.EnableMuxes()
		y = Gia_ManAppendMuxReal(src, c^((signs>>2)&1), x, r0^((signs>>3)&1))
	} else {
		x = Gia_ManAppendAnd(src, a^(signs&1), b^((signs>>1)&1))
		y = Gia_ManAppendAnd(src, x, c^((signs>>2)&1))
	}
	z := Gia_ManAppendAnd(src, y, r1)
	Gia_ManAppendCo(src, y^((signs>>4)&1))
	Gia_ManAppendCo(src, x^1)
	Gia_ManAppendCo(src, r0)
	Gia_ManAppendCo(src, z^1)
	Gia_ManSetRegNum(src, regs)
	if Gia_ManPoNum(src) > 0 {
		src.nConstrs = 1
	}
	src.SetIdentity("attributes", "attributes.aig")
	if layout {
		src = Gia_ManDupOrderDfsReverse(src)
	}
	return src
}

func TestDuplicateAllPortRolesAndPhases(t *testing.T) {
	for _, typed := range []bool{false, true} {
		for _, layout := range []bool{false, true} {
			for _, regs := range []int32{0, 2, 4} {
				for signs := int32(0); signs < 32; signs++ {
					src := attributesReviewGraph(signs, typed, layout, regs)
					for _, copy := range []func(*Gia_Man_t) *Gia_Man_t{Gia_ManDup, Gia_ManDupWithAttributes} {
						out := copy(src)
						if out.nRegs != regs || out.nConstrs != src.nConstrs || out.nObjs != src.nObjs || out.nXors != src.nXors || out.nMuxes != src.nMuxes {
							t.Fatal("duplicate changed graph/state/constraint representation")
						}
						for i, obj := range src.pObjs {
							if obj.Value != uint32(i<<1) {
								t.Fatal("source object IDs did not retain identity mapping")
							}
						}
						for mask := 0; mask < 32; mask++ {
							a, b, c, r0, r1 := mask&1 != 0, mask&2 != 0, mask&4 != 0, mask&8 != 0, mask&16 != 0
							x := (a != (signs&1 != 0)) && (b != (signs&2 != 0))
							if typed {
								x = (a != (signs&1 != 0)) != (b != (signs&2 != 0))
							}
							y := x && (c != (signs&4 != 0))
							if typed {
								y = r0 != (signs&8 != 0)
								if c != (signs&4 != 0) {
									y = x
								}
							}
							checkQuantificationOutputs(t, out, []bool{a, b, c, r0, r1}, []bool{y != (signs&16 != 0), !x, r0, !(y && r1)})
						}
					}
				}
			}
		}
	}
}

func attributesReviewTiming() *tim.Tim_Man_t {
	p := tim.Tim_ManStart(5, 4)
	tim.Tim_ManCreateBox(p, 0, 2, 3, 2, 0, 1)
	tim.Tim_ManSetDelayTables(p, [][]float32{{0, 2, 2, 1, 2, 3, 4}})
	for i, v := range []float32{1.25, 2.5, 3.75} {
		tim.Tim_ManInitPiArrival(p, i, v)
	}
	tim.Tim_ManInitPoRequired(p, 2, 10.5)
	tim.Tim_ManInitPoRequired(p, 3, 20.5)
	tim.Tim_ManBoxSetCopy(p, 0, 42)
	tim.Tim_ManIncrementTravId(p)
	tim.Tim_ManSetCurrentTravIdBoxInputs(p, 0)
	tim.Tim_ManSetCurrentTravIdBoxOutputs(p, 0)
	return p
}

func attributesReviewInstall(src *Gia_Man_t, mask int) {
	if mask&1 != 0 {
		src.VMapping = Vec_IntStart(int(src.nObjs))
		for _, obj := range src.pObjs {
			if Gia_ObjIsAnd(obj) == 0 {
				continue
			}
			Vec_IntWriteEntry(src.VMapping, int(obj.id), int32(Vec_IntSize(src.VMapping)))
			fanins := giaObjectFanins(src, obj)
			Vec_IntPush(src.VMapping, int32(len(fanins)))
			for _, id := range fanins {
				Vec_IntPush(src.VMapping, id)
			}
			trailer := obj.id
			if Gia_ObjIsMux(src, obj) != 0 {
				trailer = -trailer
			}
			Vec_IntPush(src.VMapping, trailer)
		}
	}
	if mask&2 != 0 {
		src.VPacking = permutation(1, 3)
		for _, obj := range src.pObjs {
			if Gia_ObjIsAnd(obj) != 0 {
				Vec_IntPush(src.VPacking, obj.id)
			}
		}
	}
	if mask&4 != 0 {
		src.PManTime = attributesReviewTiming()
	}
	if mask&8 != 0 {
		extra := Gia_ManStart(1)
		a, b := Gia_ManAppendCi(extra), Gia_ManAppendCi(extra)
		Gia_ManAppendCo(extra, Gia_ManAppendAnd(extra, a, b^1))
		Gia_ManAppendCo(extra, a^1)
		Gia_ManSetRegNum(extra, 1)
		extra.SetIdentity("extra", "extra.aig")
		extra.VConfigs = permutation(123) // The source routine uses basic, not recursive attribute duplication.
		src.PAigExtra = extra
	}
	if mask&16 != 0 {
		src.vRegClasses = permutation(7, 9)
	}
	if mask&32 != 0 {
		src.vRegInits = permutation(0, 1)
	}
	if mask&64 != 0 {
		src.VConfigs = permutation(4, 0, -1, 9)
	}
	if mask&128 != 0 {
		src.PCellStr = "cell\nwith comments"
	}
	if mask&256 != 0 {
		src.nAnd2Delay = 7
	}
}

func attributesReviewRecords(t *testing.T, typed, layout bool) string {
	t.Helper()
	var records strings.Builder
	for mask := 0; mask < 512; mask++ {
		if mask&2 != 0 && mask&1 == 0 {
			continue
		} // Packing requires LUT mapping.
		src := attributesReviewGraph(int32(mask&31), typed, layout, 2)
		attributesReviewInstall(src, mask)
		src.pCexSeq = util.Abc_CexAlloc(2, 3, 37)
		src.pCexSeq.IFrame = 36
		src.pCexSeq.IPo = 0
		for i := range src.pCexSeq.PData {
			src.pCexSeq.PData[i] = uint32(mask + i*151)
		}
		// ABC's duplication helper deliberately starts CEX registers at zero.
		src.pCexSeq.PData[0] &^= 3
		src.aigerBadOutputs = 1
		src.aigerLatchInits = []int32{0, 1}
		src.aigerComment = []byte("attr comment")
		src.aigerSymbols = map[byte][]string{'i': {"a", "b", "c", "r0", "r1"}, 'o': {"y", "x", "ri0", "ri1"}}
		src.aigerExtensions = map[byte][]byte{'z': {1, 2, 3}}
		var sourceTiming []byte
		if src.PManTime != nil {
			sourceTiming = tim.Tim_ManSave(src.PManTime.(*tim.Tim_Man_t), 0)
		}
		out := Gia_ManDupWithAttributes(src)
		if out.nRegs != 2 || out.nConstrs != 1 || out.aigerBadOutputs != 1 || !slices.Equal(out.aigerLatchInits, []int32{0, 1}) || out.nAnd2Delay != src.nAnd2Delay || out.PCellStr != src.PCellStr {
			t.Fatal("attribute copy lost roles/scalars")
		}
		name, spec := out.Identity()
		if name != "attributes" || spec != "attributes.aig" {
			t.Fatal("identity lost")
		}
		if out.pCexSeq == src.pCexSeq || out.pCexSeq.NRegs != 2 || out.pCexSeq.NPis != 3 || out.pCexSeq.IFrame != 36 || !slices.Equal(out.pCexSeq.PData, src.pCexSeq.PData) {
			t.Fatal("counterexample copy lost logical interface")
		}
		fmt.Fprintf(&records, "%d graph %d %d %d %d", mask, out.nObjs, Gia_ManCiNum(out), Gia_ManCoNum(out), out.nRegs)
		for _, pair := range []struct {
			name     string
			src, dst *Vec_Int_t
		}{{"mapping", src.VMapping, out.VMapping}, {"packing", src.VPacking, out.VPacking}, {"classes", src.vRegClasses, out.vRegClasses}, {"inits", src.vRegInits, out.vRegInits}, {"configs", src.VConfigs, out.VConfigs}} {
			fmt.Fprintf(&records, " %s", pair.name)
			if pair.src == nil {
				if pair.dst != nil {
					t.Fatal("absent attribute materialized")
				}
				fmt.Fprint(&records, " nil")
				continue
			}
			before := slices.Clone(Vec_IntArray(pair.src))
			if pair.src == pair.dst || !slices.Equal(before, Vec_IntArray(pair.dst)) {
				t.Fatalf("%s not independently copied", pair.name)
			}
			for _, v := range Vec_IntArray(pair.dst) {
				fmt.Fprintf(&records, " %d", v)
			}
			Vec_IntWriteEntry(pair.dst, 0, 999)
			if !slices.Equal(before, Vec_IntArray(pair.src)) {
				t.Fatalf("%s aliases source", pair.name)
			}
		}
		fmt.Fprintf(&records, " delay%d cell%q", out.nAnd2Delay, out.PCellStr)
		if src.PManTime == nil {
			if out.PManTime != nil {
				t.Fatal("absent timing manager copied")
			}
			fmt.Fprint(&records, " timing nil")
		} else {
			p, q := src.PManTime.(*tim.Tim_Man_t), out.PManTime.(*tim.Tim_Man_t)
			if p == q || !bytes.Equal(sourceTiming, tim.Tim_ManSave(q, 0)) || tim.Tim_ManBoxIsBlack(q, 0) != true || tim.Tim_ManCi(p, 3).TravId != 1 || tim.Tim_ManCi(q, 3).TravId != 0 {
				t.Fatal("timing copy/clear altered payload or source traversal")
			}
			fmt.Fprintf(&records, " timing %d %d %d %d %d %d", tim.Tim_ManCiNum(q), tim.Tim_ManCoNum(q), tim.Tim_ManBoxNum(q), tim.Tim_ManBlackBoxNum(q), tim.Tim_ManBoxCopy(q, 0), tim.Tim_ManBoxDelayTableId(q, 0))
			for _, v := range tim.Tim_ManBoxDelayTable(q, 0) {
				fmt.Fprintf(&records, " %.2f", v)
			}
			for _, v := range tim.Tim_ManGetArrTimes(q)[:3] {
				fmt.Fprintf(&records, " %.2f", v)
			}
			for _, v := range tim.Tim_ManGetReqTimes(q)[:2] {
				fmt.Fprintf(&records, " %.2f", v)
			}
			tim.Tim_ManBoxDelayTable(q, 0)[3] = 999
			tim.Tim_ManBox(q, 0).Inouts[0] = 1
			tim.Tim_ManInitPiArrival(q, 0, 999)
			if !bytes.Equal(sourceTiming, tim.Tim_ManSave(p, 0)) || tim.Tim_ManBox(p, 0).Inouts[0] != 0 {
				t.Fatal("timing copy aliases source")
			}
		}
		if src.PAigExtra == nil {
			if out.PAigExtra != nil {
				t.Fatal("absent extra graph copied")
			}
			fmt.Fprint(&records, " extra nil")
		} else {
			if out.PAigExtra == src.PAigExtra || out.PAigExtra.VConfigs != nil || out.PAigExtra.nRegs != 1 {
				t.Fatal("extra graph must be an independent basic duplicate")
			}
			fmt.Fprintf(&records, " extra %d %d %d %d", out.PAigExtra.nObjs, Gia_ManCiNum(out.PAigExtra), Gia_ManCoNum(out.PAigExtra), out.PAigExtra.nRegs)
			for bits := 0; bits < 4; bits++ {
				a, b := bits&1 != 0, bits&2 != 0
				checkQuantificationOutputs(t, out.PAigExtra, []bool{a, b}, []bool{a && !b, !a})
			}
			Gia_ManStop(out.PAigExtra)
			if src.PAigExtra.nObjs != 6 {
				t.Fatal("stopping extra copy changed source")
			}
		}
		fmt.Fprint(&records, " truth")
		for bits := 0; bits < 32; bits++ {
			inputs := []bool{bits&1 != 0, bits&2 != 0, bits&4 != 0, bits&8 != 0, bits&16 != 0}
			want, err := src.Evaluate(inputs)
			if err != nil {
				t.Fatal(err)
			}
			checkQuantificationOutputs(t, out, inputs, want)
			packed := 0
			for i, v := range want {
				if v {
					packed |= 1 << uint(i)
				}
			}
			fmt.Fprintf(&records, " %d", packed)
		}
		records.WriteByte('\n')
		out.pCexSeq.PData[0] ^= 1
		out.aigerLatchInits[0] = 1
		out.aigerComment[0] = 'X'
		out.aigerSymbols['i'][0] = "changed"
		out.aigerExtensions['z'][0] = 9
		if src.pCexSeq.PData[0] != uint32(mask)&^3 || src.aigerLatchInits[0] != 0 || string(src.aigerComment) != "attr comment" || src.aigerSymbols['i'][0] != "a" || src.aigerExtensions['z'][0] != 1 {
			t.Fatal("base metadata aliases source")
		}
		Gia_ManStop(out)
		if src.nObjs == 0 {
			t.Fatal("stopping main copy changed source")
		}
	}
	return records.String()
}

func TestAttributeCopyPresenceAndOwnership(t *testing.T) {
	for _, typed := range []bool{false, true} {
		for _, layout := range []bool{false, true} {
			attributesReviewRecords(t, typed, layout)
		}
	}
}

func TestAttributeCopyMatchesOriginalC(t *testing.T) {
	exe := compileGiaSourceOracle(t, "duplicate_attributes.c")
	native, err := testoracle.Command(t, exe).CombinedOutput()
	if err != nil {
		t.Fatalf("C attribute duplicate: %v: %s", err, native)
	}
	got, want := strings.Split(strings.TrimSpace(attributesReviewRecords(t, false, false)), "\n"), strings.Split(strings.TrimSpace(string(native)), "\n")
	if len(got) != len(want) {
		t.Fatalf("Go cases%d C cases%d", len(got), len(want))
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("case%d\nGo %s\nC %s", i, got[i], want[i])
		}
	}
}
