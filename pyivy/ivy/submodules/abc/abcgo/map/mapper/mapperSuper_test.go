package mapper

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"testing"

	abcmain "github.com/glycerine/abcgo/base/main"
	"github.com/glycerine/abcgo/map/mio"
	"github.com/glycerine/abcgo/misc/vec"
)

func TestSuperFormulaStepNested(t *testing.T) {
	var children [6]string
	var count int32
	name := Map_LibraryReadFormulaStep("and2(inv(a),and2(b,c))", children[:], &count)
	if name != "and2" || count != 2 || children[0] != "inv(a)" || children[1] != "and2(b,c)" {
		t.Fatalf("name=%q count=%d children=%v", name, count, children)
	}
}

func TestCutTruthsAgainstGraphEvaluation(t *testing.T) {
	old := abcmain.Abc_FrameReadLibSuper()
	abcmain.Abc_FrameSetLibSuper(&Map_SuperLib_t{nVarsMax: 5})
	defer abcmain.Abc_FrameSetLibSuper(old)
	rng := rand.New(rand.NewSource(19409))
	checked := 0
	for trial := 0; trial < 20; trial++ {
		p := Map_ManCreate(5, 1, 0)
		if p == nil {
			t.Fatal("could not create mapping manager")
		}
		nodes := append([]*Map_Node_t(nil), p.pInputs...)
		for j := 0; j < 15; j++ {
			a, b := nodes[rng.Intn(len(nodes))], nodes[rng.Intn(len(nodes))]
			if rng.Intn(2) != 0 {
				a = Map_Not(a)
			}
			if rng.Intn(2) != 0 {
				b = Map_Not(b)
			}
			nodes = append(nodes, Map_NodeAnd(p, a, b))
		}
		p.pOutputs[0] = nodes[len(nodes)-1]
		Map_MappingSetChoiceLevels(p)
		Map_MappingCuts(p)
		for _, node := range p.vMapObjs.pArray[:p.vMapObjs.nSize] {
			if Map_NodeIsAnd(node) == 0 {
				continue
			}
			for cut := node.pCuts.pNext; cut != nil; cut = cut.pNext {
				var truth [2]uint32
				Map_TruthsCutOne(p, cut, truth[:])
				for input := 0; input < 32; input++ {
					var eval func(*Map_Node_t) uint32
					eval = func(n *Map_Node_t) uint32 {
						inv := Map_IsComplement(n)
						n = Map_Regular(n)
						var v uint32
						if n == p.pConst1 {
							v = 1
						} else {
							found := false
							for i := int32(0); i < cut.nLeaves; i++ {
								if n == cut.ppLeaves[i] {
									v = uint32(input>>i) & 1
									found = true
									break
								}
							}
							if !found {
								v = eval(n.p1) & eval(n.p2)
							}
						}
						if inv {
							v ^= 1
						}
						return v
					}
					if got, want := truth[0]>>input&1, eval(node); got != want {
						t.Fatalf("trial=%d node=%d input=%d truth=%08x got=%d want=%d", trial, node.Num, input, truth[0], got, want)
					}
				}
				checked++
			}
		}
		Map_ManFree(p)
	}
	if checked == 0 {
		t.Fatal("no nontrivial cuts checked")
	}
}

func TestNodeCreatePreservesComplementedPhase(t *testing.T) {
	old := abcmain.Abc_FrameReadLibSuper()
	abcmain.Abc_FrameSetLibSuper(&Map_SuperLib_t{nVarsMax: 5})
	defer abcmain.Abc_FrameSetLibSuper(old)
	p := Map_ManCreate(2, 0, 0)
	defer Map_ManFree(p)
	gate := Map_NodeAnd(p, Map_Not(p.pInputs[0]), Map_Not(p.pInputs[1]))
	if gate.fInv != 1 {
		t.Fatal("AND of complemented inputs must be one on the all-zero assignment")
	}
	buf := Map_NodeBuf(p, Map_Not(p.pInputs[0]))
	if buf.fInv != 1 {
		t.Fatal("buffer must inherit the complemented input phase")
	}
	defer func() {
		if got := recover(); got == nil || !strings.Contains(fmt.Sprint(got), "!Map_IsComplement(p1)") {
			t.Fatalf("missing C regular-node assertion: %v", got)
		}
	}()
	Map_NodeComparePhase(Map_Not(p.pInputs[0]), p.pInputs[1])
}

