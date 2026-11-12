#include <stdio.h>
#include <stdlib.h>
#include "misc/util/abc_global.h"
#include "sat/msat/msatInt.h"
long long timeSelect = 0;
int main(int argc,char **argv){
 if(argc<2)return 2;
 FILE *f=fopen(argv[1],"r");if(!f)return 2;
 Msat_Solver_t *s=NULL;int parsed=Msat_SolverParseDimacs(f,&s,0);fclose(f);
 if(!parsed){puts("RESULT=-1");return 0;}
 Msat_IntVec_t *cone=Msat_IntVecAlloc(Msat_SolverReadVarNum(s));for(int v=0;v<Msat_SolverReadVarNum(s);v++)Msat_IntVecPush(cone,v);Msat_OrderClean(s->pOrder,cone);Msat_IntVecFree(cone);
 Msat_IntVec_t *as=Msat_IntVecAlloc(argc-2);
 for(int i=2;i<argc;i++)Msat_IntVecPush(as,atoi(argv[i]));
 int result=Msat_SolverSimplifyDB(s)?Msat_SolverSolve(s,as,0,0):-1;
 printf("RESULT=%d\n",result);
 Msat_IntVecFree(as);Msat_SolverFree(s);return 0;
}
