#include "proof/ssw/sswPairs.c"
int main(int argc,char **argv) { Vec_Int_t *v=Vec_IntAlloc(argc),**c; int i,j; for(i=1;i<argc;i++) Vec_IntPush(v,atoi(argv[i])); c=Ssw_TransformPairsIntoTempClasses(v,12); for(i=0;i<12;i++) if(c[i]) {printf("%d:",i); for(j=0;j<Vec_IntSize(c[i]);j++)printf("%s%d",j?",":"",Vec_IntEntry(c[i],j));printf(";");} puts(""); }
