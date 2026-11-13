#include <stdio.h>
#include "aig/aig/aig.h"
extern Aig_Man_t *Dch_DeriveChoiceAig(Aig_Man_t *,int);
int main(int argc, char **argv){Aig_Man_t*p=Aig_ManStart(100);Aig_Obj_t*a=Aig_ObjCreateCi(p),*b=Aig_ObjCreateCi(p),*c=Aig_ObjCreateCi(p),*d=Aig_ObjCreateCi(p);
Aig_Obj_t*x=Aig_And(p,Aig_And(p,Aig_And(p,a,b),c),d);Aig_Obj_t*y=Aig_And(p,Aig_And(p,a,b),Aig_And(p,c,d));Aig_Obj_t*z=Aig_And(p,a,Aig_And(p,b,Aig_And(p,c,d)));Aig_ObjCreateCo(p,x);Aig_ManReprStart(p,100);Aig_ObjSetRepr(p,y,x);Aig_ObjSetRepr(p,z,x);Aig_Man_t*q=Dch_DeriveChoiceAig(p,argc > 1);Aig_Obj_t*head=Aig_ObjFanin0(Aig_ManCo(q,0));for(Aig_Obj_t*o=head;o;o=Aig_ObjEquiv(q,o))printf("%d(%d,%d) ",o->Id,Aig_ObjFaninId0(o),Aig_ObjFaninId1(o));puts("");return 0;}
