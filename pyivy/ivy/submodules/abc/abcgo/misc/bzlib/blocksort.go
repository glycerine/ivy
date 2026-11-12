package bzlib


/*-------------------------------------------------------------*/
/*--- Block sorting machinery                               ---*/
/*---                                           blocksort.c ---*/
/*-------------------------------------------------------------*/

/* ------------------------------------------------------------------
   This file is part of bzip2/libbzip2, a program and library for
   lossless, block-sorting data compression.

   bzip2/libbzip2 version 1.0.5 of 10 December 2007
   Copyright (C) 1996-2007 Julian Seward <jseward@bzip.org>

   Please read the WARNING, DISCLAIMER and PATENTS sections in the 
   README file.

   This program is released under the terms of the license contained
   in the file LICENSE.
   ------------------------------------------------------------------ */


//#include "bzlib_private.h"

//ABC_NAMESPACE_IMPL_START

/*---------------------------------------------*/
/*--- Fallback O(N log(N)^2) sorting        ---*/
/*--- algorithm, for repetitive blocks      ---*/
/*---------------------------------------------*/

/*---------------------------------------------*/
static 
//__inline__
func fallbackSimpleSort(fmap *UInt32, eclass *UInt32, lo Int32, hi Int32) {
   var i, j, tmp Int32;
   var ec_tmp UInt32;

   if lo == hi { return; }

   if hi - lo > 3 {
      for  i = hi-4; i >= lo; i--  {
         tmp = fmap[i];
         ec_tmp = eclass[tmp];
         for  j = i+4; j <= hi && ec_tmp > eclass[fmap[j]]; j += 4  {
            fmap[j-4] = fmap[j];
         }
         fmap[j-4] = tmp;
      }
   }

   for  i = hi-1; i >= lo; i--  {
      tmp = fmap[i];
      ec_tmp = eclass[tmp];
      for  j = i+1; j <= hi && ec_tmp > eclass[fmap[j]]; j++  {
         fmap[j-1] = fmap[j];
      }
      fmap[j-1] = tmp;
   }
}


/*---------------------------------------------*/
// Original macro:
// #define fswap(zz1, zz2) \
//    { Int32 zztmp = zz1; zz1 = zz2; zz2 = zztmp; }


// Original macro:
// #define fvswap(zzp1, zzp2, zzn)       \
// {                                     \
//    Int32 yyp1 = (zzp1);               \
//    Int32 yyp2 = (zzp2);               \
//    Int32 yyn  = (zzn);                \
//    while (yyn > 0) {                  \
//       fswap(fmap[yyp1], fmap[yyp2]);  \
//       yyp1++; yyp2++; yyn--;          \
//    }                                  \
// }



// Original macro:
// #define fmin(a,b) ((a) < (b)) ? (a) : (b)


// Original macro:
// #define fpush(lz,hz) { stackLo[sp] = lz; \
//                        stackHi[sp] = hz; \
//                        sp++; }


// Original macro:
// #define fpop(lz,hz) { sp--;              \
//                       lz = stackLo[sp];  \
//                       hz = stackHi[sp]; }


const FALLBACK_QSORT_SMALL_THRESH = 10
const FALLBACK_QSORT_STACK_SIZE =   100


