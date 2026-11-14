package acb

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func acbReadFixture(t *testing.T, source string) (*Acb_Ntk_t, string) {
	t.Helper()
	name := filepath.Join(t.TempDir(), "design.v")
	if err := os.WriteFile(name, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	p := Acb_VerilogSimpleRead(name, "")
	if p == nil {
		t.Fatal("read failed")
	}
	t.Cleanup(func() { Acb_ManFree(p.pDesign) })
	return p, name
}
func TestVerilogExtractionAndInsertion(t *testing.T) {
	source := "module f(a,b,y);\ninput a,b;\noutput y;\nwire n;\n// and(ignored,a,b);\nand (n,a,b);\nnot (y,n);\nendmodule\n"
	p, in := acbReadFixture(t, source)
	names := Vec_PtrAlloc(1)
	Vec_PtrPush(names, "n")
	places := Acb_NtkPlaces(in, names)
	if !reflect.DeepEqual(Vec_IntArray(places), []int32{int32(strings.Index(source, "and (n")), 0}) {
		t.Fatalf("places %v", Vec_IntArray(places))
	}
	for number := int32(0); number < 2; number++ {
		for skip := int32(0); skip < 2; skip++ {
			out := filepath.Join(t.TempDir(), "inserted.v")
			Acb_NtkInsert(in, out, names, number, skip)
			bytes, err := os.ReadFile(out)
			if err != nil {
				t.Fatal(err)
			}
			text := string(bytes)
			target := "t0_n"
			if number != 0 {
				target = "t_0"
			}
			if !strings.Contains(text, "wire  "+target+";") {
				t.Fatalf("bad wire declaration:\n%s", text)
			}
			if number == 0 { // Only the source t_ prefix marks an ECO target.
				continue
			}
			q := Acb_VerilogSimpleRead(out, "")
			if q == nil {
				t.Fatal("inserted file unreadable")
			}
			defer Acb_ManFree(q.pDesign)
			if Vec_IntSize(&q.vTargets) != 1 {
				t.Fatalf("target count %d", Vec_IntSize(&q.vTargets))
			}
			var g, h *Gia_Man_t
			var nodes *Vec_Int_t
			var reverse *Vec_Ptr_t
			var polar *Vec_Bit_t
			if Acb_NtkExtract(out, in, 0, 0, &g, &h, 0, &nodes, &reverse, &polar) != 1 {
				t.Fatal("extraction failed")
			}
			for a := 0; a < 8; a++ {
				args := []bool{a&1 != 0, a&2 != 0, a&4 != 0}
				got, err := g.Evaluate(args)
				if err != nil || got[0] == args[2] {
					t.Fatalf("target extraction %v %v", got, err)
				}
				want, err := h.Evaluate(args)
				if err != nil || want[0] == (args[0] && args[1]) {
					t.Fatalf("reference extraction %v %v", want, err)
				}
			}
			Gia_ManStop(g)
			Gia_ManStop(h)
		}
	}
	counts := make([]int32, 5)
	empty := Vec_PtrAlloc(0)
	if got := Acb_NtkCollectMfsGates(in, empty, names, counts); got != 1 || counts[4] != 1 {
		t.Fatalf("MFFC gates %d %v", got, counts)
	}
	g := Acb_NtkToGia2(p, 0, 0, nil, 0)
	defer Gia_ManStop(g)
	if got := Acb_NtkCollectUser(p, names); Vec_IntSize(got) != 1 {
		t.Fatal("named node selection")
	}
}

func TestInputOrderAndCounterexampleFile(t *testing.T) {
	p, _ := acbReadFixture(t, "module f(b,a,y);\ninput b,a;\noutput y;\nand(y,a,b);\nendmodule\n")
	q, _ := acbReadFixture(t, "module f(a,b,y);\ninput a,b;\noutput y;\nand(y,a,b);\nendmodule\n")
	if Acb_NtkCheckPiOrder(p, q) != 0 || Acb_NtkCheckPiOrder(p, q) != 1 {
		t.Fatal("PI reorder")
	}
	out := filepath.Join(t.TempDir(), "cex.txt")
	model := []int32{1, 0}
	Acb_OutputFile(out, p, &model[0])
	b, err := os.ReadFile(out)
	if err != nil || !strings.Contains(string(b), "a 1\nb 0\n") {
		t.Fatalf("counterexample: %q %v", b, err)
	}
}

func TestSupportSearchVariants(t *testing.T) {
	s := sat_solver_new()
	defer sat_solver_delete(s)
	sat_solver_addclause(s, []int32{0, 2})
	sat_solver_addclause(s, []int32{2, 4})
	weights := Vec_IntStart(3)
	for i := int32(0); i < 3; i++ {
		Vec_IntWriteEntry(weights, i, 1)
	}
	var pats *Vec_Wrd_t
	var count int32
	support := Acb_FindSupportStart(s, 0, weights, &pats, &count)
	check := func(v *Vec_Int_t) {
		t.Helper()
		if v == nil || sat_solver_solve(s, Vec_IntArray(v), nil, 0, 0, 0, 0) != l_False {
			t.Fatalf("invalid support %v", v)
		}
	}
	check(support)
	minimal := Acb_FindSupportMin(s, 0, pats, &count, support)
	check(minimal)
	for i := int32(0); i < Vec_IntSize(minimal); i++ {
		ls := append([]int32(nil), Vec_IntArray(minimal)...)
		ls = append(ls[:i], ls[i+1:]...)
		if sat_solver_solve(s, ls, nil, 0, 0, 0, 0) != l_True {
			t.Fatal("support not subset-minimal")
		}
	}
	mask := make([]word, NWORDS)
	Abc_TtConst(mask, NWORDS, 1)
	if v := Acb_FindArgMaxUnderMask2(pats, mask, weights, count); v < 0 || v >= 3 {
		t.Fatal("weighted pattern selection")
	}
}

func TestMultiTargetMiterUpdates(t *testing.T) {
	for _, cisOnly := range []int32{0, 1} {
		t.Run(fmt.Sprint(cisOnly), func(t *testing.T) {
			m := Gia_ManStart(32)
			Gia_ManHashAlloc(m)
			a, b := Gia_ManAppendCi(m), Gia_ManAppendCi(m)
			t0, t1 := Gia_ManAppendCi(m), Gia_ManAppendCi(m)
			and := Gia_ManHashAnd(m, a, b)
			xor := Gia_ManHashXor(m, a, b)
			Gia_ManAppendCo(m, Gia_ManHashOr(m, Gia_ManHashXor(m, t0, and), Gia_ManHashXor(m, t1, xor)))
			Gia_ManAppendCo(m, a)
			Gia_ManAppendCo(m, b)
			defer func() { Gia_ManStop(m) }()
			for target := int32(0); target < 2; target++ {
				patch := Gia_ManStart(16)
				Gia_ManHashAlloc(patch)
				x, y := Gia_ManAppendCi(patch), Gia_ManAppendCi(patch)
				f := Gia_ManHashAnd(patch, x, y)
				if target != 0 {
					f = Gia_ManHashXor(patch, x, y)
				}
				Gia_ManAppendCo(patch, f)
				next := Acb_UpdateMiter(m, patch, target, 2, Vec_IntStartNatural(2), cisOnly)
				Gia_ManStop(patch)
				Gia_ManStop(m)
				m = next
				for assignment := 0; assignment < 16; assignment++ {
					args := []bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0, assignment&8 != 0}
					got, err := m.Evaluate(args)
					want := false
					if target == 0 {
						want = args[3] != (args[0] != args[1])
					}
					if err != nil || !reflect.DeepEqual(got, []bool{want, args[0], args[1]}) {
						t.Fatalf("target%d assignment%d got%v err%v", target, assignment, got, err)
					}
				}
			}
			cnf := Acb_NtkDeriveMiterCnf(m, 1, 2, 0)
			defer Cnf_DataFree(cnf)
			if Acb_CheckMiter(cnf) != 1 {
				t.Fatal("repaired miter is not UNSAT")
			}
		})
	}
}

