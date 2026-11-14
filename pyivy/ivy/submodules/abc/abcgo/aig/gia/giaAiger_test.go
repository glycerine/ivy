package gia

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/glycerine/abcgo/misc/vec"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func aigerOutputs(p *Gia_Man_t, input []bool) []bool {
	values := make([]bool, p.nObjs)
	for i, id := range vec.Vec_IntArray(p.vCis) {
		values[id] = input[i]
	}
	literal := func(lit int32) bool { return values[lit>>1] != (lit&1 != 0) }
	for id, o := range p.pObjs {
		if Gia_ObjIsAnd(o) != 0 {
			values[id] = literal(Gia_ObjFaninLit0(o, int32(id))) && literal(Gia_ObjFaninLit1(o, int32(id)))
		}
	}
	result := make([]bool, Gia_ManCoNum(p))
	for i, id := range vec.Vec_IntArray(p.vCos) {
		result[i] = literal(Gia_ObjFaninLit0(p.pObjs[id], id))
	}
	return result
}
func aigerSequentialFixture() *Gia_Man_t {
	p := Gia_ManStart(32)
	p.fGiaSimple = 1
	a, lo := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	and := Gia_ManAppendAnd(p, a, lo)
	next := Gia_ManAppendXor(p, a, lo)
	Gia_ManAppendCo(p, and)
	Gia_ManAppendCo(p, a^1)
	Gia_ManAppendCo(p, a^1)
	Gia_ManAppendCo(p, next)
	p.nRegs = 1
	p.nConstrs = 1
	p.aigerBadOutputs = 1
	p.pName = "state-machine"
	p.aigerSymbols = map[byte][]string{'i': {"request"}, 'l': {"state"}, 'o': {"result"}, 'b': {"bad"}, 'c': {"assume"}}
	return p
}
func TestAigerStandardCompactRoundTrip(t *testing.T) {
	for _, compact := range []bool{false, true} {
		for _, newline := range []bool{false, true} {
			p := aigerSequentialFixture()
			before := make([]uint32, p.nObjs)
			for i, o := range p.pObjs {
				before[i] = o.Value
			}
			p.SetAigerComment([]byte("\nA comment from the application.\n"))
			data, err := MarshalAiger(p, AigerWriteOptions{Compact: compact, CommentNewline: newline})
			if err != nil {
				t.Fatal(err)
			}
			q, err := ParseAiger(data, AigerReadOptions{SkipStrash: true})
			if err != nil {
				t.Fatalf("%v: %q", err, data)
			}
			if q.nRegs != 1 || q.nConstrs != 1 || q.aigerBadOutputs != 1 || q.pName != p.pName || q.aigerSymbols['c'][0] != "assume" {
				t.Fatal("metadata roundtrip", q.nRegs, q.nConstrs, q.pName, q.aigerSymbols)
			}
			for input := 0; input < 4; input++ {
				bits := []bool{input&1 != 0, input&2 != 0}
				if !reflect.DeepEqual(aigerOutputs(p, bits), aigerOutputs(q, bits)) {
					t.Fatal("sequential combinational relation differs", compact, input)
				}
			}
			for i, o := range p.pObjs {
				if o.Value != before[i] {
					t.Fatal("writer modified scratch values")
				}
			}
		}
	}
}
func TestAigerLatchInitialization(t *testing.T) {
	for _, state := range []int32{0, 1, -1} {
		p := aigerSequentialFixture()
		if err := p.SetAigerLatchInitialStates([]int32{state}); err != nil {
			t.Fatal(err)
		}
		data, err := MarshalAiger(p, AigerWriteOptions{})
		if err != nil {
			t.Fatal(err)
		}
		q, err := ParseAiger(data, AigerReadOptions{SkipStrash: true})
		if err != nil || !reflect.DeepEqual(q.AigerLatchInitialStates(), []int32{state}) {
			t.Fatal(err, q)
		}
		if state != 0 {
			if _, err := MarshalAiger(p, AigerWriteOptions{Compact: true}); err == nil {
				t.Fatal("compact form dropped reset")
			}
		}
	}
	if err := aigerSequentialFixture().SetAigerLatchInitialStates([]int32{2}); err == nil {
		t.Fatal("invalid reset accepted")
	}
}
func TestAigerAsciiAndEmptyCompact(t *testing.T) {
	data := []byte("aag 8 2 1 1 1 1 1\n2\n8\n12 16 12\n16\n3\n8\n16 12 2\ni0 first\ni1 second\nl0 state\no0 out\nb0 bad\nc0 assume\nc\nnormal comment\n")
	p, err := ParseAiger(data, AigerReadOptions{SkipStrash: true})
	if err != nil {
		t.Fatal(err)
	}
	if p.nRegs != 1 || Gia_ManPiNum(p) != 2 || Gia_ManPoNum(p) != 3 || p.aigerLatchInits[0] != -1 {
		t.Fatal(p.nRegs, p.aigerLatchInits)
	}
	for bits := 0; bits < 8; bits++ {
		out := aigerOutputs(p, []bool{bits&1 != 0, bits&2 != 0, bits&4 != 0})
		if out[0] != (bits&1 != 0 && bits&4 != 0) || out[1] != (bits&1 == 0) || out[2] != (bits&2 == 0) {
			t.Fatal(bits, out)
		}
	}
	for _, magic := range []string{"aig", "aig2"} {
		p, err := ParseAiger([]byte(magic+" 0 0 0 0 0\nc"), AigerReadOptions{SkipStrash: true})
		if err != nil || p.nObjs != 1 {
			t.Fatal(err, p)
		}
	}
}
func TestAigerHashModesAndMalformed(t *testing.T) {
	// Two identical gates; a third unused gate tests cleanup on hashed reads.
	raw := []byte("aig 5 2 0 1 3\n8\n")
	raw = append(raw, 2, 2, 4, 2, 8, 0)
	raw = append(raw, 'c')
	p, err := ParseAiger(raw, AigerReadOptions{SkipStrash: true})
	if err != nil {
		t.Fatal(err)
	}
	q, err := ParseAiger(raw, AigerReadOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if Gia_ManAndNum(p) != 3 || Gia_ManAndNum(q) != 1 {
		t.Fatal(Gia_ManAndNum(p), Gia_ManAndNum(q))
	}
	for _, data := range [][]byte{nil, []byte("aig 0 0 0 0 0"), []byte("wat 0 0 0 0 0\n"), []byte("aig 0 0 0 0 0 0 0 1\n"), []byte("aig 2 0 0 0 0\n"), []byte("aig 1 1 0 1 0\n8\n"), []byte("aig 1 0 0 1 1\n2\n\x00\x00"), []byte("aig 1 1 0 1 0\n2\ni3 nope\n"), []byte("aig 1 0 1 0 0\n0 3\n")} {
		if _, err := ParseAiger(data, AigerReadOptions{MaxObjects: 100}); err == nil {
			t.Fatalf("malformed accepted %q", data)
		}
	}
	if _, err := ParseAiger([]byte("aig 100 100 0 0 0\nc"), AigerReadOptions{MaxObjects: 10}); err == nil {
		t.Fatal("resource limit ignored")
	}
}
func TestAigerFilesAndPartExport(t *testing.T) {
	p := aigerSequentialFixture()
	dir := t.TempDir()
	path := filepath.Join(dir, "network.aig.gz")
	if err := Gia_AigerWrite(p, path, 1, 0, 0); err != nil {
		t.Fatal(err)
	}
	q, err := Gia_AigerRead(path, 0, 1, 1)
	if err != nil || q.PSpec != path || q.pName != p.pName {
		t.Fatal(err)
	}
	part, err := Gia_AigerWriteIntoMemoryStrPart(p, p.vCis, aigerVecInts(aigerAllAnds(p)), p.vCos, 1)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ParseAiger(vec.Vec_StrArray(part), AigerReadOptions{SkipStrash: true}); err != nil {
		t.Fatal(err)
	}
	if _, err := Gia_AigerWriteIntoMemoryStrPart(p, aigerVecInts(nil), aigerVecInts(aigerAllAnds(p)), p.vCos, 0); err == nil {
		t.Fatal("non-closed part accepted")
	}
	if err := Gia_AigerWriteSimple(p, filepath.Join(dir, "simple.aig")); err != nil {
		t.Fatal(err)
	}
	if err := Gia_DumpAiger(p, filepath.Join(dir, "dump-"), 3, 4); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "dump-0003.aig")); err != nil {
		t.Fatal(err)
	}
}
func TestAigerMetadataExtensions(t *testing.T) {
	p := aigerSequentialFixture()
	n := int(p.nObjs)
	p.aigerExtensions = map[byte][]byte{}
	payloads := map[byte][]byte{'h': {1, 2, 3}, 't': {4, 5}, 'z': {9, 8, 7}, 'd': binary.BigEndian.AppendUint32(nil, 17), 'r': aigerBEArray([]int32{2, 7, 8}), 's': aigerBEArray([]int32{1, 1}), 'b': append([]byte("cell\x00"), aigerBEArray([]int32{5, 6})...)}
	p.PReprs = make([]Gia_Rpr_t, n)
	for i := range p.PReprs {
		p.PReprs[i].iRepr = GIA_VOID
	}
	p.PReprs[3] = Gia_Rpr_t{iRepr: 1, fProved: 1}
	eq, err := Gia_WriteEquivClassesInt(p, nil)
	if err != nil {
		t.Fatal(err)
	}
	payloads['e'] = eq
	class := make([]int32, n)
	switching := make([]byte, n)
	placement := make([]byte, n*4)
	for i := range class {
		class[i] = int32(i + 3)
		switching[i] = byte(i + 7)
		binary.LittleEndian.PutUint32(placement[i*4:], uint32(i)<<1|1)
	}
	payloads['g'] = aigerLEArray(class)
	payloads['v'] = aigerLEArray(class)
	payloads['u'] = switching
	payloads['p'] = placement
	payloads['f'] = aigerLEArray([]int32{23})
	payloads['i'] = binary.LittleEndian.AppendUint32(nil, math.Float32bits(1.5))
	payloads['o'] = binary.LittleEndian.AppendUint32(nil, math.Float32bits(2.5))
	mapping, _ := aigerMakeMapping(p.nObjs, []aigerLUT{{3, []int32{1, 2}}})
	p.VMapping = aigerVecInts(mapping)
	doc, err := Gia_AigerWriteMappingDoc(p)
	if err != nil {
		t.Fatal(err)
	}
	payloads['m'] = append([]byte(nil), vec.Vec_StrArray(doc)...)
	payloads['k'] = aigerBEArray([]int32{1, 1, 3})
	payloads['q'] = aigerBEArray([]int32{1, 4, 3})
	payloads['w'] = aigerBEArray([]int32{1, 1, 3})
	for tag, payload := range payloads {
		if err := p.SetAigerExtension(tag, payload); err != nil {
			t.Fatalf("tag %c: %v", tag, err)
		}
	}
	before := p.pObjs[3].Value
	data, err := MarshalAiger(p, AigerWriteOptions{})
	if err != nil {
		t.Fatal(err)
	}
	q, err := ParseAiger(data, AigerReadOptions{SkipStrash: true})
	if err != nil {
		t.Fatal(err)
	}
	for tag, payload := range payloads {
		if got := q.AigerExtension(tag); !bytes.Equal(got, payload) {
			t.Fatalf("tag %c got %x want %x", tag, got, payload)
		}
	}
	if q.pSwitching[3] != switching[3] || q.PReprs[3].fProved != 1 || q.pPlacement[3].xCoord != 3 || vec.Vec_IntEntry(q.VConfigs, 1) != 6 || q.pObjs[3].Value != before {
		t.Fatal("native metadata not populated")
	}
	payload := q.AigerExtension('z')
	payload[0] = 0
	if q.AigerExtension('z')[0] != 9 {
		t.Fatal("metadata getter aliased")
	}
}