static
func fallbackQSort3(fmap *UInt32, eclass *UInt32, loSt Int32, hiSt Int32) {
   var unLo, unHi, ltLo, gtHi, n, m Int32;
   var sp, lo, hi Int32;
   var med, r, r3 UInt32;
   var stackLo [FALLBACK_QSORT_STACK_SIZE]Int32;
   var stackHi [FALLBACK_QSORT_STACK_SIZE]Int32;

   r = 0;

   sp = 0;
   // inlining of macro fpush
   { stackLo[sp] = loSt;                        stackHi[sp] = hiSt;                        sp++; };

   for sp > 0 {

      AssertH ( sp < FALLBACK_QSORT_STACK_SIZE - 1, 1004 );

      // inlining of macro fpop
      { sp--;                                    lo = stackLo[sp];                        hi = stackHi[sp]; };
      if hi - lo < FALLBACK_QSORT_SMALL_THRESH {
         fallbackSimpleSort ( fmap, eclass, lo, hi );
         continue;
      }

      /* Random partitioning.  Median of 3 sometimes fails to
         avoid bad cases.  Median of 9 seems to help but 
         looks rather expensive.  This too seems to work but
         is cheaper.  Guidance for the magic constants 
         7621 and 32768 is taken from Sedgewick's algorithms
         book, chapter 35.
      */
      r = ((r * 7621) + 1) % 32768;
      r3 = r % 3;
      if r3 == 0 { med = eclass[fmap[lo]]; } else
      if r3 == 1 { med = eclass[fmap[(lo+hi)>>1]]; } else {
                   med = eclass[fmap[hi]]; }

      unLo = ltLo = lo;
      unHi = gtHi = hi;

      for {
         for {
            if unLo > unHi { break; }
            n = (Int32)eclass[fmap[unLo]] - (Int32)med;
            if n == 0 { 
               // inlining of macro fswap
               { var zztmp Int32  = fmap[unLo]; fmap[unLo] = fmap[ltLo]; fmap[ltLo] = zztmp; }; 
               ltLo++; unLo++; 
               continue; 
            };
            if n > 0 { break; }
            unLo++;
         }
         for {
            if unLo > unHi { break; }
            n = (Int32)eclass[fmap[unHi]] - (Int32)med;
            if n == 0 { 
               // inlining of macro fswap
               { var zztmp Int32  = fmap[unHi]; fmap[unHi] = fmap[gtHi]; fmap[gtHi] = zztmp; }; 
               gtHi--; unHi--; 
               continue; 
            };
            if n < 0 { break; }
            unHi--;
         }
         if unLo > unHi { break; }
         // inlining of macro fswap
         { var zztmp Int32  = fmap[unLo]; fmap[unLo] = fmap[unHi]; fmap[unHi] = zztmp; }; unLo++; unHi--;
      }

      AssertD ( unHi == unLo-1, "fallbackQSort3(2)" );

      if gtHi < ltLo { continue; }

      // inlining of macro fmin
      
      if (ltLo-lo) < (unLo-ltLo) {
          n = (ltLo-lo);
      } else {
          n = (unLo-ltLo);
      // inlining of macro fvswap
      // inlining of macro fswap
      } {                                        var yyp1 Int32  = (lo);                  var yyp2 Int32  = (unLo-n);                  var yyn Int32   = (n);                   for yyn > 0 {                        { var zztmp Int32  = fmap[yyp1]; fmap[yyp1] = fmap[yyp2]; fmap[yyp2] = zztmp; };        yyp1++; yyp2++; yyn--;             }                                  };
      // inlining of macro fmin
      
      if (hi-gtHi) < (gtHi-unHi) {
          m = (hi-gtHi);
      } else {
          m = (gtHi-unHi);
      // inlining of macro fvswap
      // inlining of macro fswap
      } {                                        var yyp1 Int32  = (unLo);                  var yyp2 Int32  = (hi-m+1);                  var yyn Int32   = (m);                   for yyn > 0 {                        { var zztmp Int32  = fmap[yyp1]; fmap[yyp1] = fmap[yyp2]; fmap[yyp2] = zztmp; };        yyp1++; yyp2++; yyn--;             }                                  };

      n = lo + unLo - ltLo - 1;
      m = hi - (gtHi - unHi) + 1;

      if n - lo > hi - m {
         // inlining of macro fpush
         { stackLo[sp] = lo;                        stackHi[sp] = n;                        sp++; };
         // inlining of macro fpush
         { stackLo[sp] = m;                        stackHi[sp] = hi;                        sp++; };
      } else {
         // inlining of macro fpush
         { stackLo[sp] = m;                        stackHi[sp] = hi;                        sp++; };
         // inlining of macro fpush
         { stackLo[sp] = lo;                        stackHi[sp] = n;                        sp++; };
      }
   }
}

#undef fmin
#undef fpush
#undef fpop
#undef fswap
#undef fvswap
#undef FALLBACK_QSORT_SMALL_THRESH
#undef FALLBACK_QSORT_STACK_SIZE


/*---------------------------------------------*/
/* Pre:
      nblock > 0
      eclass exists for [0 .. nblock-1]
      ((UChar*)eclass) [0 .. nblock-1] holds block
      ptr exists for [0 .. nblock-1]

   Post:
      ((UChar*)eclass) [0 .. nblock-1] holds block
      All other areas of eclass destroyed
      fmap [0 .. nblock-1] holds sorted order
      bhtab [ 0 .. 2+(nblock/32) ] destroyed
*/

// Original macro:
// #define       SET_BH(zz)  bhtab[(zz) >> 5] |= (1 << ((zz) & 31))

// Original macro:
// #define     CLEAR_BH(zz)  bhtab[(zz) >> 5] &= ~(1 << ((zz) & 31))

// Original macro:
// #define     ISSET_BH(zz)  (bhtab[(zz) >> 5] & (1 << ((zz) & 31)))

// Original macro:
// #define      WORD_BH(zz)  bhtab[(zz) >> 5]

// Original macro:
// #define UNALIGNED_BH(zz)  ((zz) & 0x01f)


