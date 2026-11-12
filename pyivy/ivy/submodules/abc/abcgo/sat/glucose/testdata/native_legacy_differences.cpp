#include "sat/glucose/Solver.h"
#include <stdio.h>
using namespace Gluco;
int main(){Solver a;a.newVar();a.setConfBudget(0);vec<Lit> as;printf("zeroBudget=%d\n",toInt(a.solveLimited(as)));Solver b;b.newVar();b.interrupt();printf("interruptedTrivial=%d\n",toInt(b.solveLimited(as)));Solver c;c.newVar(true,false);printf("sat=%d ",c.solve());printf("nonDecisionModel=%d\n",toInt(c.model[0]));}
