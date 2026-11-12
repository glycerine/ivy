package gia

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenReadAndCompare(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "patterns.pla")
	var data strings.Builder
	data.WriteString(".i 2\n.o 1\n")
	for i := 0; i < 64; i++ {
		if i&3 == 3 {
			data.WriteString("11 1\n")
		} else if i&3 == 2 {
			data.WriteString("01 0\n")
		} else if i&3 == 1 {
			data.WriteString("10 0\n")
		} else {
			data.WriteString("00 0\n")
		}
	}
	data.WriteString(".e\n")
	if err := os.WriteFile(path, []byte(data.String()), 0600); err != nil {
		t.Fatal(err)
	}
	p := Gia_ManStart(8)
	a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	Gia_ManAppendCo(p, Gia_ManAppendAnd(p, a, b))
	stats := filepath.Join(dir, "stats.json")
	Gia_ManTestOneFile(p, path, stats)
	out, err := os.ReadFile(stats)
	if err != nil {
		t.Fatal(err)
	}
	var got struct {
		Total   int `json:"total"`
		Correct int `json:"correct"`
		Guess   int `json:"guess"`
	}
	if err = json.Unmarshal(out, &got); err != nil || got.Total != 64 || got.Correct != 64 || got.Guess != 48 {
		t.Fatalf("stats=%+v err=%v", got, err)
	}
}
