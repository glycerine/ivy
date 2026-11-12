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
files='set setc cubestr cofactor compl irred expand contain unate matrix rows cols mincov solution indep gimpel part dominate cvrm cvrmisc globals sparse cvrout sminterf cvrin pair opo hack sharp'.split()
for n in files:(out/(n+'.c')).write_text((src/(n+'.c')).read_text())
(out/'main.c').write_text(r'''#include "espresso.h"
long util_cpu_time(void){return 0;} long ptime(void){return 0;} char *util_print_time(long t){return "0";}
char *util_strsav(const char *s){char *p=malloc(strlen(s)+1);strcpy(p,s);return p;}
void output(pcover f){int i,b;printf("[");for(i=0;i<f->count;i++){unsigned mask=0;pcube p=GETSET(f,i);for(b=0;b<cube.size;b++)if(is_in_set(p,b))mask|=1u<<b;if(i)printf(",");printf("%u",mask);}printf("]");}
int main(){pPLA p=new_PLA();read_pla(stdin,1,1,FD_type,&p);printf("{\"on\":");output(p->F);printf(",\"dc\":");output(p->D);printf(",\"off\":");output(p->R);printf("}\n");return 0;}
''')
r=subprocess.run(['cc','-std=gnu89','-w','-O2','-ffunction-sections','-fdata-sections','-include',str(out/'compat.h'),'-I'+str(out)]+[str(out/(n+'.c')) for n in files+['main']]+['-Wl,--gc-sections','-o',str(out/'oracle')],capture_output=True,text=True)
if r.returncode: raise RuntimeError(r.stderr)
import random,json
rng=random.Random(113)
cases=[]
for trial in range(60):
 n=3;outputs=2
 text=f".i {n}\n.o {outputs}\n"
 if trial%2==0:text+='.symbolic 0 2 ; A B C D ;\n'
 else:text+='.symbolic-output 0 1 ; A B C D ;\n'
 for value in range(1<<n):
  text+=''.join(str((value>>v)&1) for v in range(n))+' '+''.join(rng.choice('01') for _ in range(outputs))+'\n'
 text+='.end\n'
 result=subprocess.run([str(out/'oracle')],input=text,text=True,capture_output=True,check=True)
 case=dict(input=text);case.update(json.loads(result.stdout));cases.append(case)
Path(__file__).with_name('symbolic_oracle.json').write_text(json.dumps(cases,separators=(',',':'))+'\n')
print("wrote",len(cases),"fixtures; original-C build",out)
