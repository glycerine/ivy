package espresso

import "fmt"

/*
 * Revision Control Information
 *
 * $Source$
 * $Author$
 * $Revision$
 * $Date$
 *
 */
//#include "espresso.h"

//ABC_NAMESPACE_IMPL_START


static pcube Gcube;
var Gminterm pset;

pset minterms(T)
pcover T;
{
    int size, var;
    register pcube last;

    size = 1;
    for(var = 0; var < cube.num_vars; var++)
    size *= cube.part_size[var];
    // inlining of macro set_new
    Gminterm = set_clear(ALLOC(unsigned int, SET_SIZE(size)), size);

    // inlining of macro foreach_set
    for(Gcube=T.data,last=Gcube+T.count*T.wsize;Gcube<last;Gcube+=T.wsize)
    explode(cube.num_vars-1, 0);

    return Gminterm;
}


void explode(var, z)
var var, z int32;
{
    int i, last = cube.last_part[var];
    for(i=cube.first_part[var], z *= cube.part_size[var]; i<=last; i++, z++)
    // inlining of macro is_in_set
    // inlining of macro WHICH_WORD
    // inlining of macro WHICH_BIT
    if Gcube[(((i) >> LOGBPI) + 1)] & (1 << ((i) & (BPI-1))) {
        if (var == 0)
        // inlining of macro set_insert
        // inlining of macro WHICH_WORD
        // inlining of macro WHICH_BIT
        (Gminterm[(((z) >> LOGBPI) + 1)] |= 1 << ((z) & (BPI-1)));
        else
        explode(var-1, z);
    }
}


var mapindex [16][16]int32 = [16][16]int32{
    {  0,  1,  3,  2,   16, 17, 19, 18,      80, 81, 83, 82,   64, 65, 67, 66},
    {  4,  5,  7,  6,   20, 21, 23, 22,      84, 85, 87, 86,   68, 69, 71, 70},
    { 12, 13, 15, 14,   28, 29, 31, 30,      92, 93, 95, 94,   76, 77, 79, 78},
    {  8,  9, 11, 10,   24, 25, 27, 26,      88, 89, 91, 90,   72, 73, 75, 74},

    { 32, 33, 35, 34,   48, 49, 51, 50,     112,113,115,114,   96, 97, 99, 98},
    { 36, 37, 39, 38,   52, 53, 55, 54,     116,117,119,118,  100,101,103,102},
    { 44, 45, 47, 46,   60, 61, 63, 62,     124,125,127,126,  108,109,111,110},
    { 40, 41, 43, 42,   56, 57, 59, 58,     120,121,123,122,  104,105,107,106},


    {160,161,163,162,  176,177,179,178,     240,241,243,242,  224,225,227,226},
    {164,165,167,166,  180,181,183,182,     244,245,247,246,  228,229,231,230},
    {172,173,175,174,  188,189,191,190,     252,253,255,254,  236,237,239,238},
    {168,169,171,170,  184,185,187,186,     248,249,251,250,  232,233,235,234},

    {128,129,131,130,  144,145,147,146,     208,209,211,210,  192,193,195,194},
    {132,133,135,134,  148,149,151,150,     212,213,215,214,  196,197,199,198},
    {140,141,143,142,  156,157,159,158,     220,221,223,222,  204,205,207,206},
    {136,137,139,138,  152,153,155,154,     216,217,219,218,  200,201,203,202},
};

// Original macro:
// #define POWER2(n) (1 << n)

void map(T)
pcover T;
{
    int j, k, l, other_input_offset, output_offset, outnum, ind;
    int largest_input_ind,  numout;
    char c;
    pset m;
    bool some_output;

    m = minterms(T);
    // inlining of macro POWER2
    largest_input_ind = (1 << cube.num_binary_vars);
    numout = cube.part_size[cube.num_vars-1];

    for outnum = 0; outnum < numout; outnum++ {
    output_offset = outnum * largest_input_ind;
    fmt.Printf("\n\nOutput space # %d\n", outnum);
    for l = 0; l <= MAX(cube.num_binary_vars - 8, 0); l++ {
        other_input_offset = l * 256;
        for k = 0; k < 16; k++ {
        some_output = FALSE;
        for j = 0; j < 16; j++ {
            ind = mapindex[k][j] + other_input_offset;
            if ind < largest_input_ind {
            // inlining of macro is_in_set
            // inlining of macro WHICH_WORD
            // inlining of macro WHICH_BIT
            c = (m[(((ind+output_offset) >> LOGBPI) + 1)] & (1 << ((ind+output_offset) & (BPI-1)))) ? '1' : '.';
            putchar(c);
            some_output = TRUE;
            }
            if ((j+1)%4 == 0)
            putchar(' ');
            if ((j+1)%8 == 0)
            fmt.Printf("  ");
        }
        if (some_output)
            putchar('\n');
        if (k+1)%4 == 0 {
            if (k != 15 && mapindex[k+1][0] >= largest_input_ind)
            break;
            putchar('\n');
        }
        if ((k+1)%8 == 0)
            putchar('\n');
        }
    }
    }
    // inlining of macro set_free
    FREE(m);
}
//ABC_NAMESPACE_IMPL_END