func TestAigerPropertyOverflow(t *testing.T) {
	p := aigerSequentialFixture()
	if err := p.SetAigerProperties(math.MaxInt32, math.MaxInt32); err == nil {
		t.Fatal("overflowing property counts accepted")
	}
}
func TestAigerNormalizationRemapsObjectMetadata(t *testing.T) {
	for _, headProved := range []uint32{0, 1} {
		p := Gia_ManStart(16)
		p.fGiaSimple = 1
		a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
		x := Gia_ManAppendAnd(p, a, b)
		Gia_ManAppendCo(p, x)
		c := Gia_ManAppendCi(p)
		y := Gia_ManAppendAnd(p, a, c)
		Gia_ManAppendCo(p, y)
		p.PReprs = make([]Gia_Rpr_t, p.nObjs)
		for i := range p.PReprs {
			p.PReprs[i].iRepr = GIA_VOID
		}
		p.PReprs[c>>1] = Gia_Rpr_t{iRepr: uint32(x >> 1), fProved: headProved}
		p.PReprs[y>>1] = Gia_Rpr_t{iRepr: uint32(x >> 1), fProved: 1}
		p.pSwitching = make([]byte, p.nObjs)
		class := make([]int32, p.nObjs)
		for i := range class {
			p.pSwitching[i] = byte(i + 11)
			class[i] = int32(i + 20)
		}
		p.VGateClasses = aigerVecInts(class)
		p.VObjClasses = aigerVecInts(class)
		mapping, _ := aigerMakeMapping(p.nObjs, []aigerLUT{{x >> 1, []int32{a >> 1, b >> 1}}, {y >> 1, []int32{a >> 1, c >> 1}}})
		p.VMapping = aigerVecInts(mapping)
		p.VPacking = aigerVecInts([]int32{1, 2, x >> 1, y >> 1})
		data, err := MarshalAiger(p, AigerWriteOptions{})
		if err != nil {
			t.Fatal(err)
		}
		q, err := ParseAiger(data, AigerReadOptions{SkipStrash: true})
		if err != nil {
			t.Fatal(err)
		}
		if q.pSwitching[3] != p.pSwitching[c>>1] || vec.Vec_IntEntry(q.VGateClasses, 4) != class[x>>1] || q.PReprs[4].iRepr != 3 || q.PReprs[4].fProved != headProved || q.PReprs[5].fProved != headProved {
			t.Fatal("object metadata not remapped/proof weakened", q.pSwitching, q.PReprs)
		}
		luts, err := aigerLUTs(vec.Vec_IntArray(q.VMapping), q.nObjs)
		if err != nil || len(luts) != 2 || luts[1].root != 5 || !reflect.DeepEqual(luts[1].fans, []int32{1, 3}) {
			t.Fatal(err, luts)
		}
		if !reflect.DeepEqual(vec.Vec_IntArray(q.VPacking), []int32{1, 2, 4, 5}) {
			t.Fatal("packing not remapped")
		}
		for assignment := 0; assignment < 8; assignment++ {
			input := []bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0}
			if !reflect.DeepEqual(aigerOutputs(p, input), aigerOutputs(q, input)) {
				t.Fatal("normalization changed relation")
			}
		}
		if err := p.SetAigerExtension('z', []byte{1, 2}); err != nil {
			t.Fatal(err)
		}
		if _, err := MarshalAiger(p, AigerWriteOptions{}); err == nil {
			t.Fatal("opaque indexed payload silently renumbered")
		}
	}
}
func TestAigerCommentsAndMalformedExtensions(t *testing.T) {
	for _, comment := range []string{"note: arbitrary text\n", "hello world\n", "some comment\n"} {
		data := []byte("aig 0 0 0 0 0\nc\n" + comment)
		p, err := ParseAiger(data, AigerReadOptions{SkipStrash: true})
		if err != nil || string(p.AigerComment()) != comment {
			t.Fatal(err, p)
		}
	}
	for _, tail := range [][]byte{{'u', 0, 0}, {'u', 0, 0, 0, 10, 1}, {'z', 0, 0, 0, 1}} {
		data := append([]byte("aig 0 0 0 0 0\nc"), tail...)
		if _, err := ParseAiger(data, AigerReadOptions{SkipStrash: true}); err == nil {
			t.Fatal("truncated extension accepted", tail)
		}
	}
}

