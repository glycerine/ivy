from pathlib import Path
import tempfile,subprocess,random,json
root=Path(__file__).resolve().parents[4];tmp=Path(tempfile.mkdtemp(prefix='espresso-mincov-oracle-'))
(tmp/'compat.h').write_text('''#ifndef ORACLE_COMPAT_H
#define ORACLE_COMPAT_H
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <assert.h>
#include <time.h>
#define ABC_NAMESPACE_HEADER_START
#define ABC_NAMESPACE_HEADER_END
#define ABC_NAMESPACE_IMPL_START
#define ABC_NAMESPACE_IMPL_END
#define ALLOC(type,num) ((type*)malloc(sizeof(type)*(num)))
#define REALLOC(type,obj,num) ((type*)realloc((obj),sizeof(type)*(num)))
#define FREE(obj) free(obj)
#define NIL(type) ((type*)0)
#define MAX(a,b) ((a)>(b)?(a):(b))
#define MIN(a,b) ((a)<(b)?(a):(b))
#define WEIGHT(weight,column) ((weight)==NULL?1:(weight)[column])
static long util_cpu_time(){return clock();}
static char* util_print_time(long elapsed){static char s[64];snprintf(s,sizeof(s),"%.3f",(double)elapsed/CLOCKS_PER_SEC);return s;}
#endif
''')
for name in ['sparse.h','sparse_int.h','mincov.h','mincov_int.h']:
 text=(root/'src/misc/espresso'/name).read_text().replace('#include "util_hack.h"','#include "compat.h"')
 if name=='sparse_int.h':text=text.replace('#define ABC__misc__espresso__sparse_int_h\n    if','#define sorted_insert(type,first,last,count,next,prev,value,newval,e) \\\n    if')
 (tmp/name).write_text(text)
for name in ['matrix','rows','cols','part','dominate','mincov','solution','indep','gimpel']:(tmp/(name+'.c')).write_text((root/'src/misc/espresso'/(name+'.c')).read_text())
(tmp/'main.c').write_text('''#include "mincov_int.h"
int main(){int rows,cols,weighted,heuristic;while(scanf("%d %d %d %d",&rows,&cols,&weighted,&heuristic)==4){int weights[32];for(int i=0;i<cols;i++)scanf("%d",&weights[i]);sm_matrix*A=sm_alloc();for(int r=0;r<rows;r++){unsigned mask;scanf("%u",&mask);for(int c=0;c<cols;c++)if(mask&(1u<<c))sm_insert(A,r,c);}sm_row*cover=sm_minimum_cover(A,weighted?weights:NULL,heuristic,0);int cost=0;for(sm_element*p=cover->first_col;p;p=p->next_col)cost+=weighted?weights[p->col_num]:1;printf("%d",cost);for(sm_element*p=cover->first_col;p;p=p->next_col)printf(" %d",p->col_num);printf("\\n");sm_row_free(cover);sm_free(A);}}
''')
cmd=['cc','-O2','-Wno-implicit-function-declaration','-include',str(tmp/'compat.h'),'-I'+str(tmp),*[str(tmp/(n+'.c')) for n in ['main','matrix','rows','cols','part','dominate','mincov','solution','indep','gimpel']],'-o',str(tmp/'reference')]
r=subprocess.run(cmd,capture_output=True,text=True)
if r.returncode:print(r.stderr[-7000:]);print(tmp);raise SystemExit(r.returncode)
rng=random.Random(91);cases=[]
for index in range(1000):
 rows=rng.randrange(1,9);cols=rng.randrange(1,9);weighted=index%2;heuristic=(index//2)%2;weights=[rng.randrange(1,8) for _ in range(cols)];masks=[rng.randrange(1,1<<cols) for _ in range(rows)];cases.append(dict(rows=rows,cols=cols,weighted=weighted,heuristic=heuristic,weights=weights,masks=masks))
text=''.join(f"{x['rows']} {x['cols']} {x['weighted']} {x['heuristic']} "+' '.join(map(str,x['weights']+x['masks']))+'\n' for x in cases)
r=subprocess.run([str(tmp/'reference')],input=text,capture_output=True,text=True,timeout=30)
if r.returncode:print(r.stderr,r.stdout[-500:]);print(tmp);raise SystemExit(r.returncode)
lines=r.stdout.splitlines()
if len(lines)!=len(cases):print('bad lines',len(lines));raise SystemExit(1)
for case,line in zip(cases,lines):
 values=list(map(int,line.split()));case['cost']=values[0];case['cover']=values[1:]
(tmp/'oracle.json').write_text(json.dumps(cases));print('Built original C oracle; 1000 exact/heuristic weighted/unit cases generated:',tmp/'oracle.json')

fixture=Path(__file__).with_name("mincov_oracle.json")
fixture.write_text(json.dumps(cases[:120],separators=(",",":"))+"\n")
print("Wrote permanent first-120-case fixture:",fixture)
