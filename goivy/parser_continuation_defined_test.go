package goivy

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Python inst_mod prefixes with to_pref = module.defined, which for a
// continuation object ("object obj = { ... }") includes the names it
// inherited from the earlier obj, so "var x : foo" refers to obj.foo
// (test/cont1.ivy).
func TestContinuationObjectPrefixesInheritedNames(t *testing.T) {
	src := `#lang ivy1.7

module mod = {
    object obj = {
        type foo
    }
}

instantiate mod

object obj = { ...
   var x : foo
}
`
	fn := filepath.Join(t.TempDir(), "cont.ivy")
	if err := os.WriteFile(fn, []byte(src), 0644); err != nil {
		t.Fatal(err)
	}
	res, err := ReadModule(fn, false, New().Cfg)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, d := range res.Decls {
		c := string(d.Canon())
		if strings.Contains(c, "rep:obj.x") {
			found = true
			if !strings.Contains(c, "aSort:obj.foo") {
				t.Fatalf("obj.x sort not prefixed: %s", c)
			}
		}
	}
	if !found {
		t.Fatal("no obj.x declaration")
	}
}

// Python include sets parent_object = "this", so the included file's
// top-level Ivy shares the including module's defined dict; an object
// declared in one included file is continued, with its names, in another
// (test/cont2.ivy).
func TestIncludedFilesShareDefinedNames(t *testing.T) {
	dir := t.TempDir()
	files := map[string]string{
		"conta.ivy": "#lang ivy1.7\n\nobject foo = {\n    object obj = {\n        type foo\n    }\n}\n",
		"contb.ivy": "#lang ivy1.7\n\nobject foo = { ...\n    object obj = { ...\n        var x : foo\n    }\n}\n",
		"main.ivy":  "#lang ivy1.7\n\ninclude conta\ninclude contb\n",
	}
	for n, src := range files {
		if err := os.WriteFile(filepath.Join(dir, n), []byte(src), 0644); err != nil {
			t.Fatal(err)
		}
	}
	t.Chdir(dir) // Python import_module looks in the current directory
	res, err := ReadModule("main.ivy", false, New().Cfg)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, d := range res.Decls {
		c := string(d.Canon())
		if strings.Contains(c, "rep:foo.obj.x") {
			found = true
			if !strings.Contains(c, "aSort:foo.obj.foo") {
				t.Fatalf("foo.obj.x sort not prefixed: %s", c)
			}
		}
	}
	if !found {
		t.Fatal("no foo.obj.x declaration")
	}
}
