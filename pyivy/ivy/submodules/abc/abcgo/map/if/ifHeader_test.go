package if_

import "testing"

func TestComplementProxyIsInvolutiveAndStable(t *testing.T) {
	obj := &If_Obj_t_{Type: IF_AND, Id: 7}
	neg := If_Not(obj)
	if neg == nil || neg == obj {
		t.Fatal("complement must have a distinct proxy")
	}
	if If_Not(obj) != neg {
		t.Fatal("repeated complement must preserve edge identity")
	}
	if If_Not(neg) != obj || If_Regular(neg) != obj {
		t.Fatal("double complement must recover the regular object")
	}
	if If_IsComplement(obj) != 0 || If_IsComplement(neg) != 1 {
		t.Fatal("complement state was not preserved")
	}
}

func TestCutCopyOwnsVariableLengthStorage(t *testing.T) {
	source := &If_Cut_t_{
		nLimit: 6, nLeaves: 3, pLeaves: []int32{2, 5, 8},
		permutation: []byte{1, 0, 2}, data: "source", dataInt: 9,
	}
	var dest If_Cut_t_
	If_CutCopy(nil, &dest, source)
	dest.pLeaves[0] = 99
	dest.permutation[0] = 99
	dest.data = "destination"
	if source.pLeaves[0] != 2 || source.permutation[0] != 1 || source.data != "source" {
		t.Fatal("cut copy aliases source storage")
	}
	if dest.dataInt != 9 || dest.nLeaves != 3 || dest.nLimit != 6 {
		t.Fatalf("cut metadata changed during copy: %+v", dest)
	}
}

func TestCutSupportMaskBoundaries(t *testing.T) {
	for _, tc := range []struct {
		leaves uint32
		want   uint32
	}{{0, 0}, {1, 1}, {5, 31}, {31, 0x7fffffff}, {32, 0xffffffff}} {
		cut := &If_Cut_t_{nLeaves: tc.leaves}
		if got := If_CutSuppMask(cut); got != tc.want {
			t.Errorf("support mask for %d leaves = %#x, want %#x", tc.leaves, got, tc.want)
		}
	}
}
