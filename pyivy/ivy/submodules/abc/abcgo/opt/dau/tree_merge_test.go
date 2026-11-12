package dau

import (
	"bytes"
	"math/rand"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func evalNetwork(p *Dss_Ntk_t, obj *Dss_Obj_t, bits int) bool {
	phase := Dss_IsComplement(obj) != 0
	o := Dss_Regular(obj)
	value := false
	switch o.Type {
	case DAU_DSD_CONST0:
	case DAU_DSD_VAR:
		value = bits&(1<<uint(o.Var)) != 0
	case DAU_DSD_AND:
		value = true
		for _, lit := range o.Fans {
			child := evalNetwork(p, Dss_Lit2Obj(p.Objects, lit), bits)
			value = value && child
		}
	case DAU_DSD_XOR:
		for _, lit := range o.Fans {
			value = value != evalNetwork(p, Dss_Lit2Obj(p.Objects, lit), bits)
		}
	case DAU_DSD_MUX:
		a, b, c := evalNetwork(p, Dss_Lit2Obj(p.Objects, o.Fans[0]), bits), evalNetwork(p, Dss_Lit2Obj(p.Objects, o.Fans[1]), bits), evalNetwork(p, Dss_Lit2Obj(p.Objects, o.Fans[2]), bits)
		value = a && b || !a && c
	case DAU_DSD_PRIME:
		index := 0
		for i, lit := range o.Fans {
			if evalNetwork(p, Dss_Lit2Obj(p.Objects, lit), bits) {
				index |= 1 << uint(i)
			}
		}
		value = dssBit(o.Truth, index)
	}
	return value != phase
}
func TestDssNetworkParseTransformRebuildAllThreeVar(t *testing.T) {
	manager := Dss_ManAlloc(3, 0)
	for truth := uint64(0); truth < 256; truth++ {
		stretched := truth
		for shift := 8; shift < 64; shift *= 2 {
			stretched |= stretched << uint(shift)
		}
		formula, _ := DecomposeDSD([]uint64{stretched}, 3)
		ntk, err := ParseDssNetwork(formula, 3, nil)
		if err != nil {
			t.Fatal(formula, err)
		}
		for bits := 0; bits < 8; bits++ {
			if got := evalNetwork(ntk, ntk.Root, bits); got != (truth&(1<<bits) != 0) {
				t.Fatalf("parse table%d string%s bits%d", truth, formula, bits)
			}
		}
		permutation := make([]int32, 3)
		Dss_NtkTransform(ntk, permutation)
		lit := Dss_NtkRebuild(manager, ntk)
		got := Dss_ManComputeTruth(manager, lit, 3, permutation)
		if got[0] != stretched {
			t.Fatalf("rebuild table%d string%s perm%v got%x want%x", truth, formula, permutation, got[0], stretched)
		}
	}
}
func TestDssConstantOneAndPrimePrefixPhase(t *testing.T) {
	for _, text := range []string{"1", "!0", "!!1", "!96{abc}", "96{abc}"} {
		p, err := ParseDssNetwork(text, 3, nil)
		if err != nil {
			t.Fatal(text, err)
		}
		want := Dau_DsdToTruth(text, 3)
		for bits := 0; bits < 8; bits++ {
			if got := evalNetwork(p, p.Root, bits); got != dssBit(want, bits) {
				t.Fatal(text, bits, got, want)
			}
		}
	}
}
func TestDssOperationAndPermutations(t *testing.T) {
	m := Dss_ManAlloc(6, 0)
	and := Dss_ManOperation(m, DAU_DSD_AND, []int32{2, 3}, 2, nil, nil)
	xor := Dss_ManOperation(m, DAU_DSD_XOR, []int32{2, 3}, 2, nil, nil)
	mux := Dss_ManOperation(m, DAU_DSD_MUX, []int32{3, 3, 2}, 3, nil, nil)
	for _, lit := range []int32{and, xor, mux} {
		truth := Dss_ManComputeTruth(m, lit, 6, nil)
		var b bytes.Buffer
		b.WriteString(Dss_ManExpression(m, lit, nil))
		want := Dau_DsdToTruth(b.String(), 6)
		if !reflect.DeepEqual(truth, want) {
			t.Fatal(lit, b.String(), truth, want)
		}
	}
	perm := make([]byte, 3)
	combined := Dss_ManOperation(m, DAU_DSD_AND, []int32{3, and}, 2, perm, nil)
	if Dss_VecLitSuppSize(m.Objects, combined) != 3 {
		t.Fatal("flattened support")
	}
	mapped := make([]int32, 3)
	for i, v := range perm {
		mapped[i] = int32(v)
	}
	got := Dss_ManComputeTruth(m, combined, 6, mapped)
	want := Dau_DsdToTruth("(!a(b!c))", 6)
	if !reflect.DeepEqual(got, want) {
		t.Fatal(perm, got, want)
	}
}
func TestDsdMergeRandomSemanticEquivalence(t *testing.T) {
	rng := rand.New(rand.NewSource(861))
	pool := []string{"a", "!a", "(ab)", "[ab]", "<abc>", "!(a[bc])", "(a!b)", "[a(bc)]", "CA{abc}", "!E8{abc}"}
	for trial := 0; trial < 500; trial++ {
		a, b := pool[rng.Intn(len(pool))], pool[rng.Intn(len(pool))]
		permA, permB := []int32{0, 1, 2}, []int32{2, 1, 3}
		ca, cb := int32(rng.Intn(2)), int32(rng.Intn(2))
		merged, err := MergeDSD(a, permA, b, permB, ca, cb, 4)
		if err != nil {
			t.Fatal(a, b, err)
		}
		x, _ := dssParse(a, nil)
		y, _ := dssParse(b, nil)
		dssExprMap(x, permA)
		dssExprMap(y, permB)
		x.phase = x.phase != (ca != 0)
		y.phase = y.phase != (cb != 0)
		got := Dau_DsdToTruth(merged, 4)
		for bits := 0; bits < 16; bits++ {
			if want := dssEvalExpr(x, bits) && dssEvalExpr(y, bits); dssBit(got, bits) != want {
				t.Fatalf("%s %s merged%s bits%d", a, b, merged, bits)
			}
		}
	}
}
func TestMergeConstantIdentityMustApplyPermutation(t *testing.T) {
	got, err := MergeDSD("1", nil, "a", []int32{2}, 0, 0, 3)
	if err != nil || got != "c" {
		t.Fatal(got, err)
	}
}
func TestRemoveBracesDoesNotRenameLeaves(t *testing.T) {
	for text, want := range map[string]string{"((ab)c)": "(abc)", "[a![bc]]": "[a!bc]", "(a!(bc))": "(a!(bc))"} {
		if got := Dau_DsdRemoveBraces(text, nil); got != want {
			t.Fatal(text, got, want)
		}
	}
}
func TestDssSharedMergeCacheAndSupport(t *testing.T) {
	m := Dss_ManAlloc(4, 0)
	and := Dss_ManOperation(m, DAU_DSD_AND, []int32{2, 2}, 2, nil, nil)
	xor := Dss_ManOperation(m, DAU_DSD_XOR, []int32{2, 2}, 2, nil, nil)
	lits := []int32{and, xor}
	if and > xor {
		t.Fatal("allocation ordering")
	}
	fans := [][]int32{{0, 2}, {2, 4}}
	counts := []int32{2, 2}
	Dss_ManCacheAlloc(m)
	for repeat := 0; repeat < 2; repeat++ {
		perm := make([]byte, 4)
		lit := Dss_ManMerge(m, lits, counts, fans, 1<<1, 4, perm, nil)
		if lit < 0 {
			t.Fatal("merge rejected")
		}
		mapped := make([]int32, Dss_VecLitSuppSize(m.Objects, lit))
		for i := range mapped {
			mapped[i] = int32(perm[i])
		}
		truth := Dss_ManComputeTruth(m, lit, 4, mapped)
		want := Dau_DsdToTruth("((ab)[bc])", 4)
		if !reflect.DeepEqual(truth, want) {
			t.Fatalf("shared merge lit%d perm%v got%x want%x", lit, perm, truth, want)
		}
	}
	if m.CacheHits[1] != 1 || m.CacheMisses[1] != 1 {
		t.Fatal(m.CacheHits, m.CacheMisses)
	}
}
func TestDssParserRejectsMalformed(t *testing.T) {
	for _, text := range []string{"", "(ab]", "<ab>", "a!b", "8{a}", "a trailing"} {
		if _, err := ParseDssNetwork(text, 3, nil); err == nil {
			t.Fatal("accepted malformed", text)
		}
	}
	if strings.Contains(Dss_NtkPrint(Dss_NtkCreate("1", 1, nil)), "0") {
		t.Fatal("constant one lost")
	}
}

func TestDssOriginalCStructuralOracle(t *testing.T) {
	m := Dss_ManAlloc(6, 0)
	x := []int32{2, 3, 2}
	a := Dss_ManOperation(m, DAU_DSD_AND, x, 2, nil, nil)
	b := Dss_ManOperation(m, DAU_DSD_XOR, x, 2, nil, nil)
	c := Dss_ManOperation(m, DAU_DSD_MUX, x, 3, nil, nil)
	wantLits := []int32{4, 7, 9}
	wantTruth := []uint64{0x2222222222222222, 0x9999999999999999, 0x7272727272727272}
	wantNames := []string{"(a!b)", "![ab]", "!<ab!c>"}
	for i, lit := range []int32{a, b, c} {
		got := Dss_ManComputeTruth(m, lit, 6, []int32{0, 2, 4, 6, 8, 10})
		if lit != wantLits[i] || got[0] != wantTruth[i] || Dss_ManExpression(m, lit, nil) != wantNames[i] {
			t.Fatal(i, lit, got, Dss_ManExpression(m, lit, nil))
		}
	}
	perm := make([]byte, 6)
	combined := Dss_ManOperation(m, DAU_DSD_AND, []int32{3, a}, 2, perm, nil)
	if combined != 10 || !reflect.DeepEqual(perm[:3], []byte{1, 2, 5}) {
		t.Fatal("native combined permutation", combined, perm)
	}
}

func TestDssOriginalCSharedMergeOracle(t *testing.T) {
	m := Dss_ManAlloc(4, 0)
	a := Dss_ManOperation(m, DAU_DSD_AND, []int32{2, 2}, 2, nil, nil)
	b := Dss_ManOperation(m, DAU_DSD_XOR, []int32{2, 2}, 2, nil, nil)
	perm := make([]byte, 4)
	lit := Dss_ManMerge(m, []int32{a, b}, []int32{2, 2}, [][]int32{{0, 2}, {2, 4}}, 2, 4, perm, nil)
	if lit != 8 || !reflect.DeepEqual(perm[:3], []byte{5, 2, 0}) {
		t.Fatal("native shared merge mapping", lit, perm)
	}
	truth := Dss_ManComputeTruth(m, lit, 4, []int32{5, 2, 0})
	if truth[0] != 0x0808080808080808 {
		t.Fatal(truth)
	}
}

func TestDsdStringMergeOriginalCOracle(t *testing.T) {
	data, err := os.ReadFile("test_vectors/native_merges.tsv")
	if err != nil {
		t.Fatal(err)
	}
	for i, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		fields := strings.Split(line, "\t")
		if len(fields) != 5 {
			t.Fatal("bad native fixture row", i)
		}
		ca, _ := strconv.Atoi(fields[2])
		cb, _ := strconv.Atoi(fields[3])
		got, err := MergeDSD(fields[0], []int32{0, 1, 2}, fields[1], []int32{2, 1, 3}, int32(ca), int32(cb), 4)
		if err != nil || got != fields[4] {
			t.Fatalf("native row%d %s AND %s got%s want%s err%v", i, fields[0], fields[1], got, fields[4], err)
		}
	}
}

