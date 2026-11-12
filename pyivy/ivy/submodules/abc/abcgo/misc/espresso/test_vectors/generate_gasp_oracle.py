"""Rebuild exact reduction/gasp fixtures from original Espresso C.

The first build preserves algorithms exactly. A second explicitly labeled
reference fixes only all_primes RAISE|FREESET, matching the separately verified
Go bug fix. Both results are retained. Compatibility allocator/timing wrappers
and the missing sorted_insert declaration are the other build adaptations.
"""
from pathlib import Path
import tempfile,subprocess
root=Path(__file__).resolve().parents[4]
src=root/'src/misc/espresso'; out=Path(tempfile.mkdtemp(prefix='espresso-cover-'))
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
files='set setc cubestr cofactor compl irred expand contain unate matrix rows cols mincov solution indep gimpel part dominate cvrm cvrmisc globals sparse cvrout sminterf reduce gasp'.split()
for n in files:(out/(n+'.c')).write_text((src/(n+'.c')).read_text())
(out/'main.c').write_text(r'''#include "espresso.h"
long util_cpu_time(void){return 0;} long ptime(void){return 0;} char *util_print_time(long t){return "0";}
pcover readcover(){int n,i,b;unsigned mask;scanf("%d",&n);pcover f=new_cover(n);for(i=0;i<n;i++){scanf("%u",&mask);pcube p=GETSET(f,f->count++);set_clear(p,cube.size);for(b=0;b<cube.size;b++)if(mask&(1u<<b))set_insert(p,b);}return f;}
void output(pcover f){int i,b;printf("[");for(i=0;i<f->count;i++){unsigned mask=0;pcube p=GETSET(f,i);for(b=0;b<cube.size;b++)if(is_in_set(p,b))mask|=1u<<b;if(i)printf(",");printf("%u",mask);}printf("]");}
int main(){int n;scanf("%d",&n);while(n--){cost_t cost;cube.num_vars=4;cube.num_binary_vars=2;cube.part_size=ALLOC(int,4);cube.part_size[2]=3;cube.part_size[3]=2;cube_setup();cube.sparse[2]=cube.sparse[3]=1;pcover f=readcover(),d=readcover(),r=readcover();f=irredundant(expand(f,r,0),d);printf("{\"base\":");output(f);pcover reduced=reduce(sf_save(f),d),last=last_gasp(sf_save(f),d,r,&cost),super=super_gasp(sf_save(f),d,r,&cost);printf(",\"reduced\":");output(reduced);printf(",\"last\":");output(last);printf(",\"super\":");output(super);printf("}\n");free_cover(f);free_cover(reduced);free_cover(last);free_cover(super);free_cover(d);free_cover(r);setdown_cube();}return 0;}
''')
command=['cc','-std=gnu89','-w','-O2','-ffunction-sections','-fdata-sections','-include',str(out/'compat.h'),'-I'+str(out)]+[str(out/(n+'.c')) for n in files+['main']]+['-Wl,--gc-sections','-o',str(out/'oracle')]
r=subprocess.run(command,capture_output=True,text=True)
if r.returncode: raise RuntimeError(r.stderr)
import random,json,itertools
rng=random.Random(937)
cases=[]; inputs=[]
parts=[range(0,2),range(2,4),range(4,7),range(7,9)]
points=[sum(1<<b for b in bs) for bs in itertools.product(*parts)]
for trial in range(120):
 f=[]
 for _ in range(rng.randrange(1,10)):
  mask=0
  for bits in parts:
   sub=rng.randrange(1,1<<len(bits))
   mask|=sum(1<<b for i,b in enumerate(bits) if sub&(1<<i))
  f.append(mask)
 d=[p for p in points if rng.randrange(8)==0]
 off=[p for p in points if not any(p&m==p for m in f+d)]
 cases.append(dict(on=f,dc=d,off=off))
 inputs.append(" ".join(str(x) for family in [f,d,off] for x in [len(family),*family]))
result=subprocess.run([str(out/'oracle')],input=str(len(cases))+"\n"+"\n".join(inputs),text=True,capture_output=True,check=True)
for case,line in zip(cases,result.stdout.splitlines(),strict=True):case.update(json.loads(line))
# The raw-source result stays in 'super'. Rebuild with the independently
# verified one-line FREESET fix to check exact output of corrected all_primes.
expand=(out/'expand.c').read_text()
old='B1 = new_cover(1);\n    p = GETSET(B1, B1->count++);\n    (void) set_copy(p, RAISE);'
new=old.replace('set_copy(p, RAISE)','set_or(p, RAISE, FREESET)')
assert expand.count(old)==1
(out/'expand.c').write_text(expand.replace(old,new))
subprocess.run(command,check=True,capture_output=True,text=True)
corrected=subprocess.run([str(out/'oracle')],input=str(len(cases))+"\n"+"\n".join(inputs),text=True,capture_output=True,check=True)
for case,line in zip(cases,corrected.stdout.splitlines(),strict=True):case['prime_super']=json.loads(line)['super']
Path(__file__).with_name('gasp_oracle.json').write_text(json.dumps(cases,separators=(',',':'))+'\n')
print("wrote",len(cases),"fixtures; original-C build",out)
