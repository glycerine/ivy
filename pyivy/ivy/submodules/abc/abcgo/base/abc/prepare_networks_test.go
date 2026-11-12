package abc

import (
	"bytes"
	"testing"
)

func TestPrepareTwoNetworks(t *testing.T) {
	networkReader.RLock()
	old := networkReader.read
	networkReader.RUnlock()
	defer RegisterNetworkReader(old)
	makeNetwork := func() *Network {
		n := Abc_NtkAlloc(ABC_NTK_LOGIC, ABC_FUNC_SOP, 0)
		in := Abc_NtkCreatePi(n)
		out := Abc_NtkCreatePo(n)
		Abc_ObjAddFanin(out, in)
		Abc_NtkOrderCisCos(n)
		return n
	}
	var names []string
	RegisterNetworkReader(func(name string, check int32) *Network {
		if check != 1 {
			t.Fatalf("check = %d", check)
		}
		names = append(names, name)
		return makeNetwork()
	})
	var first, second *Network
	var del1, del2 int32
	if Abc_NtkPrepareTwoNtks(&bytes.Buffer{}, makeNetwork(), []string{"input.blif"}, 1, &first, &second, &del1, &del2, 1) != 1 {
		t.Fatal("preparation failed")
	}
	if len(names) != 1 || names[0] != "input.blif" || del1 != 1 || del2 != 1 || Abc_NtkIsStrash(first) == 0 || Abc_NtkIsStrash(second) == 0 {
		t.Fatalf("names=%v flags=%d/%d", names, del1, del2)
	}
	Abc_NtkDelete(first)
	Abc_NtkDelete(second)
	if Abc_NtkPrepareTwoNtks(&bytes.Buffer{}, nil, nil, 0, &first, &second, &del1, &del2, 1) != 0 {
		t.Fatal("missing current network accepted")
	}
}
