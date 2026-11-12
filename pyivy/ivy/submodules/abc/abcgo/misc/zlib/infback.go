package zlib

import "os"

/* infback.c -- inflate using a call-back interface
 * Copyright (C) 1995-2009 Mark Adler
 * For conditions of distribution and use, see copyright notice in zlib.h
 */

/*
   This code is largely copied from inflate.c.  Normally either infback.o or
   inflate.o would be linked into an application--not both.  The interface
   with inffast.c is retained so that optimized assembler-coded versions of
   inflate_fast() can be used with either inflate.c or infback.c.
 */

//#include <stdio.h>
//#include <stdlib.h>
//#include <string.h>
//#include "misc/util/abc_global.h"

//#include "zutil.h"
//#include "inftrees.h"
//#include "inflate.h"
//#include "inffast.h"

//ABC_NAMESPACE_IMPL_START

/* function prototypes */
// local void fixedtables OF((struct inflate_state FAR *state));


/*
   strm provides memory allocation functions in zalloc and zfree, or
   Z_NULL to use the library memory allocation functions.

   windowBits is in the range 8..15, and window is a user-supplied
   window and output buffer that is 2**windowBits bytes.
 */
int ZEXPORT inflateBackInit_(z_streamp strm, int windowBits, unsigned char FAR *window, char *version, int stream_size)
{
    struct inflate_state FAR *state;

    if version == Z_NULL || version[0] != ZLIB_VERSION[0] ||
        stream_size != int32((sizeof(z_stream))) {
        return Z_VERSION_ERROR; }
    if strm == Z_NULL || window == Z_NULL ||
        windowBits < 8 || windowBits > 15 {
        return Z_STREAM_ERROR; }
    strm.msg = Z_NULL;                 /* in case we return an error */
    if strm.zalloc == (alloc_func)0 {
        strm.zalloc = zcalloc;
        strm.opaque = (voidpf)0;
    }
    if strm.zfree == (free_func)0 { strm.zfree = zcfree; }
    // inlining of macro ZALLOC
    state = (struct inflate_state FAR *)(*((strm).zalloc))((strm).opaque, (1), (sizeof(struct inflate_state)));
    if state == Z_NULL { return Z_MEM_ERROR; }
    Tracev((os.Stderr, "inflate: allocated\n"));
    strm.state = (struct internal_state FAR *)state;
    state.dmax = uint32(32768);
    state.wbits = windowBits;
    state.wsize = uint32(1) << windowBits;
    state.window = window;
    state.wnext = 0;
    state.whave = 0;
    return Z_OK;
}

/*
   Return state with length and distance decoding tables and index sizes set to
   fixed code decoding.  Normally this returns fixed tables from inffixed.h.
   If BUILDFIXED is defined, then instead this routine builds the tables the
   first time it's called, and returns those tables the first time and
   thereafter.  This reduces the size of the code by about 2K bytes, in
   exchange for a little execution time.  However, BUILDFIXED should not be
   used for threaded applications, since the rewriting of the tables and virgin
   may not be thread-safe.
 */
local void fixedtables(struct inflate_state FAR *state)
{
#ifdef BUILDFIXED
    static int virgin = 1;
    static code *lenfix, *distfix;
    static code fixed[544];

    /* build fixed huffman tables if first call (may not be thread safe) */
    if (virgin) {
        unsigned sym, bits;
        static code *next;

        /* literal/length table */
        sym = 0;
        while (sym < 144) state->lens[sym++] = 8;
        while (sym < 256) state->lens[sym++] = 9;
        while (sym < 280) state->lens[sym++] = 7;
        while (sym < 288) state->lens[sym++] = 8;
        next = fixed;
        lenfix = next;
        bits = 9;
        inflate_table(LENS, state->lens, 288, &(next), &(bits), state->work);

        /* distance table */
        sym = 0;
        while (sym < 32) state->lens[sym++] = 5;
        distfix = next;
        bits = 5;
        inflate_table(DISTS, state->lens, 32, &(next), &(bits), state->work);

        /* do this just once */
        virgin = 0;
    }
#else /* !BUILDFIXED */
#   include "inffixed.h"
#endif /* BUILDFIXED */
    state.lencode = lenfix;
    state.lenbits = 9;
    state.distcode = distfix;
    state.distbits = 5;
}

