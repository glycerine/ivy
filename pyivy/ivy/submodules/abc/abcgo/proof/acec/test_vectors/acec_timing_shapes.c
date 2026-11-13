#include <stdio.h>
#include "aig/gia/gia.h"
#include "misc/tim/tim.h"
extern Tim_Man_t *Gia_ManGenerateTim(int,int,int,int,int);
int main(void){int sh[][5]={{1,2,2,2,1},{3,1,2,4,3},{0,0,3,5,4},{2,3,0,3,2}};for(int k=0;k<4;k++){int*pi=sh[k];Tim_Man_t*t=Gia_ManGenerateTim(pi[0],pi[1],pi[2],pi[3],pi[4]);printf("case %d: ci=%d co=%d boxes=%d",k,Tim_ManCiNum(t),Tim_ManCoNum(t),Tim_ManBoxNum(t));for(int i=0;i<Tim_ManBoxNum(t);i++){float*v=Tim_ManBoxDelayTable(t,i);printf(" box=%d,%d,%d,%d neg=",Tim_ManBoxInputFirst(t,i),Tim_ManBoxInputNum(t,i),Tim_ManBoxOutputFirst(t,i),Tim_ManBoxOutputNum(t,i));for(int j=3;j<3+pi[3]*pi[4];j++)if(v[j]<0)printf("%d,",j-3);}puts("");Tim_ManStop(t);}return 0;}
