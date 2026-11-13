#include <stdio.h>
#include "proof/acec/acecPolyn.c"
int main(void){int in[][5]={{1000000000,1000000000,-1000000000,0,0},{-1000000000,-1000000000,999999999,0,0},{1000000000,-1000000000,29,29,-30},{-900000000,899999999,0,0,0}};for(int j=0;j<4;j++){Vec_Int_t*v=Vec_IntAlloc(5);for(int i=0;i<5&&in[j][i];i++)Gia_PolynMergeConstOne(v,in[j][i]);printf("row%d:",j);for(int i=0;i<Vec_IntSize(v);i++)printf(" %d",Vec_IntEntry(v,i));puts("");}return 0;}
