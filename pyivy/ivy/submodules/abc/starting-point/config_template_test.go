package startingpoint

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTemplateExcludesNearTopHeaderGuards(t *testing.T) {
	root := t.TempDir()
	header := filepath.Join(root, "guard.h")
	source := `// License and explanatory comments.
#pragma once
#include "preamble.h"
#ifndef Glucose_Alg_h /* file guard */
/* a comment does not interrupt the guard pair */
#define Glucose_Alg_h
#if WIN32
int windows;
#endif
#ifndef FEATURE_h
int optional;
#endif
#endif
int outside_guard;
`
	if err := os.WriteFile(header, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(root, "_starting.cfg")
	if _, err := WriteConfigTemplate(path, []string{header}); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "Glucose_Alg_h") {
		t.Fatal(string(data))
	}
	cfg, err := ParseConfig(string(data))
	if err != nil || len(cfg) != 2 {
		t.Fatal(cfg, err)
	}
	for _, name := range []string{"WIN32", "FEATURE_h"} {
		if _, ok := cfg[name]; !ok {
			t.Fatal(cfg)
		}
	}
}

func TestTemplateReportedGlucoseGuards(t *testing.T) {
	path := filepath.Join(t.TempDir(), "_starting.cfg")
	sources := []string{"../src/sat/glucose2/ParseUtils.h", "../src/sat/glucose2/Alg.h", "../src/sat/glucose2/Alloc.h", "../src/sat/glucose2/CGlucoseCore.h"}
	if _, err := WriteConfigTemplate(path, sources); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, guard := range []string{"Glucose2_ParseUtils_h", "Glucose_Alg_h", "Glucose_Alloc_h", "Glucose_CGlucoseCore_h"} {
		if strings.Contains(string(data), guard) {
			t.Fatal("retained", guard, string(data))
		}
	}
}

func TestConfigGuardBoundaries(t *testing.T) {
	for _, source := range []string{
		"#ifndef FEATURE\n#define FEATURE 0\n#endif\n",
		"int code;\n#ifndef LATE\n#define LATE\n#endif\n",
		"#if WIN32\n#ifndef NESTED\n#define NESTED\n#endif\n#endif\n",
		"#ifndef FIRST\n#define SECOND\n#endif\n",
	} {
		ts, err := Lex(source)
		if err != nil {
			t.Fatal(err)
		}
		if guards := configHeaderGuards(ts, "example.h"); len(guards) != 0 {
			t.Fatal(source, guards)
		}
	}
}

func TestTemplateIrregularABCGuards(t *testing.T) {
	path := filepath.Join(t.TempDir(), "_starting.cfg")
	sources := []string{"../src/misc/espresso/espresso.h", "../src/phys/place/place_base.h"}
	if _, err := WriteConfigTemplate(path, sources); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	cfg, err := ParseConfig(string(data))
	if err != nil {
		t.Fatal(err)
	}
	for _, guard := range []string{"ABC__misc__espresso__espresso_h", "ABC__phys__place__place_base_h", "PLACE_BASE_H_"} {
		if _, ok := cfg[guard]; ok {
			t.Fatal("retained guard", guard)
		}
	}
	for _, feature := range []string{"IBM_WATC", "IBMPC", "BPI"} {
		if _, ok := cfg[feature]; !ok {
			t.Fatal("lost feature", feature)
		}
	}
}

func TestHeaderSpecificGuardBoundaries(t *testing.T) {
	source := `#define FEATURE
#define example_h 1
#define ABC__other__unrelated_h
#ifndef FEATURE
#endif
#ifndef example_h
#endif
#ifndef ABC__other__unrelated_h
#endif
`
	ts, err := Lex(source)
	if err != nil {
		t.Fatal(err)
	}
	if guards := configHeaderGuards(ts, "example.h"); len(guards) != 0 {
		t.Fatal(guards)
	}
	source = "#if !defined (example_h)\n#define example_h\nint value;\n#endif\n"
	ts, err = Lex(source)
	if err != nil {
		t.Fatal(err)
	}
	if guards := configHeaderGuards(ts, "example.h"); len(guards) != 1 {
		t.Fatal(guards)
	}
}
