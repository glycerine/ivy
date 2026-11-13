#include <stdio.h>
#include "sat/bsat/satSolver2.h"
int main(void) {
 Prf_Man_t *p=Prf_ManAlloc(); int i;
 Prf_ManRestart(p, NULL, 0, 1);
 for(i=0;i<1500;i++){ Prf_ManClearNewInfo(p); Vec_WrdWriteEntry(p->vInfo,i,1); }
 Prf_ManGrow(p,64);
 printf("grow words=%d rows=%d cap=%d\n",p->nWords,Prf_ManSize(p),Vec_WrdCap(p->vInfo));
 for(i=0;i<1500;i+=2) Prf_ManAddSaved(p,i,i/2);
 Prf_ManCompact(p,750);
 printf("compact words=%d rows=%d cap=%d first=%llu last=%llu\n",p->nWords,Prf_ManSize(p),Vec_WrdCap(p->vInfo),(unsigned long long)Prf_ManClauseInfo(p,0)[0],(unsigned long long)Prf_ManClauseInfo(p,749)[0]);
 Prf_ManStop(p); return 0;
}
