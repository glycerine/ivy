#include "opt/rwr/rwr.h"
#include <stdio.h>
#include "base/main/main.h"
int main(void) {
 Abc_Start(); Rwr_Man_t *p=Rwr_ManStart(0);
 printf("%d %d\n",Vec_PtrSize(p->vForest),Vec_VecSize(p->vClasses));
 for(int c=0;c<Vec_VecSize(p->vClasses);c++) {
  Vec_Ptr_t *level=Vec_VecEntry(p->vClasses,c);printf("%d",Vec_PtrSize(level));
  for(int i=0;i<Vec_PtrSize(level);i++){Rwr_Node_t *n=Vec_PtrEntry(level,i);printf(" %d:%u",n->Id,n->uTruth);} puts("");
 }
 Rwr_ManStop(p);Abc_Stop();return 0;
}
