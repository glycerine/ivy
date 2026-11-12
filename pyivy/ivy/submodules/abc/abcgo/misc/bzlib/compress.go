package bzlib


/*-------------------------------------------------------------*/
/*--- Compression machinery (not incl block sorting)        ---*/
/*---                                            compress.c ---*/
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


/* CHANGES
    0.9.0    -- original version.
    0.9.0a/b -- no changes in this file.
    0.9.0c   -- changed setting of nGroups in sendMTFValues() 
                so as to do a bit better on small files
*/

//#include "bzlib_private.h"

//ABC_NAMESPACE_IMPL_START



/*---------------------------------------------------*/
/*--- Bit stream I/O                              ---*/
/*---------------------------------------------------*/

/*---------------------------------------------------*/
func BZ2_bsInitWrite(s *EState) {
   s.bsLive = 0;
   s.bsBuff = 0;
}


/*---------------------------------------------------*/
static
func bsFinishWrite(s *EState) {
   for s.bsLive > 0 {
      s.zbits[s.numZ] = (UChar)(s.bsBuff >> 24);
      s.numZ++;
      s.bsBuff <<= 8;
      s.bsLive -= 8;
   }
}


/*---------------------------------------------------*/
// Original macro:
// #define bsNEEDW(nz)                           \
// {                                             \
//    while (s->bsLive >= 8) {                   \
//       s->zbits[s->numZ]                       \
//          = (UChar)(s->bsBuff >> 24);          \
//       s->numZ++;                              \
//       s->bsBuff <<= 8;                        \
//       s->bsLive -= 8;                         \
//    }                                          \
// }



/*---------------------------------------------------*/
static
//__inline__
func bsW(s *EState, n Int32, v UInt32) {
   // inlining of macro bsNEEDW
   {                                                for s.bsLive >= 8 {                         s.zbits[s.numZ]                                = (UChar)(s.bsBuff >> 24);                s.numZ++;                                    s.bsBuff <<= 8;                              s.bsLive -= 8;                            }                                          };
   s.bsBuff |= (v << (32 - s.bsLive - n));
   s.bsLive += n;
}


/*---------------------------------------------------*/
static
func bsPutUInt32(s *EState, u UInt32) {
   bsW ( s, 8, (u >> 24) & int64(0xff) );
   bsW ( s, 8, (u >> 16) & int64(0xff) );
   bsW ( s, 8, (u >>  8) & int64(0xff) );
   bsW ( s, 8,  u        & int64(0xff) );
}


/*---------------------------------------------------*/
static
func bsPutUChar(s *EState, c UChar) {
   bsW( s, 8, (UInt32)c );
}


/*---------------------------------------------------*/
/*--- The back end proper                         ---*/
/*---------------------------------------------------*/

/*---------------------------------------------------*/
static
func makeMaps_e(s *EState) {
   var i Int32;
   s.nInUse = 0;
   for i = 0; i < 256; i++ {
      if s.inUse[i] {
         s.unseqToSeq[i] = s.nInUse;
         s.nInUse++;
      }
   }
}


/*---------------------------------------------------*/
static
func generateMTFValues(s *EState) {
   var yy [256]UChar;
   var i, j Int32;
   var zPend Int32;
   var wr Int32;
   var EOB Int32;

   /* 
      After sorting (eg, here),
         s->arr1 [ 0 .. s->nblock-1 ] holds sorted order,
         and
         ((UChar*)s->arr2) [ 0 .. s->nblock-1 ] 
         holds the original block data.

      The first thing to do is generate the MTF values,
      and put them in
         ((UInt16*)s->arr1) [ 0 .. s->nblock-1 ].
      Because there are strictly fewer or equal MTF values
      than block values, ptr values in this area are overwritten
      with MTF values only when they are no longer needed.

      The final compressed bitstream is generated into the
      area starting at
         (UChar*) (&((UChar*)s->arr2)[s->nblock])

      These storage aliases are set up in bzCompressInit(),
      except for the last one, which is arranged in 
      compressBlock().
   */
   var ptr *UInt32    = s.ptr;
   var block *UChar   = s.block;
   var mtfv *UInt16   = s.mtfv;

   makeMaps_e ( s );
   EOB = s.nInUse+1;

   for i = 0; i <= EOB; i++ { s.mtfFreq[i] = 0; }

   wr = 0;
   zPend = 0;
   for i = 0; i < s.nInUse; i++ { yy[i] = (UChar) i; }

   for i = 0; i < s.nblock; i++ {
      var ll_i UChar;
      AssertD ( wr <= i, "generateMTFValues(1)" );
      j = ptr[i]-1; if j < 0 { j += s.nblock; }
      ll_i = s.unseqToSeq[block[j]];
      AssertD ( ll_i < s.nInUse, "generateMTFValues(2a)" );

      if yy[0] == ll_i { 
         zPend++;
      } else {

         if zPend > 0 {
            zPend--;
            for True {
               if zPend & 1 {
                  mtfv[wr] = BZ_RUNB; wr++; 
                  s.mtfFreq[BZ_RUNB]++; 
               } else {
                  mtfv[wr] = BZ_RUNA; wr++; 
                  s.mtfFreq[BZ_RUNA]++; 
               }
               if zPend < 2 { break; }
               zPend = (zPend - 2) / 2;
            };
            zPend = 0;
         }
         {
            register UChar  rtmp;
            register UChar* ryy_j;
            register UChar  rll_i;
            rtmp  = yy[1];
            yy[1] = yy[0];
            ryy_j = &(yy[1]);
            rll_i = ll_i;
            for  rll_i != rtmp  {
               register UChar rtmp2;
               ryy_j++;
               rtmp2  = rtmp;
               rtmp   = *ryy_j;
               *ryy_j = rtmp2;
            };
            yy[0] = rtmp;
            j = ryy_j - &(yy[0]);
            mtfv[wr] = j+1; wr++; s.mtfFreq[j+1]++;
         }

      }
   }

   if zPend > 0 {
      zPend--;
      for True {
         if zPend & 1 {
            mtfv[wr] = BZ_RUNB; wr++; 
            s.mtfFreq[BZ_RUNB]++; 
         } else {
            mtfv[wr] = BZ_RUNA; wr++; 
            s.mtfFreq[BZ_RUNA]++; 
         }
         if zPend < 2 { break; }
         zPend = (zPend - 2) / 2;
      };
      zPend = 0;
   }

   mtfv[wr] = EOB; wr++; s.mtfFreq[EOB]++;

   s.nMTF = wr;
}


