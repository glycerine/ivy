"""Regenerate sparse-matrix fixtures using the original C algorithms."""
from pathlib import Path
import json
import random
import subprocess
import tempfile

root = Path(__file__).resolve().parents[4]
build = Path(tempfile.mkdtemp(prefix="espresso-sparse-oracle-"))
(build / "compat.h").write_text(r'''
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <assert.h>
#define ABC_NAMESPACE_HEADER_START
#define ABC_NAMESPACE_HEADER_END
#define ABC_NAMESPACE_IMPL_START
#define ABC_NAMESPACE_IMPL_END
#define ALLOC(type,n) ((type*)malloc(sizeof(type)*(n)))
#define REALLOC(type,obj,n) ((type*)realloc((obj),sizeof(type)*(n)))
#define FREE(obj) free(obj)
#define NIL(type) ((type*)0)
#define MAX(a,b) ((a)>(b)?(a):(b))
#define MIN(a,b) ((a)<(b)?(a):(b))
''')
for name in ["sparse.h", "sparse_int.h", "mincov.h", "mincov_int.h"]:
    text = (root / "src/misc/espresso" / name).read_text()
    text = text.replace('#include "util_hack.h"', '')
    if name == "sparse_int.h":
        # The checked-in header lost only this declaration. Keep the original
        # field-substitution replacement text, rather than rewriting its logic.
        text = text.replace(
            '#define ABC__misc__espresso__sparse_int_h\n    if',
            '#define sorted_insert(type,first,last,count,next,prev,value,newval,e) \\\n    if')
    (build / name).write_text(text)
for name in ["matrix", "rows", "cols", "dominate", "part"]:
    (build / (name + ".c")).write_text((root / "src/misc/espresso" / (name + ".c")).read_text())
(build / "main.c").write_text(r'''
#include "mincov_int.h"
static void entries(sm_matrix *a) {
 int first=1;printf("[");
 for(sm_row*r=a->first_row;r;r=r->next_row)for(sm_element*e=r->first_col;e;e=e->next_col){if(!first)printf(",");printf("[%d,%d]",e->row_num,e->col_num);first=0;}
 printf("]");
}
static void vector(sm_row*r){int first=1;printf("[");for(sm_element*e=r->first_col;e;e=e->next_col){if(!first)printf(",");printf("%d",e->col_num);first=0;}printf("]");}
int main(){int n;while(scanf("%d",&n)==1){
 sm_matrix*a=sm_alloc(),*left=NULL,*right=NULL;int weights[32];
 for(int i=0;i<n;i++){char op;int r,c;scanf(" %c %d %d",&op,&r,&c);if(op=='I')sm_insert(a,r,c);else if(op=='R')sm_remove(a,r,c);else if(op=='D')sm_delrow(a,r);else sm_delcol(a,c);}
 for(int i=0;i<32;i++)scanf("%d",&weights[i]);
 sm_row*x=sm_row_alloc(),*y=sm_row_alloc();sm_col*cx=sm_col_alloc(),*cy=sm_col_alloc();int na,nb,v;scanf("%d %d",&na,&nb);
 for(int i=0;i<na;i++){scanf("%d",&v);sm_row_insert(x,v);sm_col_insert(cx,v);}for(int i=0;i<nb;i++){scanf("%d",&v);sm_row_insert(y,v);sm_col_insert(cy,v);}
 printf("{\"entries\":");entries(a);printf(",\"row_count\":%d,\"col_count\":%d",a->nrows,a->ncols);
 sm_row*lr=sm_longest_row(a);sm_col*lc=sm_longest_col(a);printf(",\"longest_row\":%d,\"longest_col\":%d",lr?lr->row_num:-1,lc?lc->col_num:-1);
 int cmp=sm_row_compare(x,y),ccmp=sm_col_compare(cx,cy);printf(",\"subset\":%d,\"intersects\":%d,\"compare\":%d,\"col_subset\":%d,\"col_intersects\":%d,\"col_compare\":%d,\"hash\":%d",sm_row_contains(x,y),sm_row_intersects(x,y),(cmp>0)-(cmp<0),sm_col_contains(cx,cy),sm_col_intersects(cx,cy),(ccmp>0)-(ccmp<0),sm_row_hash(x,1009));
 sm_row*intersection=sm_row_and(x,y);printf(",\"intersection\":");vector(intersection);sm_row_free(intersection);
 sm_matrix*b=sm_dup(a);int rd=sm_row_dominance(b),cd=sm_col_dominance(b,NULL);printf(",\"removed_rows\":%d,\"removed_cols\":%d,\"dominated\":",rd,cd);entries(b);sm_free(b);
 b=sm_dup(a);int wd=sm_col_dominance(b,weights);printf(",\"weighted_removed\":%d,\"weighted_dominated\":",wd);entries(b);sm_free(b);
 int split=sm_block_partition(a,&left,&right);printf(",\"split\":%d,\"left\":",split);if(split)entries(left);else printf("[]");printf(",\"right\":");if(split)entries(right);else printf("[]");printf("}\n");
 if(split){sm_free(left);sm_free(right);}sm_row_free(x);sm_row_free(y);sm_col_free(cx);sm_col_free(cy);sm_free(a);
}}
''')
command = ["cc", "-O2", "-Wno-implicit-function-declaration", "-include", str(build / "compat.h"), "-I" + str(build)]
command += [str(build / (name + ".c")) for name in ["main", "matrix", "rows", "cols", "dominate", "part"]]
command += ["-o", str(build / "reference")]
subprocess.run(command, check=True, capture_output=True, text=True)
rng = random.Random(7921)
cases = []
for case in range(160):
    operations = []
    for _ in range(80):
        operations.append(dict(op=rng.choice(["I", "I", "I", "R", "D", "C"]), row=rng.randrange(12), col=rng.randrange(16)))
    cases.append(dict(operations=operations, weights=[rng.randrange(1,9) for _ in range(32)], first=[rng.randrange(24) for _ in range(rng.randrange(20))], second=[rng.randrange(24) for _ in range(rng.randrange(20))]))
text = ""
for case in cases:
    text += str(len(case["operations"])) + "\n"
    for op in case["operations"]:
        text += f"{op['op']} {op['row']} {op['col']}\n"
    text += " ".join(map(str, case["weights"])) + "\n"
    text += f"{len(case['first'])} {len(case['second'])}\n"
    text += " ".join(map(str, case["first"] + case["second"])) + "\n"
result = subprocess.run([str(build / "reference")], input=text, capture_output=True, text=True, check=True)
lines = result.stdout.splitlines()
assert len(lines) == len(cases)
for case, line in zip(cases, lines):
    case["expected"] = json.loads(line)
    case["operations"] = [[ord(op["op"]), op["row"], op["col"]] for op in case["operations"]]
output = Path(__file__).with_name("sparse_oracle.json")
output.write_text(json.dumps(cases, separators=(",", ":")) + "\n")
print(f"Wrote {len(cases)} original-C sparse fixtures: {output}")