func TestCommonPatchSupport(t *testing.T) {
	sops := Vec_PtrAlloc(3)
	Vec_PtrPush(sops, "1- 1\n")
	Vec_PtrPush(sops, "11 1\n")
	Vec_PtrPush(sops, " 0\n")
	supports := Vec_WecAlloc(3)
	Vec_WecInit(supports, 3)
	Vec_IntPushTwo(Vec_WecEntry(supports, 0), 2, 1)
	Vec_IntPushTwo(Vec_WecEntry(supports, 1), 0, 2)
	var used *Vec_Int_t
	got := Acb_TransformPatchFunctions(sops, supports, &used, 3)
	if !reflect.DeepEqual(Vec_IntArray(used), []int32{2, 0}) {
		t.Fatalf("common order %v", Vec_IntArray(used))
	}
	for i, want := range []string{"1- 1\n", "11 1\n", "-- 0\n"} {
		if Vec_PtrEntry(got, int32(i)).(string) != want {
			t.Fatalf("function%d %q", i, Vec_PtrEntry(got, int32(i)))
		}
	}
}

func TestSatokoSupportAndInsufficientDivisors(t *testing.T) {
	p := acbNetwork(t, 2)
	Acb_NtkCleanObjWeights(p)
	Acb_ObjSetWeight(p, 1, 1)
	Acb_ObjSetWeight(p, 2, 1)
	divs := Vec_IntAlloc(2)
	Vec_IntPushTwo(divs, 1, 2)
	for _, enough := range []bool{true, false} {
		m := Gia_ManStart(32)
		Gia_ManHashAlloc(m)
		a, b, target := Gia_ManAppendCi(m), Gia_ManAppendCi(m), Gia_ManAppendCi(m)
		Gia_ManAppendCo(m, Gia_ManHashXor(m, target, Gia_ManHashAnd(m, a, b)))
		Gia_ManAppendCo(m, a)
		nDiv := int32(1)
		if enough {
			Gia_ManAppendCo(m, b)
			nDiv = 2
		}
		cnf := Mf_ManGenerateCnf(m, 8, 0, 0, 0, 0)
		used := Acb_DerivePatchSupportS(cnf, 1, nDiv, divs, p, nil, 0)
		if enough {
			if used == nil || !reflect.DeepEqual(Vec_IntArray(used), []int32{0, 1}) {
				t.Fatalf("Satoko support %v", used)
			}
		} else if used != nil {
			t.Fatalf("accepted insufficient divisors %v", Vec_IntArray(used))
		}
		Cnf_DataFree(cnf)
		Gia_ManStop(m)
	}
}

