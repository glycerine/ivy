package gia

import (
	"fmt"
	"slices"
	"testing"

	abcutil "github.com/glycerine/abcgo/misc/util"
)

func mappingVerifyReviewGraph(recursive bool) (*Gia_Man_t, []int32, []int32) {
	p := Gia_ManStart(16)
	a, b, c := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	leftMissing := Gia_ManAppendAnd(p, a, b)
	left := leftMissing
	if recursive {
		left = Gia_ManAppendAnd(p, leftMissing, c)
	}
	Gia_ManAppendCo(p, left)
	rightMissing := Gia_ManAppendAnd(p, b, c)
	right := rightMissing
	if recursive {
		right = Gia_ManAppendAnd(p, rightMissing, a)
	}
	buffer := Gia_ManAppendBuf(p, right)
	Gia_ManAppendCo(p, buffer)
	Gia_ManAppendCo(p, left) // Direct missing drivers repeat; recursive visits do not.
	Gia_ManAppendCo(p, a)
	p.VMapping = Vec_IntStart(int(p.nObjs))
	if recursive {
		for _, entry := range [][3]int32{{left, leftMissing, c}, {right, rightMissing, a}} {
			Vec_IntWriteEntry(p.VMapping, int(entry[0]>>1), int32(Vec_IntSize(p.VMapping)))
			for _, value := range []int32{2, entry[1] >> 1, entry[2] >> 1, entry[0] >> 1} {
				Vec_IntPush(p.VMapping, value)
			}
		}
	}
	return p, []int32{rightMissing >> 1, leftMissing >> 1}, []int32{left >> 1, right >> 1}
}

func TestMappingVerifyRecursionStopsAtFirstFailure(t *testing.T) {
	p := Gia_ManStart(16)
	defer Gia_ManStop(p)
	a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x, y := Gia_ManAppendAnd(p, a, b), Gia_ManAppendAnd(p, a, b^1)
	root := Gia_ManAppendAnd(p, x, y)
	Gia_ManAppendCo(p, root)
	p.VMapping = Vec_IntStart(int(p.nObjs))
	offset := Vec_IntSize(p.VMapping)
	Vec_IntWriteEntry(p.VMapping, int(root>>1), int32(offset))
	for _, value := range []int32{2, x >> 1, y >> 1, root >> 1} {
		Vec_IntPush(p.VMapping, value)
	}
	for pass := 0; pass < 2; pass++ {
		first, skipped := x>>1, y>>1
		if pass != 0 {
			first, skipped = skipped, first
			Vec_IntWriteEntry(p.VMapping, offset+1, first)
			Vec_IntWriteEntry(p.VMapping, offset+2, skipped)
		}
		Gia_ManIncrementTravId(p)
		output := hashReviewStdout(t, func() {
			if Gia_ManMappingVerify_rec(p, Gia_ManObj(p, root>>1)) != 0 {
				t.Fatal("missing internal mapping accepted")
			}
		})
		want := fmt.Sprintf("Error: Gia_ManMappingVerify: Internal node %d does not have mapping.\n", first)
		if output != want || Gia_ObjIsTravIdCurrent(p, Gia_ManObj(p, skipped)) != 0 {
			t.Fatalf("wrong first failure or traversal: %q", output)
		}
		// Source marks before failure; another visit in the same epoch returns1.
		if output := hashReviewStdout(t, func() {
			if Gia_ManMappingVerify_rec(p, Gia_ManObj(p, root>>1)) != 1 {
				t.Fatal("source revisit semantics changed")
			}
		}); output != "" {
			t.Fatalf("revisit repeated a diagnostic: %q", output)
		}
	}
}

