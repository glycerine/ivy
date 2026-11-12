package zlib

/* adler32.c -- compute the Adler-32 checksum of a data stream
 * Copyright (C) 1995-2007 Mark Adler
 * For conditions of distribution and use, see copyright notice in zlib.h
 */

/* @(#) $Id$ */

//#include <stdio.h>
//#include <stdlib.h>
//#include <string.h>
//#include "misc/util/abc_global.h"

//#include "zutil.h"

//ABC_NAMESPACE_IMPL_START

#define local static

// local uLong adler32_combine_(uLong adler1, uLong adler2, z_off64_t len2);


const BASE uint64 = 65521    /* largest prime smaller than 65536 */
const NMAX = 5552
/* NMAX is the largest n such that 255n(n+1)/2 + (n+1)(BASE-1) <= 2^32-1 */

// Original macro:
// #define DO1(buf,i)  {adler += (buf)[i]; sum2 += adler;}

// Original macro:
// #define DO2(buf,i)  DO1(buf,i); DO1(buf,i+1);

// Original macro:
// #define DO4(buf,i)  DO2(buf,i); DO2(buf,i+2);

// Original macro:
// #define DO8(buf,i)  DO4(buf,i); DO4(buf,i+4);

// Original macro:
// #define DO16(buf)   DO8(buf,0); DO8(buf,8);


/* use NO_DIVIDE if your processor does not do division in hardware */
#ifdef NO_DIVIDE
#  define MOD(a) \
    do { \
        if (a >= (BASE << 16)) a -= (BASE << 16); \
        if (a >= (BASE << 15)) a -= (BASE << 15); \
        if (a >= (BASE << 14)) a -= (BASE << 14); \
        if (a >= (BASE << 13)) a -= (BASE << 13); \
        if (a >= (BASE << 12)) a -= (BASE << 12); \
        if (a >= (BASE << 11)) a -= (BASE << 11); \
        if (a >= (BASE << 10)) a -= (BASE << 10); \
        if (a >= (BASE << 9)) a -= (BASE << 9); \
        if (a >= (BASE << 8)) a -= (BASE << 8); \
        if (a >= (BASE << 7)) a -= (BASE << 7); \
        if (a >= (BASE << 6)) a -= (BASE << 6); \
        if (a >= (BASE << 5)) a -= (BASE << 5); \
        if (a >= (BASE << 4)) a -= (BASE << 4); \
        if (a >= (BASE << 3)) a -= (BASE << 3); \
        if (a >= (BASE << 2)) a -= (BASE << 2); \
        if (a >= (BASE << 1)) a -= (BASE << 1); \
        if (a >= BASE) a -= BASE; \
    } while (0)
#  define MOD4(a) \
    do { \
        if (a >= (BASE << 4)) a -= (BASE << 4); \
        if (a >= (BASE << 3)) a -= (BASE << 3); \
        if (a >= (BASE << 2)) a -= (BASE << 2); \
        if (a >= (BASE << 1)) a -= (BASE << 1); \
        if (a >= BASE) a -= BASE; \
    } while (0)
#else
#  define MOD(a) a %= BASE
#  define MOD4(a) a %= BASE
#endif