/*---------------------------------------------------*/
const BZ_LESSER_ICOST =  0
const BZ_GREATER_ICOST = 15

static
func sendMTFValues(s *EState) {
   var v, t, i, j, gs, ge, totc, bt, bc, iter Int32;
   var nSelectors, alphaSize, minLen, maxLen, selCtr Int32;
   var nGroups, nBytes Int32;

   /*--
   UChar  len [BZ_N_GROUPS][BZ_MAX_ALPHA_SIZE];
   is a global since the decoder also needs it.

   Int32  code[BZ_N_GROUPS][BZ_MAX_ALPHA_SIZE];
   Int32  rfreq[BZ_N_GROUPS][BZ_MAX_ALPHA_SIZE];
   are also globals only used in this proc.
   Made global to keep stack frame size small.
   --*/


   var cost [BZ_N_GROUPS]UInt16;
   var fave [BZ_N_GROUPS]Int32;

   var mtfv *UInt16  = s.mtfv;

   if s.verbosity >= 3 {
      VPrintf3( "      %d in block, %d after MTF & 1-2 coding, "
                "%d+2 syms in use\n", 
                s.nblock, s.nMTF, s.nInUse ); }

   alphaSize = s.nInUse+2;
   for t = 0; t < BZ_N_GROUPS; t++ {
      for v = 0; v < alphaSize; v++ {
         s.len[t][v] = BZ_GREATER_ICOST;
      }
   }

   /*--- Decide how many coding tables to use ---*/
   AssertH ( s.nMTF > 0, 3001 );
   if s.nMTF < 200 {  nGroups = 2; } else
   if s.nMTF < 600 {  nGroups = 3; } else
   if s.nMTF < 1200 { nGroups = 4; } else
   if s.nMTF < 2400 { nGroups = 5; } else {
                       nGroups = 6; }

   /*--- Generate an initial set of coding tables ---*/
   { 
      var nPart, remF, tFreq, aFreq Int32;

      nPart = nGroups;
      remF  = s.nMTF;
      gs = 0;
      for nPart > 0 {
         tFreq = remF / nPart;
         ge = gs-1;
         aFreq = 0;
         for aFreq < tFreq && ge < alphaSize-1 {
            ge++;
            aFreq += s.mtfFreq[ge];
         }

         if ge > gs 
             && nPart != nGroups && nPart != 1 
             && ((nGroups-nPart) % 2 == 1) {
            aFreq -= s.mtfFreq[ge];
            ge--;
         }

         if s.verbosity >= 3 {
            VPrintf5( "      initial group %d, [%d .. %d], "
                      "has %d syms (%4.1f%%)\n",
                      nPart, gs, ge, aFreq, 
                      (100.0 * float32(aFreq)) / float32((s.nMTF)) ); }
 
         for v = 0; v < alphaSize; v++ {
            if v >= gs && v <= ge { 
               s.len[nPart-1][v] = BZ_LESSER_ICOST; } else {
               s.len[nPart-1][v] = BZ_GREATER_ICOST; }
         }
 
         nPart--;
         gs = ge+1;
         remF -= aFreq;
      }
   }

   /*--- 
      Iterate up to BZ_N_ITERS times to improve the tables.
   ---*/
   for iter = 0; iter < BZ_N_ITERS; iter++ {

      for t = 0; t < nGroups; t++ { fave[t] = 0; }

      for t = 0; t < nGroups; t++ {
         for v = 0; v < alphaSize; v++ {
            s.rfreq[t][v] = 0;
         }
      }

      /*---
        Set up an auxiliary length table which is used to fast-track
    the common case (nGroups == 6). 
      ---*/
      if nGroups == 6 {
         for v = 0; v < alphaSize; v++ {
            s.len_pack[v][0] = (s.len[1][v] << 16) | s.len[0][v];
            s.len_pack[v][1] = (s.len[3][v] << 16) | s.len[2][v];
            s.len_pack[v][2] = (s.len[5][v] << 16) | s.len[4][v];
     }
      }

      nSelectors = 0;
      totc = 0;
      gs = 0;
      for True {

         /*--- Set group start & end marks. --*/
         if gs >= s.nMTF { break; }
         ge = gs + BZ_G_SIZE - 1; 
         if ge >= s.nMTF { ge = s.nMTF-1; }

         /*-- 
            Calculate the cost of this group as coded
            by each of the coding tables.
         --*/
         for t = 0; t < nGroups; t++ { cost[t] = 0; }

         if nGroups == 6 && 50 == ge-gs+1 {
            /*--- fast track the common case ---*/
            register UInt32 cost01, cost23, cost45;
            register UInt16 icv;
            cost01 = cost23 = cost45 = 0;

// Original macro:
// #           define BZ_ITER(nn)                \
//                icv = mtfv[gs+(nn)];           \
//                cost01 += s->len_pack[icv][0]; \
//                cost23 += s->len_pack[icv][1]; \
//                cost45 += s->len_pack[icv][2]; \
// 

            // inlining of macro BZ_ITER
            icv = mtfv[gs+(0)];                          cost01 += s.len_pack[icv][0];                cost23 += s.len_pack[icv][1];                cost45 += s.len_pack[icv][2]; ;  icv = mtfv[gs+(1)];                          cost01 += s.len_pack[icv][0];                cost23 += s.len_pack[icv][1];                cost45 += s.len_pack[icv][2]; ;  icv = mtfv[gs+(2)];                          cost01 += s.len_pack[icv][0];                cost23 += s.len_pack[icv][1];                cost45 += s.len_pack[icv][2]; ;  icv = mtfv[gs+(3)];                          cost01 += s.len_pack[icv][0];                cost23 += s.len_pack[icv][1];                cost45 += s.len_pack[icv][2]; ;  icv = mtfv[gs+(4)];                          cost01 += s.len_pack[icv][0];                cost23 += s.len_pack[icv][1];                cost45 += s.len_pack[icv][2]; ;
            // inlining of macro BZ_ITER
            icv = mtfv[gs+(5)];                          cost01 += s.len_pack[icv][0];                cost23 += s.len_pack[icv][1];                cost45 += s.len_pack[icv][2]; ;  icv = mtfv[gs+(6)];                          cost01 += s.len_pack[icv][0];                cost23 += s.len_pack[icv][1];                cost45 += s.len_pack[icv][2]; ;  icv = mtfv[gs+(7)];                          cost01 += s.len_pack[icv][0];                cost23 += s.len_pack[icv][1];                cost45 += s.len_pack[icv][2]; ;  icv = mtfv[gs+(8)];                          cost01 += s.len_pack[icv][0];                cost23 += s.len_pack[icv][1];                cost45 += s.len_pack[icv][2]; ;  icv = mtfv[gs+(9)];                          cost01 += s.len_pack[icv][0];                cost23 += s.len_pack[icv][1];                cost45 += s.len_pack[icv][2]; ;
            // inlining of macro BZ_ITER
            icv = mtfv[gs+(10)];                          cost01 += s.len_pack[icv][0];                cost23 += s.len_pack[icv][1];                cost45 += s.len_pack[icv][2]; ; icv = mtfv[gs+(11)];                          cost01 += s.len_pack[icv][0];                cost23 += s.len_pack[icv][1];                cost45 += s.len_pack[icv][2]; ; icv = mtfv[gs+(12)];                          cost01 += s.len_pack[icv][0];                cost23 += s.len_pack[icv][1];                cost45 += s.len_pack[icv][2]; ; icv = mtfv[gs+(13)];                          cost01 += s.len_pack[icv][0];                cost23 += s.len_pack[icv][1];                cost45 += s.len_pack[icv][2]; ; icv = mtfv[gs+(14)];                          cost01 += s.len_pack[icv][0];                cost23 += s.len_pack[icv][1];                cost45 += s.len_pack[icv][2]; ;
            // inlining of macro BZ_ITER
            icv = mtfv[gs+(15)];                          cost01 += s.len_pack[icv][0];                cost23 += s.len_pack[icv][1];                cost45 += s.len_pack[icv][2]; ; icv = mtfv[gs+(16)];                          cost01 += s.len_pack[icv][0];                cost23 += s.len_pack[icv][1];                cost45 += s.len_pack[icv][2]; ; icv = mtfv[gs+(17)];                          cost01 += s.len_pack[icv][0];                cost23 += s.len_pack[icv][1];                cost45 += s.len_pack[icv][2]; ; icv = mtfv[gs+(18)];                          cost01 += s.len_pack[icv][0];                cost23 += s.len_pack[icv][1];                cost45 += s.len_pack[icv][2]; ; icv = mtfv[gs+(19)];                          cost01 += s.len_pack[icv][0];                cost23 += s.len_pack[icv][1];                cost45 += s.len_pack[icv][2]; ;
            // inlining of macro BZ_ITER
            icv = mtfv[gs+(20)];                          cost01 += s.len_pack[icv][0];                cost23 += s.len_pack[icv][1];                cost45 += s.len_pack[icv][2]; ; icv = mtfv[gs+(21)];                          cost01 += s.len_pack[icv][0];                cost23 += s.len_pack[icv][1];                cost45 += s.len_pack[icv][2]; ; icv = mtfv[gs+(22)];                          cost01 += s.len_pack[icv][0];                cost23 += s.len_pack[icv][1];                cost45 += s.len_pack[icv][2]; ; icv = mtfv[gs+(23)];                          cost01 += s.len_pack[icv][0];                cost23 += s.len_pack[icv][1];                cost45 += s.len_pack[icv][2]; ; icv = mtfv[gs+(24)];                          cost01 += s.len_pack[icv][0];                cost23 += s.len_pack[icv][1];                cost45 += s.len_pack[icv][2]; ;
            // inlining of macro BZ_ITER
            icv = mtfv[gs+(25)];                          cost01 += s.len_pack[icv][0];                cost23 += s.len_pack[icv][1];                cost45 += s.len_pack[icv][2]; ; icv = mtfv[gs+(26)];                          cost01 += s.len_pack[icv][0];                cost23 += s.len_pack[icv][1];                cost45 += s.len_pack[icv][2]; ; icv = mtfv[gs+(27)];                          cost01 += s.len_pack[icv][0];                cost23 += s.len_pack[icv][1];                cost45 += s.len_pack[icv][2]; ; icv = mtfv[gs+(28)];                          cost01 += s.len_pack[icv][0];                cost23 += s.len_pack[icv][1];                cost45 += s.len_pack[icv][2]; ; icv = mtfv[gs+(29)];                          cost01 += s.len_pack[icv][0];                cost23 += s.len_pack[icv][1];                cost45 += s.len_pack[icv][2]; ;
            // inlining of macro BZ_ITER
            icv = mtfv[gs+(30)];                          cost01 += s.len_pack[icv][0];                cost23 += s.len_pack[icv][1];                cost45 += s.len_pack[icv][2]; ; icv = mtfv[gs+(31)];                          cost01 += s.len_pack[icv][0];                cost23 += s.len_pack[icv][1];                cost45 += s.len_pack[icv][2]; ; icv = mtfv[gs+(32)];                          cost01 += s.len_pack[icv][0];                cost23 += s.len_pack[icv][1];                cost45 += s.len_pack[icv][2]; ; icv = mtfv[gs+(33)];                          cost01 += s.len_pack[icv][0];                cost23 += s.len_pack[icv][1];                cost45 += s.len_pack[icv][2]; ; icv = mtfv[gs+(34)];                          cost01 += s.len_pack[icv][0];                cost23 += s.len_pack[icv][1];                cost45 += s.len_pack[icv][2]; ;
            // inlining of macro BZ_ITER
            icv = mtfv[gs+(35)];                          cost01 += s.len_pack[icv][0];                cost23 += s.len_pack[icv][1];                cost45 += s.len_pack[icv][2]; ; icv = mtfv[gs+(36)];                          cost01 += s.len_pack[icv][0];                cost23 += s.len_pack[icv][1];                cost45 += s.len_pack[icv][2]; ; icv = mtfv[gs+(37)];                          cost01 += s.len_pack[icv][0];                cost23 += s.len_pack[icv][1];                cost45 += s.len_pack[icv][2]; ; icv = mtfv[gs+(38)];                          cost01 += s.len_pack[icv][0];                cost23 += s.len_pack[icv][1];                cost45 += s.len_pack[icv][2]; ; icv = mtfv[gs+(39)];                          cost01 += s.len_pack[icv][0];                cost23 += s.len_pack[icv][1];                cost45 += s.len_pack[icv][2]; ;
            // inlining of macro BZ_ITER
            icv = mtfv[gs+(40)];                          cost01 += s.len_pack[icv][0];                cost23 += s.len_pack[icv][1];                cost45 += s.len_pack[icv][2]; ; icv = mtfv[gs+(41)];                          cost01 += s.len_pack[icv][0];                cost23 += s.len_pack[icv][1];                cost45 += s.len_pack[icv][2]; ; icv = mtfv[gs+(42)];                          cost01 += s.len_pack[icv][0];                cost23 += s.len_pack[icv][1];                cost45 += s.len_pack[icv][2]; ; icv = mtfv[gs+(43)];                          cost01 += s.len_pack[icv][0];                cost23 += s.len_pack[icv][1];                cost45 += s.len_pack[icv][2]; ; icv = mtfv[gs+(44)];                          cost01 += s.len_pack[icv][0];                cost23 += s.len_pack[icv][1];                cost45 += s.len_pack[icv][2]; ;
            // inlining of macro BZ_ITER
            icv = mtfv[gs+(45)];                          cost01 += s.len_pack[icv][0];                cost23 += s.len_pack[icv][1];                cost45 += s.len_pack[icv][2]; ; icv = mtfv[gs+(46)];                          cost01 += s.len_pack[icv][0];                cost23 += s.len_pack[icv][1];                cost45 += s.len_pack[icv][2]; ; icv = mtfv[gs+(47)];                          cost01 += s.len_pack[icv][0];                cost23 += s.len_pack[icv][1];                cost45 += s.len_pack[icv][2]; ; icv = mtfv[gs+(48)];                          cost01 += s.len_pack[icv][0];                cost23 += s.len_pack[icv][1];                cost45 += s.len_pack[icv][2]; ; icv = mtfv[gs+(49)];                          cost01 += s.len_pack[icv][0];                cost23 += s.len_pack[icv][1];                cost45 += s.len_pack[icv][2]; ;

#           undef BZ_ITER

            cost[0] = cost01 & 0xffff; cost[1] = cost01 >> 16;
            cost[2] = cost23 & 0xffff; cost[3] = cost23 >> 16;
            cost[4] = cost45 & 0xffff; cost[5] = cost45 >> 16;

         } else {
        /*--- slow version which correctly handles all situations ---*/
            for i = gs; i <= ge; i++ { 
               var icv UInt16  = mtfv[i];
               for t = 0; t < nGroups; t++ { cost[t] += s.len[t][icv]; }
            }
         }
 
         /*-- 
            Find the coding table which is best for this group,
            and record its identity in the selector table.
         --*/
         bc = 999999999; bt = -1;
         for t = 0; t < nGroups; t++ {
            if cost[t] < bc { bc = cost[t]; bt = t; }
         };
         totc += bc;
         fave[bt]++;
         s.selector[nSelectors] = bt;
         nSelectors++;

         /*-- 
            Increment the symbol frequencies for the selected table.
          --*/
         if nGroups == 6 && 50 == ge-gs+1 {
            /*--- fast track the common case ---*/

// Original macro:
// #           define BZ_ITUR(nn) s->rfreq[bt][ mtfv[gs+(nn)] ]++


            // inlining of macro BZ_ITUR
            s.rfreq[bt][ mtfv[gs+(0)] ]++;  s.rfreq[bt][ mtfv[gs+(1)] ]++;  s.rfreq[bt][ mtfv[gs+(2)] ]++;  s.rfreq[bt][ mtfv[gs+(3)] ]++;  s.rfreq[bt][ mtfv[gs+(4)] ]++;
            // inlining of macro BZ_ITUR
            s.rfreq[bt][ mtfv[gs+(5)] ]++;  s.rfreq[bt][ mtfv[gs+(6)] ]++;  s.rfreq[bt][ mtfv[gs+(7)] ]++;  s.rfreq[bt][ mtfv[gs+(8)] ]++;  s.rfreq[bt][ mtfv[gs+(9)] ]++;
            // inlining of macro BZ_ITUR
            s.rfreq[bt][ mtfv[gs+(10)] ]++; s.rfreq[bt][ mtfv[gs+(11)] ]++; s.rfreq[bt][ mtfv[gs+(12)] ]++; s.rfreq[bt][ mtfv[gs+(13)] ]++; s.rfreq[bt][ mtfv[gs+(14)] ]++;
            // inlining of macro BZ_ITUR
            s.rfreq[bt][ mtfv[gs+(15)] ]++; s.rfreq[bt][ mtfv[gs+(16)] ]++; s.rfreq[bt][ mtfv[gs+(17)] ]++; s.rfreq[bt][ mtfv[gs+(18)] ]++; s.rfreq[bt][ mtfv[gs+(19)] ]++;
            // inlining of macro BZ_ITUR
            s.rfreq[bt][ mtfv[gs+(20)] ]++; s.rfreq[bt][ mtfv[gs+(21)] ]++; s.rfreq[bt][ mtfv[gs+(22)] ]++; s.rfreq[bt][ mtfv[gs+(23)] ]++; s.rfreq[bt][ mtfv[gs+(24)] ]++;
            // inlining of macro BZ_ITUR
            s.rfreq[bt][ mtfv[gs+(25)] ]++; s.rfreq[bt][ mtfv[gs+(26)] ]++; s.rfreq[bt][ mtfv[gs+(27)] ]++; s.rfreq[bt][ mtfv[gs+(28)] ]++; s.rfreq[bt][ mtfv[gs+(29)] ]++;
            // inlining of macro BZ_ITUR
            s.rfreq[bt][ mtfv[gs+(30)] ]++; s.rfreq[bt][ mtfv[gs+(31)] ]++; s.rfreq[bt][ mtfv[gs+(32)] ]++; s.rfreq[bt][ mtfv[gs+(33)] ]++; s.rfreq[bt][ mtfv[gs+(34)] ]++;
            // inlining of macro BZ_ITUR
            s.rfreq[bt][ mtfv[gs+(35)] ]++; s.rfreq[bt][ mtfv[gs+(36)] ]++; s.rfreq[bt][ mtfv[gs+(37)] ]++; s.rfreq[bt][ mtfv[gs+(38)] ]++; s.rfreq[bt][ mtfv[gs+(39)] ]++;
            // inlining of macro BZ_ITUR
            s.rfreq[bt][ mtfv[gs+(40)] ]++; s.rfreq[bt][ mtfv[gs+(41)] ]++; s.rfreq[bt][ mtfv[gs+(42)] ]++; s.rfreq[bt][ mtfv[gs+(43)] ]++; s.rfreq[bt][ mtfv[gs+(44)] ]++;
            // inlining of macro BZ_ITUR
            s.rfreq[bt][ mtfv[gs+(45)] ]++; s.rfreq[bt][ mtfv[gs+(46)] ]++; s.rfreq[bt][ mtfv[gs+(47)] ]++; s.rfreq[bt][ mtfv[gs+(48)] ]++; s.rfreq[bt][ mtfv[gs+(49)] ]++;

#           undef BZ_ITUR

         } else {
        /*--- slow version which correctly handles all situations ---*/
            for i = gs; i <= ge; i++ {
               s.rfreq[bt][ mtfv[i] ]++;
            }
         }

         gs = ge+1;
      }
      if s.verbosity >= 3 {
         VPrintf2 ( "      pass %d: size is %d, grp uses are ", 
                   iter+1, totc/8 );
         for t = 0; t < nGroups; t++ {
            VPrintf1 ( "%d ", fave[t] );
         }
         VPrintf0 ( "\n" );
      }

      /*--
        Recompute the tables based on the accumulated frequencies.
      --*/
      /* maxLen was changed from 20 to 17 in bzip2-1.0.3.  See 
         comment in huffman.c for details. */
      for t = 0; t < nGroups; t++ {
         BZ2_hbMakeCodeLengths ( &(s.len[t][0]), &(s.rfreq[t][0]), 
                                 alphaSize, 17 /*20*/ );
      }
   }


   AssertH( nGroups < 8, 3002 );
   AssertH( nSelectors < 32768 &&
            nSelectors <= (2 + (900000 / BZ_G_SIZE)),
            3003 );


   /*--- Compute MTF values for the selectors. ---*/
   {
      var pos [BZ_N_GROUPS]UChar;
      var ll_i UChar;
      var tmp2 UChar;
      var tmp UChar;
      for i = 0; i < nGroups; i++ { pos[i] = i; }
      for i = 0; i < nSelectors; i++ {
         ll_i = s.selector[i];
         j = 0;
         tmp = pos[j];
         for  ll_i != tmp  {
            j++;
            tmp2 = tmp;
            tmp = pos[j];
            pos[j] = tmp2;
         };
         pos[0] = tmp;
         s.selectorMtf[i] = j;
      }
   };

   /*--- Assign actual codes for the tables. --*/
   for t = 0; t < nGroups; t++ {
      minLen = 32;
      maxLen = 0;
      for i = 0; i < alphaSize; i++ {
         if s.len[t][i] > maxLen { maxLen = s.len[t][i]; }
         if s.len[t][i] < minLen { minLen = s.len[t][i]; }
      }
      AssertH ( !(maxLen > 17 /*20*/ ), 3004 );
      AssertH ( !(minLen < 1),  3005 );
      BZ2_hbAssignCodes ( &(s.code[t][0]), &(s.len[t][0]), 
                          minLen, maxLen, alphaSize );
   }

   /*--- Transmit the mapping table. ---*/
   { 
      var inUse16 [16]Bool;
      for i = 0; i < 16; i++ {
          inUse16[i] = False;
          for j = 0; j < 16; j++ {
             if s.inUse[i * 16 + j] { inUse16[i] = True; }
          }
      }
     
      nBytes = s.numZ;
      for i = 0; i < 16; i++ {
         if inUse16[i] { bsW(s,1,1); } else { bsW(s,1,0); }
      }

      for i = 0; i < 16; i++ {
         if inUse16[i] {
            for j = 0; j < 16; j++ {
               if s.inUse[i * 16 + j] { bsW(s,1,1); } else { bsW(s,1,0); }
            } }
      }

      if s.verbosity >= 3 { 
         VPrintf1( "      bytes: mapping %d, ", s.numZ-nBytes ); }
   }

   /*--- Now the selectors. ---*/
   nBytes = s.numZ;
   bsW ( s, 3, nGroups );
   bsW ( s, 15, nSelectors );
   for i = 0; i < nSelectors; i++ { 
      for j = 0; j < s.selectorMtf[i]; j++ { bsW(s,1,1); }
      bsW(s,1,0);
   }
   if s.verbosity >= 3 {
      VPrintf1( "selectors %d, ", s.numZ-nBytes ); }

   /*--- Now the coding tables. ---*/
   nBytes = s.numZ;

   for t = 0; t < nGroups; t++ {
      var curr Int32  = s.len[t][0];
      bsW ( s, 5, curr );
      for i = 0; i < alphaSize; i++ {
         for curr < s.len[t][i] { bsW(s,2,2); curr++; /* 10 */ };
         for curr > s.len[t][i] { bsW(s,2,3); curr--; /* 11 */ };
         bsW ( s, 1, 0 );
      }
   }

   if s.verbosity >= 3 {
      VPrintf1 ( "code lengths %d, ", s.numZ-nBytes ); }

   /*--- And finally, the block data proper ---*/
   nBytes = s.numZ;
   selCtr = 0;
   gs = 0;
   for True {
      if gs >= s.nMTF { break; }
      ge = gs + BZ_G_SIZE - 1; 
      if ge >= s.nMTF { ge = s.nMTF-1; }
      AssertH ( s.selector[selCtr] < nGroups, 3006 );

      if nGroups == 6 && 50 == ge-gs+1 {
            /*--- fast track the common case ---*/
            var mtfv_i UInt16;
            var s_len_sel_selCtr *UChar  
               = &(s.len[s.selector[selCtr]][0]);
            var s_code_sel_selCtr *Int32 
               = &(s.code[s.selector[selCtr]][0]);

// Original macro:
// #           define BZ_ITAH(nn)                      \
//                mtfv_i = mtfv[gs+(nn)];              \
//                bsW ( s,                             \
//                      s_len_sel_selCtr[mtfv_i],      \
//                      s_code_sel_selCtr[mtfv_i] )


            // inlining of macro BZ_ITAH
            mtfv_i = mtfv[gs+(0)];                             bsW ( s,                                                  s_len_sel_selCtr[mtfv_i],                           s_code_sel_selCtr[mtfv_i] );  mtfv_i = mtfv[gs+(1)];                             bsW ( s,                                                  s_len_sel_selCtr[mtfv_i],                           s_code_sel_selCtr[mtfv_i] );  mtfv_i = mtfv[gs+(2)];                             bsW ( s,                                                  s_len_sel_selCtr[mtfv_i],                           s_code_sel_selCtr[mtfv_i] );  mtfv_i = mtfv[gs+(3)];                             bsW ( s,                                                  s_len_sel_selCtr[mtfv_i],                           s_code_sel_selCtr[mtfv_i] );  mtfv_i = mtfv[gs+(4)];                             bsW ( s,                                                  s_len_sel_selCtr[mtfv_i],                           s_code_sel_selCtr[mtfv_i] );
            // inlining of macro BZ_ITAH
            mtfv_i = mtfv[gs+(5)];                             bsW ( s,                                                  s_len_sel_selCtr[mtfv_i],                           s_code_sel_selCtr[mtfv_i] );  mtfv_i = mtfv[gs+(6)];                             bsW ( s,                                                  s_len_sel_selCtr[mtfv_i],                           s_code_sel_selCtr[mtfv_i] );  mtfv_i = mtfv[gs+(7)];                             bsW ( s,                                                  s_len_sel_selCtr[mtfv_i],                           s_code_sel_selCtr[mtfv_i] );  mtfv_i = mtfv[gs+(8)];                             bsW ( s,                                                  s_len_sel_selCtr[mtfv_i],                           s_code_sel_selCtr[mtfv_i] );  mtfv_i = mtfv[gs+(9)];                             bsW ( s,                                                  s_len_sel_selCtr[mtfv_i],                           s_code_sel_selCtr[mtfv_i] );
            // inlining of macro BZ_ITAH
            mtfv_i = mtfv[gs+(10)];                             bsW ( s,                                                  s_len_sel_selCtr[mtfv_i],                           s_code_sel_selCtr[mtfv_i] ); mtfv_i = mtfv[gs+(11)];                             bsW ( s,                                                  s_len_sel_selCtr[mtfv_i],                           s_code_sel_selCtr[mtfv_i] ); mtfv_i = mtfv[gs+(12)];                             bsW ( s,                                                  s_len_sel_selCtr[mtfv_i],                           s_code_sel_selCtr[mtfv_i] ); mtfv_i = mtfv[gs+(13)];                             bsW ( s,                                                  s_len_sel_selCtr[mtfv_i],                           s_code_sel_selCtr[mtfv_i] ); mtfv_i = mtfv[gs+(14)];                             bsW ( s,                                                  s_len_sel_selCtr[mtfv_i],                           s_code_sel_selCtr[mtfv_i] );
            // inlining of macro BZ_ITAH
            mtfv_i = mtfv[gs+(15)];                             bsW ( s,                                                  s_len_sel_selCtr[mtfv_i],                           s_code_sel_selCtr[mtfv_i] ); mtfv_i = mtfv[gs+(16)];                             bsW ( s,                                                  s_len_sel_selCtr[mtfv_i],                           s_code_sel_selCtr[mtfv_i] ); mtfv_i = mtfv[gs+(17)];                             bsW ( s,                                                  s_len_sel_selCtr[mtfv_i],                           s_code_sel_selCtr[mtfv_i] ); mtfv_i = mtfv[gs+(18)];                             bsW ( s,                                                  s_len_sel_selCtr[mtfv_i],                           s_code_sel_selCtr[mtfv_i] ); mtfv_i = mtfv[gs+(19)];                             bsW ( s,                                                  s_len_sel_selCtr[mtfv_i],                           s_code_sel_selCtr[mtfv_i] );
            // inlining of macro BZ_ITAH
            mtfv_i = mtfv[gs+(20)];                             bsW ( s,                                                  s_len_sel_selCtr[mtfv_i],                           s_code_sel_selCtr[mtfv_i] ); mtfv_i = mtfv[gs+(21)];                             bsW ( s,                                                  s_len_sel_selCtr[mtfv_i],                           s_code_sel_selCtr[mtfv_i] ); mtfv_i = mtfv[gs+(22)];                             bsW ( s,                                                  s_len_sel_selCtr[mtfv_i],                           s_code_sel_selCtr[mtfv_i] ); mtfv_i = mtfv[gs+(23)];                             bsW ( s,                                                  s_len_sel_selCtr[mtfv_i],                           s_code_sel_selCtr[mtfv_i] ); mtfv_i = mtfv[gs+(24)];                             bsW ( s,                                                  s_len_sel_selCtr[mtfv_i],                           s_code_sel_selCtr[mtfv_i] );
            // inlining of macro BZ_ITAH
            mtfv_i = mtfv[gs+(25)];                             bsW ( s,                                                  s_len_sel_selCtr[mtfv_i],                           s_code_sel_selCtr[mtfv_i] ); mtfv_i = mtfv[gs+(26)];                             bsW ( s,                                                  s_len_sel_selCtr[mtfv_i],                           s_code_sel_selCtr[mtfv_i] ); mtfv_i = mtfv[gs+(27)];                             bsW ( s,                                                  s_len_sel_selCtr[mtfv_i],                           s_code_sel_selCtr[mtfv_i] ); mtfv_i = mtfv[gs+(28)];                             bsW ( s,                                                  s_len_sel_selCtr[mtfv_i],                           s_code_sel_selCtr[mtfv_i] ); mtfv_i = mtfv[gs+(29)];                             bsW ( s,                                                  s_len_sel_selCtr[mtfv_i],                           s_code_sel_selCtr[mtfv_i] );
            // inlining of macro BZ_ITAH
            mtfv_i = mtfv[gs+(30)];                             bsW ( s,                                                  s_len_sel_selCtr[mtfv_i],                           s_code_sel_selCtr[mtfv_i] ); mtfv_i = mtfv[gs+(31)];                             bsW ( s,                                                  s_len_sel_selCtr[mtfv_i],                           s_code_sel_selCtr[mtfv_i] ); mtfv_i = mtfv[gs+(32)];                             bsW ( s,                                                  s_len_sel_selCtr[mtfv_i],                           s_code_sel_selCtr[mtfv_i] ); mtfv_i = mtfv[gs+(33)];                             bsW ( s,                                                  s_len_sel_selCtr[mtfv_i],                           s_code_sel_selCtr[mtfv_i] ); mtfv_i = mtfv[gs+(34)];                             bsW ( s,                                                  s_len_sel_selCtr[mtfv_i],                           s_code_sel_selCtr[mtfv_i] );
            // inlining of macro BZ_ITAH
            mtfv_i = mtfv[gs+(35)];                             bsW ( s,                                                  s_len_sel_selCtr[mtfv_i],                           s_code_sel_selCtr[mtfv_i] ); mtfv_i = mtfv[gs+(36)];                             bsW ( s,                                                  s_len_sel_selCtr[mtfv_i],                           s_code_sel_selCtr[mtfv_i] ); mtfv_i = mtfv[gs+(37)];                             bsW ( s,                                                  s_len_sel_selCtr[mtfv_i],                           s_code_sel_selCtr[mtfv_i] ); mtfv_i = mtfv[gs+(38)];                             bsW ( s,                                                  s_len_sel_selCtr[mtfv_i],                           s_code_sel_selCtr[mtfv_i] ); mtfv_i = mtfv[gs+(39)];                             bsW ( s,                                                  s_len_sel_selCtr[mtfv_i],                           s_code_sel_selCtr[mtfv_i] );
            // inlining of macro BZ_ITAH
            mtfv_i = mtfv[gs+(40)];                             bsW ( s,                                                  s_len_sel_selCtr[mtfv_i],                           s_code_sel_selCtr[mtfv_i] ); mtfv_i = mtfv[gs+(41)];                             bsW ( s,                                                  s_len_sel_selCtr[mtfv_i],                           s_code_sel_selCtr[mtfv_i] ); mtfv_i = mtfv[gs+(42)];                             bsW ( s,                                                  s_len_sel_selCtr[mtfv_i],                           s_code_sel_selCtr[mtfv_i] ); mtfv_i = mtfv[gs+(43)];                             bsW ( s,                                                  s_len_sel_selCtr[mtfv_i],                           s_code_sel_selCtr[mtfv_i] ); mtfv_i = mtfv[gs+(44)];                             bsW ( s,                                                  s_len_sel_selCtr[mtfv_i],                           s_code_sel_selCtr[mtfv_i] );
            // inlining of macro BZ_ITAH
            mtfv_i = mtfv[gs+(45)];                             bsW ( s,                                                  s_len_sel_selCtr[mtfv_i],                           s_code_sel_selCtr[mtfv_i] ); mtfv_i = mtfv[gs+(46)];                             bsW ( s,                                                  s_len_sel_selCtr[mtfv_i],                           s_code_sel_selCtr[mtfv_i] ); mtfv_i = mtfv[gs+(47)];                             bsW ( s,                                                  s_len_sel_selCtr[mtfv_i],                           s_code_sel_selCtr[mtfv_i] ); mtfv_i = mtfv[gs+(48)];                             bsW ( s,                                                  s_len_sel_selCtr[mtfv_i],                           s_code_sel_selCtr[mtfv_i] ); mtfv_i = mtfv[gs+(49)];                             bsW ( s,                                                  s_len_sel_selCtr[mtfv_i],                           s_code_sel_selCtr[mtfv_i] );

#           undef BZ_ITAH

      } else {
     /*--- slow version which correctly handles all situations ---*/
         for i = gs; i <= ge; i++ {
            bsW ( s, 
                  s.len  [s.selector[selCtr]] [mtfv[i]],
                  s.code [s.selector[selCtr]] [mtfv[i]] );
         }
      }


      gs = ge+1;
      selCtr++;
   }
   AssertH( selCtr == nSelectors, 3007 );

   if s.verbosity >= 3 {
      VPrintf1( "codes %d\n", s.numZ-nBytes ); }
}