func TestVecStrGetsCopiesSourceAndTracksPosition(t *testing.T) {
	input := []byte("abc\ndefgh\ntail")
	source := vec.Vec_StrAllocArrayCopy(input, len(input))
	buffer := make([]byte, 5)
	var pos int32
	if Vec_StrGets(buffer, 4, source, &pos) != 1 || string(buffer[:4]) != "abc\n" || buffer[4] != 0 || pos != 4 {
		t.Fatalf("first read %q pos=%d", buffer, pos)
	}
	if Vec_StrGets(buffer, 4, source, &pos) != 1 || string(buffer[:4]) != "defg" || buffer[4] != 0 || pos != 8 {
		t.Fatalf("bounded read %q pos=%d", buffer, pos)
	}
	if Vec_StrGets(buffer, 4, source, &pos) != 1 || string(buffer[:2]) != "h\n" || buffer[2] != 0 || pos != 10 {
		t.Fatalf("third read %q pos=%d", buffer, pos)
	}
	if Vec_StrGets(buffer, 4, source, &pos) != 1 || string(buffer[:4]) != "tail" || pos != 14 {
		t.Fatalf("fourth read %q pos=%d", buffer, pos)
	}
	if Vec_StrGets(buffer, 4, source, &pos) != 0 || buffer[0] != 0 {
		t.Fatal("missing end of input")
	}
	partial := []byte("xy")
	source = vec.Vec_StrAllocArrayCopy(partial, len(partial))
	pos = 0
	if Vec_StrGets(buffer, 4, source, &pos) != 0 || pos != 0 || string(buffer[:2]) != "xy" {
		t.Fatalf("unterminated short read = %q pos=%d", buffer, pos)
	}
}

func TestMapTwoInputAnd(t *testing.T) {
	genlib, err := mio.ReadLibrary("library.genlib", "GATE buf 1 O=a; PIN * NONINV 1 999 1 0 1 0\nGATE inv 1 O=!a; PIN * INV 1 999 1 0 1 0\nGATE and2 2 O=a*b; PIN * NONINV 1 999 1 0 1 0\n.end\n", nil)
	if err != nil {
		t.Fatal(err)
	}
	text := []byte("library.genlib\n5\n1\n6\n* and2 0 1\n")
	source := vec.Vec_StrAllocArrayCopy(text, len(text))
	lib := Map_SuperLibCreate(genlib, source, "library.super", "", 1, 0)
	if lib == nil {
		t.Fatal("could not build supergate library")
	}
	old := abcmain.Abc_FrameReadLibSuper()
	abcmain.Abc_FrameSetLibSuper(lib)
	defer abcmain.Abc_FrameSetLibSuper(old)
	p := Map_ManCreate(2, 1, 0)
	if p == nil {
		t.Fatal("could not create mapping manager")
	}
	defer Map_ManFree(p)
	p.pInputArrivals = make([]Map_Time_t, 2)
	root := Map_NodeAnd(p, p.pInputs[0], p.pInputs[1])
	p.pOutputs[0] = root
	if Map_Mapping(p) != 1 {
		t.Fatal("AND mapping failed")
	}
	if root.pCutBest[1] == nil && root.pCutBest[0] == nil {
		t.Fatal("AND has no mapped cut")
	}
	cut := root.pCutBest[1]
	if cut == nil || cut.M[0].uPhaseBest != 0x88888888 || cut.M[1].pSuperBest == nil || Mio_GateReadName(cut.M[1].pSuperBest.pRoot) != "and2" {
		t.Fatalf("positive AND cut has wrong truth or gate: %+v", cut)
	}
	// Exercise reconvergence and complemented input edges through all source
	// area-recovery modes, not just one isolated cut.
	q := Map_ManCreate(3, 1, 0)
	if q == nil {
		t.Fatal("could not create second mapping manager")
	}
	defer Map_ManFree(q)
	q.pInputArrivals = make([]Map_Time_t, 3)
	ab := Map_NodeAnd(q, q.pInputs[0], q.pInputs[1])
	ac := Map_NodeAnd(q, q.pInputs[0], Map_Not(q.pInputs[2]))
	root2 := Map_NodeAnd(q, ab, ac)
	q.pOutputs[0] = root2
	q.fAreaRecovery = 1
	if Map_Mapping(q) != 1 {
		t.Fatal("reconvergent mapping failed")
	}
	if root2.pCutBest[1] == nil && root2.pCutBest[0] == nil {
		t.Fatal("reconvergent root has no mapped cut")
	}
	for _, node := range []*Map_Node_t{ab, ac, root2} {
		for phase, c := range node.pCutBest {
			if c == nil {
				continue
			}
			match := c.M[phase]
			if match.pSuperBest == nil {
				t.Fatal("selected cut has no gate")
			}
			for input := 0; input < 32; input++ {
				index := input ^ int(match.uPhaseBest)
				got := match.pSuperBest.uTruth[0] >> index & 1
				var eval func(*Map_Node_t) uint32
				eval = func(n *Map_Node_t) uint32 {
					inverse := Map_IsComplement(n)
					n = Map_Regular(n)
					var value uint32
					for i := int32(0); i < c.nLeaves; i++ {
						if n == c.ppLeaves[i] {
							value = uint32(input>>i) & 1
							if inverse {
								value ^= 1
							}
							return value
						}
					}
					value = eval(n.p1) & eval(n.p2)
					if inverse {
						value ^= 1
					}
					return value
				}
				expected := eval(node)
				if phase == 0 {
					expected ^= 1
				}
				if got != expected {
					t.Fatalf("node %d phase %d input %d got %d want %d gate=%08x inputphase=%x leaves=%d", node.Num, phase, input, got, expected, match.pSuperBest.uTruth[0], match.uPhaseBest, c.nLeaves)
				}
			}
		}
	}
}