func TestEcoSynthesisAndQuantifiedPatches(t *testing.T) {
	m := Gia_ManStart(64)
	Gia_ManHashAlloc(m)
	a, b := Gia_ManAppendCi(m), Gia_ManAppendCi(m)
	t0, t1 := Gia_ManAppendCi(m), Gia_ManAppendCi(m)
	Gia_ManAppendCo(m, Gia_ManHashOr(m, Gia_ManHashXor(m, t0, Gia_ManHashAnd(m, a, b)), Gia_ManHashXor(m, t1, Gia_ManHashXor(m, a, b))))
	defer Gia_ManStop(m)
	synth := Acb_NtkEcoSynthesize(m)
	defer Gia_ManStop(synth)
	for assignment := 0; assignment < 16; assignment++ {
		args := []bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0, assignment&8 != 0}
		want, _ := m.Evaluate(args)
		got, err := synth.Evaluate(args)
		if err != nil || !reflect.DeepEqual(got, want) {
			t.Fatalf("synthesis changed assignment%d: %v/%v err%v", assignment, got, want, err)
		}
	}
	// Universal quantification removes the earlier target for target 1.
	patch := Acb_NtkDeriveMiterCnfInter(m, 1, 2)
	defer Gia_ManStop(patch)
	for a := 0; a < 4; a++ {
		args := []bool{a&1 != 0, a&2 != 0}
		got, err := patch.Evaluate(args)
		if err != nil || !reflect.DeepEqual(got, []bool{args[0] != args[1]}) {
			t.Fatalf("quantified patch %v %v", got, err)
		}
	}
}

