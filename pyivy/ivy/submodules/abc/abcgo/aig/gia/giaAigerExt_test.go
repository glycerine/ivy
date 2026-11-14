package gia

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"github.com/glycerine/abcgo/misc/vec"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func aigerTestManager(n int) *Gia_Man_t {
	p := Gia_ManStart(int32(n))
	for i := 1; i < n; i++ {
		Gia_ManAppendCi(p)
	}
	return p
}
func TestAigerEquivalenceGolden(t *testing.T) {
	p := aigerTestManager(9)
	p.PReprs = make([]Gia_Rpr_t, 9)
	for i := range p.PReprs {
		p.PReprs[i].iRepr = GIA_VOID
	}
	p.PReprs[1] = Gia_Rpr_t{iRepr: 0, fProved: 1}
	p.PReprs[3] = Gia_Rpr_t{iRepr: 2}
	p.PReprs[4] = Gia_Rpr_t{iRepr: 2, fProved: 1}
	p.PReprs[8] = Gia_Rpr_t{iRepr: 2, fProved: 1}
	data, err := Gia_WriteEquivClassesInt(p, nil)
	if err != nil {
		t.Fatal(err)
	}
	want := []byte{0, 0, 0, 10, 1, 6, 5, 4, 6, 18}
	if !bytes.Equal(data, want) {
		t.Fatal(data, want)
	}
	position := append(data, 99)
	decoded, err := Gia_AigerReadEquivClasses(&position, 9)
	if err != nil || !reflect.DeepEqual(decoded, p.PReprs) || !bytes.Equal(position, []byte{99}) {
		t.Fatal(err, decoded, position)
	}
	for _, bad := range [][]byte{{}, {0, 0, 0, 3}, {0, 0, 0, 6, 1}, {0, 0, 0, 5, 0}, {0, 0, 0, 6, 1, 0}, {0, 0, 0, 6, 19, 4}} {
		position := bad
		if _, err := Gia_AigerReadEquivClasses(&position, 9); err == nil {
			t.Fatal("invalid equivalence accepted", bad)
		}
	}
}
func TestAigerMappingFormats(t *testing.T) {
	p := aigerTestManager(9)
	mapping, err := aigerMakeMapping(9, []aigerLUT{{5, []int32{1, 3}}, {8, []int32{2, 4, 5}}})
	if err != nil {
		t.Fatal(err)
	}
	p.VMapping = aigerVecInts(mapping)
	compressed, err := Gia_AigerWriteMappingInt(p, nil)
	if err != nil {
		t.Fatal(err)
	}
	want := []byte{0, 0, 0, 13, 2, 3, 5, 5, 3, 6, 5, 3, 7}
	if !bytes.Equal(compressed, want) {
		t.Fatal(compressed, want)
	}
	position := append(compressed, 77)
	decoded, err := Gia_AigerReadMapping(&position, 9)
	if err != nil || !reflect.DeepEqual(decoded, mapping) || !bytes.Equal(position, []byte{77}) {
		t.Fatal(err, decoded, position)
	}
	doc, err := Gia_AigerWriteMappingDoc(p)
	if err != nil {
		t.Fatal(err)
	}
	position = append(append([]byte{}, vec.Vec_StrArray(doc)...), 77)
	docMap, err := Gia_AigerReadMappingDoc(&position, 9)
	if err != nil || !reflect.DeepEqual(vec.Vec_IntArray(docMap), mapping) || !bytes.Equal(position, []byte{77}) {
		t.Fatal(err, position)
	}
	raw, err := Gia_AigerWriteMappingSimple(p)
	if err != nil {
		t.Fatal(err)
	}
	position = append(append([]byte{}, vec.Vec_StrArray(raw)...), 77)
	simple, err := Gia_AigerReadMappingSimple(&position, int32(len(mapping)*4))
	if err != nil || !reflect.DeepEqual(simple, mapping) || !bytes.Equal(position, []byte{77}) {
		t.Fatal(err, simple, position)
	}
	bad := append([]byte{}, compressed...)
	bad[len(bad)-1] = 127
	position = bad
	if _, err := Gia_AigerReadMapping(&position, 9); err == nil {
		t.Fatal("invalid mapping root accepted")
	}
}
func TestAigerPackingAndLiterals(t *testing.T) {
	packing := aigerVecInts([]int32{2, 1, 5, 3, 2, 4, 8})
	encoded := Gia_WritePacking(packing)
	position := append(append([]byte{}, vec.Vec_StrArray(encoded)...), 9)
	decoded, err := Gia_AigerReadPacking(&position, int32(vec.Vec_StrSize(encoded)))
	if err != nil || !reflect.DeepEqual(vec.Vec_IntArray(decoded), vec.Vec_IntArray(packing)) || !bytes.Equal(position, []byte{9}) {
		t.Fatal(err, position)
	}
	for _, values := range [][]int32{nil, {0}, {0, 2, 1, 1000, 0, 2147483647, 0}, {10, 10, 10, 2, 3}} {
		encoded, err := Gia_AigerWriteLiterals(aigerVecInts(values))
		if err != nil {
			t.Fatal(err)
		}
		data := append(append([]byte{}, vec.Vec_StrArray(encoded)...), 99)
		decoded, err := Gia_AigerReadLiterals(&data, int32(len(values)))
		if err != nil || !reflect.DeepEqual(vec.Vec_IntArray(decoded), values) && len(values) != 0 || !bytes.Equal(data, []byte{99}) {
			t.Fatal(err, values, vec.Vec_IntArray(decoded), data)
		}
	}
	var data [5]byte
	positionSize, err := Gia_AigerWriteDiffValue(data[:], 0, 5, 2)
	if err != nil || positionSize != 1 || data[0] != 6 {
		t.Fatal(err, data)
	}
	position = data[:positionSize]
	if value, err := Gia_AigerReadDiffValue(&position, 5); err != nil || value != 2 {
		t.Fatal(value, err)
	}
	if _, err := Gia_AigerWriteDiffValue(nil, 0, 1, 2); err == nil {
		t.Fatal("short buffer accepted")
	}
	position = []byte{0, 0, 0, 0, 0}
	if _, err := Gia_AigerReadMappingSimple(&position, 3); err == nil {
		t.Fatal("unaligned mapping accepted")
	}
	if _, err := Gia_AigerReadPacking(&position, 8); err == nil {
		t.Fatal("short packing accepted")
	}
}
func TestAigerMalformedMappingDoc(t *testing.T) {
	for _, words := range [][]uint32{{}, {1, 2}, {1, 2, 99, 0}, {2, 0, 1, 0, 1, 0}, {1, 0, 1, 1, 2}} {
		var data []byte
		for _, word := range words {
			data = binary.BigEndian.AppendUint32(data, word)
		}
		position := data
		if _, err := Gia_AigerReadMappingDoc(&position, 4); err == nil {
			t.Fatal("invalid mapping accepted", words)
		}
	}
}

