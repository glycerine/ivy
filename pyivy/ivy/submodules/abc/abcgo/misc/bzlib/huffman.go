package bzlib


/*-------------------------------------------------------------*/
/*--- Huffman coding low-level stuff                        ---*/
/*---                                             huffman.c ---*/
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
// Original macro:
// #define WEIGHTOF(zz0)  ((zz0) & 0xffffff00)

// Original macro:
// #define DEPTHOF(zz1)   ((zz1) & 0x000000ff)

// Original macro:
// #define MYMAX(zz2,zz3) ((zz2) > (zz3) ? (zz2) : (zz3))


// Original macro:
// #define ADDWEIGHTS(zw1,zw2)                           \
//    (WEIGHTOF(zw1)+WEIGHTOF(zw2)) |                    \
//    (1 + MYMAX(DEPTHOF(zw1),DEPTHOF(zw2)))


// Original macro:
// #define UPHEAP(z)                                     \
// {                                                     \
//    Int32 zz, tmp;                                     \
//    zz = z; tmp = heap[zz];                            \
//    while (weight[tmp] < weight[heap[zz >> 1]]) {      \
//       heap[zz] = heap[zz >> 1];                       \
//       zz >>= 1;                                       \
//    }                                                  \
//    heap[zz] = tmp;                                    \
// }


// Original macro:
// #define DOWNHEAP(z)                                   \
// {                                                     \
//    Int32 zz, yy, tmp;                                 \
//    zz = z; tmp = heap[zz];                            \
//    while (True) {                                     \
//       yy = zz << 1;                                   \
//       if (yy > nHeap) break;                          \
//       if (yy < nHeap &&                               \
//           weight[heap[yy+1]] < weight[heap[yy]])      \
//          yy++;                                        \
//       if (weight[tmp] < weight[heap[yy]]) break;      \
//       heap[zz] = heap[yy];                            \
//       zz = yy;                                        \
//    }                                                  \
//    heap[zz] = tmp;                                    \
// }



