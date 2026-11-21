package ivy

import "testing"

func TestFraigClassHashCollisionsAndConstantRepresentative(t *testing.T) {
	net := Ivy_ManStart()
	defer Ivy_ManStop(net)
	inputs := make([]*Ivy_Obj_t, 8)
	for i := range inputs {
		inputs[i] = Ivy_ObjCreatePi(net)
	}
	p := ivyFraigStartSimulation(net, &Ivy_FraigParams_t{nSimWords: 2})
	defer ivyFraigStopSimulation(p)
	// Each pair hashes to zero, including both distinct nonzero signatures.
	patterns := [][2]uint32{{}, {1049, 1009}, {2098, 2018}, {}, {2098, 2018}, {1049, 1009}, {5, 1}, {5, 1}}
	for i, n := range inputs {
		copy(Ivy_ObjSim(n).pData, patterns[i][:])
		Ivy_ObjSetNodeHashNext(n, inputs[0])
		if i < 6 && Ivy_NodeHash(p, n) != 0 {
			t.Fatal("fixture needs distinct signatures with a zero hash")
		}
	}
	copy(Ivy_ObjSim(net.pConst1).pData, patterns[6][:])
	Ivy_FraigCreateClasses(p)
	want := [][]*Ivy_Obj_t{{net.pConst1, inputs[0], inputs[3]}, {inputs[1], inputs[5]}, {inputs[2], inputs[4]}, {inputs[6], inputs[7]}}
	class := p.lClasses.pHead
	for _, group := range want {
		if class != group[0] {
			t.Fatalf("incorrect representative for class beginning at ID %d", group[0].Id)
		}
		node := class
		for _, member := range group {
			if node != member {
				t.Fatalf("incorrect member for class %d, expected ID %d", class.Id, member.Id)
			}
			node = Ivy_ObjClassNodeNext(node)
		}
		if node != nil {
			t.Fatal("class merged unequal signatures")
		}
		class = Ivy_ObjEquivListNext(class)
	}
	if class != nil || p.lClasses.nItems != 4 {
		t.Fatal("unexpected class count")
	}
	for _, n := range inputs {
		if Ivy_ObjNodeHashNext(n) != nil {
			t.Fatal("completed class construction retained a temporary hash link")
		}
	}
}

func BenchmarkFraigCreateClasses(b *testing.B) {
	net := Ivy_ManStart()
	defer Ivy_ManStop(net)
	for range 1024 {
		Ivy_ObjCreatePi(net)
	}
	p := ivyFraigStartSimulation(net, &Ivy_FraigParams_t{nSimWords: 32})
	defer ivyFraigStopSimulation(p)
	for i, n := range net.vPis {
		for word := range p.simWords(n) {
			p.simWords(n)[word] = uint32(i/2+1)*uint32(word*2+1)
		}
	}
	b.ReportAllocs()
	for b.Loop() {
		Ivy_FraigCreateClasses(p)
		if p.lClasses.nItems != 512 || Ivy_ObjClassNodeNext(p.lClasses.pHead) != net.vPis[1] {
			b.Fatal("wrong pair classes")
		}
		p.lClasses = Ivy_FraigList_t{}
		for _, n := range net.vPis {
			n.fraig.classLink, n.fraig.classNext = nil, nil
			n.fraig.listNext, n.fraig.listPrev = nil, nil
		}
	}
}
