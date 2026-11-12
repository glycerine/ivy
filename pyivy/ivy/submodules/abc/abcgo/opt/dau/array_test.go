package dau

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

func dauArrayTruth(nodes []Dau_DsdArrayNode) uint64 {
	index := 0
	var eval func() uint64
	eval = func() uint64 {
		n := nodes[index]
		index++
		var t uint64
		switch n.Type {
		case DAU_DSD_CONST0:
			t = 0
		case DAU_DSD_VAR:
			t = dauTruthVariables[n.iVar]
		case DAU_DSD_AND:
			t = ^uint64(0)
			for i := int32(0); i < n.nFans; i++ {
				t &= eval()
			}
		case DAU_DSD_XOR:
			for i := int32(0); i < n.nFans; i++ {
				t ^= eval()
			}
		case DAU_DSD_MUX:
			a, b, c := eval(), eval(), eval()
			t = a&b | ^a&c
		default:
			panic("unsupported test node")
		}
		if n.fCompl != 0 {
			t = ^t
		}
		return t
	}
	return eval()
}
func TestArrayFlattenPreservesComplementedXor(t *testing.T) {
	for _, s := range []string{"[a![bc]]", "![![ab]![cd]]", "(a(b(cd)))", "[a[b[cd]]]", "<a(bc)![de]>"} {
		store := make([]Dau_DsdArrayNode, 128)
		n := Dau_DsdConstruct(s, store)
		want := dauArrayTruth(store)
		Dau_DsdRemoveUseless(store)
		Dau_DsdDepth(store)
		if got := dauArrayTruth(store); got != want {
			t.Fatalf("%s changed truth from%x to%x", s, want, got)
		}
		for _, node := range store[:n] {
			if node.Type == DAU_DSD_NONE {
				break
			}
			if node.fUnused != 0 {
				t.Fatal("unused node retained")
			}
		}
	}
	store := make([]Dau_DsdArrayNode, 4)
	for _, s := range []string{"0", "1"} {
		Dau_DsdConstruct(s, store)
		if store[1].Type != DAU_DSD_NONE || DauDsdArrayString(store) != s {
			t.Fatal("constant sentinel/printing")
		}
	}
}
func TestArrayOriginalC(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("ELF section removal required")
	}
	cc, err := exec.LookPath("cc")
	if err != nil {
		t.Skip("optional C oracle requires cc")
	}
	_, file, _, _ := runtime.Caller(0)
	src := filepath.Clean(filepath.Join(filepath.Dir(file), "../../../src"))
	dir := t.TempDir()
	main, bin := filepath.Join(dir, "main.c"), filepath.Join(dir, "oracle")
	code := `#include "opt/dau/dauArray.c"
void dump(Dau_Dsd_t *p){int n=0;while(p[n].Type)n++;printf("%d %d",n,Dau_DsdCountAnd(p));for(int i=0;i<n;i++)printf(" %u %u %u %u %u %u %u %u",p[i].iVar,p[i].nFans,p[i].Depth,p[i].Offset,p[i].Data,p[i].Type,p[i].fCompl,p[i].fUnused);puts("");}
int main(void){char s[2000];while(scanf("%1999s",s)==1){Dau_Dsd_t p[1024]={0};Dau_DsdConstruct(s,p);dump(p);Dau_DsdRemoveUseless(p);dump(p);}return 0;}`
	if err := os.WriteFile(main, []byte(code), 0644); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(cc, "-O2", "-DABC_USE_STDINT_H", "-ffunction-sections", "-fdata-sections", "-Wl,--gc-sections", "-I", src, main, "-o", bin)
	if b, err := cmd.CombinedOutput(); err != nil {
		t.Fatal(err, string(b))
	}
	cases := []string{"0", "1", "a", "!b", "(ab)", "[abc]", "<abc>", "{abcd}", "[(abd)[cf](f(sg(he)))]", "!(a!(b(cd)))", "[a[b[cd]]]", "<a(bc)!(de)>"}
	var input bytes.Buffer
	var expected [][]int64
	for _, s := range cases {
		fmt.Fprintln(&input, s)
		p := make([]Dau_DsdArrayNode, 1024)
		Dau_DsdConstruct(s, p)
		for step := 0; step < 2; step++ {
			var values []int64
			n := 0
			for _, v := range p {
				if v.Type == DAU_DSD_NONE {
					break
				}
				n++
				values = append(values, int64(v.iVar), int64(v.nFans), int64(v.Depth), int64(v.Offset), int64(v.Data), int64(v.Type), int64(v.fCompl), int64(v.fUnused))
			}
			expected = append(expected, append([]int64{int64(n), int64(Dau_DsdCountAnd(p))}, values...))
			Dau_DsdRemoveUseless(p)
		}
	}
	cmd = exec.Command(bin)
	cmd.Stdin = &input
	b, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatal(err, string(b))
	}
	scan := bufio.NewScanner(bytes.NewReader(b))
	for sample, want := range expected {
		if !scan.Scan() {
			t.Fatal("missing source sample", sample)
		}
		f := strings.Fields(scan.Text())
		if len(f) != len(want) {
			t.Fatal("source diagnostic", scan.Text())
		}
		for i, s := range f {
			v, err := strconv.ParseInt(s, 10, 64)
			if err != nil || v != want[i] {
				t.Fatalf("sample%d field%d Go%d C%s", sample, i, want[i], s)
			}
		}
	}
}
