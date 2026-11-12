package gia

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestGiaAigStatsJSON(t *testing.T) {
	g := Gia_ManStart(8)
	g.pName = "name \"quoted\""
	a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	x := Gia_ManAppendAnd(g, a, b)
	Gia_ManAppendCo(g, x)
	path := filepath.Join(t.TempDir(), "stats.json")
	Gia_ManLogAigStats(g, path)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var got struct {
		Name                      string `json:"name"`
		Input, Output, And, Level int32
	}
	if err = json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got.Name != g.pName || got.Input != 2 || got.Output != 1 || got.And != 1 || got.Level != 1 {
		t.Fatalf("stats %+v", got)
	}
}
