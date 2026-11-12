package place

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func bookshelfFixture(t *testing.T) (string, string, string) {
	t.Helper()
	dir := t.TempDir()
	files := map[string]string{
		"test.nodes": "UCLA nodes 1.0\n# example\nNumNodes : 4\nNumTerminals : 2\na 1 1\nb 1 1\np 1 1 terminal\nq 1 1 terminal\n",
		"test.nets":  "UCLA nets 1.0\nNumNets : 2\nNumPins : 5\nNetDegree : 3 n1\na O : 0 0\nb I\np I\nNetDegree : 2 n2\nb O\nq I\n",
		"test.pl":    "UCLA pl 1.0\na 1 1 : N\nb 2 1 : N\np 0 0 : N /FIXED\nq 4 4 : N \\FIXED\n",
	}
	for name, text := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
	}
	return filepath.Join(dir, "test.nodes"), filepath.Join(dir, "test.nets"), filepath.Join(dir, "test.pl")
}
func TestBookshelfRoundTripAndTransactionalRead(t *testing.T) {
	resetPlacementTest()
	defer resetPlacementTest()
	nodes, nets, pl := bookshelfFixture(t)
	if err := ReadBookshelfFiles(nodes, nets, pl); err != nil {
		t.Fatal(err)
	}
	if g_place_numCells != 4 || g_place_numNets != 2 || !g_place_concreteCells[2].m_fixed || !g_place_concreteCells[3].m_fixed {
		t.Fatal("database import")
	}
	before := TotalWirelength()
	prefix := filepath.Join(t.TempDir(), "saved")
	if err := WriteBookshelf(prefix); err != nil {
		t.Fatal(err)
	}
	if err := ReadBookshelfFiles(prefix+".nodes", prefix+".nets", prefix+".pl"); err != nil {
		t.Fatal(err)
	}
	if TotalWirelength() != before || g_place_concreteCells[0].m_label != "a" {
		t.Fatal("roundtrip")
	}
	old := g_place_concreteCells[0]
	if err := os.WriteFile(nets, []byte("UCLA nets 1.0\nNumNets : 1\nNumPins : 1\nNetDegree : 1\nmissing I\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if ReadBookshelfFiles(nodes, nets, pl) == nil {
		t.Fatal("accepted unknown pin")
	}
	if g_place_concreteCells[0] != old {
		t.Fatal("failed read changed database")
	}
	if err := WriteBookshelf(filepath.Join(t.TempDir(), "missing", "out")); err == nil {
		t.Fatal("lost writer error")
	}
}
func TestBookshelfWorkflow(t *testing.T) {
	resetPlacementTest()
	defer resetPlacementTest()
	nodes, nets, pl := bookshelfFixture(t)
	if err := RunBookshelf([]string{nodes, nets, pl}); err != nil {
		t.Fatal(err)
	}
	if err := ReadBookshelfPlacement(pl); err != nil {
		t.Fatal(err)
	}
	for _, c := range g_place_concreteCells {
		if c == nil {
			continue
		}
		if !qpFinite(c.m_x) || !qpFinite(c.m_y) {
			t.Fatal("nonfinite workflow output")
		}
	}
	if RunBookshelf(nil) == nil {
		t.Fatal("missing usage error")
	}
}
func TestBookshelfMalformedAndConnectionDeletion(t *testing.T) {
	resetPlacementTest()
	defer resetPlacementTest()
	nodes, nets, pl := bookshelfFixture(t)
	if err := ReadBookshelfFiles(nodes, nets, pl); err != nil {
		t.Fatal(err)
	}
	a := g_place_concreteCells[0]
	n := NewNet(8, []*ConcreteCell{a, a, a, g_place_concreteCells[1]}, 1)
	AddNet(n)
	delNetConnections(a)
	if len(n.m_terms) != 1 || n.m_numTerms != 1 || n.m_terms[0] == a {
		t.Fatal("duplicate connection deletion")
	}
	contents, err := os.ReadFile(nodes)
	if err != nil {
		t.Fatal(err)
	}
	for _, replacement := range []string{"a NaN 1", "a -1 1", "b 1 1"} {
		bad := strings.Replace(string(contents), "a 1 1", replacement, 1)
		if err := os.WriteFile(nodes, []byte(bad), 0600); err != nil {
			t.Fatal(err)
		}
		if ReadBookshelfNodes(nodes) == nil {
			t.Fatal("accepted malformed node", replacement)
		}
	}
	oldX := a.m_x
	if err := os.WriteFile(pl, []byte("UCLA pl 1.0\na 99 1 : N\nmissing 2 3 : N\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if ReadBookshelfPlacement(pl) == nil || a.m_x != oldX {
		t.Fatal("placement updates not transactional")
	}
}
