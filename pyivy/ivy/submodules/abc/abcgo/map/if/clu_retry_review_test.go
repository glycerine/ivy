package if_

import (
	"slices"
	"testing"
)

// Fixed vectors from an independent scalar enumeration of the C retry loops.
// Counts use sets of Boolean cofactor rows; no Go decomposition routine or C
// compiler produced these expectations. Every table has full eight-input support.
func TestCluReducedBoundRetrySourceOrder(t *testing.T) {
	type step struct {
		tag                  string
		bound, size, classes int32
		pins, order          []int32
	}
	cases := []struct {
		name  string
		truth []uint64
		steps []step
	}{
		{name: "all_retries_fail", truth: []uint64{0x5577bb66d5972b96, 0x5d79b2695577bb66, 0x5d79b2695d79b269, 0x5d79b269dd992299}, steps: []step{
			{tag: "initial", bound: 6, size: 0, classes: 5, pins: []int32{1, 0, 2, 3, 4, 5}, order: []int32{7, 6, 1, 0, 2, 3, 4, 5}},
			{tag: "minus1", bound: 5, size: 0, classes: 5, pins: []int32{6, 7, 0, 2, 3}, order: []int32{1, 5, 4, 6, 7, 0, 2, 3}},
			{tag: "minus2", bound: 4, size: 0, classes: 5, pins: []int32{5, 1, 7, 0}, order: []int32{4, 6, 3, 2, 5, 1, 7, 0}},
			{tag: "reverse", bound: 6, size: 0, classes: 5, pins: []int32{7, 0, 1, 5, 2, 3}, order: []int32{4, 6, 7, 0, 1, 5, 2, 3}},
		}},
		{name: "smaller_disjoint_after_failure", truth: []uint64{0xfff733fffdcccef7, 0xff3733ffcdccfe37, 0xff3333ffccccff33, 0xffb733ffedccdeb7}, steps: []step{
			{tag: "initial", bound: 6, size: 0, classes: 5, pins: []int32{7, 0, 2, 3, 4, 5}, order: []int32{1, 6, 7, 0, 2, 3, 4, 5}},
			{tag: "minus1", bound: 5, size: 0, classes: 5, pins: []int32{6, 1, 0, 2, 3}, order: []int32{7, 5, 4, 6, 1, 0, 2, 3}},
			{tag: "minus2", bound: 4, size: 4, classes: 2, pins: []int32{6, 7, 2, 0}, order: []int32{4, 5, 3, 1, 6, 7, 2, 0}},
		}},
		{name: "smaller_overrides_shared", truth: []uint64{0x9339333339939993, 0xa88a88888aa8aaa8, 0x6226222226626662, 0xc77c77777cc7ccc7}, steps: []step{
			{tag: "initial", bound: 6, size: 0, classes: 5, pins: []int32{1, 0, 2, 3, 4, 5}, order: []int32{7, 6, 1, 0, 2, 3, 4, 5}},
			{tag: "minus1", bound: 5, size: 5, classes: 4, pins: []int32{2, 3, 4, 5, 0}, order: []int32{7, 6, 1, 2, 3, 4, 5, 0}},
			{tag: "minus2", bound: 4, size: 4, classes: 2, pins: []int32{2, 3, 4, 5}, order: []int32{6, 1, 7, 0, 2, 3, 4, 5}},
		}},
		{name: "reverse_shared_after_failures", truth: []uint64{0xff19ffffffffe619, 0xfffbffffffff04fb, 0xffeeffffffff11ee, 0xff79ffffffff8679}, steps: []step{
			{tag: "initial", bound: 6, size: 0, classes: 5, pins: []int32{1, 0, 2, 3, 4, 5}, order: []int32{7, 6, 1, 0, 2, 3, 4, 5}},
			{tag: "minus1", bound: 5, size: 0, classes: 5, pins: []int32{6, 7, 0, 2, 3}, order: []int32{1, 5, 4, 6, 7, 0, 2, 3}},
			{tag: "minus2", bound: 4, size: 0, classes: 4, pins: []int32{6, 1, 7, 0}, order: []int32{4, 5, 3, 2, 6, 1, 7, 0}},
			{tag: "reverse", bound: 6, size: 6, classes: 3, pins: []int32{7, 0, 1, 6, 2, 3}, order: []int32{4, 5, 7, 0, 1, 6, 2, 3}},
		}},
		{name: "reverse_overrides_disjoint", truth: []uint64{0x55555555, 0x20228affffffff, 0x55555555, 0x202082ffffffff}, steps: []step{
			{tag: "initial", bound: 6, size: 0, classes: 4, pins: []int32{1, 0, 2, 3, 4, 5}, order: []int32{7, 6, 1, 0, 2, 3, 4, 5}},
			{tag: "minus1", bound: 5, size: 5, classes: 2, pins: []int32{7, 1, 4, 2, 3}, order: []int32{6, 5, 0, 7, 1, 4, 2, 3}},
			{tag: "minus2", bound: 4, size: 0, classes: 3, pins: []int32{4, 2, 1, 3}, order: []int32{0, 6, 5, 7, 4, 2, 1, 3}},
			{tag: "reverse", bound: 6, size: 6, classes: 3, pins: []int32{1, 3, 2, 4, 7, 5}, order: []int32{0, 6, 1, 3, 2, 4, 7, 5}},
		}},
		{name: "failed_reverse_different_classes", truth: []uint64{0xfa5c3693c963399, 0xa5a829628692266, 0xc06af05a00aa00aa, 0x8095a0a500550055}, steps: []step{
			{tag: "initial", bound: 6, size: 0, classes: 5, pins: []int32{1, 0, 2, 3, 4, 5}, order: []int32{7, 6, 1, 0, 2, 3, 4, 5}},
			{tag: "minus1", bound: 5, size: 0, classes: 5, pins: []int32{6, 7, 0, 2, 3}, order: []int32{1, 5, 4, 6, 7, 0, 2, 3}},
			{tag: "minus2", bound: 4, size: 0, classes: 4, pins: []int32{5, 1, 7, 2}, order: []int32{4, 6, 3, 0, 5, 1, 7, 2}},
			{tag: "reverse", bound: 6, size: 0, classes: 5, pins: []int32{7, 2, 1, 5, 0, 3}, order: []int32{4, 6, 7, 2, 1, 5, 0, 3}},
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			const n = 8
			before := slices.Clone(tc.truth)
			if group := If_CluDecUsingCofs(tc.truth, n, 6); group.nVars != 0 {
				t.Fatalf("fixture bypasses retries through easy group%+v", group)
			}
			work := slices.Clone(tc.truth)
			v2p, p2v := cluIdentity(n)
			for _, want := range tc.steps {
				if want.tag == "reverse" {
					If_CluReverseOrder(work, n, v2p, p2v, 0)
				}
				group := If_CluFindGroup(work, n, 0, 0, v2p, p2v, want.bound, 0)
				if group.nVars != want.size || group.nMyu != want.classes || !slices.Equal(group.pVars[:want.bound], want.pins) || !slices.Equal(p2v[:n], want.order) {
					t.Fatalf("%s group%+v order%v want size%d classes%d pins%v order%v", want.tag, group, p2v[:n], want.size, want.classes, want.pins, want.order)
				}
				for position, variable := range p2v[:n] {
					if v2p[variable] != int32(position) {
						t.Fatalf("%s inverse permutation", want.tag)
					}
				}
				for x := 0; x < 256; x++ {
					original := 0
					for position, variable := range want.order {
						original |= (x >> position & 1) << variable
					}
					actual := work[x>>6] >> uint(x&63) & 1
					expected := tc.truth[original>>6] >> uint(original&63) & 1
					if actual != expected {
						t.Fatalf("%s permuted assignment%d", want.tag, x)
					}
				}
			}
			var root If_Grp_t
			root.nVars = -1
			f0, f1 := ^uint64(0), ^uint64(0)
			group := If_CluCheck(nil, tc.truth, n, 0, 0, 6, 6, &root, &f0, &f1, nil, 0)
			last := tc.steps[len(tc.steps)-1]
			if group.nVars != last.size || group.nMyu != last.classes || !slices.Equal(group.pVars[:last.bound], last.pins) {
				t.Fatalf("final group%+v want%+v", group, last)
			}
			if group.nVars == 0 {
				if root != (If_Grp_t{}) || f0 != 0 || f1 != 0 {
					t.Fatal("failed search did not clear outputs")
				}
			} else {
				if root.nVars < 1 || root.nVars > 6 {
					t.Fatalf("root%+v", root)
				}
				for x := 0; x < 256; x++ {
					index := 0
					for pin, v := range group.pVars[:group.nVars] {
						index |= (x >> v & 1) << pin
					}
					boundBit := f1 >> index & 1
					index = 0
					for pin, v := range root.pVars[:root.nVars] {
						bit := uint64(x >> v & 1)
						if v == n {
							bit = boundBit
						}
						index |= int(bit) << pin
					}
					if actual, expected := f0>>index&1, tc.truth[x>>6]>>uint(x&63)&1; actual != expected {
						t.Fatalf("assignment%d got%d want%d", x, actual, expected)
					}
				}
			}
			if !slices.Equal(tc.truth, before) {
				t.Fatal("search modified supplied truth")
			}
			if got := If_CluCheck(nil, tc.truth, n, 0, 0, 6, 6, nil, nil, nil, nil, 0); got != group {
				t.Fatalf("selection-only result changed: %+v", got)
			}
		})
	}
}
