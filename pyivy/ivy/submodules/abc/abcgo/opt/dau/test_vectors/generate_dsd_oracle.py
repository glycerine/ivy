"""Generate exact DSD outputs with the unchanged original C implementation."""
from pathlib import Path
import subprocess,tempfile,random,json
root=Path(__file__).resolve().parents[4]
build=Path(tempfile.mkdtemp(prefix='dau-dsd-oracle-'))
main=build/'main.c'
main.write_text(r'''#include "opt/dau/dauDsd.c"
#include "opt/dau/dauMerge.c"
int main(){int n,split,write,leveled,levels[16];unsigned long long w;word truth[1024];char result[2000];while(scanf("%d %d %d %d",&n,&split,&write,&leveled)==4){for(int i=0;i<Abc_TtWordNum(n);i++){if(scanf("%llx",&w)!=1)return 2;truth[i]=w;}if(leveled)for(int i=0;i<n;i++)if(scanf("%d",&levels[i])!=1)return 3;word original=truth[0];int size=leveled?Dau_DsdDecomposeLevel(truth,n,split,write,result,levels):Dau_DsdDecompose(truth,n,split,write,result);printf("%d %s %s\n",size,result,n<=6?Dau_DsdPerform(original):"-");}return 0;}''')
subprocess.run(['cc','-O2','-DABC_USE_STDINT_H','-ffunction-sections','-fdata-sections','-Wl,--gc-sections','-I',str(root/'src'),str(main),'-o',str(build/'oracle')],check=True,capture_output=True)
rng=random.Random(77127);cases=[]
for n in range(1,11):
 for trial in range(80):
  words=[rng.getrandbits(64) for _ in range(1<<max(n-6,0))]
  if n<6:
   value=words[0]&((1<<(1<<n))-1)
   for shift in range(0,64,1<<n):words[0]=(words[0] if shift else 0)|(value<<shift)
  cases.append(dict(n=n,truth=words,split=trial%2,write=1,levels=[rng.randrange(4) for _ in range(n)] if trial%4==1 else []))
text=''.join(f"{c['n']} {c['split']} {c['write']} {int(bool(c['levels']))} "+' '.join(f'{w:x}' for w in c['truth'])+' '+' '.join(map(str,c['levels']))+'\n' for c in cases)
result=subprocess.run([str(build/'oracle')],input=text,capture_output=True,text=True,check=True)
for case,line in zip(cases,result.stdout.splitlines(),strict=True):
 status,formula,perform=line.split(' ');case.update(status=int(status),formula=formula,perform=perform)
Path(__file__).with_name('dsd_oracle.json').write_text(json.dumps(cases,separators=(',',':'))+'\n')
print('wrote',len(cases),'cases from',build)
