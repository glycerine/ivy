"""Compile unchanged, defined DSD2 primitives from the unfinished C draft."""
from pathlib import Path
import random,subprocess,tempfile,json
root=Path(__file__).resolve().parents[4]
source=(root/'src/opt/dau/dauDsd2.c').read_text()
names=['Abc_Tt6HalfUnShuffleVars','Dua_DsdTryConst','Dua_DsdCheckCof0Const0','Dua_DsdCheckCofsEqualNot']
functions=[]
for name in names:
 location=source.index(name+'(') if name+'(' in source else source.index(name+'(')
 start=source.rfind('\n',0,location)+1
 brace=source.index('{',location);depth=1;end=brace+1
 while depth:
  if source[end]=='{':depth+=1
  if source[end]=='}':depth-=1
  end+=1
 functions.append(source[start:end])
build=Path(tempfile.mkdtemp(prefix='dau-dsd2-primitives-'))
main=build/'main.c'
main.write_text('#include "opt/dau/dauInt.h"\n#include "misc/util/utilTruth.h"\n'+'\n'.join(functions)+r'''
int main(){int n,v,select;word truth[64],copy[64];unsigned long long w;while(scanf("%d %d %d",&n,&v,&select)==3){int words=Abc_TtWordNum(n);for(int i=0;i<words;i++){if(scanf("%llx",&w)!=1)return 2;truth[i]=copy[i]=w;}int c=Dua_DsdTryConst(copy,n);printf("%llu %d %d %d\n",(unsigned long long)Abc_Tt6HalfUnShuffleVars(truth[0],v%6,select),Dua_DsdCheckCof0Const0(truth,words,v),Dua_DsdCheckCofsEqualNot(truth,words,v),c);}return 0;}
''')
subprocess.run(['cc','-O2','-DABC_USE_STDINT_H','-I',str(root/'src'),str(main),'-o',str(build/'oracle')],capture_output=True,check=True)
rng=random.Random(931);cases=[]
for n in range(1,13):
 for trial in range(50):cases.append(dict(n=n,v=rng.randrange(n),selector=trial%2,truth=[rng.getrandbits(64) for _ in range(1<<max(0,n-6))]))
text=''.join(f"{c['n']} {c['v']} {c['selector']} "+' '.join(f'{w:x}' for w in c['truth'])+'\n' for c in cases)
output=subprocess.run([str(build/'oracle')],input=text,capture_output=True,text=True,check=True)
for case,line in zip(cases,output.stdout.splitlines(),strict=True):case['half'],case['zero'],case['opposite'],case['complement']=map(int,line.split())
Path(__file__).with_name('dsd2_primitives.json').write_text(json.dumps(cases,separators=(',',':'))+'\n')
print('wrote',len(cases),'primitive vectors')
