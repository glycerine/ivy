package core
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

func TestPreparedCutImplementations(t *testing.T) {
	prepared := Dar_LibPrepare(5)
	for _, truth := range []uint16{0x6996, 0x8888, 0xf0aa, 0xffff} {
		match, nodes, roots := prepared.CutImplementations(truth)
		values := []uint16{0xaaaa, 0xcccc, 0xf0f0, 0xff00}
		for _, node := range nodes {
			read := func(lit uint32) uint16 {
				v := values[lit>>1]
				if lit&1 != 0 {
					v = ^v
				}
				return v
			}
			values = append(values, read(node.Fanins[0])&read(node.Fanins[1]))
		}
		for _, root := range roots {
			value := values[root.Index]
			if root.Complement {
				value = ^value
			}
			if value != match.Canonical {
				t.Fatalf("truth %04x canonical %04x root %04x", truth, match.Canonical, value)
			}
		}
	}
}
