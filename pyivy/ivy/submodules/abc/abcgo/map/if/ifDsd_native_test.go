package if_

import (
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/opt/dau"
)

func TestIfDsdComputeKeepsTruthAndLiteralPermutation(t *testing.T) {
	truth := dau.Dau_DsdToTruth("(abc)", 3)
	p := If_DsdManAlloc(3, 2)
	defer If_DsdManFree(p, 0)
	permutation := make([]byte, 3)
	lit := If_DsdManCompute(p, truth, 3, permutation, "2")
	if lit&1 != 0 {
		t.Fatalf("unexpected complemented DSD literal %d", lit)
	}
	if got, want := permutation, []byte{0, 2, 4}; string(got) != string(want) {
		t.Fatalf("pin selectors = %v, want %v", got, want)
	}
	obj := dsdObject(p, lit)
	if got := dau.Dau_DsdToTruth(obj.text, 3); got[0] != truth[0] {
		t.Fatalf("DSD text %q evaluates to %#x, want %#x", obj.text, got[0], truth[0])
	}
	if obj.nSupp != 3 {
		t.Fatalf("support = %d, want 3", obj.nSupp)
	}
}

func TestIfDsdCheckXYFindsAndMasksDecomposition(t *testing.T) {
	truth := dau.Dau_DsdToTruth("((ab)c)", 3)
	p := If_DsdManAlloc(3, 2)
	defer If_DsdManFree(p, 0)
	perm := make([]byte, 3)
	lit := If_DsdManCompute(p, truth, 3, perm, "2")
	set := If_DsdManCheckXY(p, lit, 2, 0, 0, 0, 0)
	if set == 0 {
		t.Fatal("expected a two-LUT decomposition")
	}
	if got := If_DsdManCheckXY(p, lit, 2, 0, ^uint32(0), 0, 0); got != 0 {
		t.Fatalf("masked decomposition %#x remained available", got)
	}
}

func TestIfDsdDelayRecUsesAssociativeCounter(t *testing.T) {
	text, _ := dau.DecomposeDSD(dau.Dau_DsdToTruth("(ab)", 2), 2)
	network := dau.Dss_NtkCreate(text, 2, nil)
	var support int32
	delay, profile := dsdDelayRec(network, network.Root, []byte{0, 2}, []int32{2, 5}, &support, 2)
	if support != 2 {
		t.Fatalf("visited support %d, want 2", support)
	}
	if delay != 6 {
		t.Fatalf("AND delay %d, want 6", delay)
	}
	if profile == 0 {
		t.Fatal("expected a non-empty per-pin delay profile")
	}
}

func TestDsdBalanceProducesEquivalentCompactAig(t *testing.T) {
	for _, tc := range []struct {
		text string
		n    int32
	}{
		{"(ab)", 2}, {"[abc]", 3}, {"<abc>", 3}, {"(!a[bc])", 3}, {"(!96{abc}d)", 4},
	} {
		t.Run(tc.text, func(t *testing.T) {
			network := dau.Dss_NtkCreate(tc.text, tc.n, nil)
			original := dau.Dau_DsdToTruth(tc.text, tc.n)[0]
			for _, complementPins := range []bool{false, true} {
				pins := make([]byte, tc.n)
				times := make([]int32, tc.n)
				for i := range pins {
					pins[i] = byte(2 * (int(tc.n) - 1 - i))
					if complementPins && i%2 == 0 {
						pins[i]++
					}
					times[i] = int32(i)
				}
				aig := Vec_IntAlloc(0)
				defer Vec_IntFree(aig)
				var area int32
				delay, root := ifDsdBalance(network, network.Root, pins, times, aig, tc.n, &area)
				if delay < 0 {
					t.Fatalf("cannot balance %s", tc.text)
				}
				if root>>1 != tc.n+int32(Vec_IntSize(aig)/2)-1 {
					t.Fatalf("root %d is not last in %v", root, Vec_IntArray(aig))
				}
				Vec_IntPush(aig, root&1)
				got := If_AigVerifyArray(aig, tc.n)
				var want uint64
				for assignment := 0; assignment < 64; assignment++ {
					originalAssignment := 0
					for i, pin := range pins {
						originalAssignment |= ((assignment >> uint(pin>>1) & 1) ^ int(pin&1)) << uint(i)
					}
					want |= ((original >> uint(originalAssignment)) & 1) << uint(assignment)
				}
				if got != want {
					t.Fatalf("truth %#x, want %#x (pins %v)", got, want, pins)
				}
			}
		})
	}
}

