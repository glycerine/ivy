package startingpoint

import (
	"os"
	"strings"
	"testing"
)

func TestConfigBranches(t *testing.T) {
	cfg, err := ParseConfig("# platform\nWIN32 = 0\nLINUX = 1\nLEVEL = 3\n")
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct{ condition, yes, no string }{
		{"if WIN32", "drop", "keep"}, {"ifdef WIN32", "drop", "keep"}, {"ifndef WIN32", "keep", "drop"},
		{"if defined(WIN32)", "drop", "keep"}, {"if defined LINUX && LEVEL >= 2", "keep", "drop"},
		{"if WIN32 || UNKNOWN", "unknown", "unknown"}, {"if WIN32 && UNKNOWN", "drop", "keep"},
	} {
		source := "#" + test.condition + "\nint first;\n#else\nint second;\n#endif\nint after;\n"
		got, err := selectConfiguredBranches(source, cfg)
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != len(source) || strings.Count(got, "\n") != strings.Count(source, "\n") {
			t.Fatal("offsets changed")
		}
		if test.yes == "unknown" {
			if got != source {
				t.Fatal(got)
			}
			continue
		}
		if strings.Contains(got, "int first;") != (test.yes == "keep") || strings.Contains(got, "int second;") != (test.no == "keep") || strings.Contains(got, "#endif") {
			t.Fatal(test, got)
		}
	}
}

func TestConfigElifNestedAndUnknown(t *testing.T) {
	for _, elif := range []string{"elif", "elseif"} {
		source := "#if WIN32\nint bad;\n#" + elif + " LINUX\n#if LEVEL == 3\nint chosen;\n#else\nint wrong;\n#endif\n#else\nint other;\n#endif\n#if UNKNOWN\nint untouched;\n#endif\n"
		r, err := AnalyzeWithOptions("x.c", source, Options{Defines: map[string]int64{"WIN32": 0, "LINUX": 1, "LEVEL": 3}})
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(r.Draft, "var chosen int32;") || strings.Contains(r.Draft, "bad") || strings.Contains(r.Draft, "wrong") || !strings.Contains(r.Draft, "#if UNKNOWN\nint untouched;\n#endif") {
			t.Fatal(r.Draft)
		}
		if len(r.Conditionals) != 1 || r.Conditionals[0].Location.Line != 12 {
			t.Fatal(r.Conditionals)
		}
	}
	for _, s := range []string{"WIN32 0", "WIN32 = nope", "WIN32 = 0\nWIN32 = 1", "WIN32 = 0 = 1"} {
		if _, err := ParseConfig(s); err == nil {
			t.Fatal("accepted", s)
		}
	}
}

func TestPthreadIncludeInOpaqueRegion(t *testing.T) {
	source := "#if UNKNOWN\n#include \"pthread.h\"\n#include <pthread.h>\n#include <other.h>\n#endif\n"
	r, err := Analyze("x.c", source)
	if err != nil || !strings.Contains(r.Draft, "//#include \"pthread.h\"") || !strings.Contains(r.Draft, "//#include <pthread.h>") || !strings.Contains(r.Draft, "\n#include <other.h>") {
		t.Fatal(r.Draft, err)
	}
}

func TestConfiguredStarter(t *testing.T) {
	source, err := os.ReadFile("../src/starter.c")
	if err != nil {
		t.Fatal(err)
	}
	r, err := AnalyzeWithOptions("starter.c", string(source), Options{Defines: map[string]int64{"WIN32": 0}})
	if err != nil || strings.Contains(r.Draft, "#ifdef WIN32") || strings.Contains(r.Draft, "#include \"pthread.h\"") || !strings.Contains(r.Draft, "//#include <pthread.h>") {
		t.Fatal(r.Draft, err)
	}
}

func TestConfigInsideHeaderGuard(t *testing.T) {
	source := `#ifndef HEADER_H
#define HEADER_H
#ifdef WIN32
int windows;
#else
int portable;
#endif
#endif
`
	r, err := AnalyzeWithOptions("x.h", source, Options{Defines: map[string]int64{"WIN32": 0}})
	if err != nil || strings.Contains(r.Draft, "#if") || strings.Contains(r.Draft, "windows") || !strings.Contains(r.Draft, "var portable int32;") {
		t.Fatal(r.Draft, err)
	}
}

func TestInlineConfigComments(t *testing.T) {
	cfg, err := ParseConfig("// settings\nWIN32 = 0 // disabled\nLEVEL=3// number\n")
	if err != nil || cfg["WIN32"] != 0 || cfg["LEVEL"] != 3 || len(cfg) != 2 {
		t.Fatal(cfg, err)
	}
}
