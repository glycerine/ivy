package amap

import (
	"os"
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
)

func TestAmapMapGraphSimpleAnd(t *testing.T) {
	lib := mapTestLibrary(t)
	if lib == nil || lib.pGateInv == nil {
		t.Fatal("prepare AND/inverter library")
	}
	graph := aig.NewGraph()
	a, b := graph.NewInput(), graph.NewInput()
	graph.AddOutput(graph.AND(a, b))
	pars := &Amap_Par_t{}
	Amap_ManSetDefaultParams(pars)
	mapped := Amap_ManMapGraph(graph, lib, pars)
	cleanupMappedOutput(t, mapped)
	if mapped == nil || Vec_PtrSize(mapped) < 4 {
		t.Fatalf("mapping result is empty: %v", mapped)
	}
	var hasGate, hasOutput bool
	for _, item := range Vec_PtrArray(mapped) {
		out, ok := item.(*Amap_Out_t)
		if !ok || out == nil {
			continue
		}
		hasGate = hasGate || out.Type == 0 && out.pName == "AND2"
		hasOutput = hasOutput || out.Type == 1
	}
	if !hasGate || !hasOutput {
		t.Fatalf("mapped network missing AND or output: gate=%v output=%v", hasGate, hasOutput)
	}
}

func TestAmapMapGraphXorAndMux(t *testing.T) {
	lib := mapTestLibrary(t)
	for _, tc := range []struct {
		name  string
		build func(*aig.Graph, []aig.Edge) aig.Edge
		want  string
	}{
		{"xor", func(g *aig.Graph, x []aig.Edge) aig.Edge { return g.NewXor(x[0], x[1]) }, "XOR2"},
		{"mux", func(g *aig.Graph, x []aig.Edge) aig.Edge { return g.Mux(x[0], x[1], x[2]) }, "MUX2"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			graph := aig.NewGraph()
			inputs := []aig.Edge{graph.NewInput(), graph.NewInput(), graph.NewInput()}
			graph.AddOutput(tc.build(graph, inputs))
			pars := &Amap_Par_t{}
			Amap_ManSetDefaultParams(pars)
			if tc.name == "mux" {
				// The C oracle is run with `amap -m`; enable MUX cut matching here.
				pars.FUseMuxes = 1
			}
			mapped := Amap_ManMapGraph(graph, lib, pars)
			cleanupMappedOutput(t, mapped)
			found := false
			output := false
			for _, item := range Vec_PtrArray(mapped) {
				if out, ok := item.(*Amap_Out_t); ok && out != nil {
					if tc.want != "" && out.Type == 0 && out.pName == tc.want {
						found = true
					}
					if out.Type == 1 {
						output = true
					}
				}
			}
			if !output {
				t.Fatal("mapping omitted the output node")
			}
			if tc.want != "" && !found {
				t.Fatalf("mapped output does not use %s", tc.want)
			}
		})
	}
}

func mapTestLibrary(t *testing.T) *Amap_Lib_t {
	t.Helper()
	text, err := os.ReadFile("test_vectors/amap_mux.genlib")
	if err != nil {
		t.Fatal(err)
	}
	lib := Amap_LibReadAndPrepare("test_vectors/amap_mux.genlib", string(text), 0, 0)
	if lib == nil || lib.pGateInv == nil {
		t.Fatal("prepare mapping test library")
	}
	t.Cleanup(func() { Amap_LibFree(lib) })
	return lib
}

func cleanupMappedOutput(t *testing.T, result *Vec_Ptr_t) {
	t.Helper()
	if result == nil || Vec_PtrSize(result) == 0 {
		return
	}
	t.Cleanup(func() {
		if mem, ok := Vec_PtrEntry(result, Vec_PtrSize(result)-1).(*Aig_MmFlex_t); ok {
			Aig_MmFlexStop(mem, 0)
		}
		Vec_PtrFree(result)
	})
}