func TestDssNestedPrimeComplementSurvivesCanonicalization(t *testing.T) {
	p := Dss_NtkCreate("(!96{abc}d)", 4, nil)
	want := Dau_DsdToTruth("(!96{abc}d)", 4)
	for bits := 0; bits < 16; bits++ {
		if evalNetwork(p, p.Root, bits) != dssBit(want, bits) {
			t.Fatal("nested prime complement lost", bits)
		}
	}
	manager := Dss_ManAlloc(4, 0)
	perm := make([]int32, 4)
	Dss_NtkTransform(p, perm)
	lit := Dss_NtkRebuild(manager, p)
	if got := Dss_ManComputeTruth(manager, lit, 4, perm)[0]; got != 0x6900690069006900 {
		t.Fatalf("got%x", got)
	}
}
func TestDssNonDecomposableLimit(t *testing.T) {
	m := Dss_ManAlloc(3, 2)
	prime := Dss_ObjFindOrAdd(m, DAU_DSD_PRIME, []int32{2, 2, 2}, []uint64{0xe8e8e8e8e8e8e8e8})
	lit := prime.ID * 2
	got := Dss_ManMerge(m, []int32{lit, lit}, []int32{3, 3}, [][]int32{{0, 2, 4}, {0, 2, 4}}, 7, 3, make([]byte, 3), nil)
	if got != -1 {
		t.Fatal("prime limit ignored", got)
	}
}
func TestDssSharedMapRejectsAbsentFirstVariable(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("invalid shared variable accepted")
		}
	}()
	m := Dss_ManAlloc(4, 0)
	Dss_ManSharedMap(m, []int32{2, 2}, []int32{2, 2}, [][]int32{{0, 2}, {4, 6}}, 1<<2)
}

