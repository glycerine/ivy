#include <stdio.h>
#include "base/cba/cba.h"
extern void Cba_BlastConst(Cba_Ntk_t *, Vec_Int_t *, int, int, int);
int main(void) {
 Cba_Man_t m={0}; Cba_Ntk_t p={0}; p.pDesign=&m; m.pFuns=Abc_NamStart(16,16);
 char *text[]={"4'b1010","4'b1x0z","7'h2F","6'd19","3'sb101","'b1"};int n[]={4,4,7,6,8,1};
 Vec_Int_t *out=Vec_IntAlloc(16);
 for(int i=0;i<6;i++){int id=Abc_NamStrFindOrAdd(m.pFuns,text[i],NULL);Cba_BlastConst(&p,out,-id-1,n[i],i==4);for(int j=0;j<Vec_IntSize(out);j++)printf("%d",Vec_IntEntry(out,j));puts("");}
 return 0;
}