/* Macros for inflateBack(): */

/* Load returned state from inflate_fast() */
// Original macro:
// #define LOAD() \
//     do { \
//         put = strm->next_out; \
//         left = strm->avail_out; \
//         next = strm->next_in; \
//         have = strm->avail_in; \
//         hold = state->hold; \
//         bits = state->bits; \
//     } while (0)


/* Set state from registers for inflate_fast() */
// Original macro:
// #define RESTORE() \
//     do { \
//         strm->next_out = put; \
//         strm->avail_out = left; \
//         strm->next_in = next; \
//         strm->avail_in = have; \
//         state->hold = hold; \
//         state->bits = bits; \
//     } while (0)


/* Clear the input bit accumulator */
// Original macro:
// #define INITBITS() \
//     do { \
//         hold = 0; \
//         bits = 0; \
//     } while (0)


/* Assure that some input is available.  If input is requested, but denied,
   then return a Z_BUF_ERROR from inflateBack(). */
// Original macro:
// #define PULL() \
//     do { \
//         if (have == 0) { \
//             have = in(in_desc, &next); \
//             if (have == 0) { \
//                 next = Z_NULL; \
//                 ret = Z_BUF_ERROR; \
//                 goto inf_leave; \
//             } \
//         } \
//     } while (0)


/* Get a byte of input into the bit accumulator, or return from inflateBack()
   with an error if there is no input available. */
// Original macro:
// #define PULLBYTE() \
//     do { \
//         PULL(); \
//         have--; \
//         hold += (unsigned long)(*next++) << bits; \
//         bits += 8; \
//     } while (0)


/* Assure that there are at least n bits in the bit accumulator.  If there is
   not enough available input to do that, then return from inflateBack() with
   an error. */
// Original macro:
// #define NEEDBITS(n) \
//     do { \
//         while (bits < (unsigned)(n)) \
//             PULLBYTE(); \
//     } while (0)


/* Return the low n bits of the bit accumulator (n < 16) */
// Original macro:
// #define BITS(n) \
//     ((unsigned)hold & ((1U << (n)) - 1))


/* Remove n bits from the bit accumulator */
// Original macro:
// #define DROPBITS(n) \
//     do { \
//         hold >>= (n); \
//         bits -= (unsigned)(n); \
//     } while (0)


/* Remove zero to seven bits as needed to go to a byte boundary */
// Original macro:
// #define BYTEBITS() \
//     do { \
//         hold >>= bits & 7; \
//         bits -= bits & 7; \
//     } while (0)


/* Assure that some output space is available, by writing out the window
   if it's full.  If the write fails, return from inflateBack() with a
   Z_BUF_ERROR. */
// Original macro:
// #define ROOM() \
//     do { \
//         if (left == 0) { \
//             put = state->window; \
//             left = state->wsize; \
//             state->whave = left; \
//             if (out(out_desc, put, left)) { \
//                 ret = Z_BUF_ERROR; \
//                 goto inf_leave; \
//             } \
//         } \
//     } while (0)


/*
   strm provides the memory allocation functions and window buffer on input,
   and provides information on the unused input on return.  For Z_DATA_ERROR
   returns, strm will also provide an error message.

   in() and out() are the call-back input and output functions.  When
   inflateBack() needs more input, it calls in().  When inflateBack() has
   filled the window with output, or when it completes with data in the
   window, it calls out() to write out the data.  The application must not
   change the provided input until in() is called again or inflateBack()
   returns.  The application must not change the window/output buffer until
   inflateBack() returns.

   in() and out() are called with a descriptor parameter provided in the
   inflateBack() call.  This parameter can be a structure that provides the
   information required to do the read or write, as well as accumulated
   information on the input and output such as totals and check values.

   in() should return zero on failure.  out() should return non-zero on
   failure.  If either in() or out() fails, than inflateBack() returns a
   Z_BUF_ERROR.  strm->next_in can be checked for Z_NULL to see whether it
   was in() or out() that caused in the error.  Otherwise,  inflateBack()
   returns Z_STREAM_END on success, Z_DATA_ERROR for an deflate format
   error, or Z_MEM_ERROR if it could not allocate memory for the state.
   inflateBack() can also return Z_STREAM_ERROR if the input parameters
   are not correct, i.e. strm is Z_NULL or the state was not initialized.
 */
