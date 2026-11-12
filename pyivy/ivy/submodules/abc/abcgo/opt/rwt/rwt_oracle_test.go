package rwt

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

func TestRewriteTablesMatchOriginalC(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("optional original-C oracle uses ELF section removal")
	}
	cc, err := exec.LookPath("cc")
	if err != nil {
		t.Skip("optional original-C oracle requires cc")
	}
	_, file, _, _ := runtime.Caller(0)
	source := filepath.Clean(filepath.Join(filepath.Dir(file), "../../../src"))
	dir := t.TempDir()
	main, binary := filepath.Join(dir, "main.c"), filepath.Join(dir, "oracle")
	code := `#include "opt/rwt/rwtUtil.c"
int main(void) {
  unsigned long long h=1469598103934665603ULL;
  int i, n=0;
  for (i=0; !(s_RwtAigSubgraphs[2*i]==0 && s_RwtAigSubgraphs[2*i+1]==0); i++) {
    h ^= s_RwtAigSubgraphs[2*i]; h *= 1099511628211ULL;
    h ^= s_RwtAigSubgraphs[2*i+1]; h *= 1099511628211ULL;
    n++;
  }
  printf("%d %llu %u %u\n",n,h,s_RwtPracticalClasses[0],s_RwtPracticalClasses[1]);
  return 0;
}`
	if err := os.WriteFile(main, []byte(code), 0644); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(cc, "-O2", "-DABC_USE_STDINT_H", "-ffunction-sections", "-fdata-sections", "-Wl,--gc-sections", "-I", source, main, "-o", binary)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatal(err, string(output))
	}
	output, err := exec.Command(binary).CombinedOutput()
	if err != nil {
		t.Fatal(err, string(output))
	}
	fields := strings.Fields(string(output))
	if len(fields) != 4 {
		t.Fatalf("unexpected C oracle output: %s", output)
	}
	n, err := strconv.Atoi(fields[0])
	if err != nil {
		t.Fatal(err)
	}
	hash, err := strconv.ParseUint(fields[1], 10, 64)
	if err != nil {
		t.Fatal(err)
	}
	first, _ := strconv.ParseUint(fields[2], 10, 16)
	second, _ := strconv.ParseUint(fields[3], 10, 16)
	if first != uint64(rwtPracticalClasses[0]) || second != uint64(rwtPracticalClasses[1]) {
		t.Fatalf("practical-class table prefix differs: Go %d,%d C %d,%d", rwtPracticalClasses[0], rwtPracticalClasses[1], first, second)
	}
	goHash := uint64(1469598103934665603)
	goCount := 0
	for i := 0; i+1 < len(rwtSubgraphs); i += 2 {
		a, b := rwtSubgraphs[i], rwtSubgraphs[i+1]
		if a == 0 && b == 0 {
			break
		}
		goHash ^= uint64(a)
		goHash *= 1099511628211
		goHash ^= uint64(b)
		goHash *= 1099511628211
		goCount++
	}
	if goCount != n || goHash != hash {
		t.Fatalf("rewrite table differs from C: Go count/hash %d/%d C %d/%d", goCount, goHash, n, hash)
	}
	if n == 0 {
		t.Fatal(fmt.Errorf("original rewrite table was empty"))
	}
}
