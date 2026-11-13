#include "misc/util/util_hack.h"
#include "aig/aig/aig.h"
#include "aig/saig/saig.h"
#include "misc/vec/vec.h"
#include <stdio.h>
extern Aig_Man_t *createNewAigWith0LivePo(Aig_Man_t *,Vec_Ptr_t *,int *);
int main(void){Aig_Man_t *p=Aig_ManStart(20);p->pName=Abc_UtilStrsav("fixture");p->fCatchExor=1; Aig_Obj_t *a=Aig_ObjCreateCi(p),*b=Aig_ObjCreateCi(p);Aig_Obj_t *x=Aig_Oper(p,a,b,AIG_OBJ_EXOR);Aig_ObjCreateCo(p,x); Aig_ObjCreateCo(p,a); Aig_ManSetRegNum(p,1); Vec_Ptr_t *v=Vec_PtrAlloc(1);Vec_PtrPush(v,x);int i;Aig_Man_t *q=createNewAigWith0LivePo(p,v,&i);Aig_Obj_t *driver=Aig_ObjFanin0(Aig_ManCo(q,0));printf("origxor=%d copiedand=%d copiedxor=%d complemented=%d\n",Aig_ObjIsExor(x),Aig_ObjIsAnd(driver),Aig_ObjIsExor(driver),Aig_ObjFaninC0(Aig_ManCo(q,0)));return 0;}