static
func fallbackSort(fmap *UInt32, eclass *UInt32, bhtab *UInt32, nblock Int32, verb Int32) {
   var ftab [257]Int32;
   var ftabCopy [256]Int32;
   var H, i, j, k, l, r, cc, cc1 Int32;
   var nNotDone Int32;
   var nBhtab Int32;
   var eclass8 *UChar  = (UChar*)eclass;

   /*--
      Initial 1-char radix sort to generate
      initial fmap and initial BH bits.
   --*/
   if verb >= 4 {
      VPrintf0 ( "        bucket sorting ...\n" ); }
   for i = 0; i < 257;    i++ { ftab[i] = 0; }
   for i = 0; i < nblock; i++ { ftab[eclass8[i]]++; }
   for i = 0; i < 256;    i++ { ftabCopy[i] = ftab[i]; }
   for i = 1; i < 257;    i++ { ftab[i] += ftab[i-1]; }

   for i = 0; i < nblock; i++ {
      j = eclass8[i];
      k = ftab[j] - 1;
      ftab[j] = k;
      fmap[k] = i;
   }

   nBhtab = 2 + (nblock / 32);
   for i = 0; i < nBhtab; i++ { bhtab[i] = 0; }
   // inlining of macro SET_BH
   for i = 0; i < 256; i++ { bhtab[(ftab[i]) >> 5] |= (1 << ((ftab[i]) & 31)); }

   /*--
      Inductively refine the buckets.  Kind-of an
      "exponential radix sort" (!), inspired by the
      Manber-Myers suffix array construction algorithm.
   --*/

   /*-- set sentinel bits for block-end detection --*/
   for i = 0; i < 32; i++ { 
      // inlining of macro SET_BH
      bhtab[(nblock + 2*i) >> 5] |= (1 << ((nblock + 2*i) & 31));
      // inlining of macro CLEAR_BH
      bhtab[(nblock + 2*i + 1) >> 5] &= ~(1 << ((nblock + 2*i + 1) & 31));
   }

   /*-- the log(N) loop --*/
   H = 1;
   for {

      if verb >= 4 { 
         VPrintf1 ( "        depth %6d has ", H ); }

      j = 0;
      for i = 0; i < nblock; i++ {
         // inlining of macro ISSET_BH
         if bhtab[(i) >> 5] & (1 << ((i) & 31)) { j = i; }
         k = fmap[i] - H; if k < 0 { k += nblock; }
         eclass[k] = j;
      }

      nNotDone = 0;
      r = -1;
      for {

     /*-- find the next non-singleton bucket --*/
         k = r + 1;
         // inlining of macro ISSET_BH
         // inlining of macro UNALIGNED_BH
         for (bhtab[(k) >> 5] & (1 << ((k) & 31))) && ((k) & 0x01f) { k++; }
         // inlining of macro ISSET_BH
         if bhtab[(k) >> 5] & (1 << ((k) & 31)) {
            // inlining of macro WORD_BH
            for bhtab[(k) >> 5] == 0xffffffff { k += 32; }
            // inlining of macro ISSET_BH
            for bhtab[(k) >> 5] & (1 << ((k) & 31)) { k++; }
         }
         l = k - 1;
         if l >= nblock { break; }
         // inlining of macro ISSET_BH
         // inlining of macro UNALIGNED_BH
         for !(bhtab[(k) >> 5] & (1 << ((k) & 31))) && ((k) & 0x01f) { k++; }
         // inlining of macro ISSET_BH
         if !(bhtab[(k) >> 5] & (1 << ((k) & 31))) {
            // inlining of macro WORD_BH
            for bhtab[(k) >> 5] == 0x00000000 { k += 32; }
            // inlining of macro ISSET_BH
            for !(bhtab[(k) >> 5] & (1 << ((k) & 31))) { k++; }
         }
         r = k - 1;
         if r >= nblock { break; }

         /*-- now [l, r] bracket current bucket --*/
         if r > l {
            nNotDone += (r - l + 1);
            fallbackQSort3 ( fmap, eclass, l, r );

            /*-- scan bucket and generate header bits-- */
            cc = -1;
            for i = l; i <= r; i++ {
               cc1 = eclass[fmap[i]];
               // inlining of macro SET_BH
               if cc != cc1 { bhtab[(i) >> 5] |= (1 << ((i) & 31)); cc = cc1; };
            }
         }
      }

      if verb >= 4 { 
         VPrintf1 ( "%6d unresolved strings\n", nNotDone ); }

      H *= 2;
      if H > nblock || nNotDone == 0 { break; }
   }

   /*-- 
      Reconstruct the original block in
      eclass8 [0 .. nblock-1], since the
      previous phase destroyed it.
   --*/
   if verb >= 4 {
      VPrintf0 ( "        reconstructing block ...\n" ); }
   j = 0;
   for i = 0; i < nblock; i++ {
      for ftabCopy[j] == 0 { j++; }
      ftabCopy[j]--;
      eclass8[fmap[i]] = (UChar)j;
   }
   AssertH ( j < 256, 1005 );
}

#undef       SET_BH
#undef     CLEAR_BH
#undef     ISSET_BH
#undef      WORD_BH
#undef UNALIGNED_BH


/*---------------------------------------------*/
/*--- The main, O(N^2 log(N)) sorting       ---*/
/*--- algorithm.  Faster for "normal"       ---*/
/*--- non-repetitive blocks.                ---*/
/*---------------------------------------------*/

