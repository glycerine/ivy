package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestResubSearchSimpleTruths(t *testing.T) {
	a, b := word(0xaaaaaaaaaaaaaaaa), word(0xcccccccccccccccc)
	cases := []struct {
		name       string
		on         word
		limit, xor int32
	}{{"constant", 0, 0, 0}, {"buffer", a, 0, 0}, {"and", a & b, 1, 0}, {"or", a | b, 1, 0}, {"xor", a ^ b, 1, 1}}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			divs := vec.Vec_PtrAlloc(4)
			for _, v := range []word{^tc.on, tc.on, a, b} {
				vec.Vec_PtrPush(divs, []word{v})
			}
			p := Gia_ResbAlloc(1)
			Gia_ManResubPerform(p, divs, 1, tc.limit, 20, 0, tc.xor, 0, 0)
			out := []word{0}
			if Gia_ManResubVerify(p, out) != 1 || out[0] != tc.on {
				t.Fatalf("gates %v output %x want %x", vec.Vec_IntArray(p.vGates), out[0], tc.on)
			}
		})
	}
}

func TestResubSearchThreeInputFunctions(t *testing.T) {
	a, b, c := word(0xaaaaaaaaaaaaaaaa), word(0xcccccccccccccccc), word(0xf0f0f0f0f0f0f0f0)
	for truth := 0; truth < 256; truth++ {
		on := word(0)
		for bit := 0; bit < 64; bit++ {
			if truth&(1<<uint(bit&7)) != 0 {
				on |= word(1) << uint(bit)
			}
		}
		divs := vec.Vec_PtrAlloc(5)
		for _, v := range []word{^on, on, a, b, c} {
			vec.Vec_PtrPush(divs, []word{v})
		}
		p := Gia_ResbAlloc(1)
		Gia_ManResubPerform(p, divs, 1, 3, 20, 0, 1, 0, 0)
		if vec.Vec_IntSize(p.vGates) == 0 {
			continue
		}
		result := []word{0}
		if Gia_ManResubVerify(p, result) != 1 || result[0] != on {
			t.Fatalf("truth %02x gates %v result %x", truth, vec.Vec_IntArray(p.vGates), result[0])
		}
	}
}

func TestResubSearchMux(t *testing.T) {
	a, b, c := word(0xaaaaaaaaaaaaaaaa), word(0xcccccccccccccccc), word(0xf0f0f0f0f0f0f0f0)
	on := (a & b) | (^a & c)
	divs := vec.Vec_PtrAlloc(5)
	for _, v := range []word{^on, on, a, b, c} {
		vec.Vec_PtrPush(divs, []word{v})
	}
	p := Gia_ResbAlloc(1)
	Gia_ResbInit(p, divs, 1, 3, 20, 0, 1, 0, 0, 0)
	vec.Vec_IntPush(p.vBinateVars, 2)
	top := Gia_ManResubPerformMux_rec(p, 3)
	if top < 0 {
		t.Fatal("mux search failed")
	}
	vec.Vec_IntPush(p.vGates, top)
	result := []word{0}
	if Gia_ManResubVerify(p, result) != 1 || result[0] != on {
		t.Fatalf("mux %v %x", vec.Vec_IntArray(p.vGates), result[0])
	}
}