func TestTreeLibraryExcludesGateAndDescendants(t *testing.T) {
	genlib, err := mio.ReadLibrary("library.genlib", "GATE inv 1 O=!a; PIN * INV 1 999 1 0 1 0\nGATE and2 2 O=a*b; PIN * NONINV 1 999 1 0 1 0\n.end\n", nil)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	path, exclude := filepath.Join(dir, "library.super"), filepath.Join(dir, "exclude.txt")
	if err := os.WriteFile(path, []byte("library.genlib\n2\n2\n4\n* and2 0 1\n* inv 2\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(exclude, []byte("and2\n"), 0600); err != nil {
		t.Fatal(err)
	}
	lib := &Map_SuperLib_t{}
	Map_MappingSetupTruthTables(lib.uTruths[:])
	lib.tTableC = Map_SuperTableCreate(lib)
	if Map_LibraryReadTree(lib, genlib, path, exclude) != 1 {
		t.Fatal("tree library rejected")
	}
	if lib.ppSupers[2].fExclude == 0 || lib.ppSupers[3].fExclude == 0 || lib.tTableC.nEntries != 0 {
		t.Fatalf("excluded gate remained available: %#v %#v", lib.ppSupers[2], lib.ppSupers[3])
	}
	if Map_LibraryReadTree(&Map_SuperLib_t{}, genlib, path, filepath.Join(dir, "missing")) != 0 {
		t.Fatal("missing exclude file accepted")
	}
}

func TestTreeLibraryAcceptsSixInputGate(t *testing.T) {
	genlib, err := mio.ReadLibrary("library.genlib", "GATE and6 6 O=a*b*c*d*e*f; PIN * NONINV 1 999 1 0 1 0\n.end\n", nil)
	if err != nil {
		t.Fatal(err)
	}
	data := []byte("library.genlib\n6\n1\n7\n* and6 0 1 2 3 4 5 #and6(a,b,c,d,e,f)\n")
	source := vec.Vec_StrAllocArrayCopy(data, len(data))
	lib := &Map_SuperLib_t{}
	Map_MappingSetupTruthTables(lib.uTruths[:])
	lib.tTableC = Map_SuperTableCreate(lib)
	if Map_LibraryReadFileTreeStr(lib, genlib, source, "library.super") != 1 {
		t.Fatal("six-input tree rejected")
	}
	if Map_LibraryDeriveGateInfo(lib, nil) != 1 {
		t.Fatal("six-input derivation failed")
	}
	gate := lib.ppSupers[6]
	if gate.nFanins != 6 || gate.uTruth != [2]uint32{0, 0x80000000} || gate.pFormula != "and6(a,b,c,d,e,f)" {
		t.Fatalf("incorrect six-input gate: %+v", gate)
	}
}

func TestLegacySuperLibraryRead(t *testing.T) {
	dir := t.TempDir()
	gen := filepath.Join(dir, "library.genlib")
	text := "GATE zero 0 O=CONST0;\nGATE one 0 O=CONST1;\nGATE buf 1 O=a; PIN * NONINV 1 999 1 0 1 0\nGATE inv 1 O=!a; PIN * INV 1 999 1 0 1 0\nGATE and2 2 O=a*b; PIN * NONINV 1 999 1 0 1 0\n.end\n"
	if err := os.WriteFile(gen, []byte(text), 0600); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "library.super")
	data := "# library\n" + gen + "\n2\n1\n1 " + strings.Repeat("1000", 8) + " 1.0 1.0 1.0 2.0 and2(a,b)\n"
	if err := os.WriteFile(path, []byte(data), 0600); err != nil {
		t.Fatal(err)
	}
	// C's old reader zeros the upper truth word for fewer than six inputs,
	// while Mio_DeriveTruthTable produces both repeated words. Its source
	// assertion therefore fails for this nonzero two-input function.
	for _, name := range []string{"direct", "constructor"} {
		t.Run(name, func(t *testing.T) {
			defer func() {
				if got := recover(); got != "mapper: incorrect supergate truth" {
					t.Fatalf("got panic %v, want source truth assertion", got)
				}
			}()
			if name == "constructor" {
				Map_SuperLibCreate(nil, nil, path, "", 0, 0)
				return
			}
			lib := &Map_SuperLib_t{pName: path}
			Map_MappingSetupTruthTables(lib.uTruths[:])
			lib.tTableC = Map_SuperTableCreate(lib)
			Map_LibraryRead(lib, path)
		})
	}
}
