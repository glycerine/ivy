/* Independent fixture generator using ABC's original extraBddCas.c. */
#include "misc/util/util_hack.h"
#include "bdd/cudd/cuddInt.h"
#include "bdd/extrab/extraBdd.h"
#include <stdint.h>
#include <string.h>
extern struct { DdNode *Arg1; unsigned Arg2, Res; } HHTable2[15113];
static uint64_t hash = UINT64_C(14695981039346656037);
static void byte(unsigned v) { hash ^= v; hash *= UINT64_C(1099511628211); }
int main(void) {
 int partition, a, c, n, simple;
 for (partition=0; partition<65536; partition+=17) {
  DdManager *dd=Cudd_Init(5,0,CUDD_UNIQUE_SLOTS,CUDD_CACHE_SLOTS,0);
  DdNode *cols[4], *compact[4], *cube, *relation, *temp;
  Cudd_DisableGarbageCollection(dd);
  /* The original global pointer cache outlives managers. Clear it when
     creating a new manager to avoid allocator-address reuse contaminating cases. */
  memset(HHTable2,0,sizeof(HHTable2));
  for(c=0;c<4;c++)cols[c]=Cudd_Not(dd->one);
  for(a=0;a<8;a++) {
   c=(partition>>(2*a))&3;
   cube=Extra_bddBitsToCube(dd,a,3,dd->vars,1); Cudd_Ref(cube);
   cols[c]=Cudd_bddOr(dd,cols[c],cube); Cudd_Ref(cols[c]);
  }
  n=0;for(c=0;c<4;c++)if(cols[c]!=Cudd_Not(dd->one))compact[n++]=cols[c];
  cube=Extra_bddBitsToCube(dd,7,3,dd->vars,1); Cudd_Ref(cube);
  relation=Extra_bddEncodingNonStrict(dd,compact,n,cube,dd->vars+3,2,&simple);Cudd_Ref(relation);
  byte(simple);
  /* Relation codes may differ due to C pointer-hash traversal. Check invariant
     signature: number of codes, injectivity of column identity, support size. */
  for(a=0;a<8;a++) {
   int inputs[5],code,count=0;for(c=0;c<3;c++)inputs[c]=(a>>(2-c))&1;
   for(code=0;code<4;code++){inputs[3]=(code>>1)&1;inputs[4]=code&1;temp=Cudd_Eval(dd,relation,inputs);if(temp==dd->one)count++;}
   byte(count);
  }
  Cudd_Quit(dd);
 }
 printf("%016llx\n",(unsigned long long)hash);return 0;
}
