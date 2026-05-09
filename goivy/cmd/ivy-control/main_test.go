package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPrepareStaticDirMaterializesEmbeddedAssetsIntoRunweb(t *testing.T) {
	tmp := t.TempDir()
	previous, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Chdir(previous); err != nil {
			t.Fatalf("restore working directory: %v", err)
		}
	}()

	staticDir, err := prepareStaticDir("", ".runweb")
	if err != nil {
		t.Fatalf("prepare static dir: %v", err)
	}
	if staticDir != ".runweb" {
		t.Fatalf("staticDir = %q, want .runweb", staticDir)
	}
	if false { // debug
		entries, err := os.ReadDir(staticDir)
		panicOn(err)
		var files []string
		for _, de := range entries {
			files = append(files, de.Name())
		}
		vv("staticDir = '%v'; entries= '%#v'", staticDir, files)
	}
	for _, name := range []string{
		"index.html",
		"z3-471-api.js",
		//filepath.Join("dist", "ivywebvue.js"),
	} {
		info, err := os.Stat(filepath.Join(tmp, ".runweb", name))
		if err != nil {
			t.Fatalf("materialized %s: %v", name, err)
		}
		if info.IsDir() {
			t.Fatalf("materialized %s as a directory", name)
		}
	}
}

func TestPrepareStaticDirHonorsExplicitStaticDir(t *testing.T) {
	tmp := t.TempDir()
	staticDir := filepath.Join(tmp, "static")
	got, err := prepareStaticDir(staticDir, filepath.Join(tmp, ".runweb"))
	if err != nil {
		t.Fatalf("prepare static dir: %v", err)
	}
	if got != staticDir {
		t.Fatalf("staticDir = %q, want %q", got, staticDir)
	}
	if _, err := os.Stat(filepath.Join(tmp, ".runweb")); !os.IsNotExist(err) {
		t.Fatalf("runweb dir was created despite explicit static-dir: %v", err)
	}
}