func TestMappingVerifyUsesCoOrderAndPrintBridge(t *testing.T) {
	p, missing, roots := mappingVerifyReviewGraph(false)
	defer Gia_ManStop(p)
	Gia_ManAppendCo(p, roots[1]<<1)
	Vec_IntPush(p.VMapping, 0) // The fixture has no records; extend its object index.
	// Reorder the actual CO interface, preserving Cio IDs.
	cos := Vec_IntArray(p.vCos)
	cos[0], cos[len(cos)-1] = cos[len(cos)-1], cos[0]
	for i, id := range cos {
		Gia_ObjSetCioId(Gia_ManObj(p, id), int32(i))
	}
	old := abcutil.PrintBridge
	defer func() { abcutil.PrintBridge = old }()
	var messages []string
	abcutil.PrintBridge = func(message string) { messages = append(messages, message) }
	output := hashReviewStdout(t, func() { Gia_ManMappingVerify(p) })
	var want []string
	for _, id := range []int32{missing[0], missing[0], missing[1], missing[1]} {
		want = append(want, fmt.Sprintf("Error: Gia_ManMappingVerify: CO driver %d does not have mapping.\n", id))
	}
	if output != "" || !slices.Equal(messages, want) {
		t.Fatalf("stdout%q bridge%q want%q", output, messages, want)
	}
}

func TestMappingVerifyCompleteCutsAndTerminalBoundaries(t *testing.T) {
	p, missing, _ := mappingVerifyReviewGraph(true)
	defer Gia_ManStop(p)
	for _, id := range missing {
		obj := Gia_ManObj(p, id)
		Vec_IntWriteEntry(p.VMapping, int(id), int32(Vec_IntSize(p.VMapping)))
		for _, value := range []int32{2, Gia_ObjFaninId0(obj, id), Gia_ObjFaninId1(obj, id), id} {
			Vec_IntPush(p.VMapping, value)
		}
	}
	before := slices.Clone(Vec_IntArray(p.VMapping))
	for pass := 0; pass < 2; pass++ {
		if output := hashReviewStdout(t, func() { Gia_ManMappingVerify(p) }); output != "" {
			t.Fatalf("valid mapping rejected: %q", output)
		}
		for _, obj := range p.pObjs {
			if Gia_ObjIsCi(obj) == 0 && Gia_ObjIsConst0(obj) == 0 && Gia_ObjIsBuf(obj) == 0 {
				continue
			}
			Gia_ManIncrementTravId(p)
			if output := hashReviewStdout(t, func() {
				if Gia_ManMappingVerify_rec(p, obj) != 1 {
					t.Fatal("terminal boundary rejected")
				}
			}); output != "" {
				t.Fatalf("terminal diagnosed: %q", output)
			}
			if Gia_ObjIsTravIdCurrent(p, obj) == 0 {
				t.Fatal("terminal not marked visited")
			}
		}
	}
	if !slices.Equal(Vec_IntArray(p.VMapping), before) {
		t.Fatal("valid verification changed mapping")
	}
}

func TestMappingVerifyRequiresMapping(t *testing.T) {
	p := Gia_ManStart(1)
	defer Gia_ManStop(p)
	var caught any
	output := hashReviewStdout(t, func() {
		defer func() { caught = recover() }()
		Gia_ManMappingVerify(p)
	})
	if caught != "gia: mapping is absent" || output != "" || p.nTravIds != 0 {
		t.Fatalf("missing-mapping precondition: panic%v output%q epoch%d", caught, output, p.nTravIds)
	}
}

func TestMappingVerifyBuffersPrecedeOutputs(t *testing.T) {
	for _, recursive := range []bool{false, true} {
		t.Run(fmt.Sprint(recursive), func(t *testing.T) {
			p, missing, _ := mappingVerifyReviewGraph(recursive)
			defer Gia_ManStop(p)
			kind := "CO driver"
			if recursive {
				kind = "Internal node"
			}
			want := fmt.Sprintf("Error: Gia_ManMappingVerify: %s %d does not have mapping.\nError: Gia_ManMappingVerify: %s %d does not have mapping.\n", kind, missing[0], kind, missing[1])
			if !recursive {
				want += fmt.Sprintf("Error: Gia_ManMappingVerify: CO driver %d does not have mapping.\n", missing[1])
			}
			before := slices.Clone(Vec_IntArray(p.VMapping))
			for pass := 0; pass < 2; pass++ {
				output := hashReviewStdout(t, func() { Gia_ManMappingVerify(p) })
				if output != want {
					t.Errorf("pass%d: output%q want%q", pass, output, want)
				}
				if !slices.Equal(Vec_IntArray(p.VMapping), before) {
					t.Fatal("verification changed mapping")
				}
			}
		})
	}
}
