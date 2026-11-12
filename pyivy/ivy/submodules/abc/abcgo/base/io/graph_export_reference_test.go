package io

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestOriginalCGraphExports(t *testing.T) {
	source := os.Getenv("GRAPH_SOURCE_ROOT")
	if source == "" {
		t.Skip("set GRAPH_SOURCE_ROOT to original src directory")
	}
	cc, e := exec.LookPath("cc")
	if e != nil {
		t.Skip("C compiler unavailable")
	}
	cxx, e := exec.LookPath("g++")
	if e != nil {
		t.Skip("C++ linker unavailable")
	}
	if _, e = os.Stat(filepath.Join(source, "base/io/ioWriteGml.o")); e != nil {
		t.Skip("original ABC object files unavailable")
	}
	dir := t.TempDir()
	c, o, exe := filepath.Join(dir, "oracle.c"), filepath.Join(dir, "oracle.o"), filepath.Join(dir, "oracle")
	code := `#include "base/io/ioWriteBook.c"
int main(int argc,char**argv){char path[4096];char*inputs[]={"a","b"};Abc_Obj_t*a,*b,*g,*z;Abc_Ntk_t*n=Abc_NtkAlloc(ABC_NTK_STRASH,ABC_FUNC_AIG,1),*book;FILE*f;
n->pName=Abc_UtilStrsav("graph");a=Abc_NtkCreatePi(n);Abc_ObjAssignName(a,"a",0);b=Abc_NtkCreatePi(n);Abc_ObjAssignName(b,"b",0);g=Abc_AigAnd((Abc_Aig_t*)n->pManFunc,a,Abc_ObjNot(b));Abc_ObjAssignName(g,"and",0);z=Abc_NtkCreatePo(n);Abc_ObjAssignName(z,"z",0);Abc_ObjAddFanin(z,g);sprintf(path,"%s/gml",argv[1]);Io_WriteGml(n,path);sprintf(path,"%s/list",argv[1]);Io_WriteList(n,path,1);
book=Abc_NtkStartRead("book");Io_ReadCreatePi(book,"a");Io_ReadCreatePi(book,"b");Io_ReadCreatePo(book,"z");g=Io_ReadCreateNode(book,"z",inputs,2);g->pData=Abc_SopCreateAnd2((Mem_Flex_t*)book->pManFunc,0,1);sprintf(path,"%s/nodes",argv[1]);f=fopen(path,"w");Io_NtkWriteNodes(f,book);fclose(f);sprintf(path,"%s/nets",argv[1]);f=fopen(path,"w");Io_NtkWriteNets(f,book);fclose(f);return 0;}`
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
		if !d.IsDir() && strings.HasSuffix(p, ".o") && p != filepath.Join(source, "base/main/main.o") && p != filepath.Join(source, "base/io/ioWriteBook.o") {
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
	if out, e := exec.Command(exe, dir).CombinedOutput(); e != nil {
		t.Fatalf("oracle %v %s", e, out)
	}
	n := exportFixture()
	var gml, list bytes.Buffer
	WriteGML(&gml, n)
	WriteList(&list, n, true)
	data, e := os.ReadFile(filepath.Join(dir, "gml"))
	if e != nil {
		t.Fatal(e)
	}
	nodes := regexp.MustCompile(`node \[ id\s+(\d+) label "([^"]*)"`)
	edges := regexp.MustCompile(`edge \[ source\s+(\d+)\s+target\s+(\d+)`)
	if fmt.Sprint(nodes.FindAllStringSubmatch(gml.String(), -1)) != fmt.Sprint(nodes.FindAllStringSubmatch(string(data), -1)) || fmt.Sprint(edges.FindAllStringSubmatch(gml.String(), -1)) != fmt.Sprint(edges.FindAllStringSubmatch(string(data), -1)) {
		t.Fatalf("GML graph differs\nGo%s\nC%s", gml.String(), data)
	}
	data, e = os.ReadFile(filepath.Join(dir, "list"))
	if e != nil {
		t.Fatal(e)
	}
	strip := func(s string) string {
		var lines []string
		for _, l := range strings.Split(s, "\n") {
			if !strings.HasPrefix(l, "#") {
				lines = append(lines, l)
			}
		}
		return strings.Join(lines, "\n")
	}
	if strip(string(data)) != strip(list.String()) {
		t.Fatalf("list differs\nGo%s\nC%s", list.String(), data)
	}
	for _, name := range []string{"nodes", "nets"} {
		data, e = os.ReadFile(filepath.Join(dir, name))
		if e != nil {
			t.Fatal(e)
		}
		want := []string{"NumNodes : 4", "NumTerminals : 3", " 3 1"}
		if name == "nets" {
			want = []string{"NumNets : 3", "NumPins : 6"}
		}
		for _, part := range want {
			if !strings.Contains(string(data), part) {
				t.Fatalf("original Book %s missing %s", name, part)
			}
		}
	}
}