func TestSourceInterpolation(t *testing.T) {
	on := Gia_ManStart(16)
	Gia_ManHashAlloc(on)
	a, b := Gia_ManAppendCi(on), Gia_ManAppendCi(on)
	Gia_ManAppendCo(on, Gia_ManHashAnd(on, a, b))
	defer Gia_ManStop(on)
	off := Gia_ManStart(16)
	Gia_ManHashAlloc(off)
	a, b = Gia_ManAppendCi(off), Gia_ManAppendCi(off)
	Gia_ManAppendCo(off, Gia_ManHashAnd(off, a^1, b))
	defer Gia_ManStop(off)
	inter := Gia_ManInterOne(on, off, 0)
	if inter == nil {
		t.Fatal("interpolation failed")
	}
	defer Gia_ManStop(inter)
	for assignment := 0; assignment < 4; assignment++ {
		args := []bool{assignment&1 != 0, assignment&2 != 0}
		got, err := inter.Evaluate(args)
		if err != nil {
			t.Fatal(err)
		}
		if args[1] && got[0] != args[0] {
			t.Fatalf("interpolant violates on/off at %v: %v", args, got)
		}
	}
}

func TestPatchFilesHaveNoCStringTerminators(t *testing.T) {
	p, in := acbReadFixture(t, "module f(a,b,y);\ninput a,b;\noutput y;\nwire t_fix;\nbuf(y,t_fix);\nendmodule\n")
	divs := Vec_IntAlloc(2)
	Vec_IntPushTwo(divs, Acb_NtkCi(p, 0), Acb_NtkCi(p, 1))
	used := Vec_IntStartNatural(2)
	instance := Acb_GenerateInstance(p, divs, used, &p.vTargets)
	patch := Vec_StrAlloc(0)
	Vec_StrAppend(patch, "module patch(t_fix,a,b);\ninput a,b;\noutput t_fix;\nand(t_fix,a,b);\nendmodule\n")
	Vec_StrPush(patch, 0)
	dir := t.TempDir()
	standalone := filepath.Join(dir, "patch.v")
	combined := filepath.Join(dir, "combined.v")
	Acb_GenerateFilePatch(patch, standalone)
	Acb_GenerateFileOut(instance, in, combined, patch)
	for _, file := range []string{standalone, combined} {
		b, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		if strings.ContainsRune(string(b), 0) {
			t.Fatalf("C NUL written to %s", file)
		}
	}
	b, _ := os.ReadFile(combined)
	if !strings.Contains(string(b), ".t_fix(t_fix), .a(a), .b(b)") {
		t.Fatalf("instance %s", b)
	}
}

func TestDivisorFilteringAndNoSolution(t *testing.T) {
	p, in := acbReadFixture(t, "module f(a,b,c,y);\ninput a,b,c;\noutput y;\nwire n,m,t_fix;\nand(n,a,b);\nand(m,n,c);\nor(y,n,t_fix);\nendmodule\n")
	var block *Vec_Bit_t
	roots := Acb_NtkFindRoots(p, &p.vTargets, &block)
	supp := Acb_NtkFindSupp(p, roots)
	divs := Acb_NtkFindDivs(p, supp, block, 0)
	for _, id := range Vec_IntArray(divs) {
		if Vec_BitEntry(block, id) != 0 || Acb_ObjNameStr(p, id) == "c" || Acb_ObjNameStr(p, id) == "m" {
			t.Fatalf("invalid divisor %s", Acb_ObjNameStr(p, id))
		}
	}
	if Vec_IntSize(divs) != 3 {
		t.Fatalf("want a,b,n divisors: %v", Vec_IntArray(divs))
	}
	// The reference demands zero, but a=b=1 fixes implementation output at 1.
	q, good := acbReadFixture(t, "module f(a,b,c,y);\ninput a,b,c;\noutput y;\nbuf(y,1'b0);\nendmodule\n")
	out := filepath.Join(t.TempDir(), "should_not_exist.v")
	if result := Acb_NtkEcoPerform(p, q, []string{in, good, "", out}, 0, 0, 1, 1, 0, 0); result != 0 {
		t.Fatal("unsolvable ECO accepted")
	}
	if _, err := os.Stat(out); !os.IsNotExist(err) {
		t.Fatal("failed ECO wrote a patch")
	}
}

