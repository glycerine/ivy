package abci

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"
)

// The int command (interpolation, intCore.c/intM114.c/satInterA.c) against
// native ABC: verbose progress lines (interpolant sizes, conflicts), the
// result, print_status and the CEX file.

var (
	intTiming  = regexp.MustCompile(` *[0-9]+\.[0-9]+ sec`)
	intMB      = regexp.MustCompile(` *[0-9]+(\.[0-9]+)? MB`)
	intPercent = regexp.MustCompile(`\( *[0-9]+\.[0-9]+ %\)`)
)

func intNormalize(out, dir string) string {
	out = intTiming.ReplaceAllString(out, " T sec")
	out = intMB.ReplaceAllString(out, " M MB")
	out = intPercent.ReplaceAllString(out, "(P %)")
	var lines []string
	for _, line := range strings.Split(out, "\n") {
		if strings.HasPrefix(line, "ABC command line") || strings.TrimSpace(line) == "" {
			continue
		}
		lines = append(lines, strings.TrimRight(strings.ReplaceAll(line, dir, "DIR"), " "))
	}
	return strings.Join(lines, "\n")
}

// intFiles appends the digests of the files the script wrote, then removes them.
func intFiles(dir string) string {
	var s string
	for _, name := range []string{"cex.out", "inv.aig"} {
		data, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			continue
		}
		// Native Ioa_WriteAiger appends a time-stamped comment section,
		// which the Go writer does not write.
		if i := bytes.LastIndex(data, []byte("c\nThis file was produced by the IOA package")); i >= 0 {
			data = data[:i]
		}
		s += fmt.Sprintf("\n%s %x", name, md5.Sum(data))
		os.Remove(filepath.Join(dir, name))
	}
	return s
}

func intDesigns() []string {
	var pre []string
	for _, f := range []string{"test_vectors/seq/s713.blif", "test_vectors/seq/s1494.blif", "test_vectors/native_s27_drw.blif"} {
		pre = append(pre, "read "+f+"; strash", "read "+f+"; strash; dc2; miter "+f)
	}
	return pre
}

// intSlow lists the runs that take native ABC more than a few seconds
// (int -c on the s713 dc2-miter: 126 interpolation steps, 94 s; every run on
// the s1494 dc2-miter: about 9 s), which are compared only under -v.
func intSlow(p, o string) bool {
	miter := strings.Contains(p, "miter")
	return miter && strings.Contains(p, "s713") && strings.Contains(o, "-c") ||
		miter && strings.Contains(p, "s1494") && o != "-v" && o != "-b -t -v"
}

func intCompare(t *testing.T, pre []string, opts []string) {
	dir := t.TempDir()
	for _, p := range pre {
		for _, o := range opts {
			if intSlow(p, o) {
				continue
			}
			o = strings.ReplaceAll(o, "INV", filepath.Join(dir, "inv.aig"))
			script := p + "; int " + o + "; print_status; write_cex -n " + filepath.Join(dir, "cex.out")
			native, done := runNativeABCTimeout(t, script, 120*time.Second)
			if !done {
				t.Errorf("native timed out: %s", script)
				continue
			}
			want := intNormalize(native, dir) + intFiles(dir)
			out, _ := runScriptStatus(t, script)
			got := intNormalize(out, dir) + intFiles(dir)
			if got != want {
				t.Errorf("%s:\n--- native\n%s\n--- port\n%s", script, want, got)
			}
		}
	}
}

func TestIntLikeNative(t *testing.T) {
	intCompare(t, intDesigns(), []string{"", "-v", "-k -v", "-k -d -v", "-C 20 -v", "-F 3 -v", "-K 2 -v", "-r -v", "-c -v", "-g -v", "-q -v", "-p -v", "-o -v", "-m -v", "-i -I INV -v"})
}

func TestIntTransLoopLikeNative(t *testing.T) {
	if testing.Short() {
		t.Skip("slow")
	}
	intCompare(t, intDesigns(), []string{"-t -v", "-b -t -v"})
}

// CBUGS_FIXED.md #42: "int -b" without -t mixed the transition relation of
// Inter_ManStartDuplicated with the backward initial interpolant of
// Inter_ManStartOneOutput (one more CI), so the connector clauses were
// wrong: safe designs were reported as failing ("Output -1 ... was asserted
// in frame 1") or the command looked for a nonexistent CEX with bmc3 forever.
func TestNativeIntBackwardProves(t *testing.T) {
	for _, script := range []string{
		"read test_vectors/seq/int_backward_safe.blif; strash; int -b; print_status",
		"read test_vectors/native_s27_drw.blif; strash; dc2; miter test_vectors/native_s27_drw.blif; int -b; print_status",
	} {
		out, done := runNativeABCTimeout(t, script, 60*time.Second)
		if !done || !strings.Contains(out, "Property proved.") {
			t.Errorf("native %s:\n%s", script, out)
		}
		if out, _ := runScriptStatus(t, script); !strings.Contains(out, "Property proved.") {
			t.Errorf("port %s:\n%s", script, out)
		}
	}
}
