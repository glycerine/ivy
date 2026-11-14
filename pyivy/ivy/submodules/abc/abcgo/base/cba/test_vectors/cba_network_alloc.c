#include <stdio.h>
#include "base/cba/cba.h"
static void dump(char *name, Vec_Int_t *v) {printf("%s",name);for(int i=0;i<Vec_IntSize(v);i++)printf(" %d",Vec_IntEntry(v,i));puts("");}
int main(void) {
 Cba_Man_t *m=Cba_ManAlloc("sample.v",2,NULL,NULL,NULL,NULL);
 Cba_Ntk_t *n=Cba_NtkAlloc(m,1,2,1,8,8,8);
 int a=Cba_ObjAlloc(n,CBA_OBJ_PI,0,1),b=Cba_ObjAlloc(n,CBA_OBJ_PI,0,1);
 int g=Cba_ObjAlloc(n,CBA_BOX_AND,2,1),out=Cba_ObjAlloc(n,CBA_OBJ_PO,1,0);
 Cba_ObjSetFinFon(n,g,0,Cba_ObjFon0(n,a));Cba_ObjSetFinFon(n,g,1,Cba_ObjFon0(n,b));Cba_ObjSetFinFon(n,out,0,Cba_ObjFon0(n,g));
 printf("objects %d %d %d %d\n",a,b,g,out);
 dump("inputs",&n->vInputs);dump("outputs",&n->vOutputs);dump("fin_offsets",&n->vObjFin0);dump("fon_offsets",&n->vObjFon0);dump("fanin_fons",&n->vFinFon);dump("fon_objs",&n->vFonObj);
 Cba_ManFree(m);return 0;
}