/*---------------------------------------------*/
static
//__inline__
func mainGtU(i1 UInt32, i2 UInt32, block *UChar, quadrant *UInt16, nblock UInt32, budget *Int32) Bool {
   var k Int32;
   var c1, c2 UChar;
   var s1, s2 UInt16;

   AssertD ( i1 != i2, "mainGtU" );
   /* 1 */
   c1 = block[i1]; c2 = block[i2];
   if c1 != c2 { return (c1 > c2); }
   i1++; i2++;
   /* 2 */
   c1 = block[i1]; c2 = block[i2];
   if c1 != c2 { return (c1 > c2); }
   i1++; i2++;
   /* 3 */
   c1 = block[i1]; c2 = block[i2];
   if c1 != c2 { return (c1 > c2); }
   i1++; i2++;
   /* 4 */
   c1 = block[i1]; c2 = block[i2];
   if c1 != c2 { return (c1 > c2); }
   i1++; i2++;
   /* 5 */
   c1 = block[i1]; c2 = block[i2];
   if c1 != c2 { return (c1 > c2); }
   i1++; i2++;
   /* 6 */
   c1 = block[i1]; c2 = block[i2];
   if c1 != c2 { return (c1 > c2); }
   i1++; i2++;
   /* 7 */
   c1 = block[i1]; c2 = block[i2];
   if c1 != c2 { return (c1 > c2); }
   i1++; i2++;
   /* 8 */
   c1 = block[i1]; c2 = block[i2];
   if c1 != c2 { return (c1 > c2); }
   i1++; i2++;
   /* 9 */
   c1 = block[i1]; c2 = block[i2];
   if c1 != c2 { return (c1 > c2); }
   i1++; i2++;
   /* 10 */
   c1 = block[i1]; c2 = block[i2];
   if c1 != c2 { return (c1 > c2); }
   i1++; i2++;
   /* 11 */
   c1 = block[i1]; c2 = block[i2];
   if c1 != c2 { return (c1 > c2); }
   i1++; i2++;
   /* 12 */
   c1 = block[i1]; c2 = block[i2];
   if c1 != c2 { return (c1 > c2); }
   i1++; i2++;

   k = nblock + 8;

   for {
      /* 1 */
      c1 = block[i1]; c2 = block[i2];
      if c1 != c2 { return (c1 > c2); }
      s1 = quadrant[i1]; s2 = quadrant[i2];
      if s1 != s2 { return (s1 > s2); }
      i1++; i2++;
      /* 2 */
      c1 = block[i1]; c2 = block[i2];
      if c1 != c2 { return (c1 > c2); }
      s1 = quadrant[i1]; s2 = quadrant[i2];
      if s1 != s2 { return (s1 > s2); }
      i1++; i2++;
      /* 3 */
      c1 = block[i1]; c2 = block[i2];
      if c1 != c2 { return (c1 > c2); }
      s1 = quadrant[i1]; s2 = quadrant[i2];
      if s1 != s2 { return (s1 > s2); }
      i1++; i2++;
      /* 4 */
      c1 = block[i1]; c2 = block[i2];
      if c1 != c2 { return (c1 > c2); }
      s1 = quadrant[i1]; s2 = quadrant[i2];
      if s1 != s2 { return (s1 > s2); }
      i1++; i2++;
      /* 5 */
      c1 = block[i1]; c2 = block[i2];
      if c1 != c2 { return (c1 > c2); }
      s1 = quadrant[i1]; s2 = quadrant[i2];
      if s1 != s2 { return (s1 > s2); }
      i1++; i2++;
      /* 6 */
      c1 = block[i1]; c2 = block[i2];
      if c1 != c2 { return (c1 > c2); }
      s1 = quadrant[i1]; s2 = quadrant[i2];
      if s1 != s2 { return (s1 > s2); }
      i1++; i2++;
      /* 7 */
      c1 = block[i1]; c2 = block[i2];
      if c1 != c2 { return (c1 > c2); }
      s1 = quadrant[i1]; s2 = quadrant[i2];
      if s1 != s2 { return (s1 > s2); }
      i1++; i2++;
      /* 8 */
      c1 = block[i1]; c2 = block[i2];
      if c1 != c2 { return (c1 > c2); }
      s1 = quadrant[i1]; s2 = quadrant[i2];
      if s1 != s2 { return (s1 > s2); }
      i1++; i2++;

      if i1 >= nblock { i1 -= nblock; }
      if i2 >= nblock { i2 -= nblock; }

      k -= 8;
      (*budget)--;
   
if !(k >= 0) { break }
}
      

   return False;
}


/*---------------------------------------------*/
/*--
   Knuth's increments seem to work better
   than Incerpi-Sedgewick here.  Possibly
   because the number of elems to sort is
   usually small, typically <= 20.
--*/
var incs [14]Int32 = [14]Int32{ 1, 4, 13, 40, 121, 364, 1093, 3280,
                   9841, 29524, 88573, 265720,
                   797161, 2391484 };

static
func mainSimpleSort(ptr *UInt32, block *UChar, quadrant *UInt16, nblock Int32, lo Int32, hi Int32, d Int32, budget *Int32) {
   var i, j, h, bigN, hp Int32;
   var v UInt32;

   bigN = hi - lo + 1;
   if bigN < 2 { return; }

   hp = 0;
   for incs[hp] < bigN { hp++; }
   hp--;

   for ; hp >= 0; hp-- {
      h = incs[hp];

      i = lo + h;
      for True {

         /*-- copy 1 --*/
         if i > hi { break; }
         v = ptr[i];
         j = i;
         for  mainGtU ( 
                    ptr[j-h]+d, v+d, block, quadrant, nblock, budget 
                 )  {
            ptr[j] = ptr[j-h];
            j = j - h;
            if j <= (lo + h - 1) { break; }
         }
         ptr[j] = v;
         i++;

         /*-- copy 2 --*/
         if i > hi { break; }
         v = ptr[i];
         j = i;
         for  mainGtU ( 
                    ptr[j-h]+d, v+d, block, quadrant, nblock, budget 
                 )  {
            ptr[j] = ptr[j-h];
            j = j - h;
            if j <= (lo + h - 1) { break; }
         }
         ptr[j] = v;
         i++;

         /*-- copy 3 --*/
         if i > hi { break; }
         v = ptr[i];
         j = i;
         for  mainGtU ( 
                    ptr[j-h]+d, v+d, block, quadrant, nblock, budget 
                 )  {
            ptr[j] = ptr[j-h];
            j = j - h;
            if j <= (lo + h - 1) { break; }
         }
         ptr[j] = v;
         i++;

         if *budget < 0 { return; }
      }
   }
}


/*---------------------------------------------*/
/*--
   The following is an implementation of
   an elegant 3-way quicksort for strings,
   described in a paper "Fast Algorithms for
   Sorting and Searching Strings", by Robert
   Sedgewick and Jon L. Bentley.
--*/

// Original macro:
// #define mswap(zz1, zz2) \
//    { Int32 zztmp = zz1; zz1 = zz2; zz2 = zztmp; }