func TestAigerOriginalCExtensionDecoders(t *testing.T) {
	cc, err := exec.LookPath("cc")
	if err != nil {
		t.Skip("C compiler unavailable")
	}
	source, err := os.ReadFile("../../../src/aig/gia/giaAigerExt.c")
	if err != nil {
		t.Fatal(err)
	}
	text := string(source)
	extract := func(marker string) string {
		start := strings.Index(text, marker)
		if start < 0 {
			t.Fatal(marker)
		}
		open := start + strings.IndexByte(text[start:], '{')
		depth, end := 1, open+1
		for depth > 0 {
			if text[end] == '{' {
				depth++
			}
			if text[end] == '}' {
				depth--
			}
			end++
		}
		return text[start:end]
	}
	code := `#include <stdio.h>
#include <stdlib.h>
#include <assert.h>
#define ABC_CALLOC(T,n) ((T*)calloc(n,sizeof(T)))
#define GIA_VOID 0xfffffff
#define ABC_NAMESPACE_IMPL_START
#define ABC_NAMESPACE_IMPL_END
typedef struct { unsigned iRepr:28,fProved:1,fFailed:1,fColorA:1,fColorB:1; } Gia_Rpr_t;
int Gia_AigerReadInt(unsigned char *p){return (unsigned)p[0]<<24|(unsigned)p[1]<<16|(unsigned)p[2]<<8|p[3];}
unsigned Gia_AigerReadUnsigned(unsigned char **p){unsigned x=0,s=0;do{x|=(unsigned)(**p&127)<<s;s+=7;}while(*(*p)++&128);return x;}
` + extract("Gia_Rpr_t * Gia_AigerReadEquivClasses(") + "\n" + extract("static inline unsigned Gia_AigerReadDiffValue(") + "\n" + extract("int * Gia_AigerReadMapping(") + `
int main(int argc,char **argv){int kind=atoi(argv[1]),n=atoi(argv[2]);unsigned char data[65536],*p=data;int k=0,v;
while(scanf("%2x",&v)==1){data[k++]=v;}
if(kind==0){Gia_Rpr_t *r=Gia_AigerReadEquivClasses(&p,n);for(int i=0;i<n;i++)printf("%u:%u ",r[i].iRepr,r[i].fProved);free(r);}
else{int *map=Gia_AigerReadMapping(&p,n),size=n;for(int i=0;i<n;i++)if(map[i]&&map[i]+map[map[i]]+2>size)size=map[i]+map[map[i]]+2;for(int i=0;i<size;i++)printf("%d ",map[i]);free(map);}
printf("used=%d",(int)(p-data));return 0;}
`
	dir := t.TempDir()
	path := filepath.Join(dir, "reference.c")
	binary := filepath.Join(dir, "reference")
	if err := os.WriteFile(path, []byte(code), 0600); err != nil {
		t.Fatal(err)
	}
	if out, err := exec.Command(cc, path, "-o", binary).CombinedOutput(); err != nil {
		t.Fatalf("C oracle: %v: %s", err, out)
	}
	p := aigerTestManager(20)
	p.PReprs = make([]Gia_Rpr_t, p.nObjs)
	for i := range p.PReprs {
		p.PReprs[i].iRepr = GIA_VOID
	}
	for i := 2; i < 20; i++ {
		p.PReprs[i] = Gia_Rpr_t{iRepr: uint32(i % 2), fProved: uint8(i % 3 & 1)}
	}
	equivalences, err := Gia_WriteEquivClassesInt(p, nil)
	if err != nil {
		t.Fatal(err)
	}
	var expected strings.Builder
	for _, r := range p.PReprs {
		fmt.Fprintf(&expected, "%d:%d ", r.iRepr, r.fProved)
	}
	fmt.Fprintf(&expected, "used=%d", len(equivalences))
	cmd := exec.Command(binary, "0", "20")
	cmd.Stdin = strings.NewReader(hex.EncodeToString(equivalences))
	out, err := cmd.CombinedOutput()
	if err != nil || string(out) != expected.String() {
		t.Fatal(err, string(out), expected.String())
	}
	mapping, _ := aigerMakeMapping(20, []aigerLUT{{10, []int32{1, 3, 5}}, {15, []int32{3, 9, 10}}, {19, []int32{10, 15}}})
	p.VMapping = aigerVecInts(mapping)
	encoded, err := Gia_AigerWriteMappingInt(p, nil)
	if err != nil {
		t.Fatal(err)
	}
	expected.Reset()
	for _, value := range mapping {
		fmt.Fprintf(&expected, "%d ", value)
	}
	fmt.Fprintf(&expected, "used=%d", len(encoded))
	cmd = exec.Command(binary, "1", "20")
	cmd.Stdin = strings.NewReader(hex.EncodeToString(encoded))
	out, err = cmd.CombinedOutput()
	if err != nil || string(out) != expected.String() {
		t.Fatal(err, string(out), expected.String())
	}
}
