from pathlib import Path
import tempfile,subprocess
src=Path('src/misc/espresso'); out=Path(tempfile.mkdtemp(prefix='espresso-cover-'))
for p in src.glob('*.h'):
 s=p.read_text().replace('#include "util_hack.h"','')
 if p.name=='sparse_int.h':s=s.replace('#define ABC__misc__espresso__sparse_int_h\n    if','#define sorted_insert(type,first,last,count,next,prev,value,newval,e) \\\n    if')
 (out/p.name).write_text(s)
(out/'compat.h').write_text('''#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <assert.h>
#define LIN64
#define BPI 32
#define ABC_NAMESPACE_HEADER_START
#define ABC_NAMESPACE_HEADER_END
#define ABC_NAMESPACE_IMPL_START
#define ABC_NAMESPACE_IMPL_END
#define ALLOC(t,n) ((t*)malloc(sizeof(t)*(n)))
#define REALLOC(t,p,n) ((t*)realloc(p,sizeof(t)*(n)))
#define FREE(p) free(p)
#define WEIGHT(w,c) ((w)?(w)[c]:1)
#define NIL(t) ((t*)0)
#define ABS(x) ((x)<0?-(x):(x))
#define MAX(a,b) ((a)>(b)?(a):(b))
#define MIN(a,b) ((a)<(b)?(a):(b))
long ptime(void);
char *util_print_time(long);
''')
files='set setc cubestr cofactor compl irred expand contain unate matrix rows cols mincov solution indep gimpel part dominate cvrm cvrmisc globals sparse cvrout'.split()
for n in files:(out/(n+'.c')).write_text((src/(n+'.c')).read_text())
(out/'main.c').write_text(r'''#include "espresso.h"
long util_cpu_time(void){return 0;} long ptime(void){return 0;} char *util_print_time(long t){return "0";}
void print_cover(pcover f) { int i;pcube p;printf("[");for(i=0;i<f->count;i++){p=GETSET(f,i);if(i)printf(",");printf("%u",p[1]);}printf("]");}
int main(){int n,i,j;unsigned bits;pcover f,a,b,c,d;cube.num_vars=3;cube.num_binary_vars=0;cube.part_size=ALLOC(int,3);cube.part_size[0]=2;cube.part_size[1]=3;cube.part_size[2]=2;cube_setup();
while(scanf("%d",&n)==1){f=new_cover(n);for(i=0;i<n;i++){scanf("%u",&bits);set_clear(GETSET(f,i),cube.size);GETSET(f,i)[1]=bits;}f->count=n;
a=complement(cube1list(f));b=simplify(cube1list(f));simp_comp(cube1list(f),&c,&d);printf("{\"complement\":");print_cover(a);printf(",\"simplify\":");print_cover(b);printf(",\"on\":");print_cover(c);printf(",\"off\":");print_cover(d);printf("}\n");free_cover(a);free_cover(b);free_cover(c);free_cover(d);free_cover(f);}
return 0;}
''')
r=subprocess.run(['cc','-std=gnu89','-w','-O2','-ffunction-sections','-fdata-sections','-include',str(out/'compat.h'),'-I'+str(out)]+[str(out/(n+'.c')) for n in files+['main']]+['-Wl,--gc-sections','-o',str(out/'oracle')],capture_output=True,text=True)
print(out);print(r.stderr[-10000:]);print(r.returncode)

if r.returncode:raise SystemExit(r.returncode)
import random,json
rng=random.Random(8675);cases=[]
for k in range(180):
 rows=[]
 for j in range(rng.randrange(13)):
  rows.append(rng.randrange(1,4)|(rng.randrange(1,8)<<2)|(rng.randrange(1,4)<<5))
 cases.append({"input":rows})
request="".join(str(len(c["input"]))+" "+" ".join(map(str,c["input"]))+"\n" for c in cases)
run=subprocess.run([str(out/'oracle')],input=request,text=True,capture_output=True,check=True)
for case,line in zip(cases,run.stdout.splitlines()):case.update(json.loads(line))
Path('abcgo/misc/espresso/test_vectors/compl_oracle.json').write_text(json.dumps(cases,indent=2)+"\n")
print(len(cases),'cases generated')