/*---------------------------------------------------*/
func BZ2_hbMakeCodeLengths(len *UChar, freq *Int32, alphaSize Int32, maxLen Int32) {
   /*--
      Nodes and heap entries run from 1.  Entry 0
      for both the heap and nodes is a sentinel.
   --*/
   var nNodes, nHeap, n1, n2, i, j, k Int32;
   var tooLong Bool;

   var heap [BZ_MAX_ALPHA_SIZE + 2]Int32;
   var weight [BZ_MAX_ALPHA_SIZE * 2]Int32;
   var parent [BZ_MAX_ALPHA_SIZE * 2]Int32; 

   for i = 0; i < alphaSize; i++ {
      weight[i+1] = (freq[i] == 0 ? 1 : freq[i]) << 8;
   }

   for True {

      nNodes = alphaSize;
      nHeap = 0;

      heap[0] = 0;
      weight[0] = 0;
      parent[0] = -2;

      for i = 1; i <= alphaSize; i++ {
         parent[i] = -1;
         nHeap++;
         heap[nHeap] = i;
         // inlining of macro UPHEAP
         {                                                        var zz, tmp Int32;                                        zz = nHeap; tmp = heap[zz];                               for weight[tmp] < weight[heap[zz >> 1]] {            heap[zz] = heap[zz >> 1];                             zz >>= 1;                                          }
                                                     heap[zz] = tmp;                                    };
      }

      AssertH( nHeap < (BZ_MAX_ALPHA_SIZE+2), 2001 );
   
      for nHeap > 1 {
         // inlining of macro DOWNHEAP
         n1 = heap[1]; heap[1] = heap[nHeap]; nHeap--; {                                                        var zz, yy, tmp Int32;                                    zz = 1; tmp = heap[zz];                               for True {                                           yy = zz << 1;                                         if yy > nHeap { break; }
                                if yy < nHeap &&                                         weight[heap[yy+1]] < weight[heap[yy]] {               yy++; }
                                              if weight[tmp] < weight[heap[yy]] { break; }
            heap[zz] = heap[yy];                                  zz = yy;                                           }
                                                     heap[zz] = tmp;                                    };
         // inlining of macro DOWNHEAP
         n2 = heap[1]; heap[1] = heap[nHeap]; nHeap--; {                                                        var zz, yy, tmp Int32;                                    zz = 1; tmp = heap[zz];                               for True {                                           yy = zz << 1;                                         if yy > nHeap { break; }
                                if yy < nHeap &&                                         weight[heap[yy+1]] < weight[heap[yy]] {               yy++; }
                                              if weight[tmp] < weight[heap[yy]] { break; }
            heap[zz] = heap[yy];                                  zz = yy;                                           }
                                                     heap[zz] = tmp;                                    };
         nNodes++;
         parent[n1] = parent[n2] = nNodes;
         // inlining of macro ADDWEIGHTS
         // inlining of macro WEIGHTOF
         // inlining of macro MYMAX
         // inlining of macro DEPTHOF
         weight[nNodes] = (((weight[n1]) & 0xffffff00)+((weight[n2]) & 0xffffff00)) |                       (1 + ((((weight[n1]) & 0x000000ff)) > (((weight[n2]) & 0x000000ff)) ? (((weight[n1]) & 0x000000ff)) : (((weight[n2]) & 0x000000ff))));
         parent[nNodes] = -1;
         nHeap++;
         heap[nHeap] = nNodes;
         // inlining of macro UPHEAP
         {                                                        var zz, tmp Int32;                                        zz = nHeap; tmp = heap[zz];                               for weight[tmp] < weight[heap[zz >> 1]] {            heap[zz] = heap[zz >> 1];                             zz >>= 1;                                          }
                                                     heap[zz] = tmp;                                    };
      }

      AssertH( nNodes < (BZ_MAX_ALPHA_SIZE * 2), 2002 );

      tooLong = False;
      for i = 1; i <= alphaSize; i++ {
         j = 0;
         k = i;
         for parent[k] >= 0 { k = parent[k]; j++; }
         len[i-1] = j;
         if j > maxLen { tooLong = True; }
      }
      
      if ! tooLong { break; }

      /* 17 Oct 04: keep-going condition for the following loop used
         to be 'i < alphaSize', which missed the last element,
         theoretically leading to the possibility of the compressor
         looping.  However, this count-scaling step is only needed if
         one of the generated Huffman code words is longer than
         maxLen, which up to and including version 1.0.2 was 20 bits,
         which is extremely unlikely.  In version 1.0.3 maxLen was
         changed to 17 bits, which has minimal effect on compression
         ratio, but does mean this scaling step is used from time to
         time, enough to verify that it works.

         This means that bzip2-1.0.3 and later will only produce
         Huffman codes with a maximum length of 17 bits.  However, in
         order to preserve backwards compatibility with bitstreams
         produced by versions pre-1.0.3, the decompressor must still
         handle lengths of up to 20. */

      for i = 1; i <= alphaSize; i++ {
         j = weight[i] >> 8;
         j = 1 + (j / 2);
         weight[i] = j << 8;
      }
   }
}


/*---------------------------------------------------*/
func BZ2_hbAssignCodes(code *Int32, length *UChar, minLen Int32, maxLen Int32, alphaSize Int32) {
   var n, vec, i Int32;

   vec = 0;
   for n = minLen; n <= maxLen; n++ {
      for i = 0; i < alphaSize; i++ {
         if length[i] == n { code[i] = vec; vec++; }
      };
      vec <<= 1;
   }
}


/*---------------------------------------------------*/
func BZ2_hbCreateDecodeTables(limit *Int32, base *Int32, perm *Int32, length *UChar, minLen Int32, maxLen Int32, alphaSize Int32) {
   var pp, i, j, vec Int32;

   pp = 0;
   for i = minLen; i <= maxLen; i++ {
      for j = 0; j < alphaSize; j++ {
         if length[j] == i { perm[pp] = j; pp++; }
      }
   };

   for i = 0; i < BZ_MAX_CODE_LEN; i++ { base[i] = 0; }
   for i = 0; i < alphaSize; i++ { base[length[i]+1]++; }

   for i = 1; i < BZ_MAX_CODE_LEN; i++ { base[i] += base[i-1]; }

   for i = 0; i < BZ_MAX_CODE_LEN; i++ { limit[i] = 0; }
   vec = 0;

   for i = minLen; i <= maxLen; i++ {
      vec += (base[i+1] - base[i]);
      limit[i] = vec-1;
      vec <<= 1;
   }
   for i = minLen + 1; i <= maxLen; i++ {
      base[i] = ((limit[i-1] + 1) << 1) - base[i];
   }
}


/*-------------------------------------------------------------*/
/*--- end                                         huffman.c ---*/
/*-------------------------------------------------------------*/

//ABC_NAMESPACE_IMPL_END
