package goivy

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDuoAIExportInlinesCallsAndEncodesEnums(t *testing.T) {
	const source = `#lang ivy1.8
type item = {first, second}
relation ready(X:item)
after init { ready(X) := false }
action set_ready(x:item) = { ready(x) := true }
action middle(x:item) = { call set_ready(x) }
export action step(x:item) = { call middle(x) }
invariant [safe] ready(first) | ~ready(first)
`
	mod := New()
	if err := SourceString("duoai.ivy", source, mod, mod.Sig, map[string]interface{}{"create_isolate": false}); err != nil {
		t.Fatal(err)
	}
	if err := CreateIsolate("this", mod); err != nil {
		t.Fatal(err)
	}
	mod.Name = filepath.Join(t.TempDir(), "output")
	if err := DuoAIExport(mod); err != nil {
		t.Fatal(err)
	}
	path := mod.Name + "_duoai.ivy"
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	if strings.Contains(text, "call ") || !strings.Contains(text, "axiom forall X:item. X = first | X = second") {
		t.Fatalf("DuoAI export did not inline calls or encode the enum:\n%s", text)
	}
	result := New()
	if err := SourceString(path, "#lang ivy1.8\n"+text, result, result.Sig, map[string]interface{}{"create_isolate": false}); err != nil {
		t.Fatalf("cannot load emitted Ivy: %v\n%s", err, text)
	}
}
