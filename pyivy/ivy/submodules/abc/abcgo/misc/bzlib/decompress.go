package bzlib


/*-------------------------------------------------------------*/
/*--- Decompression machinery                               ---*/
/*---                                          decompress.c ---*/
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



/*---------------------------------------------------*/
static
func makeMaps_d(s *DState) {
   var i Int32;
   s.nInUse = 0;
   for i = 0; i < 256; i++ {
      if s.inUse[i] {
         s.seqToUnseq[s.nInUse] = i;
         s.nInUse++;
      }
   }
}


/*---------------------------------------------------*/
// Original macro:
// #define RETURN(rrr)                               \
//    { retVal = rrr; goto save_state_and_return; };


// Original macro:
// #define GET_BITS(lll,vvv,nnn)                     \
//    case lll: s->state = lll;                      \
//    while (True) {                                 \
//       if (s->bsLive >= nnn) {                     \
//          UInt32 v;                                \
//          v = (s->bsBuff >>                        \
//              (s->bsLive-nnn)) & ((1 << nnn)-1);   \
//          s->bsLive -= nnn;                        \
//          vvv = v;                                 \
//          break;                                   \
//       }                                           \
//       if (s->strm->avail_in == 0) RETURN(BZ_OK);  \
//       s->bsBuff                                   \
//          = (s->bsBuff << 8) |                     \
//            ((UInt32)                              \
//               (*((UChar*)(s->strm->next_in))));   \
//       s->bsLive += 8;                             \
//       s->strm->next_in++;                         \
//       s->strm->avail_in--;                        \
//       s->strm->total_in_lo32++;                   \
//       if (s->strm->total_in_lo32 == 0)            \
//          s->strm->total_in_hi32++;                \
//    }


// Original macro:
// #define GET_UCHAR(lll,uuu)                        \
//    GET_BITS(lll,uuu,8)


// Original macro:
// #define GET_BIT(lll,uuu)                          \
//    GET_BITS(lll,uuu,1)


/*---------------------------------------------------*/
// Original macro:
// #define GET_MTF_VAL(label1,label2,lval)           \
// {                                                 \
//    if (groupPos == 0) {                           \
//       groupNo++;                                  \
//       if (groupNo >= nSelectors)                  \
//          RETURN(BZ_DATA_ERROR);                   \
//       groupPos = BZ_G_SIZE;                       \
//       gSel = s->selector[groupNo];                \
//       gMinlen = s->minLens[gSel];                 \
//       gLimit = &(s->limit[gSel][0]);              \
//       gPerm = &(s->perm[gSel][0]);                \
//       gBase = &(s->base[gSel][0]);                \
//    }                                              \
//    groupPos--;                                    \
//    zn = gMinlen;                                  \
//    GET_BITS(label1, zvec, zn);                    \
//    while (1) {                                    \
//       if (zn > 20 /* the longest code */)         \
//          RETURN(BZ_DATA_ERROR);                   \
//       if (zvec <= gLimit[zn]) break;              \
//       zn++;                                       \
//       GET_BIT(label2, zj);                        \
//       zvec = (zvec << 1) | zj;                    \
//    };                                             \
//    if (zvec - gBase[zn] < 0                       \
//        || zvec - gBase[zn] >= BZ_MAX_ALPHA_SIZE)  \
//       RETURN(BZ_DATA_ERROR);                      \
//    lval = gPerm[zvec - gBase[zn]];                \
// }