func TestAigerOriginalABCBinaryInteroperability(t *testing.T) {
	executable, err := filepath.Abs("../../../abc")
	if err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(executable)
	if err != nil || info.Mode()&0111 == 0 {
		t.Skip("ABC reference executable unavailable")
	}
	for _, compact := range []bool{false, true} {
		t.Run(fmt.Sprint(compact), func(t *testing.T) {
			p := aigerSequentialFixture()
			dir := t.TempDir()
			input, output := filepath.Join(dir, "input.aig"), filepath.Join(dir, "output.aig")
			if err := Gia_AigerWrite(p, input, 0, int32(cutBoolForAigerTest(compact)), 0); err != nil {
				t.Fatal(err)
			}
			cmd := exec.Command(executable, "-c", fmt.Sprintf("&r -s %s; &w %s", input, output))
			cmd.Dir = dir
			log, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("reference ABC: %v: %s", err, log)
			}
			q, err := Gia_AigerRead(output, 0, 1, 1)
			if err != nil {
				t.Fatalf("read reference output: %v: %s", err, log)
			}
			if q.nRegs != p.nRegs || q.nConstrs != p.nConstrs {
				t.Fatal("sequential metadata changed", q.nRegs, q.nConstrs, string(log))
			}
			for bits := 0; bits < 4; bits++ {
				input := []bool{bits&1 != 0, bits&2 != 0}
				if !reflect.DeepEqual(aigerOutputs(p, input), aigerOutputs(q, input)) {
					t.Fatal("reference roundtrip changed relation", bits)
				}
			}
			if !strings.Contains(string(log), "ABC command line") {
				t.Log(string(log))
			}
		})
	}
}
func cutBoolForAigerTest(v bool) int {
	if v {
		return 1
	}
	return 0
}

