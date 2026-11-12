package dau

import (
	"encoding/binary"
	"fmt"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

func TestNPNClassTables(t *testing.T) {
	for _, test := range []struct {
		vars    int32
		classes int
	}{{1, 2}, {2, 4}, {3, 14}, {4, 222}} {
		table, reps, err := EnumerateNPN(test.vars)
		if err != nil {
			t.Fatal(err)
		}
		if len(reps) != test.classes {
			t.Fatalf("vars%d classes%d want%d", test.vars, len(reps), test.classes)
		}
		for i, rep := range table {
			if int(rep) > i || table[rep] != rep {
				t.Fatalf("vars%d nonminimal/idempotent rep at%d", test.vars, i)
			}
		}
	}
}
func TestNPNSymmetryIndependent(t *testing.T) {
	rng := rand.New(rand.NewSource(842))
	for n := int32(1); n <= 6; n++ {
		for sample := 0; sample < 80; sample++ {
			raw := rng.Uint64()
			if n < 6 {
				mask := uint64(1)<<uint(1<<n) - 1
				raw &= mask
				for shift := 1 << n; shift < 64; shift *= 2 {
					raw |= raw << uint(shift)
				}
			}
			count, mask := int32(0), int32(0)
			for i := int32(0); i < n; i++ {
				for j := i + 1; j < n; j++ {
					symmetric := true
					for assignment := 0; assignment < 1<<n; assignment++ {
						other := assignment
						if assignment>>i&1 != assignment>>j&1 {
							other ^= 1<<i | 1<<j
						}
						if raw>>uint(assignment)&1 != raw>>uint(other)&1 {
							symmetric = false
						}
					}
					if symmetric {
						count++
						mask |= 1 << j
					}
				}
			}
			if Dau_CountSymms(raw, n) != count || Dau_CountSymms2(raw, n) != mask {
				t.Fatalf("vars%d symmetry", n)
			}
			expected := int32(0)
			for v := int32(0); v < n; v++ {
				flipped := util.Abc_Tt6Flip(raw, v)
				found := flipped == ^raw
				for k := int32(0); k < n; k++ {
					if k != v && flipped == util.Abc_Tt6Flip(raw, k) {
						found = true
					}
				}
				if found {
					expected++
				}
			}
			if Dau_CountCompl(raw, n) != expected {
				t.Fatal("complement symmetry")
			}
		}
	}
}
func TestNPNOriginalCClassesAndSymmetry(t *testing.T) {
	cc, err := exec.LookPath("cc")
	if err != nil {
		t.Skip("original C oracle requires compiler")
	}
	_, file, _, _ := runtime.Caller(0)
	sourceRoot := filepath.Clean(filepath.Join(filepath.Dir(file), "../../../src"))
	dir := t.TempDir()
	source, exe := filepath.Join(dir, "oracle.c"), filepath.Join(dir, "oracle")
	code := `#include "opt/dau/dauNpn.c"
unsigned enable_dbg_outs=1;
int Abc_FrameIsBridgeMode(void){return 0;}
int Gia_ManToBridgeText(FILE*f,int n,unsigned char*b){return (int)fwrite(b,1,n,f);}
char *vnsprintf(const char*f,va_list ap){va_list cp;int n;char*r;va_copy(cp,ap);n=vsnprintf(NULL,0,f,cp);va_end(cp);r=(char*)malloc(n+1);vsnprintf(r,n+1,f,ap);return r;}
int main(int argc,char**argv){int n;unsigned long long t;if(argc>1){if(strcmp(argv[1],"network")==0)Dau_NetworkEnum(4);else Dau_TruthEnum(4);return 0;}while(scanf("%d%llx",&n,&t)==2)printf("%d %d %d\n",Dau_CountSymms(t,n),Dau_CountSymms2(t,n),Dau_CountCompl(t,n));return 0;}`
	if err := os.WriteFile(source, []byte(code), 0600); err != nil {
		t.Fatal(err)
	}
	if out, err := exec.Command(cc, "-std=gnu99", "-DLIN64", "-ffunction-sections", "-fdata-sections", "-I"+sourceRoot, source, filepath.Join(sourceRoot, "misc/extra/extraUtilMisc.c"), "-Wl,--gc-sections", "-o", exe).CombinedOutput(); err != nil {
		t.Fatalf("C oracle compile %v\n%s", err, out)
	}
	command := exec.Command(exe, "table")
	command.Dir = dir
	if out, err := command.CombinedOutput(); err != nil {
		t.Fatalf("C table oracle %v\n%s", err, out)
	}
	data, err := os.ReadFile(filepath.Join(dir, "tableW14.data"))
	if err != nil {
		t.Fatal(err)
	}
	table, _, err := EnumerateNPN(4)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) != len(table)*4 {
		t.Fatal("C table binary length")
	}
	for i, entry := range table {
		if entry != binary.LittleEndian.Uint32(data[4*i:]) {
			t.Fatalf("original C class table differs at%x", i)
		}
	}
	command = exec.Command(exe, "network")
	command.Dir = dir
	networkOut, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("C network oracle %v\n%s", err, networkOut)
	}
	levels, partial := EnumerateNetworks(4, append([]uint32(nil), table...))
	lines := regexp.MustCompile(`Nodes =\s*(\d+)\.\s+New =\s*(\d+)\. Total =\s*(\d+)\.\s+New =\s*(\d+)`).FindAllStringSubmatch(string(networkOut), -1)
	if len(lines) != len(levels) {
		t.Fatalf("network levels%d want%d\n%s", len(levels), len(lines), networkOut)
	}
	for i, line := range lines {
		count, _ := strconv.Atoi(line[2])
		countPartial, _ := strconv.Atoi(line[4])
		if len(levels[i]) != count || len(partial[i]) != countPartial {
			t.Fatalf("network level%d new%d/%d partial%d/%d", i, len(levels[i]), count, len(partial[i]), countPartial)
		}
	}
	var request strings.Builder
	for i := uint64(0); i < 256; i++ {
		fmt.Fprintf(&request, "3 %x\n", i*0x0101010101010101)
	}
	command = exec.Command(exe)
	command.Stdin = strings.NewReader(request.String())
	out, err := command.Output()
	if err != nil {
		t.Fatal(err)
	}
	reader := strings.NewReader(string(out))
	for i := uint64(0); i < 256; i++ {
		var a, b, c int32
		if _, err := fmt.Fscan(reader, &a, &b, &c); err != nil {
			t.Fatal(err)
		}
		truth := i * 0x0101010101010101
		if a != Dau_CountSymms(truth, 3) || b != Dau_CountSymms2(truth, 3) || c != Dau_CountCompl(truth, 3) {
			t.Fatalf("original C symmetry truth%x", i)
		}
	}
}

func TestNPNFileTableAndClassDeduplication(t *testing.T) {
	path := filepath.Join(t.TempDir(), "table.data")
	if err := os.WriteFile(path, []byte{1, 2, 3, 4, 5, 6}, 0600); err != nil {
		t.Fatal(err)
	}
	words := Dau_ReadFile(path, 2)
	if len(words) != 4 || words[0] != 0x04030201 || words[1] != 0x0605 || words[2] != 0 || words[3] != 0 {
		t.Fatal("short table zero fill", words)
	}
	table, _, err := EnumerateNPN(3)
	if err != nil {
		t.Fatal(err)
	}
	classes, partial := vec.Vec_IntAlloc(4), vec.Vec_IntAlloc(4)
	truth := uint64(0x8888888888888888)
	first := Dau_AddFunction(truth, 3, table, classes, partial)
	if first == 0 || vec.Vec_IntSize(classes) != 1 || vec.Vec_IntSize(partial) != 1 {
		t.Fatal("new support-deficient class not recorded")
	}
	Dau_AddFunction(^truth, 3, table, classes, partial)
	if vec.Vec_IntSize(classes) != 1 || vec.Vec_IntSize(partial) != 1 {
		t.Fatal("class duplicate/output phase not suppressed")
	}
}
