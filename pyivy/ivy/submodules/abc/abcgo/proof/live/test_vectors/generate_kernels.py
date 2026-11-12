"""Generate real original-C absorber, combination, and arena graph traces."""
from pathlib import Path
import subprocess,tempfile,json
root=Path(__file__).resolve().parents[4]
build=Path(tempfile.mkdtemp(prefix='live-kernel-oracle-'));main=build/'main.c'
main.write_text(r'''#include "aig/aig/aig.h"
#include "aig/saig/saig.h"
extern Aig_Man_t *createNewAigWith0LivePo(Aig_Man_t*,Vec_Ptr_t*,int*);
extern Aig_Man_t *introduceAbsorberLogic(Aig_Man_t*,int*,int*,int);
extern Aig_Man_t *generateGeneralDisjunctiveTester(void*,Aig_Man_t*,int);
extern Aig_Man_t *createNewAigWith0LivePoWithDSC(Aig_Man_t*,Vec_Ptr_t*,int*,int,int,int,Vec_Ptr_t*);
struct aigPoIndices{int pending,begin,end,invariant;};
struct monotoneVectorsStruct{Vec_Int_t*known,*candidates,*hints;};
struct antecedentConsequentVectorsStruct{Vec_Int_t*antecedents,*candidates;};
extern Aig_Man_t *createMonotoneTester(Aig_Man_t*,struct aigPoIndices*,struct monotoneVectorsStruct*,int*);
extern Aig_Man_t *createDisjunctiveMonotoneTester(Aig_Man_t*,struct aigPoIndices*,struct antecedentConsequentVectorsStruct*,int*);

static void trace(Aig_Man_t*p,const char*kind,int depth){Aig_Obj_t*o,*li,*lo;int i,f,pattern;printf("{\"kind\":\"%s\",\"depth\":%d,\"pis\":%d,\"pos\":%d,\"regs\":%d,\"traces\":[",kind,depth,Saig_ManPiNum(p),Saig_ManPoNum(p),Saig_ManRegNum(p));for(pattern=0;pattern<64;pattern++){if(pattern)printf(",");printf("\"");Saig_ManForEachLo(p,o,i)o->fMarkA=0;for(f=0;f<6;f++){Aig_ManConst1(p)->fMarkA=1;Saig_ManForEachPi(p,o,i)o->fMarkA=(pattern>>(f+i))&1;Aig_ManForEachNode(p,o,i)o->fMarkA=(Aig_ObjFanin0(o)->fMarkA^Aig_ObjFaninC0(o))&(Aig_ObjFanin1(o)->fMarkA^Aig_ObjFaninC1(o));Aig_ManForEachCo(p,o,i)o->fMarkA=Aig_ObjFanin0(o)->fMarkA^Aig_ObjFaninC0(o);Saig_ManForEachPo(p,o,i)printf("%d",o->fMarkA);Saig_ManForEachLiLo(p,li,lo,i)lo->fMarkA=li->fMarkA;}printf("\"");}puts("]}");}
int main(){
 Aig_Man_t*p=Aig_ManStart(30);p->pName=Abc_UtilStrsav("fixture");Aig_Obj_t*u=Aig_ObjCreateCi(p);Aig_ObjCreateCi(p);Aig_ObjCreateCo(p,u);Aig_ObjCreateCo(p,u);Aig_ManSetRegNum(p,1);Vec_Ptr_t*v=Vec_PtrAlloc(1);Vec_PtrPush(v,u);int zero=-1,index=-1;p=createNewAigWith0LivePo(p,v,&zero);for(int depth=1;depth<=3;depth++){p=introduceAbsorberLogic(p,&zero,&index,depth-1);trace(p,"absorber",depth);}
 p=Aig_ManStart(30);p->pName=Abc_UtilStrsav("fixture");u=Aig_ObjCreateCi(p);Aig_ObjCreateCi(p);Aig_ObjCreateCi(p);Aig_ObjCreateCo(p,u);Aig_ObjCreateCo(p,u);Aig_ObjCreateCo(p,Aig_Not(u));Aig_ManSetRegNum(p,2);p=generateGeneralDisjunctiveTester(NULL,p,2);trace(p,"combination",2);
 p=Aig_ManStart(30);p->pName=Abc_UtilStrsav("fixture");u=Aig_ObjCreateCi(p);Aig_ObjCreateCi(p);Aig_ObjCreateCo(p,u);Aig_ObjCreateCo(p,Aig_ManConst1(p));Aig_ObjCreateCo(p,Aig_ManConst1(p));Aig_ObjCreateCo(p,Aig_ManConst1(p));Aig_ObjCreateCo(p,u);Aig_ObjCreateCo(p,Aig_Not(u));Aig_ObjCreateCo(p,u);Aig_ManSetRegNum(p,1);v=Vec_PtrAlloc(1);Vec_PtrPush(v,u);Vec_Ptr_t*barriers=Vec_PtrAlloc(2);Vec_Int_t*b=Vec_IntAlloc(1);Vec_IntPush(b,4);Vec_PtrPush(barriers,b);b=Vec_IntAlloc(1);Vec_IntPush(b,5);Vec_PtrPush(barriers,b);p=createNewAigWith0LivePoWithDSC(p,v,&zero,1,2,3,barriers);trace(p,"arena",2);
 p=Aig_ManStart(30);p->pName=Abc_UtilStrsav("fixture");u=Aig_ObjCreateCi(p);Aig_Obj_t*q=Aig_ObjCreateCi(p);Aig_ObjCreateCo(p,Aig_ManConst1(p));Aig_ObjCreateCo(p,q);Aig_ObjCreateCo(p,Aig_ManConst0(p));Aig_ObjCreateCo(p,Aig_Not(q));Aig_ObjCreateCo(p,Aig_ManConst1(p));Aig_ObjCreateCo(p,Aig_Or(p,u,q));Aig_ManSetRegNum(p,1);Vec_Int_t*hints=Vec_IntAlloc(2);Vec_IntPush(hints,1);Vec_IntPush(hints,3);struct aigPoIndices ids={0,1,3,4};struct monotoneVectorsStruct mono={NULL,hints,hints};int start;Aig_Man_t*tester=createMonotoneTester(p,&ids,&mono,&start);trace(tester,"monotone",1);struct antecedentConsequentVectorsStruct disj={NULL,hints};tester=createDisjunctiveMonotoneTester(p,&ids,&disj,&start);trace(tester,"disjunctive",1);
 return 0;}
''')
files=[str(p) for p in (root/'src/aig/aig').glob('*.c')]
files +=[str(root/'src/proof/live'/name) for name in ['kLiveConstraints.c','kliveness.c','combination.c','arenaViolation.c','monotone.c','disjunctiveMonotone.c']]
# The source allocates one byte too few for its double-underscore display
# name. Repair only that buffer size; keep the graph construction unchanged.
monotone=(root/'src/proof/live/monotone.c').read_text()
assert monotone.count('strlen("_monotone") + 1')==1
(build/'monotone.c').write_text(monotone.replace('strlen("_monotone") + 1','strlen("_monotone") + 2'))
files=[str(build/'monotone.c') if file==str(root/'src/proof/live/monotone.c') else file for file in files]

subprocess.run(['cc','-w','-O1','-DABC_USE_STDINT_H','-ffunction-sections','-fdata-sections','-Wl,--gc-sections','-I',str(root/'src'),str(main),*files,'-lm','-o',str(build/'oracle')],check=True,capture_output=True)
result=subprocess.run([str(build/'oracle')],check=True,capture_output=True,text=True)
cases=[json.loads(line) for line in result.stdout.splitlines() if line.startswith('{')]
Path(__file__).with_name('kernel_traces.json').write_text(json.dumps(cases,separators=(',',':'))+'\n')
print('wrote',len(cases),'actual C kernel cases',build)
