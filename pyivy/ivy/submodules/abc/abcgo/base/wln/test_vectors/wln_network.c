/* Compare object, fanout, cycle and duplication behavior with source WLN. */
#include "base/wln/wln.h"
#include <stdio.h>
int main(void) {
 Wln_Ntk_t *p=Wln_NtkAlloc("",8), *q;
 int a=Wln_ObjAlloc(p,ABC_OPER_CI,0,7,0);
 int n=Wln_ObjAlloc(p,ABC_OPER_NONE,0,7,0);
 int co=Wln_ObjAlloc(p,ABC_OPER_CO,0,7,0);
 Wln_ObjAddFanin(p,n,a);Wln_ObjAddFanin(p,co,n);
 printf("source:struct=%zu mem=%d\n",sizeof(Wln_Ntk_t),Wln_NtkMemUsage(p));
 printf("source:objects=%d ci=%d co=%d acyclic=%d\n",Wln_NtkObjNum(p),Wln_NtkCiNum(p),Wln_NtkCoNum(p),Wln_NtkIsAcyclic(p));
 Wln_NtkStaticFanoutStart(p);
 printf("source:refs=%d,%d fanouts=%d,%d\n",Wln_ObjFanoutNum(p,a),Wln_ObjFanoutNum(p,n),Wln_ObjFanout(p,a,0),Wln_ObjFanout(p,n,0));
 Wln_NtkStaticFanoutStop(p);
 q=Wln_NtkDupDfs(p);printf("source:dup=%d,%d,%d,%d\n",Wln_NtkObjNum(q),Wln_NtkCiNum(q),Wln_NtkCoNum(q),Wln_ObjFanin0(q,Wln_NtkCo(q,0)));
 Wln_NtkFree(q);Wln_NtkFree(p);return 0;
}
