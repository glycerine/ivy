#include <stdio.h>
#include "base/cba/cba.h"
int main(void) {
 int ids[]={0,1,10,1<<20};
 printf("zero %d %d\n",Cba_FonIsReal(0),Cba_FonIsConst(0));
 for(int i=0;i<4;i++){int f=Cba_FonFromConst(ids[i]);printf("%d %d %d %d\n",ids[i],f,Cba_FonIsConst(f),Cba_FonConst(f));}
 return 0;
}