func TestMultipleInsertedWireSeparators(t *testing.T) {
	_, in := acbReadFixture(t, "module f(a,b,y,z);\ninput a,b;\noutput y,z;\nand(y,a,b);\nor(z,a,b);\nendmodule\n")
	names := Vec_PtrAlloc(2)
	Vec_PtrPush(names, "y")
	Vec_PtrPush(names, "z")
	out := filepath.Join(t.TempDir(), "inserted.v")
	Acb_NtkInsert(in, out, names, 1, 0)
	b, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), "wire  t_0, t_1;") {
		t.Fatalf("separators: %s", b)
	}
}

func TestMappedPatchSerialization(t *testing.T) {
	for _, version := range []int{1, 2} {
		for _, constant := range []bool{false, true} {
			t.Run(fmt.Sprintf("version%d_constant%v", version, constant), func(t *testing.T) {
				p, _ := acbReadFixture(t, "module f(a,b,y);\ninput a,b;\noutput y;\nwire t_fix;\nbuf(y,t_fix);\nendmodule\n")
				// The mapper's PI, internal node, and PO record format, with an internal
				// wire to exercise generated names (not just a gate wired to an output).
				rows := [][]int32{{}, {}, {int32(ABC_OPER_BIT_AND), 2, 0, 1}, {int32(ABC_OPER_BIT_INV), 3, 2}, {int32(ABC_OPER_BIT_BUF), 3}}
				want := stretchTruth(7, 2)
				if constant {
					rows = [][]int32{{}, {}, {int32(ABC_OPER_CONST_T), 2}, {int32(ABC_OPER_BIT_BUF), 2}}
					want = ^uint64(0)
				}
				gates := Vec_WecAlloc(int32(len(rows)))
				for _, r := range rows {
					row := Vec_WecPushLevel(gates)
					Vec_IntPushArray(row, r, int32(len(r)))
				}
				divs := Vec_IntAlloc(2)
				Vec_IntPushTwo(divs, Acb_NtkCi(p, 0), Acb_NtkCi(p, 1))
				used := Vec_IntStartNatural(2)
				var patch *Vec_Str_t
				if version == 1 {
					patch = acbGeneratePatchFromGates(p, divs, used, &p.vTargets, gates, 1)
				} else {
					ins := Vec_PtrAlloc(2)
					Vec_PtrPush(ins, "a")
					Vec_PtrPush(ins, "b")
					outs := Vec_PtrAlloc(1)
					Vec_PtrPush(outs, "t_fix")
					patch = acbGeneratePatchFromGates2(gates, ins, outs)
				}
				file := filepath.Join(t.TempDir(), "patch.v")
				Acb_GenerateFilePatch(patch, file)
				q := Acb_VerilogSimpleRead(file, "")
				if q == nil {
					t.Fatal("patch parse")
				}
				defer Acb_ManFree(q.pDesign)
				g := Acb_NtkToGia2(q, 0, 0, nil, 0)
				defer Gia_ManStop(g)
				for a := 0; a < 4; a++ {
					got, err := g.Evaluate([]bool{a&1 != 0, a&2 != 0})
					if err != nil || !reflect.DeepEqual(got, []bool{want>>uint(a)&1 != 0}) {
						t.Fatalf("serialized truth %v err%v", got, err)
					}
				}
			})
		}
	}
}