func TestDsdBalanceManagerProducesAig(t *testing.T) {
	p, cut := ifDelayTestManager(t)
	defer If_ManStop(p)
	p.pIfDsdMan = If_DsdManAlloc(3, 4)
	defer If_DsdManFree(p.pIfDsdMan, 0)
	truth := dau.Dau_DsdToTruth("<abc>", 3)
	pins := make([]byte, 3)
	lit := If_DsdManCompute(p.pIfDsdMan, truth, 3, pins, "")
	p.vTtDsds[3] = vec.Vec_IntAllocArrayCopy([]int32{0, lit}, 2)
	p.vTtPerms[3] = vec.Vec_StrAlloc(12)
	vec.Vec_StrFill(p.vTtPerms[3], 12, 0)
	for i, value := range pins {
		vec.Vec_StrWriteEntry(p.vTtPerms[3], 6+i, value)
	}
	aig := Vec_IntAlloc(0)
	defer Vec_IntFree(aig)
	for _, complement := range []uint32{0, 1} {
		cut.fCompl = complement
		delay := If_CutDsdBalanceEval(p, cut, aig)
		if delay < 0 || cut.fUser == 0 || cut.Cost == 0 {
			t.Fatalf("delay %v cost %d", delay, cut.Cost)
		}
		want := truth[0]
		if complement != 0 {
			want = ^want
		}
		if got := If_AigVerifyArray(aig, 3); got != want {
			t.Fatalf("truth %#x, want %#x", got, want)
		}
	}
}

func TestDsdBalancedCutBuildsGia(t *testing.T) {
	p, cut := ifDelayTestManager(t)
	defer If_ManStop(p)
	p.pIfDsdMan = If_DsdManAlloc(3, 4)
	defer If_DsdManFree(p.pIfDsdMan, 0)
	truth := dau.Dau_DsdToTruth("(!a[bc])", 3)
	pins := make([]byte, 3)
	lit := If_DsdManCompute(p.pIfDsdMan, truth, 3, pins, "")
	p.vTtDsds[3] = vec.Vec_IntAllocArrayCopy([]int32{0, lit}, 2)
	p.vTtPerms[3] = vec.Vec_StrAlloc(12)
	vec.Vec_StrFill(p.vTtPerms[3], 12, 0)
	for i, value := range pins {
		vec.Vec_StrWriteEntry(p.vTtPerms[3], 6+i, value)
	}
	dst := Gia_ManStart(20)
	leaves := Vec_IntAlloc(3)
	defer Vec_IntFree(leaves)
	mini := Vec_IntAlloc(10)
	defer Vec_IntFree(mini)
	for i := 0; i < 3; i++ {
		Vec_IntPush(leaves, Gia_ManAppendCi(dst))
	}
	root := Gia_ManBuildFromMini(dst, p, cut, leaves, mini, 0, 1)
	if root < 0 {
		t.Fatal("DSD reconstruction failed")
	}
	Gia_ManAppendCo(dst, root)
	for mask := 0; mask < 8; mask++ {
		values := make([]bool, Gia_ManObjNum(dst))
		for i := int32(0); i < Gia_ManCiNum(dst); i++ {
			values[gia.Gia_ObjId(dst, Gia_ManCi(dst, i))] = mask>>uint(i)&1 != 0
		}
		for id := int32(1); id < Gia_ManObjNum(dst); id++ {
			obj := Gia_ManObj(dst, id)
			if gia.Gia_ObjIsAnd(obj) != 0 {
				a := values[gia.Gia_ObjFaninId0(obj, id)] != (gia.Gia_ObjFaninC0(obj) != 0)
				b := values[gia.Gia_ObjFaninId1(obj, id)] != (gia.Gia_ObjFaninC1(obj) != 0)
				values[id] = a && b
			}
		}
		got := values[root>>1] != (root&1 != 0)
		want := truth[0]>>uint(mask)&1 != 0
		if got != want {
			t.Fatalf("assignment %d: got %v, want %v", mask, got, want)
		}
	}
}
