package gia

import (
	"fmt"
	"testing"
)

// The C choice path is built from ordinary AND graphs. Accepting Go's typed
// XOR/MUX nodes must preserve their functions, not reinterpret their data pins.
func TestNativeChoiceMiterPreservesRealGateLogic(t *testing.T) {
	snapshots := make([]*Gia_Man_t, 3)
	for i := range snapshots {
		g := Gia_ManStart(24)
		a, b, c, d := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
		g.EnableMuxes()
		x := Gia_ManAppendXorReal(g, a, b^1)
		m := Gia_ManAppendMuxReal(g, c, b, d^1)
		Gia_ManAppendCo(g, x^int32(i&1))
		Gia_ManAppendCo(g, m^int32((i>>1)&1))
		snapshots[i] = g
	}
	out := Gia_ManChoiceMiter(snapshots)
	if Gia_ManCoNum(out) != 6 {
		t.Fatalf("nodes=%d outputs=%d", Gia_ManObjNum(out), Gia_ManCoNum(out))
	}
	for mask := 0; mask < 16; mask++ {
		in := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0, mask&8 != 0}
		x, m := in[0] != !in[1], in[2] && in[1] || !in[2] && !in[3]
		want := []bool{x, !x, x, m, m, !m}
		got, err := out.Evaluate(in)
		if err != nil {
			t.Fatal(err)
		}
		for i := range want {
			if got[i] != want[i] {
				t.Fatalf("mask %d outputs %v want %v", mask, got, want)
			}
		}
	}
}

func TestChoiceMiterCopiesControlOnlyCones(t *testing.T) {
	checkChoiceMiterControlOnlyCones(t, Gia_ManChoiceMiter)
}

func TestHcdChoiceMiterCopiesControlOnlyCones(t *testing.T) {
	checkChoiceMiterControlOnlyCones(t, func(graphs []*Gia_Man_t) *Gia_Man_t {
		networks := Vec_PtrAlloc(len(graphs))
		for _, graph := range graphs {
			Vec_PtrPush(networks, graph)
		}
		return Hcd_ManChoiceMiter(networks)
	})
}

func checkChoiceMiterControlOnlyCones(t *testing.T, build func([]*Gia_Man_t) *Gia_Man_t) {
	t.Helper()
	for signs := int32(0); signs < 32; signs++ {
		t.Run(fmt.Sprint(signs), func(t *testing.T) {
			graphs := make([]*Gia_Man_t, 3)
			controls := make([]int32, len(graphs))
			for version := range graphs {
				g := Gia_ManStart(1)
				a, b, c, d := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
				g.EnableMuxes()
				x := Gia_ManAppendXorReal(g, a^(signs&1), b^((signs>>1)&1))
				control := Gia_ManAppendBuf(g, x^((signs>>2)&1))
				root := Gia_ManAppendMuxReal(g, control, c^((signs>>3)&1), d^((signs>>4)&1))
				Gia_ManAppendCo(g, root^int32(version&1))
				Gia_ManAppendCo(g, 1)
				Gia_ManAppendCo(g, Gia_ManAppendBuf(g, a^int32(version&1)))
				graphs[version], controls[version] = g, control
			}
			miter := build(graphs)
			if Gia_ManHashStarted(miter) != 0 || Gia_ManXorNum(miter) != 0 || Gia_ManMuxNum(miter) != 0 {
				t.Fatal("miter must finish hashing and contain ordinary AND logic")
			}
			for version, graph := range graphs {
				if Gia_ManObj(graph, controls[version]>>1).Value == ^uint32(0) {
					t.Fatal("MUX control cone was not visited")
				}
			}
			for mask := 0; mask < 16; mask++ {
				inputs := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0, mask&8 != 0}
				got, err := miter.Evaluate(inputs)
				if err != nil {
					t.Fatal(err)
				}
				x := (inputs[0] != (signs&1 != 0)) != (inputs[1] != (signs&2 != 0))
				mux := inputs[3] != (signs&16 != 0)
				if x != (signs&4 != 0) {
					mux = inputs[2] != (signs&8 != 0)
				}
				for version := range graphs {
					phase := version&1 != 0
					want := []bool{mux != phase, true, inputs[0] != phase}
					for output, value := range want {
						if got[output*len(graphs)+version] != value {
							t.Fatalf("assignment %d, snapshot %d, output %d: %v, want %v", mask, version, output, got, want)
						}
					}
				}
			}
		})
	}
}
