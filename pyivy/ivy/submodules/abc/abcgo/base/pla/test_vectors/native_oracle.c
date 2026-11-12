#include "base/pla/pla.h"
#include <stdio.h>
#include <string.h>
extern int Pla_ManDist1Num(Pla_Man_t*);
extern int Pla_ManHashDistance1(Pla_Man_t*);
char * Extra_FileDesignName(char *name){char *base=strrchr(name,'/');char *p=Abc_UtilStrsav(base?base+1:name);char *dot=strrchr(p,'.');if(dot)*dot=0;return p;}
char * Extra_TimeStamp(void){return "ORACLE";}
unsigned Gia_ManRandom(int reset){static unsigned z=3716960521u,w=2174103536u;if(reset){z=3716960521u;w=2174103536u;}z=36969*(z&65535)+(z>>16);w=18000*(w&65535)+(w>>16);return(z<<16)+w;}
int main(int argc,char **argv){Pla_Man_t *p;if(argc>1){p=Pla_ReadPla(argv[1]);if(!p)return 2;}else{Gia_ManRandom(1);p=Pla_ManGenerate(8,4,5,0);}printf("%d %d %d %d\n",p->Type,p->nIns,p->nOuts,Pla_ManCubeNum(p));for(int i=0;i<Pla_ManCubeNum(p);i++){for(int k=0;k<p->nInWords;k++)printf("%016llx ",(unsigned long long)Pla_CubeIn(p,i)[k]);for(int k=0;k<p->nOutWords;k++)printf("%016llx ",(unsigned long long)Pla_CubeOut(p,i)[k]);printf("\n");}printf("dist %d hash %d\n",Pla_ManDist1Num(p),Pla_ManHashDistance1(p));Pla_ManFree(p);return 0;}