// Original macro:
// #define mvswap(zzp1, zzp2, zzn)       \
// {                                     \
//    Int32 yyp1 = (zzp1);               \
//    Int32 yyp2 = (zzp2);               \
//    Int32 yyn  = (zzn);                \
//    while (yyn > 0) {                  \
//       mswap(ptr[yyp1], ptr[yyp2]);    \
//       yyp1++; yyp2++; yyn--;          \
//    }                                  \
// }


static 
//__inline__
func mmed3(a UChar, b UChar, c UChar) UChar {
   var t UChar;
   if a > b { t = a; a = b; b = t; };
   if b > c { 
      b = c;
      if a > b { b = a; }
   }
   return b;
}

// Original macro:
// #define mmin(a,b) ((a) < (b)) ? (a) : (b)


// Original macro:
// #define mpush(lz,hz,dz) { stackLo[sp] = lz; \
//                           stackHi[sp] = hz; \
//                           stackD [sp] = dz; \
//                           sp++; }


// Original macro:
// #define mpop(lz,hz,dz) { sp--;             \
//                          lz = stackLo[sp]; \
//                          hz = stackHi[sp]; \
//                          dz = stackD [sp]; }



// Original macro:
// #define mnextsize(az) (nextHi[az]-nextLo[az])


// Original macro:
// #define mnextswap(az,bz)                                        \
//    { Int32 tz;                                                  \
//      tz = nextLo[az]; nextLo[az] = nextLo[bz]; nextLo[bz] = tz; \
//      tz = nextHi[az]; nextHi[az] = nextHi[bz]; nextHi[bz] = tz; \
//      tz = nextD [az]; nextD [az] = nextD [bz]; nextD [bz] = tz; }



const MAIN_QSORT_SMALL_THRESH = 20
#define MAIN_QSORT_DEPTH_THRESH (BZ_N_RADIX + BZ_N_QSORT)
const MAIN_QSORT_STACK_SIZE = 100

static
func mainQSort3(ptr *UInt32, block *UChar, quadrant *UInt16, nblock Int32, loSt Int32, hiSt Int32, dSt Int32, budget *Int32) {
   var unLo, unHi, ltLo, gtHi, n, m, med Int32;
   var sp, lo, hi, d Int32;

   var stackLo [MAIN_QSORT_STACK_SIZE]Int32;
   var stackHi [MAIN_QSORT_STACK_SIZE]Int32;
   var stackD [MAIN_QSORT_STACK_SIZE]Int32;

   var nextLo [3]Int32;
   var nextHi [3]Int32;
   var nextD [3]Int32;

   sp = 0;
   // inlining of macro mpush
   { stackLo[sp] = loSt;                           stackHi[sp] = hiSt;                           stackD [sp] = dSt;                           sp++; };

   for sp > 0 {

      AssertH ( sp < MAIN_QSORT_STACK_SIZE - 2, 1001 );

      // inlining of macro mpop
      { sp--;                                      lo = stackLo[sp];                          hi = stackHi[sp];                          d = stackD [sp]; };
      if hi - lo < MAIN_QSORT_SMALL_THRESH || 
          d > MAIN_QSORT_DEPTH_THRESH {
         mainSimpleSort ( ptr, block, quadrant, nblock, lo, hi, d, budget );
         if *budget < 0 { return; }
         continue;
      }

      med = (Int32) 
            mmed3 ( block[ptr[ lo         ]+d],
                    block[ptr[ hi         ]+d],
                    block[ptr[ (lo+hi)>>1 ]+d] );

      unLo = ltLo = lo;
      unHi = gtHi = hi;

      for True {
         for True {
            if unLo > unHi { break; }
            n = ((Int32)block[ptr[unLo]+d]) - med;
            if n == 0 { 
               // inlining of macro mswap
               { var zztmp Int32  = ptr[unLo]; ptr[unLo] = ptr[ltLo]; ptr[ltLo] = zztmp; }; 
               ltLo++; unLo++; continue; 
            };
            if n >  0 { break; }
            unLo++;
         }
         for True {
            if unLo > unHi { break; }
            n = ((Int32)block[ptr[unHi]+d]) - med;
            if n == 0 { 
               // inlining of macro mswap
               { var zztmp Int32  = ptr[unHi]; ptr[unHi] = ptr[gtHi]; ptr[gtHi] = zztmp; }; 
               gtHi--; unHi--; continue; 
            };
            if n <  0 { break; }
            unHi--;
         }
         if unLo > unHi { break; }
         // inlining of macro mswap
         { var zztmp Int32  = ptr[unLo]; ptr[unLo] = ptr[unHi]; ptr[unHi] = zztmp; }; unLo++; unHi--;
      }

      AssertD ( unHi == unLo-1, "mainQSort3(2)" );

      if gtHi < ltLo {
         // inlining of macro mpush
         { stackLo[sp] = lo;                           stackHi[sp] = hi;                           stackD [sp] = d+1;                           sp++; };
         continue;
      }

      // inlining of macro mmin
      
      if (ltLo-lo) < (unLo-ltLo) {
          n = (ltLo-lo);
      } else {
          n = (unLo-ltLo);
      // inlining of macro mvswap
      // inlining of macro mswap
      } {                                        var yyp1 Int32  = (lo);                  var yyp2 Int32  = (unLo-n);                  var yyn Int32   = (n);                   for yyn > 0 {                        { var zztmp Int32  = ptr[yyp1]; ptr[yyp1] = ptr[yyp2]; ptr[yyp2] = zztmp; };          yyp1++; yyp2++; yyn--;             }                                  };
      // inlining of macro mmin
      
      if (hi-gtHi) < (gtHi-unHi) {
          m = (hi-gtHi);
      } else {
          m = (gtHi-unHi);
      // inlining of macro mvswap
      // inlining of macro mswap
      } {                                        var yyp1 Int32  = (unLo);                  var yyp2 Int32  = (hi-m+1);                  var yyn Int32   = (m);                   for yyn > 0 {                        { var zztmp Int32  = ptr[yyp1]; ptr[yyp1] = ptr[yyp2]; ptr[yyp2] = zztmp; };          yyp1++; yyp2++; yyn--;             }                                  };

      n = lo + unLo - ltLo - 1;
      m = hi - (gtHi - unHi) + 1;

      nextLo[0] = lo;  nextHi[0] = n;   nextD[0] = d;
      nextLo[1] = m;   nextHi[1] = hi;  nextD[1] = d;
      nextLo[2] = n+1; nextHi[2] = m-1; nextD[2] = d+1;

      // inlining of macro mnextsize
      // inlining of macro mnextswap
      if (nextHi[0]-nextLo[0]) < (nextHi[1]-nextLo[1]) { var tz Int32;                                                       tz = nextLo[0]; nextLo[0] = nextLo[1]; nextLo[1] = tz;      tz = nextHi[0]; nextHi[0] = nextHi[1]; nextHi[1] = tz;      tz = nextD [0]; nextD [0] = nextD [1]; nextD [1] = tz; };
      // inlining of macro mnextsize
      // inlining of macro mnextswap
      if (nextHi[1]-nextLo[1]) < (nextHi[2]-nextLo[2]) { var tz Int32;                                                       tz = nextLo[1]; nextLo[1] = nextLo[2]; nextLo[2] = tz;      tz = nextHi[1]; nextHi[1] = nextHi[2]; nextHi[2] = tz;      tz = nextD [1]; nextD [1] = nextD [2]; nextD [2] = tz; };
      // inlining of macro mnextsize
      // inlining of macro mnextswap
      if (nextHi[0]-nextLo[0]) < (nextHi[1]-nextLo[1]) { var tz Int32;                                                       tz = nextLo[0]; nextLo[0] = nextLo[1]; nextLo[1] = tz;      tz = nextHi[0]; nextHi[0] = nextHi[1]; nextHi[1] = tz;      tz = nextD [0]; nextD [0] = nextD [1]; nextD [1] = tz; };

      AssertD (mnextsize(0) >= mnextsize(1), "mainQSort3(8)" );
      AssertD (mnextsize(1) >= mnextsize(2), "mainQSort3(9)" );

      // inlining of macro mpush
      { stackLo[sp] = nextLo[0];                           stackHi[sp] = nextHi[0];                           stackD [sp] = nextD[0];                           sp++; };
      // inlining of macro mpush
      { stackLo[sp] = nextLo[1];                           stackHi[sp] = nextHi[1];                           stackD [sp] = nextD[1];                           sp++; };
      // inlining of macro mpush
      { stackLo[sp] = nextLo[2];                           stackHi[sp] = nextHi[2];                           stackD [sp] = nextD[2];                           sp++; };
   }
}

