package hop

import "testing"

func dfsEvaluate(root *Hop_Obj_t, values []bool) bool {
	o := Hop_Regular(root)
	var result bool
	switch o.Type {
	case AIG_CONST1:
		result = true
	case AIG_PI:
		result = values[o.PioNum]
	case AIG_AND:
		result = dfsEvaluate(o.pFanin0, values) && dfsEvaluate(o.pFanin1, values)
	case AIG_EXOR:
		result = dfsEvaluate(o.pFanin0, values) != dfsEvaluate(o.pFanin1, values)
	default:
		panic("invalid test graph")
	}
	if Hop_IsComplement(root) != 0 {
		return !result
	}
	return result
}
func dfsAssertUnmarked(t *testing.T, p *Hop_Man_t) {
	t.Helper()
	for _, head := range p.pTable {
		for o := head; o != nil; o = o.pNext {
			if o.fMarkA != 0 {
				t.Fatal("mark leaked", o.Id)
			}
		}
	}
}
func TestDfsPostorderAndCone(t *testing.T) {
	p := Hop_ManStart()
	defer Hop_ManStop(p)
	a, b, c := Hop_IthVar(p, 0), Hop_IthVar(p, 1), Hop_IthVar(p, 2)
	x, y := Hop_And(p, a, b), Hop_And(p, a, c)
	root := Hop_And(p, x, y)
	Hop_And(p, b, c) // A disconnected node belongs to manager DFS, not root DFS.
	nodes := Hop_ManDfsNode(p, Hop_Regular(root))
	if Vec_PtrSize(nodes) != 3 || Hop_DagSize(root) != 3 || Hop_DagSize(Hop_Not(root)) != 3 {
		t.Fatal("wrong cone size")
	}
	seen := map[*Hop_Obj_t]bool{}
	for _, entry := range Vec_PtrArray(nodes) {
		o := entry.(*Hop_Obj_t)
		for _, f := range []*Hop_Obj_t{Hop_ObjFanin0(o), Hop_ObjFanin1(o)} {
			if Hop_ObjIsNode(f) != 0 && !seen[f] {
				t.Fatal("not postorder")
			}
		}
		seen[o] = true
	}
	if got := Vec_PtrSize(Hop_ManDfs(p)); got != 4 {
		t.Fatal(got)
	}
	dfsAssertUnmarked(t, p)
	if Hop_ObjFanoutCount(root, a) != 2 || Hop_ObjFanoutCount(root, b) != 1 {
		t.Fatal("wrong PI fanout count")
	}
	dfsAssertUnmarked(t, p)
	for _, entry := range Vec_PtrArray(nodes) {
		entry.(*Hop_Obj_t).pData = "scratch"
	}
	Hop_ConeCleanAndMark_rec(Hop_Regular(root))
	for _, entry := range Vec_PtrArray(nodes) {
		o := entry.(*Hop_Obj_t)
		if o.fMarkA != 1 || o.pData != nil {
			t.Fatal("clean and mark")
		}
	}
	Hop_ConeUnmark_rec(Hop_Regular(root))
	dfsAssertUnmarked(t, p)
}
func TestDfsLevelsAndReferences(t *testing.T) {
	p := Hop_ManStart()
	defer Hop_ManStop(p)
	a, b, c := Hop_IthVar(p, 0), Hop_IthVar(p, 1), Hop_IthVar(p, 2)
	x := Hop_And(p, a, b)
	root := Hop_And(p, x, c)
	Hop_ObjCreatePo(p, Hop_Not(root))
	if Hop_ManCountLevels(p) != 2 {
		t.Fatal("wrong depth")
	}
	p.fRefCount = 0
	a.nRefs = 99
	b.nRefs = 99
	c.nRefs = 99
	Hop_ManCreateRefs(p)
	if a.nRefs != 1 || b.nRefs != 1 || c.nRefs != 1 || Hop_Regular(x).nRefs != 1 || Hop_Regular(root).nRefs != 1 {
		t.Fatal("bad references")
	}
	Hop_ManCreateRefs(p)
	if a.nRefs != 1 {
		t.Fatal("refs rebuilt twice")
	}
}
func TestDfsTransferAndComposition(t *testing.T) {
	source, dest := Hop_ManStart(), Hop_ManStart()
	defer Hop_ManStop(source)
	defer Hop_ManStop(dest)
	a, b := Hop_IthVar(source, 0), Hop_IthVar(source, 1)
	root := Hop_Not(Hop_And(source, a, Hop_Not(b)))
	transferred := Hop_Transfer(source, dest, root, 2)
	for mask := 0; mask < 4; mask++ {
		inputs := []bool{mask&1 != 0, mask&2 != 0}
		if dfsEvaluate(root, inputs) != dfsEvaluate(transferred, inputs) {
			t.Fatal("transfer changed function", mask)
		}
	}
	if Hop_Transfer(source, source, root, 2) != root {
		t.Fatal("same-manager transfer")
	}
	if Hop_Transfer(source, dest, Hop_ManConst0(source), 2) != Hop_ManConst0(dest) {
		t.Fatal("constant transfer")
	}
	conjunction := Hop_And(source, a, b)
	if Hop_Compose(source, conjunction, b, 0) != b {
		t.Fatal("composition did not simplify")
	}
	complemented := Hop_Complement(source, conjunction, 0)
	for mask := 0; mask < 4; mask++ {
		inputs := []bool{mask&1 != 0, mask&2 != 0}
		if got, want := dfsEvaluate(complemented, inputs), !inputs[0] && inputs[1]; got != want {
			t.Fatal(mask, got, want)
		}
	}
	if Hop_Compose(source, conjunction, b, 2) != nil || Hop_Complement(source, conjunction, -1) != nil {
		t.Fatal("invalid variable accepted")
	}
	dfsAssertUnmarked(t, source)
}
func TestDfsSupportRemappingAndPermutation(t *testing.T) {
	p := Hop_ManStart()
	defer Hop_ManStop(p)
	a, b, c := Hop_IthVar(p, 0), Hop_IthVar(p, 1), Hop_IthVar(p, 2)
	root := Hop_And(p, a, c)
	if got := Hop_Remap(p, root, 5, 3); got != Hop_And(p, a, b) {
		t.Fatal("support not compacted")
	}
	if got := Hop_Remap(p, root, 0, 3); got != Hop_ManConst0(p) {
		t.Fatal("zero support")
	}
	function := Hop_And(p, a, Hop_Not(b))
	permuted := Hop_Permute(p, function, 2, []int32{1, 0})
	for mask := 0; mask < 4; mask++ {
		inputs := []bool{mask&1 != 0, mask&2 != 0, false}
		if got, want := dfsEvaluate(permuted, inputs), inputs[1] && !inputs[0]; got != want {
			t.Fatal(mask, got, want)
		}
	}
	if Hop_Remap(p, root, 5, 4) != nil {
		t.Fatal("invalid manager variable count")
	}
	dfsAssertUnmarked(t, p)
}
func TestDfsCycleDetected(t *testing.T) {
	p := Hop_ManStart()
	defer Hop_ManStop(p)
	root := Hop_Regular(Hop_And(p, Hop_IthVar(p, 0), Hop_IthVar(p, 1)))
	old := root.pFanin0
	root.pFanin0 = root
	defer func() { root.pFanin0 = old }()
	defer func() {
		if recover() == nil {
			t.Error("cycle not detected")
		}
	}()
	Hop_ManDfsNode(p, root)
}
