package abc

import (
	"encoding/hex"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestSopOriginalCParity(t *testing.T) {
	cc, e := exec.LookPath("cc")
	if e != nil {
		t.Skip("C compiler unavailable")
	}
	cxx, e := exec.LookPath("g++")
	if e != nil {
		t.Skip("C++ linker unavailable")
	}
	_, file, _, _ := runtime.Caller(0)
	source := filepath.Clean(filepath.Join(filepath.Dir(file), "../../../src"))
	if _, e = os.Stat(filepath.Join(source, "base/abc/abcSop.o")); e != nil {
		t.Skip("original ABC objects unavailable")
	}
	d := t.TempDir()
	c, o, exe := filepath.Join(d, "oracle.c"), filepath.Join(d, "oracle.o"), filepath.Join(d, "oracle")
	code := `#include "base/abc/abc.h"
void hex(char*s){while(*s)printf("%02x",(unsigned char)*s++);puts("");}
int main(void){int n,count,phase,i,j;char p[10000],row[100];while(scanf("%d%d%d",&n,&count,&phase)==3){for(i=0;i<count;i++){scanf("%s",row);for(j=0;j<n;j++)p[i*(n+3)+j]=row[j];p[i*(n+3)+n]=' ';p[i*(n+3)+n+1]='0'+phase;p[i*(n+3)+n+2]='\n';}p[count*(n+3)]=0;printf("%d %d %d %d %d %d %d %d %d %llx\n",Abc_SopGetCubeNum(p),Abc_SopGetLitNum(p),Abc_SopGetVarNum(p),Abc_SopGetPhase(p),Abc_SopIsBuf(p),Abc_SopIsInv(p),Abc_SopIsAndType(p),Abc_SopIsOrType(p),Abc_SopIsExorType(p),(unsigned long long)Abc_SopToTruth(p,n));Abc_SopComplement(p);hex(p);Abc_SopComplementVar(p,n-1);hex(p);}return 0;}`
	if e = os.WriteFile(c, []byte(code), 0600); e != nil {
		t.Fatal(e)
	}
	if out, e := exec.Command(cc, "-std=gnu99", "-DLIN64", "-I"+source, "-c", c, "-o", o).CombinedOutput(); e != nil {
		t.Fatalf("compile %v %s", e, out)
	}
	args := []string{o}
	e = filepath.WalkDir(source, func(p string, d os.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if !d.IsDir() && strings.HasSuffix(p, ".o") && p != filepath.Join(source, "base/main/main.o") {
			args = append(args, p)
		}
		return nil
	})
	if e != nil {
		t.Fatal(e)
	}
	args = append(args, "-lreadline", "-lpthread", "-lm", "-ldl", "-lz", "-o", exe)
	if out, e := exec.Command(cxx, args...).CombinedOutput(); e != nil {
		t.Fatalf("link %v %s", e, out)
	}
	r := rand.New(rand.NewSource(718))
	var input strings.Builder
	var covers []string
	for trial := 0; trial < 500; trial++ {
		n, count, phase := 1+trial%6, 1+r.Intn(8), r.Intn(2)
		fmt.Fprintln(&input, n, count, phase)
		var s strings.Builder
		for i := 0; i < count; i++ {
			pattern := make([]byte, n)
			for j := range pattern {
				pattern[j] = "01-"[r.Intn(3)]
			}
			fmt.Fprintln(&input, string(pattern))
			fmt.Fprintf(&s, "%s %d\n", pattern, phase)
		}
		covers = append(covers, s.String())
	}
	cmd := exec.Command(exe)
	cmd.Stdin = strings.NewReader(input.String())
	out, e := cmd.CombinedOutput()
	if e != nil {
		t.Fatalf("oracle %v %s", e, out)
	}
	fields := strings.Fields(string(out))
	if len(fields) != len(covers)*12 {
		t.Fatalf("oracle token count %d", len(fields))
	}
	for sample, s := range covers {
		n := Abc_SopGetVarNum(s)
		expected := []int32{Abc_SopGetCubeNum(s), Abc_SopGetLitNum(s), n, Abc_SopGetPhase(s), Abc_SopIsBuf(s), Abc_SopIsInv(s), Abc_SopIsAndType(s), Abc_SopIsOrType(s), Abc_SopIsExorType(s)}
		for j, want := range expected {
			if fields[sample*12+j] != fmt.Sprint(want) {
				t.Fatalf("sample%d metric%d", sample, j)
			}
		}
		if fields[sample*12+9] != fmt.Sprintf("%x", Abc_SopToTruth(s, n)) {
			t.Fatalf("sample%d truth", sample)
		}
		s = Abc_SopComplement(s)
		if fields[sample*12+10] != hex.EncodeToString([]byte(s)) {
			t.Fatalf("sample%d complement", sample)
		}
		s = Abc_SopComplementVar(s, n-1)
		if fields[sample*12+11] != hex.EncodeToString([]byte(s)) {
			t.Fatalf("sample%d variable complement", sample)
		}
	}
}