func TestDssExtendedCompositionAndDuplicateValidation(t *testing.T) {
	text := "<abc>{def}"
	p, err := ParseDssNetwork(text, 6, nil)
	if err != nil {
		t.Fatal(err)
	}
	want := Dau_DsdToTruth(text, 6)
	for bits := 0; bits < 64; bits++ {
		if evalNetwork(p, p.Root, bits) != dssBit(want, bits) {
			t.Fatal("composition", bits)
		}
	}
	if _, err := ParseDssNetwork("(aa)", 2, nil); err == nil {
		t.Fatal("repeated variable violates disjoint support")
	}
	merged, err := MergeDSD(text, nil, "(ab)", nil, 0, 0, 6)
	if err != nil {
		t.Fatal(err)
	}
	truth := Dau_DsdToTruth(merged, 6)
	right := Dau_DsdToTruth("(ab)", 6)
	if truth[0] != want[0]&right[0] {
		t.Fatal("composition merge", merged)
	}
}

func TestDssPrimeHashIgnoresCallerScratchTail(t *testing.T) {
	p := Dss_ManAlloc(8, 0)
	truth := []uint64{0xe8e8e8e8e8e8e8e8, 0x1234, 0x5678, 0}
	fans := []int32{2, 2, 2}
	node := Dss_ObjFindOrAdd(p, DAU_DSD_PRIME, fans, truth)
	if got := Dss_ObjHashLookup(p, DAU_DSD_PRIME, fans, truth); got != node {
		t.Fatal("hash lookup included irrelevant scratch tail")
	}
}

