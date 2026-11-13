package rwr

import (
	"bufio"
	"fmt"
	"os"
	"testing"
)

func TestSourceManagerClassOrdering(t *testing.T) {
	f, err := os.Open("test_vectors/manager_results.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	in := bufio.NewReader(f)
	var nodes, classes int
	if _, err = fmt.Fscan(in, &nodes, &classes); err != nil {
		t.Fatal(err)
	}
	p := rwrStartLibrary()
	if len(p.forest) != nodes || len(p.classes) != classes {
		t.Fatalf("source nodes/classes %d/%d Go %d/%d", nodes, classes, len(p.forest), len(p.classes))
	}
	for class, list := range p.classes {
		var count int
		if _, err = fmt.Fscan(in, &count); err != nil {
			t.Fatal(err)
		}
		if len(list) != count {
			t.Fatalf("class %d C=%d Go=%d", class, count, len(list))
		}
		for _, id := range list {
			var sourceID int
			var truth uint16
			if _, err = fmt.Fscanf(in, " %d:%d", &sourceID, &truth); err != nil {
				t.Fatal(err)
			}
			if id != sourceID || p.forest[id].truth != truth {
				t.Fatalf("class %d C=%d:%d Go=%d:%d", class, sourceID, truth, id, p.forest[id].truth)
			}
		}
	}
}
