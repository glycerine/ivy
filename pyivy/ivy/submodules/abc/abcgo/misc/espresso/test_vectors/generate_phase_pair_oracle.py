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
char *util_strsav(const char *);
long ptime(void);
char *util_print_time(long);
''')
files='set setc cubestr cofactor compl irred expand contain unate matrix rows cols mincov solution indep gimpel part dominate cvrm cvrmisc globals sparse cvrout sminterf pair opo cvrin'.split()
for n in files:(out/(n+'.c')).write_text((src/(n+'.c')).read_text())
(out/'main.c').write_text(r'''#include "espresso.h"
long util_cpu_time(void){return 0;} long ptime(void){return 0;} char *util_print_time(long t){return "0";}
pcover readcover(){int n,i,b;unsigned mask;scanf("%d",&n);pcover f=new_cover(n);for(i=0;i<n;i++){scanf("%u",&mask);pcube p=GETSET(f,f->count++);set_clear(p,cube.size);for(b=0;b<cube.size;b++)if(mask&(1u<<b))set_insert(p,b);}return f;}
char *util_strsav(const char *s){char *p=malloc(strlen(s)+1);strcpy(p,s);return p;}
void output(pcover f){int i,b;printf("[");for(i=0;i<f->count;i++){unsigned mask=0;pcube p=GETSET(f,i);for(b=0;b<cube.size;b++)if(is_in_set(p,b))mask|=1u<<b;if(i)printf(",");printf("%u",mask);}printf("]");}
int main(){int n;scanf("%d",&n);while(n--){int op,i;scanf("%d",&op);cube.num_vars=5;cube.num_binary_vars=4;cube.part_size=ALLOC(int,5);cube.part_size[4]=2;cube_setup();pPLA p=new_PLA();p->F=readcover();p->D=readcover();p->R=readcover();if(op==0){p->pair=pair_new(1);p->pair->cnt=1;p->pair->var1[0]=1;p->pair->var2[0]=3;set_pair(p);}else if(op==1){p->phase=set_save(cube.fullset);set_remove(p->phase,cube.last_part[cube.output]);set_phase(p);}else{output_phase_setup(p,1);}printf("{\"mapped_on\":");output(p->F);printf(",\"mapped_dc\":");output(p->D);printf(",\"mapped_off\":");output(p->R);printf("}\n");setdown_cube();}return 0;}
''')
r=subprocess.run(['cc','-std=gnu89','-w','-O2','-ffunction-sections','-fdata-sections','-include',str(out/'compat.h'),'-I'+str(out)]+[str(out/(n+'.c')) for n in files+['main']]+['-Wl,--gc-sections','-o',str(out/'oracle')],capture_output=True,text=True)
if r.returncode: raise RuntimeError(r.stderr)
import random,json,itertools
rng=random.Random(941)
cases=[];inputs=[]
parts=[range(0,2),range(2,4),range(4,6),range(6,8),range(8,10)]
for trial in range(120):
 families=[]
 for _ in range(3):
  f=[]
  for row in range(rng.randrange(1,10)):
   mask=0
   for bits in parts:
    sub=rng.randrange(1,1<<len(bits));mask|=sum(1<<b for i,b in enumerate(bits) if sub&(1<<i))
   f.append(mask)
  families.append(f)
 op=trial%3;cases.append(dict(operation=op,on=families[0],dc=families[1],off=families[2]))
 inputs.append(str(op)+" "+" ".join(str(x) for family in families for x in [len(family),*family]))
result=subprocess.run([str(out/'oracle')],input=str(len(cases))+"\n"+"\n".join(inputs),text=True,capture_output=True,check=True)
for case,line in zip(cases,result.stdout.splitlines(),strict=True):case.update(json.loads(line))
Path(__file__).with_name('phase_pair_oracle.json').write_text(json.dumps(cases,separators=(',',':'))+'\n')
print("wrote",len(cases),"fixtures; original-C build",out)