func TestAigerLegacyNormalizesLatchInitialization(t *testing.T) {
	for _, reset := range []int32{1, -1} {
		p := Gia_ManStart(8)
		ro := Gia_ManAppendCi(p)
		Gia_ManAppendCo(p, ro)
		Gia_ManAppendCo(p, ro)
		Gia_ManSetRegNum(p, 1)
		if err := p.SetAigerLatchInitialStates([]int32{reset}); err != nil {
			t.Fatal(err)
		}
		data, err := MarshalAiger(p, AigerWriteOptions{})
		if err != nil {
			t.Fatal(err)
		}
		q, err := Gia_AigerReadFromMemory(data, int32(len(data)), 1, 1, 1)
		if err != nil {
			t.Fatal(err)
		}
		if reset == 1 {
			if got := aigerOutputs(q, []bool{false}); !got[0] || got[1] {
				t.Fatalf("reset one: %v", got)
			}
		} else {
			if Gia_ManPiNum(q) != 1 || q.nRegs != 2 {
				t.Fatalf("unknown reset shape: pi=%d regs=%d", Gia_ManPiNum(q), q.nRegs)
			}
			for _, x := range []bool{false, true} {
				got := aigerOutputs(q, []bool{x, false, false})
				if got[0] != x || got[1] != x || !got[2] {
					t.Fatalf("unknown initialization %v: %v", x, got)
				}
			}
		}
	}
}

func TestAigerMalformedBytesNeverPanic(t *testing.T) {
	state := uint32(1)
	for n := 0; n < 3000; n++ {
		data := make([]byte, n%173)
		for i := range data {
			state = 1664525*state + 1013904223
			data[i] = byte(state >> 24)
		}
		if n%3 == 0 {
			data = append([]byte("aig 4 2 0 1 2\n"), data...)
		}
		_, _ = ParseAiger(data, AigerReadOptions{MaxObjects: 256})
	}
}

func TestAigerLegacyNamedProperties(t *testing.T) {
	data := []byte("aag 1 1 0 3 0\n2\n2\n3\n2\no0 constraint:enable\no1 invariant:skip\no2 result\nc\n")
	p, err := Gia_AigerReadFromMemory(data, int32(len(data)), 1, 1, 1)
	if err != nil {
		t.Fatal(err)
	}
	if Gia_ManPoNum(p) != 2 || p.nConstrs != 1 {
		t.Fatalf("po=%d constraints=%d", Gia_ManPoNum(p), p.nConstrs)
	}
	if got := aigerOutputs(p, []bool{true}); !reflect.DeepEqual(got, []bool{true, false}) {
		t.Fatal(got)
	}
}