#undef mswap
#undef mvswap
#undef mpush
#undef mpop
#undef mmin
#undef mnextsize
#undef mnextswap
#undef MAIN_QSORT_SMALL_THRESH
#undef MAIN_QSORT_DEPTH_THRESH
#undef MAIN_QSORT_STACK_SIZE


/*---------------------------------------------*/
/* Pre:
      nblock > N_OVERSHOOT
      block32 exists for [0 .. nblock-1 +N_OVERSHOOT]
      ((UChar*)block32) [0 .. nblock-1] holds block
      ptr exists for [0 .. nblock-1]

   Post:
      ((UChar*)block32) [0 .. nblock-1] holds block
      All other areas of block32 destroyed
      ftab [0 .. 65536 ] destroyed
      ptr [0 .. nblock-1] holds sorted order
      if (*budget < 0), sorting was abandoned
*/

// Original macro:
// #define BIGFREQ(b) (ftab[((b)+1) << 8] - ftab[(b) << 8])

#define SETMASK (1 << 21)
#define CLEARMASK (~(SETMASK))

static
func mainSort(ptr *UInt32, block *UChar, quadrant *UInt16, ftab *UInt32, nblock Int32, verb Int32, budget *Int32) {
   var i, j, k, ss, sb Int32;
   var runningOrder [256]Int32;
   var bigDone [256]Bool;
   var copyStart [256]Int32;
   var copyEnd [256]Int32;
   var c1 UChar;
   var numQSorted Int32;
   var s UInt16;
   if verb >= 4 { VPrintf0 ( "        main sort initialise ...\n" ); }

   /*-- set up the 2-byte frequency table --*/
   for i = 65536; i >= 0; i-- { ftab[i] = 0; }

   j = block[0] << 8;
   i = nblock-1;
   for ; i >= 3; i -= 4 {
      quadrant[i] = 0;
      j = (j >> 8) | ( ((UInt16)block[i]) << 8);
      ftab[j]++;
      quadrant[i-1] = 0;
      j = (j >> 8) | ( ((UInt16)block[i-1]) << 8);
      ftab[j]++;
      quadrant[i-2] = 0;
      j = (j >> 8) | ( ((UInt16)block[i-2]) << 8);
      ftab[j]++;
      quadrant[i-3] = 0;
      j = (j >> 8) | ( ((UInt16)block[i-3]) << 8);
      ftab[j]++;
   }
   for ; i >= 0; i-- {
      quadrant[i] = 0;
      j = (j >> 8) | ( ((UInt16)block[i]) << 8);
      ftab[j]++;
   }

   /*-- (emphasises close relationship of block & quadrant) --*/
   for i = 0; i < BZ_N_OVERSHOOT; i++ {
      block   [nblock+i] = block[i];
      quadrant[nblock+i] = 0;
   }

   if verb >= 4 { VPrintf0 ( "        bucket sorting ...\n" ); }

   /*-- Complete the initial radix sort --*/
   for i = 1; i <= 65536; i++ { ftab[i] += ftab[i-1]; }

   s = block[0] << 8;
   i = nblock-1;
   for ; i >= 3; i -= 4 {
      s = (s >> 8) | (block[i] << 8);
      j = ftab[s] -1;
      ftab[s] = j;
      ptr[j] = i;
      s = (s >> 8) | (block[i-1] << 8);
      j = ftab[s] -1;
      ftab[s] = j;
      ptr[j] = i-1;
      s = (s >> 8) | (block[i-2] << 8);
      j = ftab[s] -1;
      ftab[s] = j;
      ptr[j] = i-2;
      s = (s >> 8) | (block[i-3] << 8);
      j = ftab[s] -1;
      ftab[s] = j;
      ptr[j] = i-3;
   }
   for ; i >= 0; i-- {
      s = (s >> 8) | (block[i] << 8);
      j = ftab[s] -1;
      ftab[s] = j;
      ptr[j] = i;
   }

   /*--
      Now ftab contains the first loc of every small bucket.
      Calculate the running order, from smallest to largest
      big bucket.
   --*/
   for i = 0; i <= 255; i++ {
      bigDone     [i] = False;
      runningOrder[i] = i;
   }

   {
      var vv Int32;
      var h Int32  = 1;
      do h = 3 * h + 1; while (h <= 256);
      for {
         h = h / 3;
         for i = h; i <= 255; i++ {
            vv = runningOrder[i];
            j = i;
            // inlining of macro BIGFREQ
            for  (ftab[((runningOrder[j-h])+1) << 8] - ftab[(runningOrder[j-h]) << 8]) > (ftab[((vv)+1) << 8] - ftab[(vv) << 8])  {
               runningOrder[j] = runningOrder[j-h];
               j = j - h;
               if j <= (h - 1) { goto zero; }
            }
            zero:
            runningOrder[j] = vv;
         }
      
if !(h != 1) { break }
} 
   }

   /*--
      The main sorting loop.
   --*/

   numQSorted = 0;

   for i = 0; i <= 255; i++ {

      /*--
         Process big buckets, starting with the least full.
         Basically this is a 3-step process in which we call
         mainQSort3 to sort the small buckets [ss, j], but
         also make a big effort to avoid the calls if we can.
      --*/
      ss = runningOrder[i];

      /*--
         Step 1:
         Complete the big bucket [ss] by quicksorting
         any unsorted small buckets [ss, j], for j != ss.  
         Hopefully previous pointer-scanning phases have already
         completed many of the small buckets [ss, j], so
         we don't have to sort them at all.
      --*/
      for j = 0; j <= 255; j++ {
         if j != ss {
            sb = (ss << 8) + j;
            if ! (ftab[sb] & SETMASK) {
               var lo Int32  = ftab[sb]   & CLEARMASK;
               var hi Int32  = (ftab[sb+1] & CLEARMASK) - 1;
               if hi > lo {
                  if verb >= 4 {
                     VPrintf4 ( "        qsort [0x%x, 0x%x]   "
                                "done %d   this %d\n",
                                ss, j, numQSorted, hi - lo + 1 ); }
                  mainQSort3 ( 
                     ptr, block, quadrant, nblock, 
                     lo, hi, BZ_N_RADIX, budget 
                  );   
                  numQSorted += (hi - lo + 1);
                  if *budget < 0 { return; }
               }
            }
            ftab[sb] |= SETMASK;
         }
      }

      AssertH ( !bigDone[ss], 1006 );

      /*--
         Step 2:
         Now scan this big bucket [ss] so as to synthesise the
         sorted order for small buckets [t, ss] for all t,
         including, magically, the bucket [ss,ss] too.
         This will avoid doing Real Work in subsequent Step 1's.
      --*/
      {
         for j = 0; j <= 255; j++ {
            copyStart[j] =  ftab[(j << 8) + ss]     & CLEARMASK;
            copyEnd  [j] = (ftab[(j << 8) + ss + 1] & CLEARMASK) - 1;
         }
         for j = ftab[ss << 8] & CLEARMASK; j < copyStart[ss]; j++ {
            k = ptr[j]-1; if k < 0 { k += nblock; }
            c1 = block[k];
            if !bigDone[c1] {
               ptr[ copyStart[c1]++ ] = k; }
         }
         for j = (ftab[(ss+1) << 8] & CLEARMASK) - 1; j > copyEnd[ss]; j-- {
            k = ptr[j]-1; if k < 0 { k += nblock; }
            c1 = block[k];
            if !bigDone[c1] { 
               ptr[ copyEnd[c1]-- ] = k; }
         }
      }

      AssertH ( (copyStart[ss]-1 == copyEnd[ss])
                || 
                /* Extremely rare case missing in bzip2-1.0.0 and 1.0.1.
                   Necessity for this case is demonstrated by compressing 
                   a sequence of approximately 48.5 million of character 
                   251; 1.0.0/1.0.1 will then die here. */
                (copyStart[ss] == 0 && copyEnd[ss] == nblock-1),
                1007 )

      for (j = 0; j <= 255; j++) ftab[(j << 8) + ss] |= SETMASK;

      /*--
         Step 3:
         The [ss] big bucket is now done.  Record this fact,
         and update the quadrant descriptors.  Remember to
         update quadrants in the overshoot area too, if
         necessary.  The "if (i < 255)" test merely skips
         this updating for the last bucket processed, since
         updating for the last bucket is pointless.

         The quadrant array provides a way to incrementally
         cache sort orderings, as they appear, so as to 
         make subsequent comparisons in fullGtU() complete
         faster.  For repetitive blocks this makes a big
         difference (but not big enough to be able to avoid
         the fallback sorting mechanism, exponential radix sort).

         The precise meaning is: at all times:

            for 0 <= i < nblock and 0 <= j <= nblock

            if block[i] != block[j], 

               then the relative values of quadrant[i] and 
                    quadrant[j] are meaningless.

               else {
                  if quadrant[i] < quadrant[j]
                     then the string starting at i lexicographically
                     precedes the string starting at j

                  else if quadrant[i] > quadrant[j]
                     then the string starting at j lexicographically
                     precedes the string starting at i

                  else
                     the relative ordering of the strings starting
                     at i and j has not yet been determined.
               }
      --*/
      bigDone[ss] = True;

      if i < 255 {
         var bbStart Int32   = ftab[ss << 8] & CLEARMASK;
         var bbSize Int32    = (ftab[(ss+1) << 8] & CLEARMASK) - bbStart;
         var shifts Int32;

         for (bbSize >> shifts) > 65534 { shifts++; }

         for j = bbSize-1; j >= 0; j-- {
            var a2update Int32      = ptr[bbStart + j];
            var qVal UInt16         = (UInt16)(j >> shifts);
            quadrant[a2update] = qVal;
            if a2update < BZ_N_OVERSHOOT {
               quadrant[a2update + nblock] = qVal; }
         }
         AssertH ( ((bbSize-1) >> shifts) <= 65535, 1002 );
      }

   }

   if verb >= 4 {
      VPrintf3 ( "        %d pointers, %d sorted, %d scanned\n",
                 nblock, numQSorted, nblock - numQSorted ); }
}

