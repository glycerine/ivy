package cec

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

func TestPatternClassCOracle(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("ELF section collection required")
	}
	cc, err := exec.LookPath("cc")
	if err != nil {
		t.Skip("optional original-C oracle requires cc")
	}
	source := os.Getenv("ABC_SOURCE_ROOT")
	if source == "" {
		_, file, _, _ := runtime.Caller(0)
		source = filepath.Clean(filepath.Join(filepath.Dir(file), "../../../src"))
	}
	dir := t.TempDir()
	main := filepath.Join(dir, "main.c")
	binary := filepath.Join(dir, "oracle")
	code := `#include "proof/cec/cecClass.c"
#include "proof/cec/cecPat.c"
#include "proof/cec/cecSim.c"
int main(void) {
 char mode;
 while (scanf(" %c", &mode)==1) {
  if(mode=='W') {
   int n; scanf("%d",&n); unsigned a[32],b[32]; int sa[1024]={0},sb[1024]={0};
   for(int i=0;i<n;i++) scanf("%x",a+i); for(int i=0;i<n;i++) scanf("%x",b+i);
   printf("%d %d %d %d %d",Cec_ManSimCompareConst(a,n),Cec_ManSimCompareEqual(a,b,n),Cec_ManSimCompareConstFirstBit(a,n),Cec_ManSimCompareEqualFirstBit(a,b,n),Cec_ManSimHashKey(a,n,1009));
   Cec_ManSimCompareConstScore(a,n,sa); Cec_ManSimCompareEqualScore(a,b,n,sb);
   for(int i=0;i<32*n;i++) printf(" %d %d",sa[i],sb[i]); puts("");
  } else if(mode=='N') {
   int n,x; scanf("%d",&n); Cec_ManPat_t p={0}; p.vStorage=Vec_StrAlloc(0);Vec_Int_t *v=Vec_IntAlloc(n);
   for(int i=0;i<n;i++){scanf("%d",&x);Vec_IntPush(v,x);} Cec_ManPatStore(&p,v);
   printf("%d",Vec_StrSize(p.vStorage));for(int i=0;i<Vec_StrSize(p.vStorage);i++)printf(" %u",(unsigned char)Vec_StrEntry(p.vStorage,i));puts("");Vec_StrFree(p.vStorage);Vec_IntFree(v);
  } else if(mode=='P') {
   int ni,nr,nw,n,x;scanf("%d%d%d%d",&ni,&nr,&nw,&n);Vec_Int_t *v=Vec_IntAlloc(n);for(int i=0;i<n;i++){scanf("%d",&x);Vec_IntPush(v,x);}Gia_ManRandom(1);Vec_Ptr_t *p=Cec_ManPatPackPatterns(v,ni,nr,nw);int words=Vec_PtrReadWordsSimInfo(p);printf("%d",words);for(int i=0;i<ni;i++)for(int j=0;j<words;j++)printf(" %u",((unsigned *)Vec_PtrEntry(p,i))[j]);puts("");Vec_PtrFree(p);Vec_IntFree(v);
  } else if(mode=='C') {
   int n,nw;scanf("%d%d",&n,&nw);Gia_Man_t g={0};g.nObjs=n+1;g.pReprs=calloc(n+1,sizeof(Gia_Rpr_t));g.pNexts=calloc(n+1,sizeof(int));Cec_ManSim_t p={0};p.pAig=&g;p.nWords=nw;p.pSimInfo=calloc(n+1,sizeof(int));p.pMems=calloc((n+2)*(nw+1),sizeof(unsigned));p.vClassOld=Vec_IntAlloc(n);p.vClassNew=Vec_IntAlloc(n);Vec_Int_t *v=Vec_IntAlloc(n);
   for(int i=1;i<=n;i++){Vec_IntPush(v,i);p.pSimInfo[i]=i*(nw+1);for(int j=0;j<nw;j++)scanf("%x",p.pMems+p.pSimInfo[i]+1+j);}Cec_ManSimClassCreate(&g,v);printf("%d",Cec_ManSimClassRefineOne(&p,1));for(int i=1;i<=n;i++)printf(" %d %d",Gia_ObjRepr(&g,i),Gia_ObjNext(&g,i));puts("");free(g.pReprs);free(g.pNexts);free(p.pSimInfo);free(p.pMems);Vec_IntFree(p.vClassOld);Vec_IntFree(p.vClassNew);Vec_IntFree(v);
  } else if(mode=='S'||mode=='I') {
   int nw,c0,c1;scanf("%d%d%d",&nw,&c0,&c1);Gia_Man_t g={0};g.nObjs=4;g.pObjs=calloc(4,sizeof(Gia_Obj_t));g.pObjs[0].iDiff0=g.pObjs[0].iDiff1=GIA_NONE;for(int i=1;i<3;i++){g.pObjs[i].fTerm=1;g.pObjs[i].iDiff0=GIA_NONE;g.pObjs[i].iDiff1=i-1;}g.pObjs[3].iDiff0=2;g.pObjs[3].iDiff1=1;g.pObjs[3].fCompl0=c0;g.pObjs[3].fCompl1=c1;
   Cec_ManS_t p={0};p.pAig=&g;p.nWords=nw;p.vSims=Vec_WrdStart(8*nw);p.pTemp[0]=calloc(4*nw,sizeof(word));for(int i=1;i<4;i++)p.pTemp[i]=p.pTemp[i-1]+nw;
   unsigned long long x;for(int i=0;i<SIM_RANDS;i++){scanf("%llx",&x);p.Rands[i]=x;}for(int i=0;i<8*nw;i++){scanf("%llx",&x);Vec_WrdWriteEntry(p.vSims,i,x);}
   int status=-1;if(mode=='S')status=Cec_ManSRunPropagate(&p,3);else Cec_ManSRunImply(&p,3);printf("%d %d %d %d",status,p.iRand,p.nSkipped,p.nVisited);for(int i=0;i<8*nw;i++)printf(" %llu",(unsigned long long)Vec_WrdEntry(p.vSims,i));puts("");Vec_WrdFree(p.vSims);free(p.pTemp[0]);free(g.pObjs);
  } else return 3;
 }
 return 0;
}`
	if err := os.WriteFile(main, []byte(code), 0644); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(cc, "-O2", "-DABC_USE_STDINT_H", "-ffunction-sections", "-fdata-sections", "-Wl,--gc-sections", "-I", source, main, filepath.Join(source, "aig/gia/giaUtil.c"), "-o", binary)
	if b, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build original C oracle: %v\n%s", err, b)
	}
	rng := rand.New(rand.NewSource(130))
	var input bytes.Buffer
	var expected [][]int64
	for trial := 0; trial < 80; trial++ {
		n := 1 + trial%20
		a, b := make([]uint32, n), make([]uint32, n)
		for i := range a {
			a[i] = rng.Uint32()
			b[i] = rng.Uint32()
		}
		if trial%3 == 0 {
			copy(b, a)
		}
		if trial%5 == 0 {
			for i := range b {
				b[i] = ^a[i]
			}
		}
		fmt.Fprint(&input, "W ", n)
		for _, v := range append(append([]uint32(nil), a...), b...) {
			fmt.Fprintf(&input, " %x", v)
		}
		fmt.Fprintln(&input)
		want := []int64{int64(Cec_ManSimCompareConst(a, int32(n))), int64(Cec_ManSimCompareEqual(a, b, int32(n))), int64(Cec_ManSimCompareConstFirstBit(a, int32(n))), int64(Cec_ManSimCompareEqualFirstBit(a, b, int32(n))), int64(Cec_ManSimHashKey(a, int32(n), 1009))}
		sa, sb := make([]int32, 32*n), make([]int32, 32*n)
		Cec_ManSimCompareConstScore(a, int32(n), sa)
		Cec_ManSimCompareEqualScore(a, b, int32(n), sb)
		for i := range sa {
			want = append(want, int64(sa[i]), int64(sb[i]))
		}
		expected = append(expected, want)
	}
	for _, a := range [][]int32{{0}, {0, 1, 127, 128, 129, 16384, 2147483647}, {12, 25, 23456, 2000000000}} {
		p := Cec_ManPatStart()
		Cec_ManPatStore(p, cecTestVec(a...))
		fmt.Fprint(&input, "N ", len(a))
		for _, v := range a {
			fmt.Fprint(&input, " ", v)
		}
		fmt.Fprintln(&input)
		want := []int64{int64(vec.Vec_StrSize(p.vStorage))}
		for _, b := range vec.Vec_StrArray(p.vStorage) {
			want = append(want, int64(b))
		}
		expected = append(expected, want)
	}
	for ni := 2; ni <= 6; ni++ {
		for nw := 1; nw <= 2; nw++ {
			store := cecTestVec()
			for a := 0; a < 1<<uint(ni); a++ {
				vec.Vec_IntPush(store, 0)
				vec.Vec_IntPush(store, int32(ni))
				for i := 0; i < ni; i++ {
					l := int32(2 * i)
					if a&(1<<uint(i)) == 0 {
						l++
					}
					vec.Vec_IntPush(store, l)
				}
			}
			fmt.Fprintf(&input, "P %d 0 %d %d", ni, nw, vec.Vec_IntSize(store))
			for _, v := range vec.Vec_IntArray(store) {
				fmt.Fprint(&input, " ", v)
			}
			fmt.Fprintln(&input)
			gia.Gia_ManRandom(1)
			packed := Cec_ManPatPackPatterns(store, int32(ni), 0, int32(nw))
			words := len(vec.Vec_PtrEntry(packed, 0).([]uint32))
			want := []int64{int64(words)}
			for _, row := range vec.Vec_PtrArray(packed) {
				for _, w := range row.([]uint32) {
					want = append(want, int64(w))
				}
			}
			expected = append(expected, want)
		}
	}
	for trial := 0; trial < 40; trial++ {
		n, nw := 9, 1+trial%4
		g := gia.Gia_ManStart(20)
		for i := 0; i < n; i++ {
			gia.Gia_ManAppendCi(g)
		}
		g.StartRepresentatives()
		p := Cec_ManSimStart(g, &Cec_ParSim_t{NWords: int32(nw)})
		p.pMems = make([]uint32, (n+2)*(nw+1))
		fmt.Fprintf(&input, "C %d %d", n, nw)
		class := cecTestVec()
		for i := 1; i <= n; i++ {
			vec.Vec_IntPush(class, int32(i))
			off := i * (nw + 1)
			p.pSimInfo[i] = int32(off)
			for j := 0; j < nw; j++ {
				x := rng.Uint32()
				if i > 1 && i%3 == 0 {
					x = ^p.pMems[(i-1)*(nw+1)+1+j]
				}
				p.pMems[off+1+j] = x
				fmt.Fprintf(&input, " %x", x)
			}
		}
		fmt.Fprintln(&input)
		Cec_ManSimClassCreate(g, class)
		want := []int64{int64(Cec_ManSimClassRefineOne(p, 1))}
		for i := 1; i <= n; i++ {
			want = append(want, int64(gia.Gia_ObjRepr(g, int32(i))), int64(g.Next(int32(i))))
		}
		expected = append(expected, want)
	}
	for _, nw := range []int32{1, 2, 4} {
		for flags := int32(0); flags < 4; flags++ {
			for trial := 0; trial < 6; trial++ {
				for _, mode := range []byte{'S', 'I'} {
					g := gia.Gia_ManStart(10)
					a := gia.Gia_ManAppendCi(g)
					b := gia.Gia_ManAppendCi(g)
					gia.Gia_ManAppendAnd(g, a^(flags&1), b^(flags>>1))
					p := Cec_ManSStart(g, nw)
					fmt.Fprintf(&input, "%c %d %d %d", mode, nw, flags&1, flags>>1)
					for i := range p.Rands {
						p.Rands[i] = rng.Uint64()
						fmt.Fprintf(&input, " %x", p.Rands[i])
					}
					for i := range vec.Vec_WrdArray(p.vSims) {
						x := rng.Uint64()
						if trial == 0 {
							x = 0
						}
						vec.Vec_WrdWriteEntry(p.vSims, i, x)
						fmt.Fprintf(&input, " %x", x)
					}
					fmt.Fprintln(&input)
					status := int32(-1)
					if mode == 'S' {
						status = Cec_ManSRunPropagate(p, 3)
					} else {
						Cec_ManSRunImply(p, 3)
					}
					want := []int64{int64(status), int64(p.iRand), int64(p.nSkipped), int64(p.nVisited)}
					for _, w := range vec.Vec_WrdArray(p.vSims) {
						want = append(want, int64(w))
					}
					expected = append(expected, want)
				}
			}
		}
	}
	cmd = exec.Command(binary)
	cmd.Stdin = &input
	data, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatal(err, string(data))
	}
	scan := bufio.NewScanner(bytes.NewReader(data))
	scan.Buffer(make([]byte, 4096), 1<<20)
	for trial, want := range expected {
		if !scan.Scan() {
			t.Fatalf("missing C result %d", trial)
		}
		fields := strings.Fields(scan.Text())
		if len(fields) != len(want) {
			t.Fatalf("C diagnostic at %d: %s", trial, scan.Text())
		}
		for i, s := range fields {
			v, err := strconv.ParseInt(s, 10, 64)
			if err != nil {
				var u uint64
				u, err = strconv.ParseUint(s, 10, 64)
				v = int64(u)
			}
			if err != nil || v != want[i] {
				t.Fatalf("C sample %d field %d = %s, Go = %d", trial, i, s, want[i])
			}
		}
	}
}
