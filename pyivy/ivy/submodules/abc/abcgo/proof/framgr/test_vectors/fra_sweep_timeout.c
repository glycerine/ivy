#include "proof/fra/fra.h"
#include <stdio.h>
static void cb(void *v,void *ignored){Fra_Man_t*p=v;printf("callback failures=%d logged=%d\n",p->nSatFailsReal,p->vTimeouts?Vec_PtrSize(p->vTimeouts):0);}
static Aig_Obj_t *parity(Aig_Man_t*p,Aig_Obj_t**a,int n){if(n==1)return a[0];int m=n/2;return Aig_Exor(p,parity(p,a,m),parity(p,a+m,n-m));}
int main(void){Aig_Man_t*g=Aig_ManStart(256),*out;Aig_Obj_t*v[8],*left;Fra_Par_t p;for(int i=0;i<8;i++)v[i]=Aig_ObjCreateCi(g);left=v[0];for(int i=1;i<8;i++)left=Aig_Exor(g,left,v[i]);Aig_ObjCreateCo(g,left);Aig_ObjCreateCo(g,parity(g,v,8));Fra_ParamsDefault(&p);p.nBTLimitNode=1;p.fConeBias=0;p.nSimWords=1;g->pImpFunc=cb;out=Fra_FraigPerform(g,&p);printf("done nodes=%d\n",Aig_ManNodeNum(out));return 0;}
