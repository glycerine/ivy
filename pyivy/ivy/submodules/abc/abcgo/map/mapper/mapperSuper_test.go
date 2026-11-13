package mapper

import (
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
}

func TestTreeLibraryExcludesGateAndDescendants(t *testing.T) {
	genlib, err := mio.ReadLibrary("library.genlib", "GATE and2 2 O=a*b; PIN * NONINV 1 999 1 0 1 0\n.end\n", nil)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	path, exclude := filepath.Join(dir, "library.super"), filepath.Join(dir, "exclude.txt")
	if err := os.WriteFile(path, []byte("library.genlib\n2\n2\n4\n* and2 0 1\n* and2 2 1\n"), 0600); err != nil {
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
	lib := &Map_SuperLib_t{pName: path}
	Map_MappingSetupTruthTables(lib.uTruths[:])
	lib.tTableC = Map_SuperTableCreate(lib)
	// C's old reader zeros the upper truth word for fewer than six inputs,
	// while Mio_DeriveTruthTable produces both repeated words. Its source
	// assertion therefore fails for this nonzero two-input function.
	defer func() {
		if got := recover(); got != "mapper: incorrect supergate truth" {
			t.Fatalf("got panic %v, want source truth assertion", got)
		}
	}()
	Map_LibraryRead(lib, path)
}