/* ========================================================================= */
func adler32(adler uLong, buf *Bytef, len uInt) uLong ZEXPORT {
    unsigned long sum2;
    var n uint32;

    /* split Adler-32 into component sums */
    sum2 = (adler >> 16) & 0xffff;
    adler &= 0xffff;

    /* in case user likes doing a byte at a time, keep it fast */
    if len == 1 {
        adler += buf[0];
        if adler >= BASE {
            adler -= BASE; }
        sum2 += adler;
        if sum2 >= BASE {
            sum2 -= BASE; }
        return adler | (sum2 << 16);
    }

    /* initial Adler-32 value (deferred check for len == 1 speed) */
    if buf == Z_NULL {
        return int64(1); }

    /* in case short lengths are provided, keep it somewhat fast */
    if len < 16 {
        for len-- {
            adler += *buf++;
            sum2 += adler;
        }
        if adler >= BASE {
            adler -= BASE; }
        MOD4(sum2);             /* only added so many BASE's */
        return adler | (sum2 << 16);
    }

    /* do length NMAX blocks -- requires just one modulo operation */
    for len >= NMAX {
        len -= NMAX;
        n = NMAX / 16;          /* NMAX is divisible by 16 */
        for {
            // inlining of macro DO16
            // inlining of macro DO8
            // inlining of macro DO4
            // inlining of macro DO2
            // inlining of macro DO1
            {adler += (buf)[0]; sum2 += adler;}; {adler += (buf)[0+1]; sum2 += adler;};; {adler += (buf)[0+2]; sum2 += adler;}; {adler += (buf)[0+2+1]; sum2 += adler;};;; {adler += (buf)[0+4]; sum2 += adler;}; {adler += (buf)[0+4+1]; sum2 += adler;};; {adler += (buf)[0+4+2]; sum2 += adler;}; {adler += (buf)[0+4+2+1]; sum2 += adler;};;;; {adler += (buf)[8]; sum2 += adler;}; {adler += (buf)[8+1]; sum2 += adler;};; {adler += (buf)[8+2]; sum2 += adler;}; {adler += (buf)[8+2+1]; sum2 += adler;};;; {adler += (buf)[8+4]; sum2 += adler;}; {adler += (buf)[8+4+1]; sum2 += adler;};; {adler += (buf)[8+4+2]; sum2 += adler;}; {adler += (buf)[8+4+2+1]; sum2 += adler;};;;;;          /* 16 sums unrolled */
            buf += 16;
        
if !(--n) { break }
} 
        MOD(adler);
        MOD(sum2);
    }

    /* do remaining bytes (less than NMAX, still just one modulo) */
    if len {                  /* avoid modulos if none remaining */
        for len >= 16 {
            len -= 16;
            // inlining of macro DO16
            // inlining of macro DO8
            // inlining of macro DO4
            // inlining of macro DO2
            // inlining of macro DO1
            {adler += (buf)[0]; sum2 += adler;}; {adler += (buf)[0+1]; sum2 += adler;};; {adler += (buf)[0+2]; sum2 += adler;}; {adler += (buf)[0+2+1]; sum2 += adler;};;; {adler += (buf)[0+4]; sum2 += adler;}; {adler += (buf)[0+4+1]; sum2 += adler;};; {adler += (buf)[0+4+2]; sum2 += adler;}; {adler += (buf)[0+4+2+1]; sum2 += adler;};;;; {adler += (buf)[8]; sum2 += adler;}; {adler += (buf)[8+1]; sum2 += adler;};; {adler += (buf)[8+2]; sum2 += adler;}; {adler += (buf)[8+2+1]; sum2 += adler;};;; {adler += (buf)[8+4]; sum2 += adler;}; {adler += (buf)[8+4+1]; sum2 += adler;};; {adler += (buf)[8+4+2]; sum2 += adler;}; {adler += (buf)[8+4+2+1]; sum2 += adler;};;;;;
            buf += 16;
        }
        for len-- {
            adler += *buf++;
            sum2 += adler;
        }
        MOD(adler);
        MOD(sum2);
    }

    /* return recombined sums */
    return adler | (sum2 << 16);
}

/* ========================================================================= */
func adler32_combine_(adler1 uLong, adler2 uLong, len2 z_off64_t) local uLong {
    unsigned long sum1;
    unsigned long sum2;
    var rem uint32;

    /* the derivation of this formula is left as an exercise for the reader */
    rem = uint32((len2 % BASE));
    sum1 = adler1 & 0xffff;
    sum2 = rem * sum1;
    MOD(sum2);
    sum1 += (adler2 & 0xffff) + BASE - 1;
    sum2 += ((adler1 >> 16) & 0xffff) + ((adler2 >> 16) & 0xffff) + BASE - rem;
    if sum1 >= BASE { sum1 -= BASE; }
    if sum1 >= BASE { sum1 -= BASE; }
    if sum2 >= (BASE << 1) { sum2 -= (BASE << 1); }
    if sum2 >= BASE { sum2 -= BASE; }
    return sum1 | (sum2 << 16);
}

/* ========================================================================= */
func adler32_combine(adler1 uLong, adler2 uLong, len2 z_off_t) uLong ZEXPORT {
    return adler32_combine_(adler1, adler2, len2);
}

func adler32_combine64(adler1 uLong, adler2 uLong, len2 z_off64_t) uLong ZEXPORT {
    return adler32_combine_(adler1, adler2, len2);
}


//ABC_NAMESPACE_IMPL_END

