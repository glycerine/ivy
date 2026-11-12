"""Generate parser/fragment fixtures from unchanged original C functions."""
from pathlib import Path
import subprocess,tempfile,random,json,re
root=Path(__file__).resolve().parents[4]
source=(root/'src/proof/live/ltl_parser.c').read_text()
functions=[]
for name in ['generateTypedNode','isNotVarNameSymbol','getVarName','isUnexpectedEOS','isTemporalOperator','readLtlFormula','resetGlobalVar','parseFormulaCreateAST','isNonTemporalSubformula','isWellFormed']:
 match=re.search(r'^.*\b'+name+r'\([^;]*?\)\s*\{',source,re.M)
 if not match:raise ValueError(name)
 start=match.start();opening=source.index('{',match.start());depth=1;end=opening+1
 while depth:
  if source[end]=='{':depth+=1
  if source[end]=='}':depth-=1
  end+=1
 functions.append(source[start:end])
build=Path(tempfile.mkdtemp(prefix='live-ltl-oracle-'))
prelude='''#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <assert.h>
enum ltlToken {AND,OR,NOT,IMPLY,GLOBALLY,EVENTUALLY,NEXT,UNTIL,BOOL};
typedef enum ltlToken tokenType;
typedef struct ltlNode_t {tokenType type;char *name;void*pObj;struct ltlNode_t *left,*right;}ltlNode;
int startOfSuffixString=0;
'''
main='''
void show(ltlNode*n){const char*op[]={"&","+","!",">",":G",":F",":X",":U"};if(n->type==BOOL){printf("%s",n->name);return;}printf("%s ",op[n->type]);show(n->left);if(n->right){printf(" ");show(n->right);}}
int main(){char formula[4096];while(fgets(formula,sizeof(formula),stdin)){formula[strcspn(formula,"\\n")]=0;ltlNode*n=parseFormulaCreateAST(formula);if(!n)return 2;printf("%d ",isWellFormed(n));show(n);puts("");}return 0;}
'''
(build/'main.c').write_text(prelude+'\n'.join(functions)+main)
subprocess.run(['cc','-O2',str(build/'main.c'),'-o',str(build/'oracle')],check=True,capture_output=True)
rng=random.Random(831)
def formula(depth):
 if depth==0 or rng.randrange(4)==0:return rng.choice(['p','q','ready','pending_1','true','false'])
 operator=rng.choice(['&','+','!',':G',':F',':X',':U'])
 return operator+' '+formula(depth-1)+(' '+formula(depth-1) if operator in ['&','+',':U'] else '')
cases=[dict(formula=formula(4)) for _ in range(200)]
run=subprocess.run([str(build/'oracle')],input='\n'.join(c['formula'] for c in cases)+'\n',check=True,capture_output=True,text=True)
for case,line in zip(cases,run.stdout.splitlines(),strict=True):
 well,tree=line.split(' ',1);case.update(well=int(well),tree=tree)
Path(__file__).with_name('ltl_oracle.json').write_text(json.dumps(cases,separators=(',',':'))+'\n')
print('wrote',len(cases),'original-C parser vectors')
