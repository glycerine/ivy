package ret

import (
	"fmt"
	"github.com/glycerine/abcgo/base/abc"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// Link the actual ABC objects: this exercises the complete original cut manager,
// allocation, enumeration, truth manipulation, hash lookup, and cleanup.
func TestRetimingOriginalC(t *testing.T) {
	cc, err := exec.LookPath("cc")
	if err != nil {
		t.Skip("C compiler unavailable")
	}
	cxx, err := exec.LookPath("g++")
	if err != nil {
		t.Skip("C++ linker unavailable")
	}
	_, file, _, _ := runtime.Caller(0)
	src := filepath.Clean(filepath.Join(filepath.Dir(file), "../../../src"))
	if _, err := os.Stat(filepath.Join(src, "opt/ret/retCore.o")); err != nil {
		t.Skip("original ABC object files unavailable")
	}
	dir := t.TempDir()
	source, object, exe := filepath.Join(dir, "oracle.c"), filepath.Join(dir, "oracle.o"), filepath.Join(dir, "oracle")
	code := `#include "opt/ret/retInt.h"
#include "base/main/main.h"
Abc_Obj_t*latch(Abc_Ntk_t*n,Abc_Obj_t*d,int one){Abc_Obj_t*in=Abc_NtkCreateBi(n),*l=Abc_NtkCreateLatch(n),*out=Abc_NtkCreateBo(n);Abc_ObjAddFanin(in,d);Abc_ObjAddFanin(l,in);Abc_ObjAddFanin(out,l);if(one)Abc_LatchSetInit1(l);else Abc_LatchSetInit0(l);return out;}
Abc_Ntk_t*fixture(int forward){Abc_Ntk_t*n=Abc_NtkAlloc(ABC_NTK_LOGIC,ABC_FUNC_SOP,1);Abc_Obj_t*a=Abc_NtkCreatePi(n),*b=Abc_NtkCreatePi(n),*o,*x,*y;Vec_Ptr_t*v=Vec_PtrAlloc(2);if(forward){x=latch(n,a,1);y=latch(n,b,0);Vec_PtrPush(v,x);Vec_PtrPush(v,y);o=Abc_NtkCreateNodeAnd(n,v);}else{Vec_PtrPush(v,a);Vec_PtrPush(v,b);o=latch(n,Abc_NtkCreateNodeAnd(n,v),1);}Abc_ObjAddFanin(Abc_NtkCreatePo(n),o);Vec_PtrFree(v);Abc_NtkOrderCisCos(n);Abc_NtkAddDummyPiNames(n);Abc_NtkAddDummyPoNames(n);Abc_NtkAddDummyBoxNames(n);return n;}
void dump(Abc_Ntk_t*n,int mode,int dir){Vec_Ptr_t*nodes=Abc_NtkDfs(n,0);Abc_Obj_t*o;int seq,f,i;unsigned long long truth[16]={0};for(seq=0;seq<256;seq++){Abc_NtkForEachLatch(n,o,i)Abc_ObjFanout0(o)->pCopy=(Abc_Obj_t*)(ABC_PTRUINT_T)Abc_LatchIsInit1(o);for(f=0;f<4;f++){Abc_NtkForEachPi(n,o,i)o->pCopy=(Abc_Obj_t*)(ABC_PTRUINT_T)((seq>>(2*f+i))&1);Vec_PtrForEachEntry(Abc_Obj_t*,nodes,o,i)o->pCopy=(Abc_Obj_t*)(ABC_PTRUINT_T)Abc_ObjSopSimulate(o);o=Abc_NtkPo(n,0);if(Abc_ObjFanin0(o)->pCopy)truth[(4*seq+f)/64]|=1ull<<((4*seq+f)%64);Abc_NtkForEachLatch(n,o,i)o->pCopy=Abc_ObjFanin0(Abc_ObjFanin0(o))->pCopy;Abc_NtkForEachLatch(n,o,i)Abc_ObjFanout0(o)->pCopy=o->pCopy;}}printf("RESULT %d %d %d %d",mode,dir,Abc_NtkLatchNum(n),Abc_NtkLevel(n));for(i=0;i<16;i++)printf(" %llx",truth[i]);puts("");Vec_PtrFree(nodes);}
int main(void){int m,d;Abc_Ntk_t*n;Abc_Start();for(m=1;m<=6;m++)for(d=0;d<2;d++){n=fixture(d);Abc_NtkRetime(n,m,0,0,0,0,1,0);dump(n,m,d);Abc_NtkDelete(n);}Abc_Stop();return 0;}
`
	if err := os.WriteFile(source, []byte(code), 0600); err != nil {
		t.Fatal(err)
	}
	if out, err := exec.Command(cc, "-std=gnu99", "-DLIN64", "-I"+src, "-c", source, "-o", object).CombinedOutput(); err != nil {
		t.Fatalf("compile: %v\n%s", err, out)
	}
	args := []string{object}
	err = filepath.WalkDir(src, func(path string, e os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !e.IsDir() && strings.HasSuffix(path, ".o") && path != filepath.Join(src, "base/main/main.o") {
			args = append(args, path)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	args = append(args, "-lreadline", "-lpthread", "-lm", "-ldl", "-lz", "-o", exe)
	if out, err := exec.Command(cxx, args...).CombinedOutput(); err != nil {
		t.Fatalf("link: %v\n%s", err, out)
	}

	command := exec.Command(exe)
	data, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("oracle %v\n%s", err, data)
	}
	count := 0
	for _, line := range strings.Split(string(data), "\n") {
		if !strings.HasPrefix(line, "RESULT ") {
			continue
		}
		reader := strings.NewReader(strings.TrimPrefix(line, "RESULT "))
		var mode, dir, latches, level int32
		if _, err := fmt.Fscan(reader, &mode, &dir, &latches, &level); err != nil {
			t.Fatal(err)
		}
		n := retFixture(dir != 0)
		Abc_NtkRetime(n, mode, 0, 0, 0, 0, 1, 0)
		if int32(len(retLatches(n))) != latches || abc.Abc_NtkLevel(n) != level {
			t.Fatalf("mode%d dir%d dimensions Go%d/%d C%d/%d", mode, dir, len(retLatches(n)), abc.Abc_NtkLevel(n), latches, level)
		}
		var want [16]uint64
		for i := range want {
			if _, err := fmt.Fscanf(reader, "%x", &want[i]); err != nil {
				t.Fatal(err)
			}
		}
		got := retTrace(n)
		if got != want {
			t.Fatalf("mode%d dir%d traces differ", mode, dir)
		}
		count++
	}
	if count != 12 {
		t.Fatalf("oracle returned%d results: %s", count, data)
	}
}
func retTrace(n *abc.Network) [16]uint64 {
	var truth [16]uint64
	pis, pos, latches := n.ObjectsOfKind(abc.ObjectPI), n.ObjectsOfKind(abc.ObjectPO), retLatches(n)
	for sequence := 0; sequence < 256; sequence++ {
		state := make([]bool, len(latches))
		for i, l := range latches {
			state[i] = retLatchValue(l) != 0
		}
		for frame := 0; frame < 4; frame++ {
			inputs := make(map[*abc.Object]bool)
			for i, o := range pis {
				inputs[o] = sequence&(1<<(2*frame+i)) != 0
			}
			for i, l := range latches {
				inputs[l.Fanouts()[0]] = state[i]
			}
			value, err := n.Evaluate(pos[0].Signal(), inputs)
			if err != nil {
				panic(err)
			}
			bit := 4*sequence + frame
			if value {
				truth[bit/64] |= 1 << uint(bit%64)
			}
			next := make([]bool, len(latches))
			for i, l := range latches {
				next[i], err = n.Evaluate(retFan0(l).Signal(), inputs)
				if err != nil {
					panic(err)
				}
			}
			state = next
		}
	}
	return truth
}
