#include "misc/util/util_hack.h"
#include "aig/aig/aig.h"
#include <stdio.h>
static int repr(Aig_Man_t *p, Aig_Obj_t *o) { Aig_Obj_t *r = p->pReprs[o->Id]; return r ? r->Id : -1; }
static unsigned truth(Aig_Man_t *p, int index) {
 int mask,i; unsigned result=0; Aig_Obj_t *n,*o;
 for(mask=0;mask<16;mask++) { Aig_ManConst1(p)->iData=1; Aig_ManForEachCi(p,n,i) n->iData=(mask>>i)&1; Aig_ManForEachNode(p,n,i) n->iData=(Aig_Regular(n->pFanin0)->iData^Aig_IsComplement(n->pFanin0)) & (Aig_Regular(n->pFanin1)->iData^Aig_IsComplement(n->pFanin1)); o=Aig_ManCo(p,index); result|=(Aig_Regular(o->pFanin0)->iData^Aig_IsComplement(o->pFanin0))<<mask; } return result;
}
int main(void) {
 int sizes[]={1,2,3},limits[]={0,1,1000},levels[]={0,1}; int si,ci,li,j;
 for(si=0;si<3;si++) for(ci=0;ci<3;ci++) for(li=0;li<2;li++) {
  Aig_Man_t *g=Aig_ManStart(100),*result;
  Aig_Obj_t *a=Aig_ObjCreateCi(g),*b=Aig_ObjCreateCi(g),*c=Aig_ObjCreateCi(g),*d=Aig_ObjCreateCi(g);
  Aig_Obj_t *ab=Aig_And(g,a,b),*bc=Aig_And(g,b,c),*r1=Aig_And(g,ab,c),*r2=Aig_And(g,a,bc);
  Aig_Obj_t *r3=Aig_And(g,r1,Aig_Or(g,a,d)),*r4=Aig_And(g,Aig_Not(r2),Aig_Not(d));
  Aig_ObjCreateCo(g,r1); Aig_ObjCreateCo(g,r3); Aig_ObjCreateCo(g,Aig_Not(r2)); Aig_ObjCreateCo(g,r4);
  result=Aig_ManFraigPartitioned(g,sizes[si],limits[ci],levels[li],0);
  printf("matrix %d %d %d %d %d %d %d %d %d %d",sizes[si],limits[ci],levels[li],Aig_ManNodeNum(result),repr(g,r1),repr(g,r2),repr(g,r3),repr(g,r4),Aig_ManCoNum(result),Aig_ManCiNum(result));
  for(j=0;j<4;j++) printf(" %u",truth(result,j)); puts(""); Aig_ManStop(result); Aig_ManStop(g);
 }
 for(si=0;si<2;si++) for(ci=0;ci<3;ci+=2) for(li=0;li<2;li++) {
  Vec_Ptr_t *versions=Vec_PtrAlloc(2); Aig_Man_t *result; int v;
  for(v=0;v<2;v++) {
   Aig_Man_t *g=Aig_ManStart(100);
   Aig_Obj_t *a=Aig_ObjCreateCi(g),*b=Aig_ObjCreateCi(g),*c=Aig_ObjCreateCi(g),*d=Aig_ObjCreateCi(g);
   Aig_Obj_t *r=v?Aig_And(g,a,Aig_And(g,b,c)):Aig_And(g,Aig_And(g,a,b),c);
   Aig_ObjCreateCo(g,r); Aig_ObjCreateCo(g,Aig_Not(r)); Aig_ObjCreateCo(g,Aig_And(g,r,d)); Vec_PtrPush(versions,g);
  }
  result=Aig_ManChoicePartitioned(versions,sizes[si],limits[ci],levels[li],0);
  printf("choice-matrix %d %d %d %d %d %d %d",sizes[si],limits[ci],levels[li],Aig_ManNodeNum(result),Aig_ManChoiceNum(result),Aig_ManCiNum(result),Aig_ManCoNum(result));
  for(j=0;j<3;j++) printf(" %u",truth(result,j)); puts("");
  Aig_ManStop(result); for(v=0;v<2;v++) Aig_ManStop((Aig_Man_t*)Vec_PtrEntry(versions,v)); Vec_PtrFree(versions);
 }
 return 0;
}
