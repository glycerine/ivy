package gia

import (
	"slices"
	"testing"
)

func TestRepresentativeRemappingAgainstConnectivity(t *testing.T) {
	// Four partitions include constants, singletons, and two nonconstant
	// classes. Independently compute the image partition with graph closure.
	for _, heads := range [][]int32{{0, 0, 2, 2, 2}, {0, 1, 1, 3, 3}, {0, 0, 0, 0, 0}, {0, 1, 2, 3, 4}} {
		src, dst := Gia_ManStart(1), Gia_ManStart(1)
		for i := 0; i < 4; i++ {
			Gia_ManAppendCi(src)
			if i < 3 {
				Gia_ManAppendCi(dst)
			}
		}
		src.StartRepresentatives()
		for id, head := range heads {
			if head != int32(id) {
				Gia_ObjSetRepr(src, int32(id), head)
			}
		}
		src.PNexts = Gia_ManDeriveNexts(src)
		before := slices.Clone(src.PReprs)
		beforeNext := slices.Clone(src.PNexts)
		for code := 0; code < 625; code++ {
			mapping := []int32{0, 0, 0, 0, 0}
			for id, digits := 1, code; id < 5; id, digits = id+1, digits/5 {
				mapping[id] = int32(digits%5) - 1
			}
			var connected [4][4]bool
			for i := range connected {
				connected[i][i] = true
			}
			for i, a := range mapping {
				for j, b := range mapping {
					if a >= 0 && b >= 0 && heads[i] == heads[j] {
						connected[a][b] = true
					}
				}
			}
			for k := 0; k < 4; k++ {
				for i := 0; i < 4; i++ {
					for j := 0; j < 4; j++ {
						connected[i][j] = connected[i][j] || connected[i][k] && connected[k][j]
					}
				}
			}
			if err := dst.TransferRepresentativesFrom(src, mapping); err != nil {
				t.Fatalf("heads %v map %v: %v", heads, mapping, err)
			}
			for id := int32(0); id < 4; id++ {
				head := int32(0)
				for !connected[id][head] {
					head++
				}
				repr, next := int32(GIA_VOID), int32(0)
				if head != id {
					repr = head
				}
				if head != 0 {
					for candidate := id + 1; candidate < 4; candidate++ {
						if connected[id][candidate] {
							next = candidate
							break
						}
					}
				}
				if Gia_ObjRepr(dst, id) != repr || Gia_ObjNext(dst, id) != next {
					t.Fatalf("heads %v map %v object %d: repr/next %d/%d, want %d/%d", heads, mapping, id,
						Gia_ObjRepr(dst, id), Gia_ObjNext(dst, id), repr, next)
				}
			}
			if !slices.Equal(src.PReprs, before) || !slices.Equal(src.PNexts, beforeNext) {
				t.Fatal("remapping mutated source classes")
			}
		}
	}
}

func TestRepresentativeRemappingErrorsAreAtomic(t *testing.T) {
	for _, problem := range []string{"short map", "negative ID", "large ID", "short representatives", "forward representative"} {
		t.Run(problem, func(t *testing.T) {
			src, dst := Gia_ManStart(1), Gia_ManStart(1)
			for _, g := range []*Gia_Man_t{src, dst} {
				Gia_ManAppendCi(g)
				Gia_ManAppendCi(g)
				g.StartRepresentatives()
				Gia_ObjSetRepr(g, 2, 1)
				Gia_ObjSetNext(g, 1, 2)
			}
			mapping := []int32{0, 1, 2}
			switch problem {
			case "short map":
				mapping = mapping[:2]
			case "negative ID":
				mapping[2] = -2
			case "large ID":
				mapping[2] = 3
			case "short representatives":
				src.PReprs = src.PReprs[:2]
			case "forward representative":
				src.PReprs[2].iRepr = 2
			}
			before, next := slices.Clone(dst.PReprs), slices.Clone(dst.PNexts)
			oldRepr, oldNext := &dst.PReprs[0], &dst.PNexts[0]
			if err := dst.TransferRepresentativesFrom(src, mapping); err == nil {
				t.Fatal("invalid input accepted")
			}
			if !slices.Equal(dst.PReprs, before) || !slices.Equal(dst.PNexts, next) || &dst.PReprs[0] != oldRepr || &dst.PNexts[0] != oldNext {
				t.Fatal("failed transfer replaced or modified destination")
			}
		})
	}
}

func TestNameRemappingBothInterfacesAndOwnership(t *testing.T) {
	src, dst := Gia_ManStart(1), Gia_ManStart(1)
	for i := 0; i < 3; i++ {
		Gia_ManAppendCo(src, Gia_ManAppendCi(src))
		if i < 2 {
			Gia_ManAppendCo(dst, Gia_ManAppendCi(dst))
		}
	}
	src.VNamesIn, src.VNamesOut = Vec_PtrAlloc(2), Vec_PtrAlloc(3)
	for _, name := range []string{"a", "b"} { // missing third input name is legal
		Vec_PtrPush(src.VNamesIn, name)
	}
	for _, name := range []string{"x", "y", "z"} {
		Vec_PtrPush(src.VNamesOut, name)
	}
	for code := 0; code < 256; code++ {
		mapping := make([]int32, 4)
		for i, digits := 0, code; i < 4; i, digits = i+1, digits/4 {
			mapping[i] = int32(digits%4) - 1
		}
		if err := dst.CopyIONamesFrom(src, mapping[:2], mapping[2:]); err != nil {
			t.Fatal(err)
		}
		for i, got := range append(dst.InputNames(), dst.OutputNames()...) {
			want := ""
			names := []string{"a", "b", ""}
			if i >= 2 {
				names = []string{"x", "y", "z"}
			}
			if mapping[i] >= 0 {
				want = names[mapping[i]]
			}
			if got != want {
				t.Fatalf("map %v interface slot %d: %q, want %q", mapping, i, got, want)
			}
		}
	}
	oldIn, oldOut := dst.VNamesIn, dst.VNamesOut
	// A valid input map must not be installed before output validation fails.
	if err := dst.CopyIONamesFrom(src, []int32{0, 1}, []int32{0, 3}); err == nil || dst.VNamesIn != oldIn || dst.VNamesOut != oldOut {
		t.Fatal("invalid output map changed interface names")
	}
	if err := dst.CopyIONamesFrom(src, nil, nil); err == nil {
		t.Fatal("identity map accepted unequal interface sizes")
	}
	if err := dst.CopyIONamesFrom(src, []int32{1, 0}, []int32{2, 1}); err != nil {
		t.Fatal(err)
	}
	copyNames := dst.InputNames()
	copyNames[0] = "changed"
	Vec_PtrWriteEntry(dst.VNamesOut, 0, "changed")
	if dst.InputNames()[0] != "b" || src.OutputNames()[2] != "z" {
		t.Fatal("returned or destination name storage aliases its source")
	}
}
