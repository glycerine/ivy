package ivy2cpp

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMemberTargetTestGeneratedCPPRegression(t *testing.T) {
	fixture := filepath.Join(os.Getenv("HOME"), "ivy", "ivy-lang-examples", "jea", "member_test_ivy2cpp.ivy")
	if _, err := os.Stat(fixture); err != nil {
		t.Skipf("member ivy fixture not available: %v", err)
	}

	outDir := t.TempDir()
	batch, err := CompileAndGenerateAll(fixture, map[string]string{
		"target": "test",
		"outdir": outDir,
	}, Config{})
	if err != nil {
		t.Fatalf("CompileAndGenerateAll: %v", err)
	}
	if err := WriteBatchOutput(batch, outDir); err != nil {
		t.Fatalf("WriteBatchOutput: %v", err)
	}
	if len(batch.Outputs) != 1 {
		t.Fatalf("expected one generated output, got %d", len(batch.Outputs))
	}
	out := batch.Outputs[0]
	assertNoUnsupportedCPP(t, out)

	rawImpl, err := os.ReadFile(filepath.Join(outputDirectory(outDir), out.BaseName+".cpp"))
	if err != nil {
		t.Fatalf("read generated impl from temp dir: %v", err)
	}
	impl := normalizeCPP(string(rawImpl))
	for _, bad := range []string{
		"for(auto it=reliable_membership__delivered_upcall.memo.begin(),en=reliable_membership__delivered_upcall.memo.end(); it != en; ++it)if (it->second) {\n                    if (!(!reliable_membership__delivered_upcall[member_test_ivy2cpp::__tup__unsigned__unsigned__unsigned_long_long(n, E, V)]",
		"for(auto it=reliable_membership__pending_upcall.memo.begin(),en=reliable_membership__pending_upcall.memo.end(); it != en; ++it)if (it->second) {\n                    if (!(!reliable_membership__pending_upcall[member_test_ivy2cpp::__tup__unsigned__unsigned__unsigned_long_long(n, E, V)]",
		`obj.reliable_membership__pending_upcall_t0[this->n][this->e][this->v]`,
		`obj.reliable_membership__pending_upcall_until[this->n][this->e][this->v]`,
	} {
		if strings.Contains(impl, bad) {
			t.Fatalf("generated member target=test C++ still contains known bad pattern %q:\n%s", bad, impl)
		}
	}

	for _, want := range []string{
		`auto V = it->first.arg2;`,
		`obj.reliable_membership__pending_upcall_t0[member_test_ivy2cpp::__tup__unsigned__unsigned__unsigned_long_long(this->n, this->e, this->v)]`,
		`obj.reliable_membership__pending_upcall_until[member_test_ivy2cpp::__tup__unsigned__unsigned__unsigned_long_long(this->n, this->e, this->v)]`,
	} {
		if !strings.Contains(impl, want) {
			t.Fatalf("generated member target=test C++ missing tuple-keyed state read %q:\n%s", want, impl)
		}
	}

	if SlowCppTest {
		if _, err := BuildOutput(out, outDir); err != nil {
			if isMissingZ3ToolchainError(err) {
				t.Skip(err.Error())
			}
			t.Fatalf("build generated member target=test C++: %v", err)
		}
	}
}
