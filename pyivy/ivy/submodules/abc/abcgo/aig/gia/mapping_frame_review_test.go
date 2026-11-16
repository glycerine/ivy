package gia

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	abcmain "github.com/glycerine/abcgo/base/main"
)

func TestMappingFrameTimingSetterBoundaries(t *testing.T) {
	for _, ci := range []bool{true, false} {
		t.Run(fmt.Sprintf("CI%t", ci), func(t *testing.T) {
			frame := abcmain.NewFrame()
			defer frame.Close()
			p := Gia_ManStart(8)
			a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
			Gia_ManAppendCo(p, Gia_ManAppendAnd(p, a, b))
			frame.SetResource("pGia", p)
			set := Abc_FrameMiniAigSetCoRequireds
			current := func() []int32 { return Vec_IntArray(p.vCoReqs) }
			count := 1
			absent := "Required times are not given.\n"
			if ci {
				set = Abc_FrameMiniAigSetCiArrivals
				current = func() []int32 { return Vec_IntArray(p.vCiArrs) }
				count = 2
				absent = "Arrival times are not given.\n"
			}
			for iteration := int32(1); iteration <= 3; iteration++ {
				values := make([]int32, count)
				for i := range values {
					values[i] = iteration*10 + int32(i)
				}
				want := slices.Clone(values)
				if out := hashReviewStdout(t, func() { set(frame, values) }); out != "" {
					t.Fatal(out)
				}
				values[0] = -1
				if !slices.Equal(current(), want) {
					t.Fatal("setter retained caller storage or failed replacement")
				}
				out := hashReviewStdout(t, func() { set(frame, nil) })
				if out != absent || !slices.Equal(current(), want) {
					t.Fatalf("nil input changed timing state: %q %v", out, current())
				}
				var failure any
				func() { defer func() { failure = recover() }(); set(frame, make([]int32, count+1)) }()
				if failure == nil || !slices.Equal(current(), want) {
					t.Fatal("bad dimensions accepted or changed existing state")
				}
			}
			if out := hashReviewStdout(t, func() { set(nil, nil) }); out != absent {
				t.Fatalf("nil-array validation order: %q", out)
			}
			out := hashReviewStdout(t, func() { set(nil, []int32{1}) })
			if !strings.Contains(out, "ABC framework is not initialized") {
				t.Fatalf("missing nil-frame diagnostic: %q", out)
			}
			missing := abcmain.NewFrame()
			defer missing.Close()
			out = hashReviewStdout(t, func() { set(missing, []int32{1}) })
			if out != "Current network in ABC framework is not defined.\n" {
				t.Fatalf("missing graph diagnostic: %q", out)
			}
		})
	}
}

func TestMappingFrameReadAndDelayBoundaries(t *testing.T) {
	var result []int32
	output := hashReviewStdout(t, func() { result = Abc_FrameMiniAigReadCoArrivals(nil) })
	if result != nil || !strings.Contains(output, "ABC framework is not initialized") {
		t.Fatalf("nil-frame read: %v %q", result, output)
	}
	frame := abcmain.NewFrame()
	defer frame.Close()
	output = hashReviewStdout(t, func() { result = Abc_FrameMiniAigReadCoArrivals(frame) })
	if result != nil || output != "Current network in ABC framework is not defined.\n" {
		t.Fatalf("absent graph read: %v %q", result, output)
	}
	// Safe Go error returns retain the original diagnostics where C's delay
	// setter would proceed to dereference a missing frame/network.
	for _, f := range []*abcmain.Frame{nil, frame} {
		if out := hashReviewStdout(t, func() { Abc_FrameMiniAigSetAndGateDelay(f, 9) }); out == "" {
			t.Fatal("missing invalid-frame delay diagnostic")
		}
	}
	p := Gia_ManStart(8)
	frame.SetResource("pGia", p)
	a := Gia_ManAppendCi(p)
	Gia_ManAppendCo(p, a)
	Gia_ManAppendCo(p, a^1)
	output = hashReviewStdout(t, func() { result = Abc_FrameMiniAigReadCoArrivals(frame) })
	if result != nil || output != "Current network in ABC framework has no CO arrival times.\n" {
		t.Fatalf("absent arrival read: %v %q", result, output)
	}
	for i := int32(0); i < 3; i++ {
		p.SetMappingOutputArrivals([]int32{i, i + 7})
		result = Abc_FrameMiniAigReadCoArrivals(frame)
		if !slices.Equal(result, []int32{i, i + 7}) {
			t.Fatal("arrival values changed")
		}
		result[0] = -1
		if again := Abc_FrameMiniAigReadCoArrivals(frame); !slices.Equal(again, []int32{i, i + 7}) {
			t.Fatal("reader exposed mutable owned storage")
		}
	}
	for _, delay := range []int32{-9, 0, 7, 1 << 30} {
		Abc_FrameMiniAigSetAndGateDelay(frame, delay)
		if p.And2Delay != delay {
			t.Fatalf("delay%d stored%d", delay, p.And2Delay)
		}
	}
}

func TestMappingFrameEmptyInterface(t *testing.T) {
	frame := abcmain.NewFrame()
	defer frame.Close()
	p := Gia_ManStart(2)
	frame.SetResource("pGia", p)
	output := hashReviewStdout(t, func() {
		Abc_FrameMiniAigSetCiArrivals(frame, []int32{})
		Abc_FrameMiniAigSetCoRequireds(frame, []int32{})
		p.SetMappingOutputArrivals([]int32{})
		if result := Abc_FrameMiniAigReadCoArrivals(frame); len(result) != 0 {
			t.Fatal("empty interface has output arrivals")
		}
	})
	if output != "" || p.vCiArrs == nil || p.vCoReqs == nil {
		t.Fatalf("empty arrays treated as absent: %q", output)
	}
}
