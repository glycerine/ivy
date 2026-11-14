#include "aig/gia/giaEmbed.c"

int main(void)
{
    int dims[] = {4,16,64}, d, seed, i;
    float eigen[64], v[64], result[64];
    for (d = 0; d < 3; ++d)
        for (seed = 0; seed < 8; ++seed)
        {
            int n = dims[d];
            for (i = 0; i < n; ++i)
            {
                eigen[i] = (i%2 ? -1.0f : 1.0f)/sqrt(n);
                v[i] = ((i*7+seed)%13-6)/8.0f;
            }
            Emb_ManVecOrthogonolizeOne(eigen,v,n,result);
            printf("%d %d",n,seed);
            for (i = 0; i < n; ++i)
            {
                unsigned bits;
                memcpy(&bits,result+i,sizeof(bits));
                printf(" %08x",bits);
            }
            puts("");
        }
    return 0;
}
