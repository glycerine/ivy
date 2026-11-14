package abci

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/opt/dar"
	"github.com/glycerine/abcgo/opt/sbd"
	"testing"
)

func TestDARTransformsPreserveExhaustiveOutputs(t *testing.T) {
	for _, tc := range []struct {
		name string
		run  func(*abc.Network) *abc.Network
	}{
		{"rewrite", func(n *abc.Network) *abc.Network { p := dar.DefaultRewriteParams(); return Abc_NtkDRewrite(n, &p) }},
		{"refactor", func(n *abc.Network) *abc.Network { p := dar.DefaultRefactorParams(); return Abc_NtkDRefactor(n, &p) }},
		{"dc2", func(n *abc.Network) *abc.Network { return Abc_NtkDC2(n, 1, 1, 1, 0, 0) }},
		{"drwsat", func(n *abc.Network) *abc.Network { return Abc_NtkDrwsat(n, 1, 0) }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			input := truthNetwork(t, 4, 0x6996, 0xf731, 0, 0xffff)
			defer abc.Abc_NtkDelete(input)
			output := tc.run(input)
			if output == nil {
				t.Fatal("conversion returned nil")
			}
			defer abc.Abc_NtkDelete(output)
			if abc.Abc_NtkCheck(output) == 0 {
				t.Fatal("invalid transformed network")
			}
			for bits := 0; bits < 16; bits++ {
				want, got := evaluateNetwork(t, input, bits), evaluateNetwork(t, output, bits)
				if len(got) != len(want) {
					t.Fatal("output count changed")
				}
				for i := range want {
					if got[i] != want[i] {
						t.Fatalf("bits=%d output=%d got=%v want=%v", bits, i, got, want)
					}
				}
			}
		})
	}
}

func TestSbdCutRemapsConstantsDuplicatesAndPhases(t *testing.T) {
	g := aig.NewGraph()
	defer aig.Aig_ManStop(g)
	a, b := g.NewInput(), g.NewInput()
	edges := []aig.Edge{a.Not(), b, g.Const0(), a, g.Const1()}
	for _, complement := range []bool{false, true} {
		truth := []uint64{0x13579bdf2468ace0}
		cut := abcSbdCut(edges, truth, complement)
		if len(cut.Leaves) != 2 || cut.Leaves[0] != int32(a.NodeID()) || cut.Leaves[1] != int32(b.NodeID()) {
			t.Fatalf("bad compact leaves %v", cut.Leaves)
		}
		for bits := 0; bits < 64; bits++ {
			index := 16
			if bits&1 == 0 {
				index |= 1
			} else {
				index |= 8
			}
			if bits&2 != 0 {
				index |= 2
			}
			want := (truth[0]>>index&1 != 0) != complement
			if got := cut.Truth[0]>>bits&1 != 0; got != want {
				t.Fatalf("bits=%d complement=%v got=%v want=%v", bits, complement, got, want)
			}
		}
	}
}

func TestSbdAdapterPreservesMappedAndUnmappedFunctions(t *testing.T) {
	for _, mapped := range []bool{false, true} {
		n := truthNetwork(t, 3, 0x96, 0xe8)
		defer abc.Abc_NtkDelete(n)
		graph := abc.Abc_NtkToDar(n, 0, 0)
		defer aig.Aig_ManStop(graph)
		p := gia.Gia_ManFromAigSimple(graph)
		defer gia.Gia_ManStop(p)
		if mapped {
			data := make([]int32, gia.Gia_ManObjNum(p))
			for id := int32(1); id < gia.Gia_ManObjNum(p); id++ {
				a, b, _, kind := p.Gate(id)
				if kind != 2 {
					continue
				}
				data[id] = int32(len(data))
				data = append(data, 2, a>>1, b>>1, id)
			}
			if err := p.SetMappedData(vec.Vec_IntAllocArray(data, len(data)), nil, nil, ""); err != nil {
				t.Fatal(err)
			}
		}
		var parameters sbd.Parameters
		sbd.Sbd_ParSetDefault(&parameters)
		parameters.Mapping = mapped
		output, err := abcSbdPerform(p, &parameters)
		if err != nil {
			t.Fatal(err)
		}
		defer gia.Gia_ManStop(output)
		g := gia.Gia_ManToAigSimple(output)
		defer aig.Aig_ManStop(g)
		for bits := 0; bits < 8; bits++ {
			want := evaluateNetwork(t, n, bits)
			got, err := g.Evaluate([]bool{bits&1 != 0, bits&2 != 0, bits&4 != 0})
			if err != nil {
				t.Fatal(err)
			}
			for i := range want {
				if got[i] != want[i] {
					t.Fatalf("mapped=%v bits=%d output=%d got=%v want=%v", mapped, bits, i, got, want)
				}
			}
		}
	}
}