/*---------------------------------------------------*/
func BZ2_compressBlock(s *EState, is_last_block Bool) {
   if s.nblock > 0 {

      BZ_FINALISE_CRC ( s.blockCRC );
      s.combinedCRC = (s.combinedCRC << 1) | (s.combinedCRC >> 31);
      s.combinedCRC ^= s.blockCRC;
      if s.blockNo > 1 { s.numZ = 0; }

      if s.verbosity >= 2 {
         VPrintf4( "    block %d: crc = 0x%08x, "
                   "combined CRC = 0x%08x, size = %d\n",
                   s.blockNo, s.blockCRC, s.combinedCRC, s.nblock ); }

      BZ2_blockSort ( s );
   }

   s.zbits = (UChar*) (&((UChar*)s.arr2)[s.nblock]);

   /*-- If this is the first block, create the stream header. --*/
   if s.blockNo == 1 {
      BZ2_bsInitWrite ( s );
      bsPutUChar ( s, BZ_HDR_B );
      bsPutUChar ( s, BZ_HDR_Z );
      bsPutUChar ( s, BZ_HDR_h );
      bsPutUChar ( s, (UChar)(BZ_HDR_0 + s.blockSize100k) );
   }

   if s.nblock > 0 {

      bsPutUChar ( s, 0x31 ); bsPutUChar ( s, 0x41 );
      bsPutUChar ( s, 0x59 ); bsPutUChar ( s, 0x26 );
      bsPutUChar ( s, 0x53 ); bsPutUChar ( s, 0x59 );

      /*-- Now the block's CRC, so it is in a known place. --*/
      bsPutUInt32 ( s, s.blockCRC );

      /*-- 
         Now a single bit indicating (non-)randomisation. 
         As of version 0.9.5, we use a better sorting algorithm
         which makes randomisation unnecessary.  So always set
         the randomised bit to 'no'.  Of course, the decoder
         still needs to be able to handle randomised blocks
         so as to maintain backwards compatibility with
         older versions of bzip2.
      --*/
      bsW(s,1,0);

      bsW ( s, 24, s.origPtr );
      generateMTFValues ( s );
      sendMTFValues ( s );
   }


   /*-- If this is the last block, add the stream trailer. --*/
   if is_last_block {

      bsPutUChar ( s, 0x17 ); bsPutUChar ( s, 0x72 );
      bsPutUChar ( s, 0x45 ); bsPutUChar ( s, 0x38 );
      bsPutUChar ( s, 0x50 ); bsPutUChar ( s, 0x90 );
      bsPutUInt32 ( s, s.combinedCRC );
      if s.verbosity >= 2 {
         VPrintf1( "    final combined CRC = 0x%08x\n   ", s.combinedCRC ); }
      bsFinishWrite ( s );
   }
}


/*-------------------------------------------------------------*/
/*--- end                                        compress.c ---*/
/*-------------------------------------------------------------*/
//ABC_NAMESPACE_IMPL_END

