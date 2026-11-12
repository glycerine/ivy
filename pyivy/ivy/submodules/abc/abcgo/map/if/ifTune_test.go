package if_

import "testing"

func TestIfnParseExpressionNetwork(t *testing.T) {
	for _, tc := range []struct {
		name            string
		text            string
		inputs, objects int32
		rootType        int32
	}{
		{name: "and", text: "(ab)", inputs: 2, objects: 3, rootType: int32(IFN_DSD_AND)},
		{name: "nested xor", text: "[(ab)c]", inputs: 3, objects: 5, rootType: int32(IFN_DSD_XOR)},
		{name: "mux", text: "<abc>", inputs: 3, objects: 4, rootType: int32(IFN_DSD_MUX)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			network := Ifn_NtkParse(tc.text)
			if network == nil {
				t.Fatalf("Ifn_NtkParse(%q) returned nil", tc.text)
			}
			if network.nInps != tc.inputs || network.nObjs != tc.objects {
				t.Fatalf("dimensions = (%d inputs, %d objects), want (%d, %d)", network.nInps, network.nObjs, tc.inputs, tc.objects)
			}
			if got := network.Nodes[network.nObjs-1].Type; got != tc.rootType {
				t.Fatalf("root type = %d, want %d", got, tc.rootType)
			}
		})
	}
}

func TestIfnParseDefinedNetworkAndConstraints(t *testing.T) {
	network := Ifn_NtkParse("c=(ab);AB")
	if network == nil {
		t.Fatal("Ifn_NtkParse returned nil")
	}
	if network.nInps != 2 || network.nObjs != 3 {
		t.Fatalf("dimensions = (%d, %d), want (2, 3)", network.nInps, network.nObjs)
	}
	root := network.Nodes[2]
	if root.Type != int32(IFN_DSD_AND) || root.nFanins != 2 || root.Fanins[0] != 0 || root.Fanins[1] != 1 {
		t.Fatalf("parsed definition = %+v", root)
	}
	if network.nConstr != 1 || network.pConstr[0] != 1 {
		t.Fatalf("constraints = %v (count %d), want AB", network.pConstr[:network.nConstr], network.nConstr)
	}
}

func TestIfnConfigWordsPreservesPackedBitOrder(t *testing.T) {
	packed := Vec_IntAlloc(4)
	defer Vec_IntFree(packed)
	low, high, last := uint32(0x89abcdef), uint32(0x01234567), uint32(0xfedcba98)
	Vec_IntPush(packed, int32(low))
	Vec_IntPush(packed, int32(high))
	Vec_IntPush(packed, int32(last))

	got := ifnConfigWords(packed, 0, 3)
	want := []word{0x0123456789abcdef, 0x00000000fedcba98}
	if len(got) != len(want) {
		t.Fatalf("got %d words, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("word %d = %#016x, want %#016x", i, got[i], want[i])
		}
	}
}
