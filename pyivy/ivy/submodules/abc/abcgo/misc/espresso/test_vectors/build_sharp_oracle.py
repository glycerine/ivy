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
files='set setc cubestr cofactor compl irred expand contain unate matrix rows cols mincov solution indep gimpel part dominate cvrm cvrmisc globals sparse cvrout essen equiv sharp'.split()
for n in files:(out/(n+'.c')).write_text((src/(n+'.c')).read_text())
(out/'main.c').write_text(r'''#include "espresso.h"
long util_cpu_time(void){return 0;} long ptime(void){return 0;} char *util_print_time(long t){return "0";}
void print_cover(pcover f){int i;printf("[");for(i=0;i<f->count;i++){if(i)printf(",");printf("%u",GETSET(f,i)[1]);}printf("]");free_cover(f);}
int main(){int n,m,i;unsigned bits;pcover f,d;pcube query;cube.num_vars=3;cube.num_binary_vars=1;cube.part_size=ALLOC(int,3);cube.part_size[1]=3;cube.part_size[2]=2;cube_setup();
while(scanf("%d%d%u",&n,&m,&bits)==3){query=new_cube();query[1]=bits;f=new_cover(n);d=new_cover(m);
for(i=0;i<n;i++){scanf("%u",&bits);set_clear(GETSET(f,i),cube.size);GETSET(f,i)[1]=bits;}f->count=n;
for(i=0;i<m;i++){scanf("%u",&bits);set_clear(GETSET(d,i),cube.size);GETSET(d,i)[1]=bits;}d->count=m;
printf("{\"sharp\":");print_cover(cv_sharp(f,d));printf(",\"dsharp\":");print_cover(cv_dsharp(f,d));printf(",\"intersection\":");print_cover(cv_intersect(f,d));printf(",\"disjoint\":");print_cover(make_disjoint(f));printf(",\"cube_sharp\":");print_cover(cb_sharp(query,d));printf(",\"cube_dsharp\":");print_cover(cb_dsharp(query,d));printf("}\n");free_cover(f);free_cover(d);free_cube(query);}
return 0;}
''')
r=subprocess.run(['cc','-std=gnu89','-w','-O2','-ffunction-sections','-fdata-sections','-include',str(out/'compat.h'),'-I'+str(out)]+[str(out/(n+'.c')) for n in files+['main']]+['-Wl,--gc-sections','-o',str(out/'oracle')],capture_output=True,text=True)
print(out);print(r.stderr[-10000:]);print(r.returncode)

if r.returncode:raise SystemExit(r.returncode)
import random,json
rng=random.Random(8675);cases=[]
for k in range(120):
 def row():return rng.randrange(1,4)|(rng.randrange(1,8)<<2)|(rng.randrange(1,4)<<5)
 cases.append({"input":[row() for _ in range(rng.randrange(10))],"dc":[row() for _ in range(rng.randrange(4))],"query":row()})
request="".join(str(len(c["input"]))+" "+str(len(c["dc"]))+" "+str(c["query"])+" "+" ".join(map(str,c["input"]+c["dc"]))+"\n" for c in cases)
run=subprocess.run([str(out/'oracle')],input=request,text=True,capture_output=True,check=True)
for case,line in zip(cases,run.stdout.splitlines()):case.update(json.loads(line))
Path('abcgo/misc/espresso/test_vectors/sharp_oracle.json').write_text(json.dumps(cases,indent=2)+"\n")
print(len(cases),'cases generated')
