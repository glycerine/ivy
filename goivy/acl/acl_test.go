package acl

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRegisterFromFile_LiteralsAndRegex(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "acl.yaml")
	content := `
ignores:
  - "safety_prop_1"
  - "regex(helper_.*)"
assumes:
  - "liveness_prop_2"
  - "regex(ext_.*)"
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := RegisterFromFile(path)
	if err != nil {
		t.Fatalf("RegisterFromFile: %v", err)
	}

	if !cfg.IsIgnored("safety_prop_1") {
		t.Error("literal ignore should match")
	}
	if !cfg.IsIgnored("helper_foo") {
		t.Error("regex ignore should match helper_foo")
	}
	if cfg.IsIgnored("xhelper_foo") {
		t.Error("regex ignore should NOT match xhelper_foo (start-of-string anchor)")
	}
	if cfg.IsIgnored("liveness_prop_2") {
		t.Error("liveness_prop_2 should not be ignored")
	}

	if !cfg.IsAssumed("liveness_prop_2") {
		t.Error("literal assume should match")
	}
	if !cfg.IsAssumed("ext_something") {
		t.Error("regex assume should match ext_something")
	}
	if cfg.IsAssumed("not_ext_something") {
		t.Error("regex assume should NOT match not_ext_something (start-of-string anchor)")
	}
}

func TestRegisterFromFile_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.yaml")
	if err := os.WriteFile(path, []byte(""), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := RegisterFromFile(path)
	if err != nil {
		t.Fatalf("RegisterFromFile on empty file: %v", err)
	}
	if cfg.IsIgnored("anything") || cfg.IsAssumed("anything") {
		t.Error("empty ACL should match nothing")
	}
}

func TestRegisterFromFile_IgnoresOnly(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "ign.yaml")
	content := `
ignores:
  - "prop_a"
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := RegisterFromFile(path)
	if err != nil {
		t.Fatalf("RegisterFromFile: %v", err)
	}
	if !cfg.IsIgnored("prop_a") {
		t.Error("prop_a should be ignored")
	}
	if cfg.IsAssumed("prop_a") {
		t.Error("prop_a should not be assumed")
	}
}

func TestRegisterFromFile_MissingFile(t *testing.T) {
	_, err := RegisterFromFile("/nonexistent/path.yaml")
	if err == nil {
		t.Fatal("should error on missing file")
	}
}

func TestRegexAnchoredToStart(t *testing.T) {
	cfg := NewConfig()
	if err := cfg.RegisterIgnores([]string{"regex(foo.*)"}); err != nil {
		t.Fatal(err)
	}
	if !cfg.IsIgnored("foobar") {
		t.Error("foobar should match regex(foo.*)")
	}
	if cfg.IsIgnored("xfoobar") {
		t.Error("xfoobar should NOT match — regex is anchored to start of string")
	}
}
