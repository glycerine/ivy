#include "proof/ssw/sswInt.h"
#include <stdio.h>
int main(void) {
 Aig_Man_t *g=Aig_ManStart(16);Aig_Obj_t *a=Aig_ObjCreateCi(g), *b=Aig_ObjCreateCi(g), *q=Aig_ObjCreateCi(g);
 Aig_ObjCreateCo(g,a);Aig_ObjCreateCo(g,Aig_Not(b));Aig_ObjCreateCo(g,q);Aig_ManSetRegNum(g,1);
 Ssw_Cla_t *c=Ssw_ClassesPrepareTargets(g);
 printf("targets=%d classes=%d a_const=%d b_const=%d q_const=%d\n",Ssw_ClassesCand1Num(c),Ssw_ClassesClassNum(c),Ssw_ObjIsConst1Cand(g,a),Ssw_ObjIsConst1Cand(g,b),Ssw_ObjIsConst1Cand(g,q));
 Ssw_ClassesStop(c);Aig_ManStop(g);
 g=Aig_ManStart(16);a=Aig_ObjCreateCi(g);b=Aig_ObjCreateCi(g);q=Aig_ObjCreateCi(g);
 Vec_Int_t*v=Vec_IntAlloc(4);Vec_IntPush(v,Aig_ObjId(a));Vec_IntPush(v,Aig_ObjId(b));
 c=Ssw_ClassesPreparePairsSimple(g,v);
 printf("pairs=%d repr=%d target=%d\n",Ssw_ClassesClassNum(c),Aig_ObjId(Aig_ObjRepr(g,b)),Aig_ObjId(a));
 Ssw_ClassesStop(c);Vec_IntFree(v);Aig_ManStop(g);return 0;
}