#undef BIGFREQ
#undef SETMASK
#undef CLEARMASK


/*---------------------------------------------*/
/* Pre:
      nblock > 0
      arr2 exists for [0 .. nblock-1 +N_OVERSHOOT]
      ((UChar*)arr2)  [0 .. nblock-1] holds block
      arr1 exists for [0 .. nblock-1]

   Post:
      ((UChar*)arr2) [0 .. nblock-1] holds block
      All other areas of block destroyed
      ftab [ 0 .. 65536 ] destroyed
      arr1 [0 .. nblock-1] holds sorted order
*/
func BZ2_blockSort(s *EState) {
   var ptr *UInt32     = s.ptr; 
   var block *UChar   = s.block;
   var ftab *UInt32    = s.ftab;
   var nblock Int32  = s.nblock;
   var verb Int32    = s.verbosity;
   var wfact Int32   = s.workFactor;
   var quadrant *UInt16;
   var budget Int32;
   var budgetInit Int32;
   var i Int32;

   if nblock < 10000 {
      fallbackSort ( s.arr1, s.arr2, ftab, nblock, verb );
   } else {
      /* Calculate the location for quadrant, remembering to get
         the alignment right.  Assumes that &(block[0]) is at least
         2-byte aligned -- this should be ok since block is really
         the first section of arr2.
      */
      i = nblock+BZ_N_OVERSHOOT;
      if i & 1 { i++; }
      quadrant = (UInt16*)(&(block[i]));

      /* (wfact-1) / 3 puts the default-factor-30
         transition point at very roughly the same place as 
         with v0.1 and v0.9.0.  
         Not that it particularly matters any more, since the
         resulting compressed stream is now the same regardless
         of whether or not we use the main sort or fallback sort.
      */
      if wfact < 1 { wfact = 1; }
      if wfact > 100 { wfact = 100; }
      budgetInit = nblock * ((wfact-1) / 3);
      budget = budgetInit;

      mainSort ( ptr, block, quadrant, ftab, nblock, verb, &budget );
      if verb >= 3 { 
         VPrintf3 ( "      %d work, %d block, ratio %5.2f\n",
                    budgetInit - budget,
                    nblock, 
                    float32((budgetInit - budget)) /
                    float32((nblock==0 ? 1 : nblock)) ); } 
      if budget < 0 {
         if verb >= 2 { 
            VPrintf0 ( "    too repetitive; using fallback"
                       " sorting algorithm\n" ); }
         fallbackSort ( s.arr1, s.arr2, ftab, nblock, verb );
      }
   }

   s.origPtr = -1;
   for i = 0; i < s.nblock; i++ {
      if ptr[i] == 0 { s.origPtr = i; break; }
   };

   AssertH( s.origPtr != -1, 1003 );
}


/*-------------------------------------------------------------*/
/*--- end                                       blocksort.c ---*/
/*-------------------------------------------------------------*/

//ABC_NAMESPACE_IMPL_END