int ZEXPORT inflateBack(z_streamp strm, in_func in, void FAR *in_desc, out_func out, void FAR *out_desc)
{
    struct inflate_state FAR *state;
    unsigned char FAR *next;    /* next input */
    unsigned char FAR *put;     /* next output */
    var have, left uint32;        /* available input and output */
    unsigned long hold;         /* bit buffer */
    var bits uint32;              /* bits in bit buffer */
    var copy uint32;              /* number of stored or match bytes to copy */
    unsigned char FAR *from;    /* where to copy match bytes from */
    var here code;                  /* current decoding table entry */
    var last code;                  /* parent table entry */
    var len uint32;               /* length to copy for repeats, bits to drop */
    var ret int32;                    /* return code */
    static unsigned short order[19] = /* permutation of code lengths */
        {16, 17, 18, 0, 8, 7, 9, 6, 10, 5, 11, 4, 12, 3, 13, 2, 14, 1, 15};

    /* Check that the strm exists and that the state was initialized */
    if strm == Z_NULL || strm.state == Z_NULL {
        return Z_STREAM_ERROR; }
    state = (struct inflate_state FAR *)strm.state;

    /* Reset the state */
    strm.msg = Z_NULL;
    state.mode = TYPE;
    state.last = 0;
    state.whave = 0;
    next = strm.next_in;
    if next != Z_NULL {
        have = strm.avail_in;
    } else {
        have = 0;
    }
    hold = 0;
    bits = 0;
    put = state.window;
    left = state.wsize;

    /* Inflate until end of block marked as last */
    for {
        switch state.mode {
        case TYPE:
            /* determine and dispatch block type */
            if state.last {
                // inlining of macro BYTEBITS
                for {         hold >>= bits & 7;         bits -= bits & 7;     
if !(false) { break }
} 
                state.mode = DONE;
                break;
            }
            // inlining of macro NEEDBITS
            // inlining of macro PULLBYTE
            // inlining of macro PULL
            for {         for bits < uint32((3)) {             for {         for {         if have == 0 {             have = in(in_desc, &next);             if have == 0 {                 next = Z_NULL;                 ret = Z_BUF_ERROR;                 goto inf_leave;             }         }     
if !(false) { break }
}
          have--;         hold += (unsigned long)(*next++) << bits;         bits += 8;     
if !(false) { break }
}  }     
if !(false) { break }
} 
            // inlining of macro BITS
            state.last = (uint32(hold) & ((uint32(1) << (1)) - 1));
            // inlining of macro DROPBITS
            for {         hold >>= (1);         bits -= uint32((1));     
if !(false) { break }
} 
            // inlining of macro BITS
            switch uint32(hold) & ((uint32(1) << (2)) - 1) {
            case 0:                             /* stored block */
                Tracev((os.Stderr, "inflate:     stored block%s\n",
                        state.last ? " (last)" : ""));
                state.mode = STORED;
                break;
            case 1:                             /* fixed block */
                fixedtables(state);
                Tracev((os.Stderr, "inflate:     fixed codes block%s\n",
                        state.last ? " (last)" : ""));
                state.mode = LEN;              /* decode codes */
                break;
            case 2:                             /* dynamic block */
                Tracev((os.Stderr, "inflate:     dynamic codes block%s\n",
                        state.last ? " (last)" : ""));
                state.mode = TABLE;
                break;
            case 3:
                strm.msg = (char *)"invalid block type";
                state.mode = BAD;
            }
            // inlining of macro DROPBITS
            for {         hold >>= (2);         bits -= uint32((2));     
if !(false) { break }
} 
            break;

        case STORED:
            /* get and verify stored block length */
            // inlining of macro BYTEBITS
            for {         hold >>= bits & 7;         bits -= bits & 7;     
if !(false) { break }
}                          /* go to byte boundary */
            // inlining of macro NEEDBITS
            // inlining of macro PULLBYTE
            // inlining of macro PULL
            for {         for bits < uint32((32)) {             for {         for {         if have == 0 {             have = in(in_desc, &next);             if have == 0 {                 next = Z_NULL;                 ret = Z_BUF_ERROR;                 goto inf_leave;             }         }     
if !(false) { break }
}
          have--;         hold += (unsigned long)(*next++) << bits;         bits += 8;     
if !(false) { break }
}  }     
if !(false) { break }
} 
            if (hold & 0xffff) != ((hold >> 16) ^ 0xffff) {
                strm.msg = (char *)"invalid stored block lengths";
                state.mode = BAD;
                break;
            }
            state.length = uint32(hold) & 0xffff;
            Tracev((os.Stderr, "inflate:       stored length %u\n",
                    state.length));
            // inlining of macro INITBITS
            for {         hold = 0;         bits = 0;     
if !(false) { break }
} 

            /* copy stored block from input to output */
            for state.length != 0 {
                copy = state.length;
                // inlining of macro PULL
                for {         if have == 0 {             have = in(in_desc, &next);             if have == 0 {                 next = Z_NULL;                 ret = Z_BUF_ERROR;                 goto inf_leave;             }         }     
if !(false) { break }
} 
                // inlining of macro ROOM
                for {         if left == 0 {             put = state.window;             left = state.wsize;             state.whave = left;             if out(out_desc, put, left) {                 ret = Z_BUF_ERROR;                 goto inf_leave;             }         }     
if !(false) { break }
} 
                if copy > have { copy = have; }
                if copy > left { copy = left; }
                zmemcpy(put, next, copy);
                have -= copy;
                next += copy;
                left -= copy;
                put += copy;
                state.length -= copy;
            }
            Tracev((os.Stderr, "inflate:       stored end\n"));
            state.mode = TYPE;
            break;

        case TABLE:
            /* get dynamic table entries descriptor */
            // inlining of macro NEEDBITS
            // inlining of macro PULLBYTE
            // inlining of macro PULL
            for {         for bits < uint32((14)) {             for {         for {         if have == 0 {             have = in(in_desc, &next);             if have == 0 {                 next = Z_NULL;                 ret = Z_BUF_ERROR;                 goto inf_leave;             }         }     
if !(false) { break }
}
          have--;         hold += (unsigned long)(*next++) << bits;         bits += 8;     
if !(false) { break }
}  }     
if !(false) { break }
} 
            // inlining of macro BITS
            state.nlen = (uint32(hold) & ((uint32(1) << (5)) - 1)) + 257;
            // inlining of macro DROPBITS
            for {         hold >>= (5);         bits -= uint32((5));     
if !(false) { break }
} 
            // inlining of macro BITS
            state.ndist = (uint32(hold) & ((uint32(1) << (5)) - 1)) + 1;
            // inlining of macro DROPBITS
            for {         hold >>= (5);         bits -= uint32((5));     
if !(false) { break }
} 
            // inlining of macro BITS
            state.ncode = (uint32(hold) & ((uint32(1) << (4)) - 1)) + 4;
            // inlining of macro DROPBITS
            for {         hold >>= (4);         bits -= uint32((4));     
if !(false) { break }
} 
#ifndef PKZIP_BUG_WORKAROUND
            if (state->nlen > 286 || state->ndist > 30) {
                strm->msg = (char *)"too many length or distance symbols";
                state->mode = BAD;
                break;
            }
#endif
            Tracev((os.Stderr, "inflate:       table sizes ok\n"));

            /* get code length code lengths (not a typo) */
            state.have = 0;
            for state.have < state.ncode {
                // inlining of macro NEEDBITS
                // inlining of macro PULLBYTE
                // inlining of macro PULL
                for {         for bits < uint32((3)) {             for {         for {         if have == 0 {             have = in(in_desc, &next);             if have == 0 {                 next = Z_NULL;                 ret = Z_BUF_ERROR;                 goto inf_leave;             }         }     
if !(false) { break }
}
          have--;         hold += (unsigned long)(*next++) << bits;         bits += 8;     
if !(false) { break }
}  }     
if !(false) { break }
} 
                // inlining of macro BITS
                state.lens[order[state.have++]] = (unsigned short)(uint32(hold) & ((uint32(1) << (3)) - 1));
                // inlining of macro DROPBITS
                for {         hold >>= (3);         bits -= uint32((3));     
if !(false) { break }
} 
            }
            for state.have < 19 {
                state.lens[order[state.have++]] = 0;
            }
            state.next = state.codes;
            state.lencode = (code FAR *)(state.next);
            state.lenbits = 7;
            ret = inflate_table(CODES, state.lens, 19, &(state.next),
                                &(state.lenbits), state.work);
            if ret != 0 {
                strm.msg = (char *)"invalid code lengths set";
                state.mode = BAD;
                break;
            }
            Tracev((os.Stderr, "inflate:       code lengths ok\n"));

            /* get length and distance code code lengths */
            state.have = 0;
            for state.have < state.nlen + state.ndist {
                for {
                    // inlining of macro BITS
                    here = state.lencode[(uint32(hold) & ((uint32(1) << (state.lenbits)) - 1))];
                    if uint32((here.bits)) <= bits { break; }
                    // inlining of macro PULLBYTE
                    // inlining of macro PULL
                    for {         for {         if have == 0 {             have = in(in_desc, &next);             if have == 0 {                 next = Z_NULL;                 ret = Z_BUF_ERROR;                 goto inf_leave;             }         }     
if !(false) { break }
}
          have--;         hold += (unsigned long)(*next++) << bits;         bits += 8;     
if !(false) { break }
} 
                }
                if here.val < 16 {
                    // inlining of macro NEEDBITS
                    // inlining of macro PULLBYTE
                    // inlining of macro PULL
                    for {         for bits < uint32((here.bits)) {             for {         for {         if have == 0 {             have = in(in_desc, &next);             if have == 0 {                 next = Z_NULL;                 ret = Z_BUF_ERROR;                 goto inf_leave;             }         }     
if !(false) { break }
}
          have--;         hold += (unsigned long)(*next++) << bits;         bits += 8;     
if !(false) { break }
}  }     
if !(false) { break }
} 
                    // inlining of macro DROPBITS
                    for {         hold >>= (here.bits);         bits -= uint32((here.bits));     
if !(false) { break }
} 
                    state.lens[state.have++] = here.val;
                
                } else {
                    if here.val == 16 {
                        // inlining of macro NEEDBITS
                        // inlining of macro PULLBYTE
                        // inlining of macro PULL
                        for {         for bits < uint32((here.bits + 2)) {             for {         for {         if have == 0 {             have = in(in_desc, &next);             if have == 0 {                 next = Z_NULL;                 ret = Z_BUF_ERROR;                 goto inf_leave;             }         }     
if !(false) { break }
}
          have--;         hold += (unsigned long)(*next++) << bits;         bits += 8;     
if !(false) { break }
}  }     
if !(false) { break }
} 
                        // inlining of macro DROPBITS
                        for {         hold >>= (here.bits);         bits -= uint32((here.bits));     
if !(false) { break }
} 
                        if state.have == 0 {
                            strm.msg = (char *)"invalid bit length repeat";
                            state.mode = BAD;
                            break;
                        }
                        len = uint32((state.lens[state.have - 1]));
                        // inlining of macro BITS
                        copy = 3 + (uint32(hold) & ((uint32(1) << (2)) - 1));
                        // inlining of macro DROPBITS
                        for {         hold >>= (2);         bits -= uint32((2));     
if !(false) { break }
} 
                    
                    } else if here.val == 17 {
                        // inlining of macro NEEDBITS
                        // inlining of macro PULLBYTE
                        // inlining of macro PULL
                        for {         for bits < uint32((here.bits + 3)) {             for {         for {         if have == 0 {             have = in(in_desc, &next);             if have == 0 {                 next = Z_NULL;                 ret = Z_BUF_ERROR;                 goto inf_leave;             }         }     
if !(false) { break }
}
          have--;         hold += (unsigned long)(*next++) << bits;         bits += 8;     
if !(false) { break }
}  }     
if !(false) { break }
} 
                        // inlining of macro DROPBITS
                        for {         hold >>= (here.bits);         bits -= uint32((here.bits));     
if !(false) { break }
} 
                        len = 0;
                        // inlining of macro BITS
                        copy = 3 + (uint32(hold) & ((uint32(1) << (3)) - 1));
                        // inlining of macro DROPBITS
                        for {         hold >>= (3);         bits -= uint32((3));     
if !(false) { break }
} 
                    
                    } else {
                        // inlining of macro NEEDBITS
                        // inlining of macro PULLBYTE
                        // inlining of macro PULL
                        for {         for bits < uint32((here.bits + 7)) {             for {         for {         if have == 0 {             have = in(in_desc, &next);             if have == 0 {                 next = Z_NULL;                 ret = Z_BUF_ERROR;                 goto inf_leave;             }         }     
if !(false) { break }
}
          have--;         hold += (unsigned long)(*next++) << bits;         bits += 8;     
if !(false) { break }
}  }     
if !(false) { break }
} 
                        // inlining of macro DROPBITS
                        for {         hold >>= (here.bits);         bits -= uint32((here.bits));     
if !(false) { break }
} 
                        len = 0;
                        // inlining of macro BITS
                        copy = 11 + (uint32(hold) & ((uint32(1) << (7)) - 1));
                        // inlining of macro DROPBITS
                        for {         hold >>= (7);         bits -= uint32((7));     
if !(false) { break }
} 
                    }
                    if state.have + copy > state.nlen + state.ndist {
                        strm.msg = (char *)"invalid bit length repeat";
                        state.mode = BAD;
                        break;
                    }
                    for copy-- {
                        state.lens[state.have++] = (unsigned short)len;
                    }
                }
            }

            /* handle error breaks in while */
            if state.mode == BAD { break; }

            /* check for end-of-block code (better have one) */
            if state.lens[256] == 0 {
                strm.msg = (char *)"invalid code -- missing end-of-block";
                state.mode = BAD;
                break;
            }

            /* build code tables -- note: do not change the lenbits or distbits
               values here (9 and 6) without reading the comments in inftrees.h
               concerning the ENOUGH constants, which depend on those values */
            state.next = state.codes;
            state.lencode = (code FAR *)(state.next);
            state.lenbits = 9;
            ret = inflate_table(LENS, state.lens, state.nlen, &(state.next),
                                &(state.lenbits), state.work);
            if ret != 0 {
                strm.msg = (char *)"invalid literal/lengths set";
                state.mode = BAD;
                break;
            }
            state.distcode = (code FAR *)(state.next);
            state.distbits = 6;
            ret = inflate_table(DISTS, state.lens + state.nlen, state.ndist,
                            &(state.next), &(state.distbits), state.work);
            if ret != 0 {
                strm.msg = (char *)"invalid distances set";
                state.mode = BAD;
                break;
            }
            Tracev((os.Stderr, "inflate:       codes ok\n"));
            state.mode = LEN;

        fallthrough
case LEN:
            /* use inflate_fast() if we have enough input and output */
            if have >= 6 && left >= 258 {
                // inlining of macro RESTORE
                for {         strm.next_out = put;         strm.avail_out = left;         strm.next_in = next;         strm.avail_in = have;         state.hold = hold;         state.bits = bits;     
if !(false) { break }
} 
                if state.whave < state.wsize {
                    state.whave = state.wsize - left; }
                inflate_fast(strm, state.wsize);
                // inlining of macro LOAD
                for {         put = strm.next_out;         left = strm.avail_out;         next = strm.next_in;         have = strm.avail_in;         hold = state.hold;         bits = state.bits;     
if !(false) { break }
} 
                break;
            }

            /* get a literal, length, or end-of-block code */
            for {
                // inlining of macro BITS
                here = state.lencode[(uint32(hold) & ((uint32(1) << (state.lenbits)) - 1))];
                if uint32((here.bits)) <= bits { break; }
                // inlining of macro PULLBYTE
                // inlining of macro PULL
                for {         for {         if have == 0 {             have = in(in_desc, &next);             if have == 0 {                 next = Z_NULL;                 ret = Z_BUF_ERROR;                 goto inf_leave;             }         }     
if !(false) { break }
}
          have--;         hold += (unsigned long)(*next++) << bits;         bits += 8;     
if !(false) { break }
} 
            }
            if here.op && (here.op & 0xf0) == 0 {
                last = here;
                for {
                    here = state.lencode[last.val +
                            // inlining of macro BITS
                            ((uint32(hold) & ((uint32(1) << (last.bits + last.op)) - 1)) >> last.bits)];
                    if uint32((last.bits + here.bits)) <= bits { break; }
                    // inlining of macro PULLBYTE
                    // inlining of macro PULL
                    for {         for {         if have == 0 {             have = in(in_desc, &next);             if have == 0 {                 next = Z_NULL;                 ret = Z_BUF_ERROR;                 goto inf_leave;             }         }     
if !(false) { break }
}
          have--;         hold += (unsigned long)(*next++) << bits;         bits += 8;     
if !(false) { break }
} 
                }
                // inlining of macro DROPBITS
                for {         hold >>= (last.bits);         bits -= uint32((last.bits));     
if !(false) { break }
} 
            }
            // inlining of macro DROPBITS
            for {         hold >>= (here.bits);         bits -= uint32((here.bits));     
if !(false) { break }
} 
            state.length = uint32(here.val);

            /* process literal */
            if here.op == 0 {
                Tracevv((os.Stderr, here.val >= 0x20 && here.val < 0x7f ?
                        "inflate:         literal '%c'\n" :
                        "inflate:         literal 0x%02x\n", here.val));
                // inlining of macro ROOM
                for {         if left == 0 {             put = state.window;             left = state.wsize;             state.whave = left;             if out(out_desc, put, left) {                 ret = Z_BUF_ERROR;                 goto inf_leave;             }         }     
if !(false) { break }
} 
                *put++ = (unsigned char)(state.length);
                left--;
                state.mode = LEN;
                break;
            }

            /* process end of block */
            if here.op & 32 {
                Tracevv((os.Stderr, "inflate:         end of block\n"));
                state.mode = TYPE;
                break;
            }

            /* invalid code */
            if here.op & 64 {
                strm.msg = (char *)"invalid literal/length code";
                state.mode = BAD;
                break;
            }

            /* length code -- get extra bits, if any */
            state.extra = uint32((here.op)) & 15;
            if state.extra != 0 {
                // inlining of macro NEEDBITS
                // inlining of macro PULLBYTE
                // inlining of macro PULL
                for {         for bits < uint32((state.extra)) {             for {         for {         if have == 0 {             have = in(in_desc, &next);             if have == 0 {                 next = Z_NULL;                 ret = Z_BUF_ERROR;                 goto inf_leave;             }         }     
if !(false) { break }
}
          have--;         hold += (unsigned long)(*next++) << bits;         bits += 8;     
if !(false) { break }
}  }     
if !(false) { break }
} 
                // inlining of macro BITS
                state.length += (uint32(hold) & ((uint32(1) << (state.extra)) - 1));
                // inlining of macro DROPBITS
                for {         hold >>= (state.extra);         bits -= uint32((state.extra));     
if !(false) { break }
} 
            }
            Tracevv((os.Stderr, "inflate:         length %u\n", state.length));

            /* get distance code */
            for {
                // inlining of macro BITS
                here = state.distcode[(uint32(hold) & ((uint32(1) << (state.distbits)) - 1))];
                if uint32((here.bits)) <= bits { break; }
                // inlining of macro PULLBYTE
                // inlining of macro PULL
                for {         for {         if have == 0 {             have = in(in_desc, &next);             if have == 0 {                 next = Z_NULL;                 ret = Z_BUF_ERROR;                 goto inf_leave;             }         }     
if !(false) { break }
}
          have--;         hold += (unsigned long)(*next++) << bits;         bits += 8;     
if !(false) { break }
} 
            }
            if (here.op & 0xf0) == 0 {
                last = here;
                for {
                    here = state.distcode[last.val +
                            // inlining of macro BITS
                            ((uint32(hold) & ((uint32(1) << (last.bits + last.op)) - 1)) >> last.bits)];
                    if uint32((last.bits + here.bits)) <= bits { break; }
                    // inlining of macro PULLBYTE
                    // inlining of macro PULL
                    for {         for {         if have == 0 {             have = in(in_desc, &next);             if have == 0 {                 next = Z_NULL;                 ret = Z_BUF_ERROR;                 goto inf_leave;             }         }     
if !(false) { break }
}
          have--;         hold += (unsigned long)(*next++) << bits;         bits += 8;     
if !(false) { break }
} 
                }
                // inlining of macro DROPBITS
                for {         hold >>= (last.bits);         bits -= uint32((last.bits));     
if !(false) { break }
} 
            }
            // inlining of macro DROPBITS
            for {         hold >>= (here.bits);         bits -= uint32((here.bits));     
if !(false) { break }
} 
            if here.op & 64 {
                strm.msg = (char *)"invalid distance code";
                state.mode = BAD;
                break;
            }
            state.offset = uint32(here.val);

            /* get distance extra bits, if any */
            state.extra = uint32((here.op)) & 15;
            if state.extra != 0 {
                // inlining of macro NEEDBITS
                // inlining of macro PULLBYTE
                // inlining of macro PULL
                for {         for bits < uint32((state.extra)) {             for {         for {         if have == 0 {             have = in(in_desc, &next);             if have == 0 {                 next = Z_NULL;                 ret = Z_BUF_ERROR;                 goto inf_leave;             }         }     
if !(false) { break }
}
          have--;         hold += (unsigned long)(*next++) << bits;         bits += 8;     
if !(false) { break }
}  }     
if !(false) { break }
} 
                // inlining of macro BITS
                state.offset += (uint32(hold) & ((uint32(1) << (state.extra)) - 1));
                // inlining of macro DROPBITS
                for {         hold >>= (state.extra);         bits -= uint32((state.extra));     
if !(false) { break }
} 
            }
            if state.offset > state.wsize - (state.whave < state.wsize ?
                                                left : 0) {
                strm.msg = (char *)"invalid distance too far back";
                state.mode = BAD;
                break;
            }
            Tracevv((os.Stderr, "inflate:         distance %u\n", state.offset));

            /* copy match from window to output */
            for {
                // inlining of macro ROOM
                for {         if left == 0 {             put = state.window;             left = state.wsize;             state.whave = left;             if out(out_desc, put, left) {                 ret = Z_BUF_ERROR;                 goto inf_leave;             }         }     
if !(false) { break }
} 
                copy = state.wsize - state.offset;
                if copy < left {
                    from = put + copy;
                    copy = left - copy;
                
                } else {
                    from = put - state.offset;
                    copy = left;
                }
                if copy > state.length { copy = state.length; }
                state.length -= copy;
                left -= copy;
                for {
                    *put++ = *from++;
                
if !(--copy) { break }
} 
            
if !(state.length != 0) { break }
} 
            break;

        case DONE:
            /* inflate stream terminated properly -- write leftover output */
            ret = Z_STREAM_END;
            if left < state.wsize {
                if out(out_desc, state.window, state.wsize - left) {
                    ret = Z_BUF_ERROR; }
            }
            goto inf_leave;

        case BAD:
            ret = Z_DATA_ERROR;
            goto inf_leave;

        default:                /* can't happen, but makes compilers happy */
            ret = Z_STREAM_ERROR;
            goto inf_leave;
        }
    }

    /* Return unused input */
  inf_leave:
    strm.next_in = next;
    strm.avail_in = have;
    return ret;
}

func inflateBackEnd(strm z_streamp) int ZEXPORT {
    if strm == Z_NULL || strm.state == Z_NULL || strm.zfree == (free_func)0 {
        return Z_STREAM_ERROR; }
    ZFREE(strm, strm.state);
    strm.state = Z_NULL;
    Tracev((os.Stderr, "inflate: end\n"));
    return Z_OK;
}


//ABC_NAMESPACE_IMPL_END

