#include "sat/glucose/SimpSolver.h"
#include <stdio.h>
using namespace Gluco;
int main(){SimpSolver s;s.use_rcheck=true;for(int i=0;i<4;i++)s.newVar();for(int i=1;i<4;i++)s.setFrozen(i,true);s.addClause(mkLit(0),mkLit(1),mkLit(2));s.addClause(~mkLit(0),mkLit(1),mkLit(3));vec<Lit> as;for(int i=1;i<4;i++)as.push(~mkLit(i));printf("%d\n",s.solve(as));}