func TestClassicRefactorPreservesNetwork(t *testing.T) {
	for _, truth := range []uint64{0x88, 0x96, 0xe8, 0xfe} {
		n := truthNetwork(t, 3, truth)
		defer abc.Abc_NtkDelete(n)
		if Abc_NtkRefactor(n, 10, 16, 0, 0, 0, 0) != 1 {
			t.Fatalf("truth=%x check=%v names=%d PI=%d PO=%d AIG=%d", truth, n.Check(), abc.Abc_NtkCheckNames(n), abc.Abc_NtkCheckPis(n), abc.Abc_NtkCheckPos(n), abc.Abc_AigCheck(n.FunctionManager.(*abc.Abc_Aig_t)))
		}
		for bits := 0; bits < 8; bits++ {
			if got := evaluateNetwork(t, n, bits)[0]; got != (truth>>bits&1 != 0) {
				t.Fatalf("truth=%x bits=%d got=%v", truth, bits, got)
			}
		}
	}
}

func TestMappedGiaRetainsLutBoundariesAndHopFunctions(t *testing.T) {
	p := gia.Gia_ManStart(16)
	defer gia.Gia_ManStop(p)
	x, y, z := gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p)
	root := gia.Gia_ManAppendAnd(p, gia.Gia_ManAppendAnd(p, x, y^1), z)
	gia.Gia_ManAppendCo(p, root^1)
	gia.Gia_ManAppendCo(p, x)
	gia.Gia_ManAppendCo(p, 0)
	p.SetIdentity("mapped", "mapped.blif")
	data := make([]int32, gia.Gia_ManObjNum(p))
	data[root>>1] = int32(len(data))
	data = append(data, 3, x>>1, y>>1, z>>1, root>>1)
	if err := p.SetMappedData(vec.Vec_IntAllocArray(data, len(data)), nil, nil, ""); err != nil {
		t.Fatal(err)
	}
	for _, buffers := range []int32{0, 1} {
		n := Abc_NtkFromMappedGia(p, 0, buffers)
		if n == nil {
			t.Fatal("mapped conversion failed")
		}
		if n.Name != "mapped" || n.Spec != "mapped.blif" || abc.Abc_NtkCheck(n) == 0 {
			t.Fatal("lost metadata or invalid network")
		}
		found := false
		for _, node := range n.ObjectsOfKind(abc.ObjectNode) {
			if len(node.Fanins()) == 3 {
				found = true
			}
			if _, ok := node.Data.(*Hop_Obj_t); !ok {
				t.Fatalf("AIG function has type %T", node.Data)
			}
		}
		if !found {
			t.Fatal("three-input LUT boundary was lost")
		}
		for bits := 0; bits < 8; bits++ {
			got := evaluateNetwork(t, n, bits)
			want := !(bits&1 != 0 && bits&2 == 0 && bits&4 != 0)
			if got[0] != want || got[1] != (bits&1 != 0) || got[2] {
				t.Fatalf("bits=%d outputs=%v", bits, got)
			}
		}
		abc.Abc_NtkDelete(n)
	}
}

func TestMappedGiaExplicitMuxXorAndEnableExtraction(t *testing.T) {
	p := gia.Gia_ManStart(16)
	defer gia.Gia_ManStop(p)
	p.EnableMuxes()
	x, y, z := gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p)
	xor := gia.Gia_ManAppendXorReal(p, x, y)
	mux := gia.Gia_ManAppendMuxReal(p, x, z^1, xor)
	gia.Gia_ManAppendCo(p, mux^1)
	n := Abc_NtkFromMappedGia(p, 0, 1)
	if n == nil {
		t.Fatal("mux conversion failed")
	}
	defer abc.Abc_NtkDelete(n)
	for bits := 0; bits < 8; bits++ {
		a, b, c := bits&1 != 0, bits&2 != 0, bits&4 != 0
		want := a != b
		if a {
			want = !c
		}
		if got := evaluateNetwork(t, n, bits)[0]; got == want {
			t.Fatalf("bits=%d got=%v want=%v", bits, got, !want)
		}
	}
	q := gia.Gia_ManStart(16)
	defer gia.Gia_ManStop(q)
	enable, input, state := gia.Gia_ManAppendCi(q), gia.Gia_ManAppendCi(q), gia.Gia_ManAppendCi(q)
	next := gia.Gia_ManAppendMux(q, enable, input, state)
	gia.Gia_ManAppendCo(q, state)
	gia.Gia_ManAppendCo(q, next)
	gia.Gia_ManSetRegNum(q, 1)
	seq := Abc_NtkFromMappedGia(q, 1, 1)
	if seq == nil {
		t.Fatal("enable conversion failed")
	}
	defer abc.Abc_NtkDelete(seq)
	if abc.Abc_NtkLatchNum(seq) != 1 {
		t.Fatal("latch lost")
	}
	for bits := 0; bits < 8; bits++ {
		e, d, s := bits&1 != 0, bits&2 != 0, bits&4 != 0
		want := s
		if e {
			want = d
		}
		got := evaluateNetwork(t, seq, bits)
		if got[0] != s || got[1] != want {
			t.Fatalf("bits=%d outputs=%v next=%v", bits, got, want)
		}
	}
}
