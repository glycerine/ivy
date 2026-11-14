#include "sat/bsat/satSolver.h"
#include <stdio.h>
extern void Wlc_BlastFullAdderCtrlCnf(sat_solver*,int,int,int,int,int*,int*,int*);
int main(void){for(int out=0;out<4;out++){sat_solver*s=sat_solver_new();sat_solver_setnvars(s,7);int n=5,c,z;Wlc_BlastFullAdderCtrlCnf(s,1,2,3,4,&c,&z,&n);int a[6]={3,5,7,9,2*c+1-(out&1),2*z+1-((out>>1)&1)};int v=sat_solver_solve(s,a,a+6,0,0,0,0);printf("%d %d\n",out,v);sat_solver_delete(s);}return 0;}
