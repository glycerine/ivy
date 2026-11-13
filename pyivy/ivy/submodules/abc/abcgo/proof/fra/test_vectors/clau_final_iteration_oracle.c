#include "aig/aig/aig.h"
#include <stdio.h>
extern int Fra_Clau(Aig_Man_t *,int,int,int);
int main(void) { for(int limit=1;limit<=2;limit++) { Aig_Man_t *g=Aig_ManStart(128); Aig_Obj_t *e[12]; for(int i=0;i<4;i++) e[i]=Aig_ObjCreateCi(g);
 e[4]=Aig_And(g,Aig_Not(e[1]),Aig_Not(e[2])); e[5]=Aig_And(g,e[3],Aig_Not(e[1])); e[6]=Aig_And(g,Aig_Not(e[2]),Aig_Not(e[2])); e[7]=Aig_And(g,Aig_Not(e[6]),e[0]); e[8]=Aig_And(g,e[1],Aig_Not(e[1])); e[9]=Aig_And(g,Aig_Not(e[4]),e[1]); e[10]=Aig_And(g,Aig_Not(e[1]),e[3]); e[11]=Aig_And(g,Aig_Not(e[6]),e[10]); Aig_ObjCreateCo(g,e[7]); Aig_ObjCreateCo(g,e[2]); Aig_ObjCreateCo(g,e[1]);Aig_ManSetRegNum(g,2);
 printf("limit=%d result=%d\n",limit,Fra_Clau(g,limit,0,0));Aig_ManStop(g); }
}
