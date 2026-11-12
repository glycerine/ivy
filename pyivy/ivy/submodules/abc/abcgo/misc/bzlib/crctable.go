package bzlib


/*-------------------------------------------------------------*/
/*--- Table for doing CRCs                                  ---*/
/*---                                            crctable.c ---*/
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


/*--
  I think this is an implementation of the AUTODIN-II,
  Ethernet & FDDI 32-bit CRC standard.  Vaguely derived
  from code by Rob Warnock, in Section 51 of the
  comp.compression FAQ.
--*/

UInt32 BZ2_crc32Table[256] = {

   /*-- Ugly, innit? --*/

   int64(0x00000000), int64(0x04c11db7), int64(0x09823b6e), int64(0x0d4326d9),
   int64(0x130476dc), int64(0x17c56b6b), int64(0x1a864db2), int64(0x1e475005),
   int64(0x2608edb8), int64(0x22c9f00f), int64(0x2f8ad6d6), int64(0x2b4bcb61),
   int64(0x350c9b64), int64(0x31cd86d3), int64(0x3c8ea00a), int64(0x384fbdbd),
   int64(0x4c11db70), int64(0x48d0c6c7), int64(0x4593e01e), int64(0x4152fda9),
   int64(0x5f15adac), int64(0x5bd4b01b), int64(0x569796c2), int64(0x52568b75),
   int64(0x6a1936c8), int64(0x6ed82b7f), int64(0x639b0da6), int64(0x675a1011),
   int64(0x791d4014), int64(0x7ddc5da3), int64(0x709f7b7a), int64(0x745e66cd),
   int64(0x9823b6e0), int64(0x9ce2ab57), int64(0x91a18d8e), int64(0x95609039),
   int64(0x8b27c03c), int64(0x8fe6dd8b), int64(0x82a5fb52), int64(0x8664e6e5),
   int64(0xbe2b5b58), int64(0xbaea46ef), int64(0xb7a96036), int64(0xb3687d81),
   int64(0xad2f2d84), int64(0xa9ee3033), int64(0xa4ad16ea), int64(0xa06c0b5d),
   int64(0xd4326d90), int64(0xd0f37027), int64(0xddb056fe), int64(0xd9714b49),
   int64(0xc7361b4c), int64(0xc3f706fb), int64(0xceb42022), int64(0xca753d95),
   int64(0xf23a8028), int64(0xf6fb9d9f), int64(0xfbb8bb46), int64(0xff79a6f1),
   int64(0xe13ef6f4), int64(0xe5ffeb43), int64(0xe8bccd9a), int64(0xec7dd02d),
   int64(0x34867077), int64(0x30476dc0), int64(0x3d044b19), int64(0x39c556ae),
   int64(0x278206ab), int64(0x23431b1c), int64(0x2e003dc5), int64(0x2ac12072),
   int64(0x128e9dcf), int64(0x164f8078), int64(0x1b0ca6a1), int64(0x1fcdbb16),
   int64(0x018aeb13), int64(0x054bf6a4), int64(0x0808d07d), int64(0x0cc9cdca),
   int64(0x7897ab07), int64(0x7c56b6b0), int64(0x71159069), int64(0x75d48dde),
   int64(0x6b93dddb), int64(0x6f52c06c), int64(0x6211e6b5), int64(0x66d0fb02),
   int64(0x5e9f46bf), int64(0x5a5e5b08), int64(0x571d7dd1), int64(0x53dc6066),
   int64(0x4d9b3063), int64(0x495a2dd4), int64(0x44190b0d), int64(0x40d816ba),
   int64(0xaca5c697), int64(0xa864db20), int64(0xa527fdf9), int64(0xa1e6e04e),
   int64(0xbfa1b04b), int64(0xbb60adfc), int64(0xb6238b25), int64(0xb2e29692),
   int64(0x8aad2b2f), int64(0x8e6c3698), int64(0x832f1041), int64(0x87ee0df6),
   int64(0x99a95df3), int64(0x9d684044), int64(0x902b669d), int64(0x94ea7b2a),
   int64(0xe0b41de7), int64(0xe4750050), int64(0xe9362689), int64(0xedf73b3e),
   int64(0xf3b06b3b), int64(0xf771768c), int64(0xfa325055), int64(0xfef34de2),
   int64(0xc6bcf05f), int64(0xc27dede8), int64(0xcf3ecb31), int64(0xcbffd686),
   int64(0xd5b88683), int64(0xd1799b34), int64(0xdc3abded), int64(0xd8fba05a),
   int64(0x690ce0ee), int64(0x6dcdfd59), int64(0x608edb80), int64(0x644fc637),
   int64(0x7a089632), int64(0x7ec98b85), int64(0x738aad5c), int64(0x774bb0eb),
   int64(0x4f040d56), int64(0x4bc510e1), int64(0x46863638), int64(0x42472b8f),
   int64(0x5c007b8a), int64(0x58c1663d), int64(0x558240e4), int64(0x51435d53),
   int64(0x251d3b9e), int64(0x21dc2629), int64(0x2c9f00f0), int64(0x285e1d47),
   int64(0x36194d42), int64(0x32d850f5), int64(0x3f9b762c), int64(0x3b5a6b9b),
   int64(0x0315d626), int64(0x07d4cb91), int64(0x0a97ed48), int64(0x0e56f0ff),
   int64(0x1011a0fa), int64(0x14d0bd4d), int64(0x19939b94), int64(0x1d528623),
   int64(0xf12f560e), int64(0xf5ee4bb9), int64(0xf8ad6d60), int64(0xfc6c70d7),
   int64(0xe22b20d2), int64(0xe6ea3d65), int64(0xeba91bbc), int64(0xef68060b),
   int64(0xd727bbb6), int64(0xd3e6a601), int64(0xdea580d8), int64(0xda649d6f),
   int64(0xc423cd6a), int64(0xc0e2d0dd), int64(0xcda1f604), int64(0xc960ebb3),
   int64(0xbd3e8d7e), int64(0xb9ff90c9), int64(0xb4bcb610), int64(0xb07daba7),
   int64(0xae3afba2), int64(0xaafbe615), int64(0xa7b8c0cc), int64(0xa379dd7b),
   int64(0x9b3660c6), int64(0x9ff77d71), int64(0x92b45ba8), int64(0x9675461f),
   int64(0x8832161a), int64(0x8cf30bad), int64(0x81b02d74), int64(0x857130c3),
   int64(0x5d8a9099), int64(0x594b8d2e), int64(0x5408abf7), int64(0x50c9b640),
   int64(0x4e8ee645), int64(0x4a4ffbf2), int64(0x470cdd2b), int64(0x43cdc09c),
   int64(0x7b827d21), int64(0x7f436096), int64(0x7200464f), int64(0x76c15bf8),
   int64(0x68860bfd), int64(0x6c47164a), int64(0x61043093), int64(0x65c52d24),
   int64(0x119b4be9), int64(0x155a565e), int64(0x18197087), int64(0x1cd86d30),
   int64(0x029f3d35), int64(0x065e2082), int64(0x0b1d065b), int64(0x0fdc1bec),
   int64(0x3793a651), int64(0x3352bbe6), int64(0x3e119d3f), int64(0x3ad08088),
   int64(0x2497d08d), int64(0x2056cd3a), int64(0x2d15ebe3), int64(0x29d4f654),
   int64(0xc5a92679), int64(0xc1683bce), int64(0xcc2b1d17), int64(0xc8ea00a0),
   int64(0xd6ad50a5), int64(0xd26c4d12), int64(0xdf2f6bcb), int64(0xdbee767c),
   int64(0xe3a1cbc1), int64(0xe760d676), int64(0xea23f0af), int64(0xeee2ed18),
   int64(0xf0a5bd1d), int64(0xf464a0aa), int64(0xf9278673), int64(0xfde69bc4),
   int64(0x89b8fd09), int64(0x8d79e0be), int64(0x803ac667), int64(0x84fbdbd0),
   int64(0x9abc8bd5), int64(0x9e7d9662), int64(0x933eb0bb), int64(0x97ffad0c),
   int64(0xafb010b1), int64(0xab710d06), int64(0xa6322bdf), int64(0xa2f33668),
   int64(0xbcb4666d), int64(0xb8757bda), int64(0xb5365d03), int64(0xb1f740b4)
};


/*-------------------------------------------------------------*/
/*--- end                                        crctable.c ---*/
/*-------------------------------------------------------------*/
//ABC_NAMESPACE_IMPL_END

