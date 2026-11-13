#include "misc/util/util_hack.h"
#include "aig/aig/aig.h"
#include "aig/saig/saig.h"
#include "base/abc/abc.h"
#include "misc/vec/vec.h"
#include <stdio.h>
extern Aig_Man_t *createNewAigWith0LivePo(Aig_Man_t *, Vec_Ptr_t *, int *);
extern Aig_Man_t *introduceAbsorberLogic(Aig_Man_t *, int *, int *, int);
extern Aig_Man_t *generateGeneralDisjunctiveTester(Abc_Ntk_t *, Aig_Man_t *, int);
int main(void) {
 Aig_Man_t *p=Aig_ManStart(20); p->pName=Abc_UtilStrsav("fixture");
 Aig_Obj_t *a=Aig_ObjCreateCi(p), *b=Aig_ObjCreateCi(p);
 Aig_ObjCreateCo(p,a); Aig_ObjCreateCo(p,Aig_Not(a)); Aig_ObjCreateCo(p,b);
 Aig_ManSetRegNum(p,1); p->nConstrs=1;
 Vec_Ptr_t *v=Vec_PtrAlloc(1); Vec_PtrPush(v,a);
 int target=-1; Aig_Man_t *q=createNewAigWith0LivePo(p,v,&target);
 printf("zero oldpos=%d oldconstraints=%d newpos=%d newconstraints=%d target=%d\n",Saig_ManPoNum(p),p->nConstrs,Saig_ManPoNum(q),q->nConstrs,target);
 int zero=0, absorber=-1; Aig_Man_t *r=introduceAbsorberLogic(p,&zero,&absorber,0);
 printf("absorber newpos=%d newconstraints=%d target=%d\n",Saig_ManPoNum(r),r->nConstrs,absorber);
 Aig_Man_t *s=generateGeneralDisjunctiveTester(NULL,p,1);
 printf("combination newpos=%d newconstraints=%d\n",Saig_ManPoNum(s),s->nConstrs);
 Aig_Man_t *d=Aig_ManDupSimple(p);
 printf("dup newpos=%d newconstraints=%d\n",Saig_ManPoNum(d),d->nConstrs);
 return 0;
}