func TestMpmFunctionMappingAndOwnedResults(t *testing.T) {
	p := Dss_ManAlloc(4, 0)
	a := Dss_ManOperation(p, DAU_DSD_AND, []int32{2, 2}, 2, nil, nil)
	b := Dss_ManOperation(p, DAU_DSD_XOR, []int32{2, 2}, 2, nil, nil)
	leaves := int32(4)
	perm := make([]int32, 4)
	lit := Mpm_FuncCompute(p, a, b, []byte{0, 2}, perm, &leaves)
	if lit < 0 || leaves != 3 {
		t.Fatal(lit, leaves)
	}
	concat := []int32{0, 2, 2, 4}
	mapped := make([]int32, leaves)
	for i, v := range perm[:leaves] {
		mapped[i] = Dss_Lit2Lit(concat, v)
	}
	if got := Dss_ManComputeTruth(p, lit, 4, mapped)[0]; got != 0x0808080808080808 {
		t.Fatalf("Mpm mapping%x %v", got, perm)
	}
	first := Dss_ManOperationFun(p, []int32{2, 3}, 2)
	saved := append([]byte(nil), first.Fans...)
	Dss_ManOperationFun(p, []int32{a, b}, 4)
	if !reflect.DeepEqual(first.Fans, saved) {
		t.Fatal("operation result aliases process-global scratch")
	}
}
