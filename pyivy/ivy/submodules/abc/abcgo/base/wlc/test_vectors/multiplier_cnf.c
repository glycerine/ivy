#include "sat/bsat/satSolver.h"
#include <stdio.h>
extern sat_solver *Wlc_BlastMultiplierCnfMain(int);
int main(void){for(int n=1;n<=3;n++){sat_solver*s=Wlc_BlastMultiplierCnfMain(n);int st=sat_solver_solve(s,NULL,NULL,0,0,0,0);printf("%d %d\n",n,st);sat_solver_delete(s);}return 0;}
