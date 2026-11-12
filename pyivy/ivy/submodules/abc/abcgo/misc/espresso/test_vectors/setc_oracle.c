#include "../../../../src/misc/espresso/espresso.h"
int bit_count[256];
void fatal(char *s){fprintf(stderr,"%s\n",s);exit(1);}
struct cube_struct cube,temp_cube_save;
struct cdata_struct cdata,temp_cdata_save;
int main(void){int trial,bit,i;
for(i=0;i<256;i++){int v=i;while(v){bit_count[i]+=v&1;v>>=1;}}
cube.num_vars=19;cube.num_binary_vars=17;cube.part_size=ALLOC(int,19);cube.part_size[17]=3;cube.part_size[18]=39;cube_setup();
for(trial=0;trial<50;trial++){pset a=new_cube(),b=new_cube(),cof=new_cube(),lower=new_cube(),result=new_cube();
for(bit=0;bit<cube.size;bit++){if((bit*7+trial*3)%11<5)set_insert(a,bit);if((bit*5+trial*7)%13<7)set_insert(b,bit);if((bit+trial)%5==0)set_insert(cof,bit);}
printf("%d %d %d %d %d %d\n",cdist(a,b),cdist01(a,b),cdist0(a,b),cactive(a),ccommon(a,b,cof),full_row(a,cof));
force_lower(lower,a,b);consensus(result,a,b);set_write(stdout,lower);set_write(stdout,result);
}return 0;}
