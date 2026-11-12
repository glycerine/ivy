package dar

import "testing"

func TestLibraryClassPrioritiesAndPreparation(t *testing.T) {
	lib := Dar_LibRead()
	prepared := lib.Prepare(5)
	roots, classes := 0, 0
	for class, group := range lib.classes {
		roots += len(group)
		if len(group) > 0 {
			classes++
		}
		want := min(5, len(group))
		if class == 1 {
			want = len(group)
		}
		if len(prepared.classes[class].roots) != want {
			t.Fatal("prepared root count", class)
		}
		for i, root := range prepared.classes[class].roots {
			if root.libraryID != group[lib.priorities[class][i]] {
				t.Fatal("source priorities changed")
			}
		}
		// Every prepared DAG reproduces its library truth using only prior slots.
		truth := []uint16{0xaaaa, 0xcccc, 0xf0f0, 0xff00}
		for _, node := range prepared.classes[class].nodes {
			a, b := node.fanins[0], node.fanins[1]
			av, bv := truth[a>>1], truth[b>>1]
			if a&1 != 0 {
				av = ^av
			}
			if b&1 != 0 {
				bv = ^bv
			}
			truth = append(truth, av&bv)
			if truth[len(truth)-1] != lib.nodes[node.libraryID].truth {
				t.Fatal("prepared DAG truth changed")
			}
		}
	}
	if roots != 24772 || classes != 221 {
		t.Fatal(roots, classes)
	}
	if lib.Prepare(5) != prepared {
		t.Fatal("immutable preparation not reused")
	}
}
