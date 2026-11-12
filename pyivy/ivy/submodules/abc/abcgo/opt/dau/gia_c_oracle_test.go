package dau

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/aig"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestDSDGraphOriginalC(t *testing.T) {
	cc, err := exec.LookPath("cc")
	if err != nil {
		t.Skip("C compiler unavailable")
	}
	cxx, err := exec.LookPath("g++")
	if err != nil {
		t.Skip("original ABC objects require C++ linker")
	}
	_, file, _, _ := runtime.Caller(0)
	src := filepath.Clean(filepath.Join(filepath.Dir(file), "../../../src"))
	if _, err := os.Stat(filepath.Join(src, "opt/dau/dauGia.o")); err != nil {
		t.Skip("original ABC object files unavailable")
	}
	dir := t.TempDir()
	source, object, exe := filepath.Join(dir, "oracle.c"), filepath.Join(dir, "oracle.o"), filepath.Join(dir, "oracle")
	code := `#include "opt/dau/dauInt.h"
#include "aig/gia/gia.h"
extern int Dau_DsdToGia(Gia_Man_t*,char*,int*,Vec_Int_t*);
int main(void){char s[8192];while(scanf("%8191s",s)==1){Gia_Man_t*p=Gia_ManStart(1000);Vec_Int_t*cover=Vec_IntAlloc(1000);int leaves[12],i,k,root;Gia_Obj_t*o;word truth=0;p->vLevels=Vec_IntStart(1000);Gia_ManHashStart(p);Gia_ManConst0(p)->Value=0;for(i=0;i<6;i++)leaves[i]=Gia_ManAppendCi(p);root=Dau_DsdToGia(p,s,leaves,cover);for(k=0;k<64;k++){Gia_ManForEachCi(p,o,i)o->Value=(k>>i)&1;Gia_ManForEachAnd(p,o,i)o->Value=(Gia_ObjFanin0(o)->Value^Gia_ObjFaninC0(o))&(Gia_ObjFanin1(o)->Value^Gia_ObjFaninC1(o));truth|=(word)(Gia_ManObj(p,Abc_Lit2Var(root))->Value^Abc_LitIsCompl(root))<<k;}printf("%016llx %d %d\n",(unsigned long long)truth,Gia_ObjLevelId(p,Abc_Lit2Var(root)),Gia_ManObjNum(p));Vec_IntFree(cover);Gia_ManStop(p);}return 0;}`
	if err := os.WriteFile(source, []byte(code), 0600); err != nil {
		t.Fatal(err)
	}
	if out, err := exec.Command(cc, "-std=gnu99", "-DLIN64", "-I"+src, "-c", source, "-o", object).CombinedOutput(); err != nil {
		t.Fatalf("compile original graph oracle%v\n%s", err, out)
	}
	args := []string{object}
	err = filepath.WalkDir(src, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() && strings.HasSuffix(path, ".o") && path != filepath.Join(src, "base/main/main.o") {
			args = append(args, path)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	args = append(args, "-lreadline", "-lpthread", "-lm", "-ldl", "-lz", "-o", exe)
	if out, err := exec.Command(cxx, args...).CombinedOutput(); err != nil {
		t.Fatalf("link original graph oracle%v\n%s", err, out)
	}
	expressions := []string{"0", "1", "!a", "(abcdef)", "[abcdef]", "![a(bc)<def>]", "<abc>{(ab)[cd](ef)}", "E8{abc}", "6996{abcd}", "!(a[bc]E8{def})"}
	command := exec.Command(exe)
	command.Stdin = strings.NewReader(strings.Join(expressions, "\n"))
	out, err := command.Output()
	if err != nil {
		t.Fatal(err)
	}
	reader := strings.NewReader(string(out))
	for _, text := range expressions {
		var truth uint64
		var level, objects int
		if _, err := fmt.Fscanf(reader, "%x %d %d\n", &truth, &level, &objects); err != nil {
			t.Fatal(err)
		}
		graph := aig.NewGraph()
		leaves := make([]aig.Edge, 6)
		for i := range leaves {
			leaves[i] = graph.NewInput()
		}
		root := Dau_DsdToGia(graph, text, leaves, nil)
		if truth != Dau_Dsd6ToTruth(text) || graph.Node(root).Level != level || graph.NodeCount() != objects {
			t.Fatalf("source graph %s: truth%x level%d/%d objects%d/%d", text, truth, graph.Node(root).Level, level, graph.NodeCount(), objects)
		}
	}
}
