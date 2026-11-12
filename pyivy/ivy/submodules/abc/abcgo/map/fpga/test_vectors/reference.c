/* Executes the unchanged original FPGA mapper. Only host-frame/progress services
   are supplied locally; cut enumeration, matching, references and timing are C. */
#include "map/fpga/fpgaInt.h"
static Fpga_LutLib_t library;
void * Abc_FrameReadLibLut(void) { return &library; }
ProgressBar * Extra_ProgressBarStart(FILE *f, int n) { return NULL; }
void Extra_ProgressBarUpdate_int(ProgressBar *p, int n, char *s) {}
void Extra_ProgressBarStop(ProgressBar *p) {}
int main(int argc,char **argv) {
 int k=atoi(argv[1]),kind=atoi(argv[2]),i;
 Fpga_Man_t *p; Fpga_Node_t *root,*a,*b;
 library.LutMax=k;
 for(i=1;i<=k;i++){library.pLutAreas[i]=1;library.pLutDelays[i][0]=1;}
 p=Fpga_ManCreate(6,2,0);
 p->pInputArrivals=calloc(6,sizeof(float));
 root=p->pInputs[0];
 if(kind==0) for(i=1;i<6;i++)root=Fpga_NodeAnd(p,root,p->pInputs[i]);
 else {
  a=Fpga_NodeExor(p,p->pInputs[0],p->pInputs[1]);
  b=Fpga_NodeAnd(p,p->pInputs[2],p->pInputs[3]);
  root=Fpga_NodeMux(p,p->pInputs[4],a,b);
  root=Fpga_NodeOr(p,root,p->pInputs[5]);
 }
 p->pOutputs[0]=root; p->pOutputs[1]=p->pInputs[0];
 if(!Fpga_Mapping(p))return 2;
 printf("RESULT %.6g %.6g %d\n",p->fAreaGlo,Fpga_TimeComputeArrivalMax(p),p->vMapping->nSize);
 Fpga_ManFree(p);return 0;
}
clock_t s_MappingTime; int s_MappingMem; unsigned enable_dbg_outs=1;
int Abc_FrameIsBridgeMode(void){return 0;}
int Gia_ManToBridgeText(FILE *f,int n,unsigned char *b){return fwrite(b,1,n,f);}
char *vnsprintf(const char *format,va_list args){va_list copy;int n;char *b;va_copy(copy,args);n=vsnprintf(NULL,0,format,copy);va_end(copy);b=malloc(n+1);vsnprintf(b,n+1,format,args);return b;}
