#include "base/main/main.h"
#include "misc/util/util_hack.h"
#include "base/abc/abc.h"
#include <stdio.h>
extern int Abc_NtkSweep(Abc_Ntk_t *,int);
int main(void) {
 Abc_Start();
 Abc_Ntk_t *n=Abc_NtkAlloc(ABC_NTK_LOGIC,ABC_FUNC_SOP,1);
 Abc_Obj_t *a=Abc_NtkCreatePi(n),*b=Abc_NtkCreatePi(n);
 Abc_Obj_t *inv=Abc_NtkCreateNodeInv(n,a);
 Vec_Ptr_t *v=Vec_PtrAlloc(2);Vec_PtrPush(v,inv);Vec_PtrPush(v,b);
 Abc_Obj_t *and=Abc_NtkCreateNodeAnd(n,v);Vec_PtrFree(v);
 Abc_Obj_t *po=Abc_NtkCreatePo(n);Abc_ObjAddFanin(po,and);
 Abc_NtkCreateNodeBuf(n,a);
 int before=Abc_NtkNodeNum(n),removed=Abc_NtkSweep(n,0);
 Abc_Obj_t *root=Abc_ObjFanin0(po);
 printf("before=%d removed=%d after=%d fanins=%d first=%d second=%d\n",before,removed,Abc_NtkNodeNum(n),Abc_ObjFaninNum(root),Abc_ObjFanin0(root)->Id,Abc_ObjFanin1(root)->Id);
 Abc_Stop();return 0;
}
