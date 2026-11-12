package scl

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// Compile the unchanged original interpolation routine with a tiny vector ABI.
// The production Go implementation has no C or compiler dependency.
func TestLookupOriginalC(t *testing.T) {
	cc, e := exec.LookPath("cc")
	if e != nil {
		t.Skip("original-C oracle needs cc")
	}
	source, e := os.ReadFile("../../../src/map/scl/sclLib.h")
	if e != nil {
		t.Fatal(e)
	}
	text := string(source)
	start := strings.Index(text, "static inline float Scl_LibLookup(")
	end := strings.Index(text[start:], "static inline void Scl_LibPinArrival(")
	if start < 0 || end < 0 {
		t.Fatal("original routine not found")
	}
	harness := `#include <stdio.h>
#include <assert.h>
typedef struct {float *p;int n;} Vec_Flt_t;
typedef struct {void **p;int n;} Vec_Ptr_t;
typedef struct {Vec_Flt_t vIndex0,vIndex1;Vec_Ptr_t vData;} SC_Surface;
#define Vec_FltSize(v) ((v)->n)
#define Vec_PtrSize(v) ((v)->n)
#define Vec_FltArray(v) ((v)->p)
#define Vec_FltEntry(v,i) ((v)->p[i])
#define Vec_PtrEntry(v,i) ((v)->p[i])
` + text[start:start+end] + `
int main(void) {float x[]={1,3,7},y[]={2,4,9};float a[]={3,9,4},b[]={5,7,11},c[]={8,1,14};Vec_Flt_t rows[]={{a,3},{b,3},{c,3}};void *ptrs[]={rows,rows+1,rows+2};SC_Surface s={{x,3},{y,3},{ptrs,3}};for(int i=-2;i<12;i++)for(int j=-1;j<13;j++)printf("%.9g\n",Scl_LibLookup(&s,(float)i,(float)j));return 0;}`
	dir := t.TempDir()
	path := filepath.Join(dir, "oracle.c")
	if e = os.WriteFile(path, []byte(harness), 0600); e != nil {
		t.Fatal(e)
	}
	binary := filepath.Join(dir, "oracle")
	if output, e := exec.Command(cc, "-std=c99", "-o", binary, path).CombinedOutput(); e != nil {
		t.Fatal(e, string(output))
	}
	output, e := exec.Command(binary).Output()
	if e != nil {
		t.Fatal(e)
	}
	values := strings.Fields(string(output))
	s := SC_Surface{Index0: []float32{1, 3, 7}, Index1: []float32{2, 4, 9}, Data: [][]float32{{3, 9, 4}, {5, 7, 11}, {8, 1, 14}}}
	k := 0
	for i := -2; i < 12; i++ {
		for j := -1; j < 13; j++ {
			want, e := strconv.ParseFloat(values[k], 32)
			if e != nil {
				t.Fatal(e)
			}
			got := Scl_LibLookup(&s, float32(i), float32(j))
			if got != float32(want) {
				t.Fatal(fmt.Sprintf("(%d,%d): Go %g C %g", i, j, got, want))
			}
			k++
		}
	}
}

func TestLibertyReaderOriginalC(t *testing.T) {
	abc, e := filepath.Abs("../../../abc")
	if e != nil {
		t.Fatal(e)
	}
	if _, e = os.Stat(abc); e != nil {
		t.Skip("original ABC executable unavailable")
	}
	// Three distinct functions satisfy the original read_lib command's minimum
	// class count. All have explicit positive tables for genlib extraction.
	text := testLiberty
	start := strings.Index(text, "cell(inv)")
	end := strings.Index(text, "cell(skip)")
	cell := text[start:end]
	text = strings.Replace(text, cell, cell+strings.ReplaceAll(strings.ReplaceAll(cell, "cell(inv)", "cell(buf)"), `"!A"`, `"A"`)+strings.ReplaceAll(strings.ReplaceAll(cell, "cell(inv)", "cell(zero)"), `"!A"`, `"0"`), 1)
	dir := t.TempDir()
	path := filepath.Join(dir, "test.lib")
	out := filepath.Join(dir, "test.scl")
	if e = os.WriteFile(path, []byte(text), 0600); e != nil {
		t.Fatal(e)
	}
	output, e := exec.Command(abc, "-s", "-c", fmt.Sprintf("read_lib %s; write_scl %s", path, out)).CombinedOutput()
	if e != nil {
		t.Fatal(e, string(output))
	}
	data, e := os.ReadFile(out)
	if e != nil {
		t.Fatal(e, string(output))
	}
	original, e := Abc_SclReadFromStr(data)
	if e != nil {
		t.Fatal(e, string(output))
	}
	ported, e := ParseLiberty(path, text)
	if e != nil {
		t.Fatal(e)
	}
	got, e := MarshalSCL(ported)
	if e != nil {
		t.Fatal(e)
	}
	if string(got) != string(data) {
		for i := 0; i < min(len(got), len(data)); i++ {
			if got[i] != data[i] {
				t.Fatalf("original C differs at byte %d: Go=%d C=%d (lengths %d/%d, cells %d/%d)", i, got[i], data[i], len(got), len(data), len(ported.Cells), len(original.Cells))
			}
		}
		t.Fatal("SCL lengths", len(got), len(data))
	}
}

func TestDownsizeOriginalC(t *testing.T) {
	abc, e := filepath.Abs("../../../abc")
	if e != nil {
		t.Fatal(e)
	}
	if _, e = os.Stat(abc); e != nil {
		t.Skip("original ABC executable unavailable")
	}
	lib, _ := downsizeFixture(t)
	inv := *lib.Cells[0]
	inv.Name = "INV"
	inv.ID = 2
	inv.Pins = append([]*SC_Pin(nil), inv.Pins...)
	out := *inv.Pins[1]
	out.Function = "!A"
	out.Truth = []uint64{0x5555555555555555}
	inv.Pins[1] = &out
	zero := &SC_Cell{Name: "ZERO", ID: 3, Outputs: 1, Pins: []*SC_Pin{{Name: "Y", Direction: ScDirOutput, Function: "0", Truth: []uint64{0}}}}
	lib.Cells = append(lib.Cells, &inv, zero)
	Abc_SclHashCells(lib)
	Abc_SclLinkCells(lib)
	text, e := LibertyText(lib)
	if e != nil {
		t.Fatal(e)
	}
	dir := t.TempDir()
	libpath := filepath.Join(dir, "test.lib")
	genpath := filepath.Join(dir, "test.genlib")
	netpath := filepath.Join(dir, "test.blif")
	outpath := filepath.Join(dir, "out.blif")
	for path, data := range map[string]string{libpath: text, genpath: Abc_SclProduceGenlibStrSimple(lib), netpath: ".model chain\n.inputs a\n.outputs y\n.gate large A=a Y=n\n.gate large A=n Y=y\n.end\n"} {
		if e = os.WriteFile(path, []byte(data), 0600); e != nil {
			t.Fatal(e)
		}
	}
	command := fmt.Sprintf("read_lib %s; read_genlib %s; read_blif %s; dnsize -I 2 -N 4 -G 1; write_blif %s", libpath, genpath, netpath, outpath)
	output, e := exec.Command(abc, "-s", "-c", command).CombinedOutput()
	if e != nil {
		t.Fatal(e, string(output))
	}
	result, e := os.ReadFile(outpath)
	if e != nil {
		t.Fatal(e, string(output))
	}
	if strings.Count(string(result), ".gate small") != 2 {
		t.Fatal("original downsizing did not select both small cells", string(result), string(output))
	}
}