/*---------------------------------------------------*/
func BZ2_decompress(s *DState) Int32 {
   var uc UChar;
   var retVal Int32;
   var minLen, maxLen Int32;
   var strm *bz_stream  = s.strm;

   /* stuff that needs to be saved/restored */
   var i Int32;
   var j Int32;
   var t Int32;
   var alphaSize Int32;
   var nGroups Int32;
   var nSelectors Int32;
   var EOB Int32;
   var groupNo Int32;
   var groupPos Int32;
   var nextSym Int32;
   var nblockMAX Int32;
   var nblock Int32;
   var es Int32;
   var N Int32;
   var curr Int32;
   var zt Int32;
   var zn Int32; 
   var zvec Int32;
   var zj Int32;
   var gSel Int32;
   var gMinlen Int32;
   var gLimit *Int32;
   var gBase *Int32;
   var gPerm *Int32;

   if s.state == BZ_X_MAGIC_1 {
      /*initialise the save area*/
      s.save_i           = 0;
      s.save_j           = 0;
      s.save_t           = 0;
      s.save_alphaSize   = 0;
      s.save_nGroups     = 0;
      s.save_nSelectors  = 0;
      s.save_EOB         = 0;
      s.save_groupNo     = 0;
      s.save_groupPos    = 0;
      s.save_nextSym     = 0;
      s.save_nblockMAX   = 0;
      s.save_nblock      = 0;
      s.save_es          = 0;
      s.save_N           = 0;
      s.save_curr        = 0;
      s.save_zt          = 0;
      s.save_zn          = 0;
      s.save_zvec        = 0;
      s.save_zj          = 0;
      s.save_gSel        = 0;
      s.save_gMinlen     = 0;
      s.save_gLimit      = nil;
      s.save_gBase       = nil;
      s.save_gPerm       = nil;
   }

   /*restore from the save area*/
   i           = s.save_i;
   j           = s.save_j;
   t           = s.save_t;
   alphaSize   = s.save_alphaSize;
   nGroups     = s.save_nGroups;
   nSelectors  = s.save_nSelectors;
   EOB         = s.save_EOB;
   groupNo     = s.save_groupNo;
   groupPos    = s.save_groupPos;
   nextSym     = s.save_nextSym;
   nblockMAX   = s.save_nblockMAX;
   nblock      = s.save_nblock;
   es          = s.save_es;
   N           = s.save_N;
   curr        = s.save_curr;
   zt          = s.save_zt;
   zn          = s.save_zn; 
   zvec        = s.save_zvec;
   zj          = s.save_zj;
   gSel        = s.save_gSel;
   gMinlen     = s.save_gMinlen;
   gLimit      = s.save_gLimit;
   gBase       = s.save_gBase;
   gPerm       = s.save_gPerm;

   retVal = BZ_OK;

   switch s.state {

      // inlining of macro GET_UCHAR
      // inlining of macro GET_BITS
      case BZ_X_MAGIC_1: s.state = BZ_X_MAGIC_1;                         for True {                                       if s.bsLive >= 8 {                              var v UInt32;                                         v = (s.bsBuff >>                                     (s.bsLive-8)) & ((1 << 8)-1);            s.bsLive -= 8;                                 uc = v;                                          break;                                         }
                                                 // inlining of macro RETURN
                                                 if s.strm.avail_in == 0 { retVal = BZ_OK; goto save_state_and_return; };;        s.bsBuff                                            = (s.bsBuff << 8) |                                ((UInt32)                                            (*((UChar*)(s.strm.next_in))));         s.bsLive += 8;                                   s.strm.next_in++;                               s.strm.avail_in--;                              s.strm.total_in_lo32++;                         if s.strm.total_in_lo32 == 0 {                     s.strm.total_in_hi32++; }                   };
      // inlining of macro RETURN
      if uc != BZ_HDR_B { retVal = BZ_DATA_ERROR_MAGIC; goto save_state_and_return; };;

      // inlining of macro GET_UCHAR
      // inlining of macro GET_BITS
      fallthrough
case BZ_X_MAGIC_2: s.state = BZ_X_MAGIC_2;                         for True {                                       if s.bsLive >= 8 {                              var v UInt32;                                         v = (s.bsBuff >>                                     (s.bsLive-8)) & ((1 << 8)-1);            s.bsLive -= 8;                                 uc = v;                                          break;                                         }
                                                 // inlining of macro RETURN
                                                 if s.strm.avail_in == 0 { retVal = BZ_OK; goto save_state_and_return; };;        s.bsBuff                                            = (s.bsBuff << 8) |                                ((UInt32)                                            (*((UChar*)(s.strm.next_in))));         s.bsLive += 8;                                   s.strm.next_in++;                               s.strm.avail_in--;                              s.strm.total_in_lo32++;                         if s.strm.total_in_lo32 == 0 {                     s.strm.total_in_hi32++; }                   };
      // inlining of macro RETURN
      if uc != BZ_HDR_Z { retVal = BZ_DATA_ERROR_MAGIC; goto save_state_and_return; };;

      // inlining of macro GET_UCHAR
      // inlining of macro GET_BITS
      fallthrough
case BZ_X_MAGIC_3: s.state = BZ_X_MAGIC_3;                         for True {                                       if s.bsLive >= 8 {                              var v UInt32;                                         v = (s.bsBuff >>                                     (s.bsLive-8)) & ((1 << 8)-1);            s.bsLive -= 8;                                 uc = v;                                          break;                                         }
                                                 // inlining of macro RETURN
                                                 if s.strm.avail_in == 0 { retVal = BZ_OK; goto save_state_and_return; };;        s.bsBuff                                            = (s.bsBuff << 8) |                                ((UInt32)                                            (*((UChar*)(s.strm.next_in))));         s.bsLive += 8;                                   s.strm.next_in++;                               s.strm.avail_in--;                              s.strm.total_in_lo32++;                         if s.strm.total_in_lo32 == 0 {                     s.strm.total_in_hi32++; }                   }
      // inlining of macro RETURN
      if uc != BZ_HDR_h { retVal = BZ_DATA_ERROR_MAGIC; goto save_state_and_return; };;

      // inlining of macro GET_BITS
      fallthrough
case BZ_X_MAGIC_4: s.state = BZ_X_MAGIC_4;                         for True {                                       if s.bsLive >= 8 {                              var v UInt32;                                         v = (s.bsBuff >>                                     (s.bsLive-8)) & ((1 << 8)-1);            s.bsLive -= 8;                                 s.blockSize100k = v;                                          break;                                         }
                                                 // inlining of macro RETURN
                                                 if s.strm.avail_in == 0 { retVal = BZ_OK; goto save_state_and_return; };;        s.bsBuff                                            = (s.bsBuff << 8) |                                ((UInt32)                                            (*((UChar*)(s.strm.next_in))));         s.bsLive += 8;                                   s.strm.next_in++;                               s.strm.avail_in--;                              s.strm.total_in_lo32++;                         if s.strm.total_in_lo32 == 0 {                     s.strm.total_in_hi32++; }                   }
      if s.blockSize100k < (BZ_HDR_0 + 1) || 
          // inlining of macro RETURN
          s.blockSize100k > (BZ_HDR_0 + 9) { retVal = BZ_DATA_ERROR_MAGIC; goto save_state_and_return; };;
      s.blockSize100k -= BZ_HDR_0;

      if s.smallDecompress {
         s.ll16 = (unsigned short *)BZALLOC( s.blockSize100k * 100000 * sizeof(UInt16) );
         s.ll4  = (unsigned char *)BZALLOC( 
                      ((1 + s.blockSize100k * 100000) >> 1) * sizeof(UChar) 
                   );
         // inlining of macro RETURN
         if s.ll16 == nil || s.ll4 == nil { retVal = BZ_MEM_ERROR; goto save_state_and_return; };;
      } else {
         s.tt  = (unsigned *)BZALLOC( s.blockSize100k * 100000 * sizeof(Int32) );
         // inlining of macro RETURN
         if s.tt == nil { retVal = BZ_MEM_ERROR; goto save_state_and_return; };;
      }

      // inlining of macro GET_UCHAR
      // inlining of macro GET_BITS
      fallthrough
case BZ_X_BLKHDR_1: s.state = BZ_X_BLKHDR_1;                         for True {                                       if s.bsLive >= 8 {                              var v UInt32;                                         v = (s.bsBuff >>                                     (s.bsLive-8)) & ((1 << 8)-1);            s.bsLive -= 8;                                 uc = v;                                          break;                                         }
                                                 // inlining of macro RETURN
                                                 if s.strm.avail_in == 0 { retVal = BZ_OK; goto save_state_and_return; };;        s.bsBuff                                            = (s.bsBuff << 8) |                                ((UInt32)                                            (*((UChar*)(s.strm.next_in))));         s.bsLive += 8;                                   s.strm.next_in++;                               s.strm.avail_in--;                              s.strm.total_in_lo32++;                         if s.strm.total_in_lo32 == 0 {                     s.strm.total_in_hi32++; }                   };

      if uc == 0x17 { goto endhdr_2; }
      // inlining of macro RETURN
      if uc != 0x31 { retVal = BZ_DATA_ERROR; goto save_state_and_return; };;
      // inlining of macro GET_UCHAR
      // inlining of macro GET_BITS
      fallthrough
case BZ_X_BLKHDR_2: s.state = BZ_X_BLKHDR_2;                         for True {                                       if s.bsLive >= 8 {                              var v UInt32;                                         v = (s.bsBuff >>                                     (s.bsLive-8)) & ((1 << 8)-1);            s.bsLive -= 8;                                 uc = v;                                          break;                                         }
                                                 // inlining of macro RETURN
                                                 if s.strm.avail_in == 0 { retVal = BZ_OK; goto save_state_and_return; };;        s.bsBuff                                            = (s.bsBuff << 8) |                                ((UInt32)                                            (*((UChar*)(s.strm.next_in))));         s.bsLive += 8;                                   s.strm.next_in++;                               s.strm.avail_in--;                              s.strm.total_in_lo32++;                         if s.strm.total_in_lo32 == 0 {                     s.strm.total_in_hi32++; }                   };
      // inlining of macro RETURN
      if uc != 0x41 { retVal = BZ_DATA_ERROR; goto save_state_and_return; };;
      // inlining of macro GET_UCHAR
      // inlining of macro GET_BITS
      fallthrough
case BZ_X_BLKHDR_3: s.state = BZ_X_BLKHDR_3;                         for True {                                       if s.bsLive >= 8 {                              var v UInt32;                                         v = (s.bsBuff >>                                     (s.bsLive-8)) & ((1 << 8)-1);            s.bsLive -= 8;                                 uc = v;                                          break;                                         }
                                                 // inlining of macro RETURN
                                                 if s.strm.avail_in == 0 { retVal = BZ_OK; goto save_state_and_return; };;        s.bsBuff                                            = (s.bsBuff << 8) |                                ((UInt32)                                            (*((UChar*)(s.strm.next_in))));         s.bsLive += 8;                                   s.strm.next_in++;                               s.strm.avail_in--;                              s.strm.total_in_lo32++;                         if s.strm.total_in_lo32 == 0 {                     s.strm.total_in_hi32++; }                   };
      // inlining of macro RETURN
      if uc != 0x59 { retVal = BZ_DATA_ERROR; goto save_state_and_return; };;
      // inlining of macro GET_UCHAR
      // inlining of macro GET_BITS
      fallthrough
case BZ_X_BLKHDR_4: s.state = BZ_X_BLKHDR_4;                         for True {                                       if s.bsLive >= 8 {                              var v UInt32;                                         v = (s.bsBuff >>                                     (s.bsLive-8)) & ((1 << 8)-1);            s.bsLive -= 8;                                 uc = v;                                          break;                                         }
                                                 // inlining of macro RETURN
                                                 if s.strm.avail_in == 0 { retVal = BZ_OK; goto save_state_and_return; };;        s.bsBuff                                            = (s.bsBuff << 8) |                                ((UInt32)                                            (*((UChar*)(s.strm.next_in))));         s.bsLive += 8;                                   s.strm.next_in++;                               s.strm.avail_in--;                              s.strm.total_in_lo32++;                         if s.strm.total_in_lo32 == 0 {                     s.strm.total_in_hi32++; }                   };
      // inlining of macro RETURN
      if uc != 0x26 { retVal = BZ_DATA_ERROR; goto save_state_and_return; };;
      // inlining of macro GET_UCHAR
      // inlining of macro GET_BITS
      fallthrough
case BZ_X_BLKHDR_5: s.state = BZ_X_BLKHDR_5;                         for True {                                       if s.bsLive >= 8 {                              var v UInt32;                                         v = (s.bsBuff >>                                     (s.bsLive-8)) & ((1 << 8)-1);            s.bsLive -= 8;                                 uc = v;                                          break;                                         }
                                                 // inlining of macro RETURN
                                                 if s.strm.avail_in == 0 { retVal = BZ_OK; goto save_state_and_return; };;        s.bsBuff                                            = (s.bsBuff << 8) |                                ((UInt32)                                            (*((UChar*)(s.strm.next_in))));         s.bsLive += 8;                                   s.strm.next_in++;                               s.strm.avail_in--;                              s.strm.total_in_lo32++;                         if s.strm.total_in_lo32 == 0 {                     s.strm.total_in_hi32++; }                   };
      // inlining of macro RETURN
      if uc != 0x53 { retVal = BZ_DATA_ERROR; goto save_state_and_return; };;
      // inlining of macro GET_UCHAR
      // inlining of macro GET_BITS
      fallthrough
case BZ_X_BLKHDR_6: s.state = BZ_X_BLKHDR_6;                         for True {                                       if s.bsLive >= 8 {                              var v UInt32;                                         v = (s.bsBuff >>                                     (s.bsLive-8)) & ((1 << 8)-1);            s.bsLive -= 8;                                 uc = v;                                          break;                                         }
                                                 // inlining of macro RETURN
                                                 if s.strm.avail_in == 0 { retVal = BZ_OK; goto save_state_and_return; };;        s.bsBuff                                            = (s.bsBuff << 8) |                                ((UInt32)                                            (*((UChar*)(s.strm.next_in))));         s.bsLive += 8;                                   s.strm.next_in++;                               s.strm.avail_in--;                              s.strm.total_in_lo32++;                         if s.strm.total_in_lo32 == 0 {                     s.strm.total_in_hi32++; }                   };
      // inlining of macro RETURN
      if uc != 0x59 { retVal = BZ_DATA_ERROR; goto save_state_and_return; };;

      s.currBlockNo++;
      if s.verbosity >= 2 {
         VPrintf1 ( "\n    [%d: huff+mtf ", s.currBlockNo ); }
 
      s.storedBlockCRC = 0;
      // inlining of macro GET_UCHAR
      // inlining of macro GET_BITS
      fallthrough
case BZ_X_BCRC_1: s.state = BZ_X_BCRC_1;                         for True {                                       if s.bsLive >= 8 {                              var v UInt32;                                         v = (s.bsBuff >>                                     (s.bsLive-8)) & ((1 << 8)-1);            s.bsLive -= 8;                                 uc = v;                                          break;                                         }
                                                 // inlining of macro RETURN
                                                 if s.strm.avail_in == 0 { retVal = BZ_OK; goto save_state_and_return; };;        s.bsBuff                                            = (s.bsBuff << 8) |                                ((UInt32)                                            (*((UChar*)(s.strm.next_in))));         s.bsLive += 8;                                   s.strm.next_in++;                               s.strm.avail_in--;                              s.strm.total_in_lo32++;                         if s.strm.total_in_lo32 == 0 {                     s.strm.total_in_hi32++; }                   };
      s.storedBlockCRC = (s.storedBlockCRC << 8) | ((UInt32)uc);
      // inlining of macro GET_UCHAR
      // inlining of macro GET_BITS
      fallthrough
case BZ_X_BCRC_2: s.state = BZ_X_BCRC_2;                         for True {                                       if s.bsLive >= 8 {                              var v UInt32;                                         v = (s.bsBuff >>                                     (s.bsLive-8)) & ((1 << 8)-1);            s.bsLive -= 8;                                 uc = v;                                          break;                                         }
                                                 // inlining of macro RETURN
                                                 if s.strm.avail_in == 0 { retVal = BZ_OK; goto save_state_and_return; };;        s.bsBuff                                            = (s.bsBuff << 8) |                                ((UInt32)                                            (*((UChar*)(s.strm.next_in))));         s.bsLive += 8;                                   s.strm.next_in++;                               s.strm.avail_in--;                              s.strm.total_in_lo32++;                         if s.strm.total_in_lo32 == 0 {                     s.strm.total_in_hi32++; }                   };
      s.storedBlockCRC = (s.storedBlockCRC << 8) | ((UInt32)uc);
      // inlining of macro GET_UCHAR
      // inlining of macro GET_BITS
      fallthrough
case BZ_X_BCRC_3: s.state = BZ_X_BCRC_3;                         for True {                                       if s.bsLive >= 8 {                              var v UInt32;                                         v = (s.bsBuff >>                                     (s.bsLive-8)) & ((1 << 8)-1);            s.bsLive -= 8;                                 uc = v;                                          break;                                         }
                                                 // inlining of macro RETURN
                                                 if s.strm.avail_in == 0 { retVal = BZ_OK; goto save_state_and_return; };;        s.bsBuff                                            = (s.bsBuff << 8) |                                ((UInt32)                                            (*((UChar*)(s.strm.next_in))));         s.bsLive += 8;                                   s.strm.next_in++;                               s.strm.avail_in--;                              s.strm.total_in_lo32++;                         if s.strm.total_in_lo32 == 0 {                     s.strm.total_in_hi32++; }                   };
      s.storedBlockCRC = (s.storedBlockCRC << 8) | ((UInt32)uc);
      // inlining of macro GET_UCHAR
      // inlining of macro GET_BITS
      fallthrough
case BZ_X_BCRC_4: s.state = BZ_X_BCRC_4;                         for True {                                       if s.bsLive >= 8 {                              var v UInt32;                                         v = (s.bsBuff >>                                     (s.bsLive-8)) & ((1 << 8)-1);            s.bsLive -= 8;                                 uc = v;                                          break;                                         }
                                                 // inlining of macro RETURN
                                                 if s.strm.avail_in == 0 { retVal = BZ_OK; goto save_state_and_return; };;        s.bsBuff                                            = (s.bsBuff << 8) |                                ((UInt32)                                            (*((UChar*)(s.strm.next_in))));         s.bsLive += 8;                                   s.strm.next_in++;                               s.strm.avail_in--;                              s.strm.total_in_lo32++;                         if s.strm.total_in_lo32 == 0 {                     s.strm.total_in_hi32++; }                   };
      s.storedBlockCRC = (s.storedBlockCRC << 8) | ((UInt32)uc);

      // inlining of macro GET_BITS
      fallthrough
case BZ_X_RANDBIT: s.state = BZ_X_RANDBIT;                         for True {                                       if s.bsLive >= 1 {                              var v UInt32;                                         v = (s.bsBuff >>                                     (s.bsLive-1)) & ((1 << 1)-1);            s.bsLive -= 1;                                 s.blockRandomised = v;                                          break;                                         }
                                                 // inlining of macro RETURN
                                                 if s.strm.avail_in == 0 { retVal = BZ_OK; goto save_state_and_return; };;        s.bsBuff                                            = (s.bsBuff << 8) |                                ((UInt32)                                            (*((UChar*)(s.strm.next_in))));         s.bsLive += 8;                                   s.strm.next_in++;                               s.strm.avail_in--;                              s.strm.total_in_lo32++;                         if s.strm.total_in_lo32 == 0 {                     s.strm.total_in_hi32++; }                   };

      s.origPtr = 0;
      // inlining of macro GET_UCHAR
      // inlining of macro GET_BITS
      fallthrough
case BZ_X_ORIGPTR_1: s.state = BZ_X_ORIGPTR_1;                         for True {                                       if s.bsLive >= 8 {                              var v UInt32;                                         v = (s.bsBuff >>                                     (s.bsLive-8)) & ((1 << 8)-1);            s.bsLive -= 8;                                 uc = v;                                          break;                                         }
                                                 // inlining of macro RETURN
                                                 if s.strm.avail_in == 0 { retVal = BZ_OK; goto save_state_and_return; };;        s.bsBuff                                            = (s.bsBuff << 8) |                                ((UInt32)                                            (*((UChar*)(s.strm.next_in))));         s.bsLive += 8;                                   s.strm.next_in++;                               s.strm.avail_in--;                              s.strm.total_in_lo32++;                         if s.strm.total_in_lo32 == 0 {                     s.strm.total_in_hi32++; }                   };
      s.origPtr = (s.origPtr << 8) | ((Int32)uc);
      // inlining of macro GET_UCHAR
      // inlining of macro GET_BITS
      fallthrough
case BZ_X_ORIGPTR_2: s.state = BZ_X_ORIGPTR_2;                         for True {                                       if s.bsLive >= 8 {                              var v UInt32;                                         v = (s.bsBuff >>                                     (s.bsLive-8)) & ((1 << 8)-1);            s.bsLive -= 8;                                 uc = v;                                          break;                                         }
                                                 // inlining of macro RETURN
                                                 if s.strm.avail_in == 0 { retVal = BZ_OK; goto save_state_and_return; };;        s.bsBuff                                            = (s.bsBuff << 8) |                                ((UInt32)                                            (*((UChar*)(s.strm.next_in))));         s.bsLive += 8;                                   s.strm.next_in++;                               s.strm.avail_in--;                              s.strm.total_in_lo32++;                         if s.strm.total_in_lo32 == 0 {                     s.strm.total_in_hi32++; }                   };
      s.origPtr = (s.origPtr << 8) | ((Int32)uc);
      // inlining of macro GET_UCHAR
      // inlining of macro GET_BITS
      fallthrough
case BZ_X_ORIGPTR_3: s.state = BZ_X_ORIGPTR_3;                         for True {                                       if s.bsLive >= 8 {                              var v UInt32;                                         v = (s.bsBuff >>                                     (s.bsLive-8)) & ((1 << 8)-1);            s.bsLive -= 8;                                 uc = v;                                          break;                                         }
                                                 // inlining of macro RETURN
                                                 if s.strm.avail_in == 0 { retVal = BZ_OK; goto save_state_and_return; };;        s.bsBuff                                            = (s.bsBuff << 8) |                                ((UInt32)                                            (*((UChar*)(s.strm.next_in))));         s.bsLive += 8;                                   s.strm.next_in++;                               s.strm.avail_in--;                              s.strm.total_in_lo32++;                         if s.strm.total_in_lo32 == 0 {                     s.strm.total_in_hi32++; }                   };
      s.origPtr = (s.origPtr << 8) | ((Int32)uc);

      if s.origPtr < 0 {
         // inlining of macro RETURN
          retVal = BZ_DATA_ERROR; goto save_state_and_return; };;
      if s.origPtr > 10 + 100000*s.blockSize100k { 
         // inlining of macro RETURN
          retVal = BZ_DATA_ERROR; goto save_state_and_return; };;

      /*--- Receive the mapping table ---*/
      for i = 0; i < 16; i++ {
         // inlining of macro GET_BIT
         // inlining of macro GET_BITS
         case BZ_X_MAPPING_1: s.state = BZ_X_MAPPING_1;                         for True {                                       if s.bsLive >= 1 {                              var v UInt32;                                         v = (s.bsBuff >>                                     (s.bsLive-1)) & ((1 << 1)-1);            s.bsLive -= 1;                                 uc = v;                                          break;                                         }
                                                 // inlining of macro RETURN
                                                 if s.strm.avail_in == 0 { retVal = BZ_OK; goto save_state_and_return; };;        s.bsBuff                                            = (s.bsBuff << 8) |                                ((UInt32)                                            (*((UChar*)(s.strm.next_in))));         s.bsLive += 8;                                   s.strm.next_in++;                               s.strm.avail_in--;                              s.strm.total_in_lo32++;                         if s.strm.total_in_lo32 == 0 {                     s.strm.total_in_hi32++; }                   };
         if uc == 1 { 
            s.inUse16[i] = True; } else { 
            s.inUse16[i] = False; }
      }

      for i = 0; i < 256; i++ { s.inUse[i] = False; }

      for i = 0; i < 16; i++ {
         if s.inUse16[i] {
            for j = 0; j < 16; j++ {
               // inlining of macro GET_BIT
               // inlining of macro GET_BITS
               case BZ_X_MAPPING_2: s.state = BZ_X_MAPPING_2;                         for True {                                       if s.bsLive >= 1 {                              var v UInt32;                                         v = (s.bsBuff >>                                     (s.bsLive-1)) & ((1 << 1)-1);            s.bsLive -= 1;                                 uc = v;                                          break;                                         }
                                                 // inlining of macro RETURN
                                                 if s.strm.avail_in == 0 { retVal = BZ_OK; goto save_state_and_return; };;        s.bsBuff                                            = (s.bsBuff << 8) |                                ((UInt32)                                            (*((UChar*)(s.strm.next_in))));         s.bsLive += 8;                                   s.strm.next_in++;                               s.strm.avail_in--;                              s.strm.total_in_lo32++;                         if s.strm.total_in_lo32 == 0 {                     s.strm.total_in_hi32++; }                   };
               if uc == 1 { s.inUse[i * 16 + j] = True; }
            } }
      }
      makeMaps_d ( s );
      // inlining of macro RETURN
      if s.nInUse == 0 { retVal = BZ_DATA_ERROR; goto save_state_and_return; };;
      alphaSize = s.nInUse+2;

      /*--- Now the selectors ---*/
      // inlining of macro GET_BITS
      fallthrough
case BZ_X_SELECTOR_1: s.state = BZ_X_SELECTOR_1;                         for True {                                       if s.bsLive >= 3 {                              var v UInt32;                                         v = (s.bsBuff >>                                     (s.bsLive-3)) & ((1 << 3)-1);            s.bsLive -= 3;                                 nGroups = v;                                          break;                                         }
                                                 // inlining of macro RETURN
                                                 if s.strm.avail_in == 0 { retVal = BZ_OK; goto save_state_and_return; };;        s.bsBuff                                            = (s.bsBuff << 8) |                                ((UInt32)                                            (*((UChar*)(s.strm.next_in))));         s.bsLive += 8;                                   s.strm.next_in++;                               s.strm.avail_in--;                              s.strm.total_in_lo32++;                         if s.strm.total_in_lo32 == 0 {                     s.strm.total_in_hi32++; }                   };
      // inlining of macro RETURN
      if nGroups < 2 || nGroups > 6 { retVal = BZ_DATA_ERROR; goto save_state_and_return; };;
      // inlining of macro GET_BITS
      fallthrough
case BZ_X_SELECTOR_2: s.state = BZ_X_SELECTOR_2;                         for True {                                       if s.bsLive >= 15 {                              var v UInt32;                                         v = (s.bsBuff >>                                     (s.bsLive-15)) & ((1 << 15)-1);            s.bsLive -= 15;                                 nSelectors = v;                                          break;                                         }
                                                 // inlining of macro RETURN
                                                 if s.strm.avail_in == 0 { retVal = BZ_OK; goto save_state_and_return; };;        s.bsBuff                                            = (s.bsBuff << 8) |                                ((UInt32)                                            (*((UChar*)(s.strm.next_in))));         s.bsLive += 8;                                   s.strm.next_in++;                               s.strm.avail_in--;                              s.strm.total_in_lo32++;                         if s.strm.total_in_lo32 == 0 {                     s.strm.total_in_hi32++; }                   };
      // inlining of macro RETURN
      if nSelectors < 1 { retVal = BZ_DATA_ERROR; goto save_state_and_return; };;
      for i = 0; i < nSelectors; i++ {
         j = 0;
         for True {
            // inlining of macro GET_BIT
            // inlining of macro GET_BITS
            case BZ_X_SELECTOR_3: s.state = BZ_X_SELECTOR_3;                         for True {                                       if s.bsLive >= 1 {                              var v UInt32;                                         v = (s.bsBuff >>                                     (s.bsLive-1)) & ((1 << 1)-1);            s.bsLive -= 1;                                 uc = v;                                          break;                                         }
                                                 // inlining of macro RETURN
                                                 if s.strm.avail_in == 0 { retVal = BZ_OK; goto save_state_and_return; };;        s.bsBuff                                            = (s.bsBuff << 8) |                                ((UInt32)                                            (*((UChar*)(s.strm.next_in))));         s.bsLive += 8;                                   s.strm.next_in++;                               s.strm.avail_in--;                              s.strm.total_in_lo32++;                         if s.strm.total_in_lo32 == 0 {                     s.strm.total_in_hi32++; }                   };
            if uc == 0 { break; }
            j++;
            // inlining of macro RETURN
            if j >= nGroups { retVal = BZ_DATA_ERROR; goto save_state_and_return; };;
         }
         s.selectorMtf[i] = j;
      }

      /*--- Undo the MTF values for the selectors. ---*/
      {
         var pos [BZ_N_GROUPS]UChar;
         var tmp UChar;
         var v UChar;
         for v = 0; v < nGroups; v++ { pos[v] = v; }
   
         for i = 0; i < nSelectors; i++ {
            v = s.selectorMtf[i];
            tmp = pos[v];
            for v > 0 { pos[v] = pos[v-1]; v--; }
            pos[0] = tmp;
            s.selector[i] = tmp;
         }
      }

      /*--- Now the coding tables ---*/
      for t = 0; t < nGroups; t++ {
         // inlining of macro GET_BITS
         case BZ_X_CODING_1: s.state = BZ_X_CODING_1;                         for True {                                       if s.bsLive >= 5 {                              var v UInt32;                                         v = (s.bsBuff >>                                     (s.bsLive-5)) & ((1 << 5)-1);            s.bsLive -= 5;                                 curr = v;                                          break;                                         }
                                                 // inlining of macro RETURN
                                                 if s.strm.avail_in == 0 { retVal = BZ_OK; goto save_state_and_return; };;        s.bsBuff                                            = (s.bsBuff << 8) |                                ((UInt32)                                            (*((UChar*)(s.strm.next_in))));         s.bsLive += 8;                                   s.strm.next_in++;                               s.strm.avail_in--;                              s.strm.total_in_lo32++;                         if s.strm.total_in_lo32 == 0 {                     s.strm.total_in_hi32++; }                   };
         for i = 0; i < alphaSize; i++ {
            for True {
               // inlining of macro RETURN
               if curr < 1 || curr > 20 { retVal = BZ_DATA_ERROR; goto save_state_and_return; };;
               // inlining of macro GET_BIT
               // inlining of macro GET_BITS
               case BZ_X_CODING_2: s.state = BZ_X_CODING_2;                         for True {                                       if s.bsLive >= 1 {                              var v UInt32;                                         v = (s.bsBuff >>                                     (s.bsLive-1)) & ((1 << 1)-1);            s.bsLive -= 1;                                 uc = v;                                          break;                                         }
                                                 // inlining of macro RETURN
                                                 if s.strm.avail_in == 0 { retVal = BZ_OK; goto save_state_and_return; };;        s.bsBuff                                            = (s.bsBuff << 8) |                                ((UInt32)                                            (*((UChar*)(s.strm.next_in))));         s.bsLive += 8;                                   s.strm.next_in++;                               s.strm.avail_in--;                              s.strm.total_in_lo32++;                         if s.strm.total_in_lo32 == 0 {                     s.strm.total_in_hi32++; }                   };
               if uc == 0 { break; }
               // inlining of macro GET_BIT
               // inlining of macro GET_BITS
               case BZ_X_CODING_3: s.state = BZ_X_CODING_3;                         for True {                                       if s.bsLive >= 1 {                              var v UInt32;                                         v = (s.bsBuff >>                                     (s.bsLive-1)) & ((1 << 1)-1);            s.bsLive -= 1;                                 uc = v;                                          break;                                         }
                                                 // inlining of macro RETURN
                                                 if s.strm.avail_in == 0 { retVal = BZ_OK; goto save_state_and_return; };;        s.bsBuff                                            = (s.bsBuff << 8) |                                ((UInt32)                                            (*((UChar*)(s.strm.next_in))));         s.bsLive += 8;                                   s.strm.next_in++;                               s.strm.avail_in--;                              s.strm.total_in_lo32++;                         if s.strm.total_in_lo32 == 0 {                     s.strm.total_in_hi32++; }                   };
               if uc == 0 { curr++; } else { curr--; }
            }
            s.len[t][i] = curr;
         }
      }

      /*--- Create the Huffman decoding tables ---*/
      for t = 0; t < nGroups; t++ {
         minLen = 32;
         maxLen = 0;
         for i = 0; i < alphaSize; i++ {
            if s.len[t][i] > maxLen { maxLen = s.len[t][i]; }
            if s.len[t][i] < minLen { minLen = s.len[t][i]; }
         }
         BZ2_hbCreateDecodeTables ( 
            &(s.limit[t][0]), 
            &(s.base[t][0]), 
            &(s.perm[t][0]), 
            &(s.len[t][0]),
            minLen, maxLen, alphaSize
         );
         s.minLens[t] = minLen;
      }

      /*--- Now the MTF values ---*/

      EOB      = s.nInUse+1;
      nblockMAX = 100000 * s.blockSize100k;
      groupNo  = -1;
      groupPos = 0;

      for i = 0; i <= 255; i++ { s.unzftab[i] = 0; }

      /*-- MTF init --*/
      {
         var ii, jj, kk Int32;
         kk = MTFA_SIZE-1;
         for ii = 256 / MTFL_SIZE - 1; ii >= 0; ii-- {
            for jj = MTFL_SIZE-1; jj >= 0; jj-- {
               s.mtfa[kk] = (UChar)(ii * MTFL_SIZE + jj);
               kk--;
            }
            s.mtfbase[ii] = kk + 1;
         }
      }
      /*-- end MTF init --*/

      nblock = 0;
      // inlining of macro GET_MTF_VAL
      // inlining of macro RETURN
      {                                                    if groupPos == 0 {                                 groupNo++;                                        if groupNo >= nSelectors                           { retVal = BZ_DATA_ERROR; goto save_state_and_return; };;                         groupPos = BZ_G_SIZE;                             gSel = s.selector[groupNo];                      gMinlen = s.minLens[gSel];                       gLimit = &(s.limit[gSel][0]);                    gPerm = &(s.perm[gSel][0]);                      gBase = &(s.base[gSel][0]);                   }
                                                 // inlining of macro GET_BITS
                                                 groupPos--;                                       zn = gMinlen;                                     case BZ_X_MTF_1: s.state = BZ_X_MTF_1;                         for True {                                       if s.bsLive >= zn {                              var v UInt32;                                         v = (s.bsBuff >>                                     (s.bsLive-zn)) & ((1 << zn)-1);            s.bsLive -= zn;                                 zvec = v;                                          break;                                         }
                                                 // inlining of macro RETURN
                                                 if s.strm.avail_in == 0 { retVal = BZ_OK; goto save_state_and_return; };;        s.bsBuff                                            = (s.bsBuff << 8) |                                ((UInt32)                                            (*((UChar*)(s.strm.next_in))));         s.bsLive += 8;                                   s.strm.next_in++;                               s.strm.avail_in--;                              s.strm.total_in_lo32++;                         if s.strm.total_in_lo32 == 0 {                     s.strm.total_in_hi32++; }                   };                       for {                                          if zn > 20 /* the longest code */                  { retVal = BZ_DATA_ERROR; goto save_state_and_return; };;                         if zvec <= gLimit[zn] { break; }
                    // inlining of macro GET_BIT
                    // inlining of macro GET_BITS
                    zn++;                                             case BZ_X_MTF_2: s.state = BZ_X_MTF_2;                         for True {                                       if s.bsLive >= 1 {                              var v UInt32;                                         v = (s.bsBuff >>                                     (s.bsLive-1)) & ((1 << 1)-1);            s.bsLive -= 1;                                 zj = v;                                          break;                                         }
                                                 // inlining of macro RETURN
                                                 if s.strm.avail_in == 0 { retVal = BZ_OK; goto save_state_and_return; };;        s.bsBuff                                            = (s.bsBuff << 8) |                                ((UInt32)                                            (*((UChar*)(s.strm.next_in))));         s.bsLive += 8;                                   s.strm.next_in++;                               s.strm.avail_in--;                              s.strm.total_in_lo32++;                         if s.strm.total_in_lo32 == 0 {                     s.strm.total_in_hi32++; }                   };                              zvec = (zvec << 1) | zj;                       };                                                if zvec - gBase[zn] < 0                              || zvec - gBase[zn] >= BZ_MAX_ALPHA_SIZE        { retVal = BZ_DATA_ERROR; goto save_state_and_return; };;                         nextSym = gPerm[zvec - gBase[zn]];                };

      for True {

         if nextSym == EOB { break; }

         if nextSym == BZ_RUNA || nextSym == BZ_RUNB {

            es = -1;
            N = 1;
            for {
               if nextSym == BZ_RUNA { es = es + (0+1) * N; } else
               if nextSym == BZ_RUNB { es = es + (1+1) * N; }
               N = N * 2;
               // inlining of macro GET_MTF_VAL
               // inlining of macro RETURN
               {                                                    if groupPos == 0 {                                 groupNo++;                                        if groupNo >= nSelectors                           { retVal = BZ_DATA_ERROR; goto save_state_and_return; };;                         groupPos = BZ_G_SIZE;                             gSel = s.selector[groupNo];                      gMinlen = s.minLens[gSel];                       gLimit = &(s.limit[gSel][0]);                    gPerm = &(s.perm[gSel][0]);                      gBase = &(s.base[gSel][0]);                   }
                                                 // inlining of macro GET_BITS
                                                 groupPos--;                                       zn = gMinlen;                                     case BZ_X_MTF_3: s.state = BZ_X_MTF_3;                         for True {                                       if s.bsLive >= zn {                              var v UInt32;                                         v = (s.bsBuff >>                                     (s.bsLive-zn)) & ((1 << zn)-1);            s.bsLive -= zn;                                 zvec = v;                                          break;                                         }
                                                 // inlining of macro RETURN
                                                 if s.strm.avail_in == 0 { retVal = BZ_OK; goto save_state_and_return; };;        s.bsBuff                                            = (s.bsBuff << 8) |                                ((UInt32)                                            (*((UChar*)(s.strm.next_in))));         s.bsLive += 8;                                   s.strm.next_in++;                               s.strm.avail_in--;                              s.strm.total_in_lo32++;                         if s.strm.total_in_lo32 == 0 {                     s.strm.total_in_hi32++; }                   };                       for {                                          if zn > 20 /* the longest code */                  { retVal = BZ_DATA_ERROR; goto save_state_and_return; };;                         if zvec <= gLimit[zn] { break; }
                    // inlining of macro GET_BIT
                    // inlining of macro GET_BITS
                    zn++;                                             case BZ_X_MTF_4: s.state = BZ_X_MTF_4;                         for True {                                       if s.bsLive >= 1 {                              var v UInt32;                                         v = (s.bsBuff >>                                     (s.bsLive-1)) & ((1 << 1)-1);            s.bsLive -= 1;                                 zj = v;                                          break;                                         }
                                                 // inlining of macro RETURN
                                                 if s.strm.avail_in == 0 { retVal = BZ_OK; goto save_state_and_return; };;        s.bsBuff                                            = (s.bsBuff << 8) |                                ((UInt32)                                            (*((UChar*)(s.strm.next_in))));         s.bsLive += 8;                                   s.strm.next_in++;                               s.strm.avail_in--;                              s.strm.total_in_lo32++;                         if s.strm.total_in_lo32 == 0 {                     s.strm.total_in_hi32++; }                   };                              zvec = (zvec << 1) | zj;                       };                                                if zvec - gBase[zn] < 0                              || zvec - gBase[zn] >= BZ_MAX_ALPHA_SIZE        { retVal = BZ_DATA_ERROR; goto save_state_and_return; };;                         nextSym = gPerm[zvec - gBase[zn]];                };
            
if !(nextSym == BZ_RUNA || nextSym == BZ_RUNB) { break }
}
               

            es++;
            uc = s.seqToUnseq[ s.mtfa[s.mtfbase[0]] ];
            s.unzftab[uc] += es;

            if s.smallDecompress {
               for es > 0 {
                  // inlining of macro RETURN
                  if nblock >= nblockMAX { retVal = BZ_DATA_ERROR; goto save_state_and_return; };;
                  s.ll16[nblock] = (UInt16)uc;
                  nblock++;
                  es--;
               }
            } else {
               for es > 0 {
                  // inlining of macro RETURN
                  if nblock >= nblockMAX { retVal = BZ_DATA_ERROR; goto save_state_and_return; };;
                  s.tt[nblock] = (UInt32)uc;
                  nblock++;
                  es--;
               } };

            continue;

         } else {

            // inlining of macro RETURN
            if nblock >= nblockMAX { retVal = BZ_DATA_ERROR; goto save_state_and_return; };;

            /*-- uc = MTF ( nextSym-1 ) --*/
            {
               var ii, jj, kk, pp, lno, off Int32;
               var nn UInt32;
               nn = (UInt32)(nextSym - 1);

               if nn < MTFL_SIZE {
                  /* avoid general-case expense */
                  pp = s.mtfbase[0];
                  uc = s.mtfa[pp+nn];
                  for nn > 3 {
                     var z Int32  = pp+nn;
                     s.mtfa[(z)  ] = s.mtfa[(z)-1];
                     s.mtfa[(z)-1] = s.mtfa[(z)-2];
                     s.mtfa[(z)-2] = s.mtfa[(z)-3];
                     s.mtfa[(z)-3] = s.mtfa[(z)-4];
                     nn -= 4;
                  }
                  for nn > 0 { 
                     s.mtfa[(pp+nn)] = s.mtfa[(pp+nn)-1]; nn--; 
                  };
                  s.mtfa[pp] = uc;
               } else { 
                  /* general case */
                  lno = nn / MTFL_SIZE;
                  off = nn % MTFL_SIZE;
                  pp = s.mtfbase[lno] + off;
                  uc = s.mtfa[pp];
                  for pp > s.mtfbase[lno] { 
                     s.mtfa[pp] = s.mtfa[pp-1]; pp--; 
                  };
                  s.mtfbase[lno]++;
                  for lno > 0 {
                     s.mtfbase[lno]--;
                     s.mtfa[s.mtfbase[lno]] 
                        = s.mtfa[s.mtfbase[lno-1] + MTFL_SIZE - 1];
                     lno--;
                  }
                  s.mtfbase[0]--;
                  s.mtfa[s.mtfbase[0]] = uc;
                  if s.mtfbase[0] == 0 {
                     kk = MTFA_SIZE-1;
                     for ii = 256 / MTFL_SIZE-1; ii >= 0; ii-- {
                        for jj = MTFL_SIZE-1; jj >= 0; jj-- {
                           s.mtfa[kk] = s.mtfa[s.mtfbase[ii] + jj];
                           kk--;
                        }
                        s.mtfbase[ii] = kk + 1;
                     }
                  }
               }
            }
            /*-- end uc = MTF ( nextSym-1 ) --*/

            s.unzftab[s.seqToUnseq[uc]]++;
            if s.smallDecompress {
               s.ll16[nblock] = (UInt16)(s.seqToUnseq[uc]); } else {
               s.tt[nblock]   = (UInt32)(s.seqToUnseq[uc]); }
            nblock++;

            // inlining of macro GET_MTF_VAL
            // inlining of macro RETURN
            {                                                    if groupPos == 0 {                                 groupNo++;                                        if groupNo >= nSelectors                           { retVal = BZ_DATA_ERROR; goto save_state_and_return; };;                         groupPos = BZ_G_SIZE;                             gSel = s.selector[groupNo];                      gMinlen = s.minLens[gSel];                       gLimit = &(s.limit[gSel][0]);                    gPerm = &(s.perm[gSel][0]);                      gBase = &(s.base[gSel][0]);                   }
                                                 // inlining of macro GET_BITS
                                                 groupPos--;                                       zn = gMinlen;                                     case BZ_X_MTF_5: s.state = BZ_X_MTF_5;                         for True {                                       if s.bsLive >= zn {                              var v UInt32;                                         v = (s.bsBuff >>                                     (s.bsLive-zn)) & ((1 << zn)-1);            s.bsLive -= zn;                                 zvec = v;                                          break;                                         }
                                                 // inlining of macro RETURN
                                                 if s.strm.avail_in == 0 { retVal = BZ_OK; goto save_state_and_return; };;        s.bsBuff                                            = (s.bsBuff << 8) |                                ((UInt32)                                            (*((UChar*)(s.strm.next_in))));         s.bsLive += 8;                                   s.strm.next_in++;                               s.strm.avail_in--;                              s.strm.total_in_lo32++;                         if s.strm.total_in_lo32 == 0 {                     s.strm.total_in_hi32++; }                   };                       for {                                          if zn > 20 /* the longest code */                  { retVal = BZ_DATA_ERROR; goto save_state_and_return; };;                         if zvec <= gLimit[zn] { break; }
                    // inlining of macro GET_BIT
                    // inlining of macro GET_BITS
                    zn++;                                             case BZ_X_MTF_6: s.state = BZ_X_MTF_6;                         for True {                                       if s.bsLive >= 1 {                              var v UInt32;                                         v = (s.bsBuff >>                                     (s.bsLive-1)) & ((1 << 1)-1);            s.bsLive -= 1;                                 zj = v;                                          break;                                         }
                                                 // inlining of macro RETURN
                                                 if s.strm.avail_in == 0 { retVal = BZ_OK; goto save_state_and_return; };;        s.bsBuff                                            = (s.bsBuff << 8) |                                ((UInt32)                                            (*((UChar*)(s.strm.next_in))));         s.bsLive += 8;                                   s.strm.next_in++;                               s.strm.avail_in--;                              s.strm.total_in_lo32++;                         if s.strm.total_in_lo32 == 0 {                     s.strm.total_in_hi32++; }                   };                              zvec = (zvec << 1) | zj;                       };                                                if zvec - gBase[zn] < 0                              || zvec - gBase[zn] >= BZ_MAX_ALPHA_SIZE        { retVal = BZ_DATA_ERROR; goto save_state_and_return; };;                         nextSym = gPerm[zvec - gBase[zn]];                };
            continue;
         }
      }

      /* Now we know what nblock is, we can do a better sanity
         check on s->origPtr.
      */
      if s.origPtr < 0 || s.origPtr >= nblock {
         // inlining of macro RETURN
          retVal = BZ_DATA_ERROR; goto save_state_and_return; };;

      /*-- Set up cftab to facilitate generation of T^(-1) --*/
      s.cftab[0] = 0;
      for i = 1; i <= 256; i++ { s.cftab[i] = s.unzftab[i-1]; }
      for i = 1; i <= 256; i++ { s.cftab[i] += s.cftab[i-1]; }
      for i = 0; i <= 256; i++ {
         if s.cftab[i] < 0 || s.cftab[i] > nblock {
            /* s->cftab[i] can legitimately be == nblock */
            // inlining of macro RETURN
            { retVal = BZ_DATA_ERROR; goto save_state_and_return; };;
         }
      }

      s.state_out_len = 0;
      s.state_out_ch  = 0;
      BZ_INITIALISE_CRC ( s.calculatedBlockCRC );
      s.state = BZ_X_OUTPUT;
      if s.verbosity >= 2 { VPrintf0 ( "rt+rld" ); }

      if s.smallDecompress {

         /*-- Make a copy of cftab, used in generation of T --*/
         for i = 0; i <= 256; i++ { s.cftabCopy[i] = s.cftab[i]; }

         /*-- compute the T vector --*/
         for i = 0; i < nblock; i++ {
            uc = (UChar)(s.ll16[i]);
            SET_LL(i, s.cftabCopy[uc]);
            s.cftabCopy[uc]++;
         }

         /*-- Compute T^(-1) by pointer reversal on T --*/
         i = s.origPtr;
         j = GET_LL(i);
         for {
            var tmp Int32  = GET_LL(j);
            SET_LL(j, i);
            i = j;
            j = tmp;
         
if !(i != s.origPtr) { break }
}
            

         s.tPos = s.origPtr;
         s.nblock_used = 0;
         if s.blockRandomised {
            BZ_RAND_INIT_MASK;
            BZ_GET_SMALL(s.k0); s.nblock_used++;
            BZ_RAND_UPD_MASK; s.k0 ^= BZ_RAND_MASK; 
         } else {
            BZ_GET_SMALL(s.k0); s.nblock_used++;
         }

      } else {

         /*-- compute the T^(-1) vector --*/
         for i = 0; i < nblock; i++ {
            uc = (UChar)(s.tt[i] & 0xff);
            s.tt[s.cftab[uc]] |= (i << 8);
            s.cftab[uc]++;
         }

         s.tPos = s.tt[s.origPtr] >> 8;
         s.nblock_used = 0;
         if s.blockRandomised {
            BZ_RAND_INIT_MASK;
            BZ_GET_FAST(s.k0); s.nblock_used++;
            BZ_RAND_UPD_MASK; s.k0 ^= BZ_RAND_MASK; 
         } else {
            BZ_GET_FAST(s.k0); s.nblock_used++;
         }

      }

      // inlining of macro RETURN
      { retVal = BZ_OK; goto save_state_and_return; };;



    endhdr_2:

      // inlining of macro GET_UCHAR
      // inlining of macro GET_BITS
      fallthrough
case BZ_X_ENDHDR_2: s.state = BZ_X_ENDHDR_2;                         for True {                                       if s.bsLive >= 8 {                              var v UInt32;                                         v = (s.bsBuff >>                                     (s.bsLive-8)) & ((1 << 8)-1);            s.bsLive -= 8;                                 uc = v;                                          break;                                         }
                                                 // inlining of macro RETURN
                                                 if s.strm.avail_in == 0 { retVal = BZ_OK; goto save_state_and_return; };;        s.bsBuff                                            = (s.bsBuff << 8) |                                ((UInt32)                                            (*((UChar*)(s.strm.next_in))));         s.bsLive += 8;                                   s.strm.next_in++;                               s.strm.avail_in--;                              s.strm.total_in_lo32++;                         if s.strm.total_in_lo32 == 0 {                     s.strm.total_in_hi32++; }                   };
      // inlining of macro RETURN
      if uc != 0x72 { retVal = BZ_DATA_ERROR; goto save_state_and_return; };;
      // inlining of macro GET_UCHAR
      // inlining of macro GET_BITS
      fallthrough
case BZ_X_ENDHDR_3: s.state = BZ_X_ENDHDR_3;                         for True {                                       if s.bsLive >= 8 {                              var v UInt32;                                         v = (s.bsBuff >>                                     (s.bsLive-8)) & ((1 << 8)-1);            s.bsLive -= 8;                                 uc = v;                                          break;                                         }
                                                 // inlining of macro RETURN
                                                 if s.strm.avail_in == 0 { retVal = BZ_OK; goto save_state_and_return; };;        s.bsBuff                                            = (s.bsBuff << 8) |                                ((UInt32)                                            (*((UChar*)(s.strm.next_in))));         s.bsLive += 8;                                   s.strm.next_in++;                               s.strm.avail_in--;                              s.strm.total_in_lo32++;                         if s.strm.total_in_lo32 == 0 {                     s.strm.total_in_hi32++; }                   };
      // inlining of macro RETURN
      if uc != 0x45 { retVal = BZ_DATA_ERROR; goto save_state_and_return; };;
      // inlining of macro GET_UCHAR
      // inlining of macro GET_BITS
      fallthrough
case BZ_X_ENDHDR_4: s.state = BZ_X_ENDHDR_4;                         for True {                                       if s.bsLive >= 8 {                              var v UInt32;                                         v = (s.bsBuff >>                                     (s.bsLive-8)) & ((1 << 8)-1);            s.bsLive -= 8;                                 uc = v;                                          break;                                         }
                                                 // inlining of macro RETURN
                                                 if s.strm.avail_in == 0 { retVal = BZ_OK; goto save_state_and_return; };;        s.bsBuff                                            = (s.bsBuff << 8) |                                ((UInt32)                                            (*((UChar*)(s.strm.next_in))));         s.bsLive += 8;                                   s.strm.next_in++;                               s.strm.avail_in--;                              s.strm.total_in_lo32++;                         if s.strm.total_in_lo32 == 0 {                     s.strm.total_in_hi32++; }                   };
      // inlining of macro RETURN
      if uc != 0x38 { retVal = BZ_DATA_ERROR; goto save_state_and_return; };;
      // inlining of macro GET_UCHAR
      // inlining of macro GET_BITS
      fallthrough
case BZ_X_ENDHDR_5: s.state = BZ_X_ENDHDR_5;                         for True {                                       if s.bsLive >= 8 {                              var v UInt32;                                         v = (s.bsBuff >>                                     (s.bsLive-8)) & ((1 << 8)-1);            s.bsLive -= 8;                                 uc = v;                                          break;                                         }
                                                 // inlining of macro RETURN
                                                 if s.strm.avail_in == 0 { retVal = BZ_OK; goto save_state_and_return; };;        s.bsBuff                                            = (s.bsBuff << 8) |                                ((UInt32)                                            (*((UChar*)(s.strm.next_in))));         s.bsLive += 8;                                   s.strm.next_in++;                               s.strm.avail_in--;                              s.strm.total_in_lo32++;                         if s.strm.total_in_lo32 == 0 {                     s.strm.total_in_hi32++; }                   };
      // inlining of macro RETURN
      if uc != 0x50 { retVal = BZ_DATA_ERROR; goto save_state_and_return; };;
      // inlining of macro GET_UCHAR
      // inlining of macro GET_BITS
      fallthrough
case BZ_X_ENDHDR_6: s.state = BZ_X_ENDHDR_6;                         for True {                                       if s.bsLive >= 8 {                              var v UInt32;                                         v = (s.bsBuff >>                                     (s.bsLive-8)) & ((1 << 8)-1);            s.bsLive -= 8;                                 uc = v;                                          break;                                         }
                                                 // inlining of macro RETURN
                                                 if s.strm.avail_in == 0 { retVal = BZ_OK; goto save_state_and_return; };;        s.bsBuff                                            = (s.bsBuff << 8) |                                ((UInt32)                                            (*((UChar*)(s.strm.next_in))));         s.bsLive += 8;                                   s.strm.next_in++;                               s.strm.avail_in--;                              s.strm.total_in_lo32++;                         if s.strm.total_in_lo32 == 0 {                     s.strm.total_in_hi32++; }                   };
      // inlining of macro RETURN
      if uc != 0x90 { retVal = BZ_DATA_ERROR; goto save_state_and_return; };;

      s.storedCombinedCRC = 0;
      // inlining of macro GET_UCHAR
      // inlining of macro GET_BITS
      fallthrough
case BZ_X_CCRC_1: s.state = BZ_X_CCRC_1;                         for True {                                       if s.bsLive >= 8 {                              var v UInt32;                                         v = (s.bsBuff >>                                     (s.bsLive-8)) & ((1 << 8)-1);            s.bsLive -= 8;                                 uc = v;                                          break;                                         }
                                                 // inlining of macro RETURN
                                                 if s.strm.avail_in == 0 { retVal = BZ_OK; goto save_state_and_return; };;        s.bsBuff                                            = (s.bsBuff << 8) |                                ((UInt32)                                            (*((UChar*)(s.strm.next_in))));         s.bsLive += 8;                                   s.strm.next_in++;                               s.strm.avail_in--;                              s.strm.total_in_lo32++;                         if s.strm.total_in_lo32 == 0 {                     s.strm.total_in_hi32++; }                   };
      s.storedCombinedCRC = (s.storedCombinedCRC << 8) | ((UInt32)uc);
      // inlining of macro GET_UCHAR
      // inlining of macro GET_BITS
      fallthrough
case BZ_X_CCRC_2: s.state = BZ_X_CCRC_2;                         for True {                                       if s.bsLive >= 8 {                              var v UInt32;                                         v = (s.bsBuff >>                                     (s.bsLive-8)) & ((1 << 8)-1);            s.bsLive -= 8;                                 uc = v;                                          break;                                         }
                                                 // inlining of macro RETURN
                                                 if s.strm.avail_in == 0 { retVal = BZ_OK; goto save_state_and_return; };;        s.bsBuff                                            = (s.bsBuff << 8) |                                ((UInt32)                                            (*((UChar*)(s.strm.next_in))));         s.bsLive += 8;                                   s.strm.next_in++;                               s.strm.avail_in--;                              s.strm.total_in_lo32++;                         if s.strm.total_in_lo32 == 0 {                     s.strm.total_in_hi32++; }                   };
      s.storedCombinedCRC = (s.storedCombinedCRC << 8) | ((UInt32)uc);
      // inlining of macro GET_UCHAR
      // inlining of macro GET_BITS
      fallthrough
case BZ_X_CCRC_3: s.state = BZ_X_CCRC_3;                         for True {                                       if s.bsLive >= 8 {                              var v UInt32;                                         v = (s.bsBuff >>                                     (s.bsLive-8)) & ((1 << 8)-1);            s.bsLive -= 8;                                 uc = v;                                          break;                                         }
                                                 // inlining of macro RETURN
                                                 if s.strm.avail_in == 0 { retVal = BZ_OK; goto save_state_and_return; };;        s.bsBuff                                            = (s.bsBuff << 8) |                                ((UInt32)                                            (*((UChar*)(s.strm.next_in))));         s.bsLive += 8;                                   s.strm.next_in++;                               s.strm.avail_in--;                              s.strm.total_in_lo32++;                         if s.strm.total_in_lo32 == 0 {                     s.strm.total_in_hi32++; }                   };
      s.storedCombinedCRC = (s.storedCombinedCRC << 8) | ((UInt32)uc);
      // inlining of macro GET_UCHAR
      // inlining of macro GET_BITS
      fallthrough
case BZ_X_CCRC_4: s.state = BZ_X_CCRC_4;                         for True {                                       if s.bsLive >= 8 {                              var v UInt32;                                         v = (s.bsBuff >>                                     (s.bsLive-8)) & ((1 << 8)-1);            s.bsLive -= 8;                                 uc = v;                                          break;                                         }
                                                 // inlining of macro RETURN
                                                 if s.strm.avail_in == 0 { retVal = BZ_OK; goto save_state_and_return; };;        s.bsBuff                                            = (s.bsBuff << 8) |                                ((UInt32)                                            (*((UChar*)(s.strm.next_in))));         s.bsLive += 8;                                   s.strm.next_in++;                               s.strm.avail_in--;                              s.strm.total_in_lo32++;                         if s.strm.total_in_lo32 == 0 {                     s.strm.total_in_hi32++; }                   };
      s.storedCombinedCRC = (s.storedCombinedCRC << 8) | ((UInt32)uc);

      s.state = BZ_X_IDLE;
      // inlining of macro RETURN
      { retVal = BZ_STREAM_END; goto save_state_and_return; };;

      fallthrough
default: AssertH ( False, 4001 );
   }

   AssertH ( False, 4002 );

   save_state_and_return:

   s.save_i           = i;
   s.save_j           = j;
   s.save_t           = t;
   s.save_alphaSize   = alphaSize;
   s.save_nGroups     = nGroups;
   s.save_nSelectors  = nSelectors;
   s.save_EOB         = EOB;
   s.save_groupNo     = groupNo;
   s.save_groupPos    = groupPos;
   s.save_nextSym     = nextSym;
   s.save_nblockMAX   = nblockMAX;
   s.save_nblock      = nblock;
   s.save_es          = es;
   s.save_N           = N;
   s.save_curr        = curr;
   s.save_zt          = zt;
   s.save_zn          = zn;
   s.save_zvec        = zvec;
   s.save_zj          = zj;
   s.save_gSel        = gSel;
   s.save_gMinlen     = gMinlen;
   s.save_gLimit      = gLimit;
   s.save_gBase       = gBase;
   s.save_gPerm       = gPerm;

   return retVal;   
}


/*-------------------------------------------------------------*/
/*--- end                                      decompress.c ---*/
/*-------------------------------------------------------------*/
//ABC_NAMESPACE_IMPL_END

