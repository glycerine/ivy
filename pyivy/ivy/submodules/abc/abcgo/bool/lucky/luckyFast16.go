package lucky

/**CFile****************************************************************

  FileName    [luckyFast16.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Semi-canonical form computation package.]

  Synopsis    [Truth table minimization procedures for up to  16 vars.]

  Author      [Jake]

  Date        [Started - September 2012]

***********************************************************************/

//#include "luckyInt.h"
//#define LUCKY_VERIFY

//ABC_NAMESPACE_IMPL_START

var SFmask [5][4]word = [5][4]word{
    {ABC_CONST(0x8888888888888888),ABC_CONST(0x4444444444444444),ABC_CONST(0x2222222222222222),ABC_CONST(0x1111111111111111)},
    {ABC_CONST(0xC0C0C0C0C0C0C0C0),ABC_CONST(0x3030303030303030),ABC_CONST(0x0C0C0C0C0C0C0C0C),ABC_CONST(0x0303030303030303)},
    {ABC_CONST(0xF000F000F000F000),ABC_CONST(0x0F000F000F000F00),ABC_CONST(0x00F000F000F000F0),ABC_CONST(0x000F000F000F000F)},
    {ABC_CONST(0xFF000000FF000000),ABC_CONST(0x00FF000000FF0000),ABC_CONST(0x0000FF000000FF00),ABC_CONST(0x000000FF000000FF)},
    {ABC_CONST(0xFFFF000000000000),ABC_CONST(0x0000FFFF00000000),ABC_CONST(0x00000000FFFF0000),ABC_CONST(0x000000000000FFFF)},   
};

// we need next two functions only for verification of lucky method in debugging mode 
func swapAndFlip(pAfter *word, nVars int32, iVarInPosition int32, jVar int32, pCanonPerm string, pUCanonPhase *uint32) {
    var Temp int32;
    swap_ij(pAfter, nVars, iVarInPosition, jVar);
    
    Temp = pCanonPerm[iVarInPosition];
    pCanonPerm[iVarInPosition] = pCanonPerm[jVar];
    pCanonPerm[jVar] = Temp;
    
    if ((*pUCanonPhase & (1 << iVarInPosition)) > 0) != ((*pUCanonPhase & (1 << jVar)) > 0) {
        *pUCanonPhase ^= (1 << iVarInPosition);
        *pUCanonPhase ^= (1 << jVar);
    }
    if (*pUCanonPhase>>iVarInPosition) & 1 {
        Kit_TruthChangePhase_64bit( pAfter, nVars, iVarInPosition ); }
    
}
func luckyCheck(pAfter *word, pBefore *word, nVars int32, pCanonPerm string, uCanonPhase uint32) int32 {
    var i, j int32;
    var tempChar byte;
    for j=0;j<nVars;j++ {
        tempChar = 'a'+ j;
        for i=j;i<nVars;i++ {
            if tempChar != pCanonPerm[i] {
                continue; }
            swapAndFlip(pAfter , nVars, j, i, pCanonPerm, &uCanonPhase);
            break;
        }
    }
    if (uCanonPhase>>nVars) & 1 {
        Kit_TruthNot_64bit(pAfter, nVars ); }
    if memcmp(pAfter, pBefore, Kit_TruthWordNum_64bit( nVars )*sizeof(word)) == 0 {
        return 0;
    } else {
        return 1; }
}

////////////////////////////////////lessThen5/////////////////////////////////////////////////////////////////////////////////////////////

// there are 4 parts in every block to compare and rearrange - quoters(0Q,1Q,2Q,3Q)
//updataInfo updates CanonPerm and CanonPhase based on what quoter in position iQ and jQ
func updataInfo(iQ int32, jQ int32, iVar int32, pCanonPerm string, pCanonPhase *uint32) {
    *pCanonPhase = adjustInfoAfterSwap(pCanonPerm, *pCanonPhase, iVar, ((abs(iQ-jQ)-1)<<2) + iQ );

}

// returns the first shift from the left in word x that has One bit in it.
// blockSize = ShiftSize/4
func firstShiftWithOneBit(x word, blockSize int32) int32 {
    var n int32;
    if blockSize == 16{ return 0;}     
    if x >= ABC_CONST(0x0000000100000000) {n = n + 32; x = x >> 32;} 
    if blockSize == 8{ return (64-n)/32;}  
    if x >= ABC_CONST(0x0000000000010000) {n = n + 16; x = x >> 16;} 
    if blockSize == 4{ return (64-n)/16;}
    if x >= ABC_CONST(0x0000000000000100) {n = n + 8; x = x >> 8;}
    if blockSize == 2{ return (64-n)/8;}
    if x >= ABC_CONST(0x0000000000000010) {n = n + 4; x = x >> 4;} 
    return (64-n)/4;    
    
}

// It rearranges InOut (swaps and flips through rearrangement of quoters)
// It updates Info at the end
func arrangeQuoters_superFast_lessThen5(pInOut *word, start int32, iQ int32, jQ int32, kQ int32, lQ int32, iVar int32, nWords int32, pCanonPerm string, pCanonPhase *uint32) {
    var i int32;
    var blockSize int32 = 1<<iVar;
//     printf("in arrangeQuoters_superFast_lessThen5\n");
//     printf("start = %d, iQ = %d,jQ = %d,kQ = %d,lQ = %d, iVar = %d, nWords = %d\n", start, iQ, jQ, kQ , lQ, iVar, nWords);
    for i=start;i>=0;i-- {   
        assert(iQ*blockSize < 64, "iQ*blockSize < 64");
        assert(jQ*blockSize < 64, "jQ*blockSize < 64");
        assert(kQ*blockSize < 64, "kQ*blockSize < 64");
        assert(lQ*blockSize < 64, "lQ*blockSize < 64");
        assert(3*blockSize < 64, "3*blockSize < 64");
        pInOut[i] = (pInOut[i] & SFmask[iVar][iQ])<<(iQ*blockSize) |
            (((pInOut[i] & SFmask[iVar][jQ])<<(jQ*blockSize))>>blockSize) |
            (((pInOut[i] & SFmask[iVar][kQ])<<(kQ*blockSize))>>2*blockSize) |
            (((pInOut[i] & SFmask[iVar][lQ])<<(lQ*blockSize))>>3*blockSize);
    }
    updataInfo(iQ, jQ, iVar, pCanonPerm, pCanonPhase);
//     printf("out arrangeQuoters_superFast_lessThen5\n");

}
// static word SFmask[5][4] = {
//     {0x8888888888888888,0x4444444444444444,0x2222222222222222,0x1111111111111111},
//     {0xC0C0C0C0C0C0C0C0,0x3030303030303030,0x0C0C0C0C0C0C0C0C,0x0303030303030303},
//     {0xF000F000F000F000,0x0F000F000F000F00,0x00F000F000F000F0,0x000F000F000F000F},
//     {0xFF000000FF000000,0x00FF000000FF0000,0x0000FF000000FF00,0x000000FF000000FF},
//     {0xFFFF000000000000,0x0000FFFF00000000,0x00000000FFFF0000,0x000000000000FFFF}   
// };
//It compares 0Q and 3Q and returns 0 if 0Q is smaller then 3Q ( comparison starts at highest bit) and visa versa
// DifStart contains the information about the first different bit in 0Q and 3Q
func minTemp0_fast(pInOut *word, iVar int32, nWords int32, pDifStart *int32) int32 {
    var i int32;
    var blockSize int32 = 1<<iVar;
    var temp word;
    for i=nWords - 1; i>=0; i-- {
        assert(3*blockSize < 64, "3*blockSize < 64");
        temp = ((pInOut[i] & SFmask[iVar][0])) ^ ((pInOut[i] & SFmask[iVar][3])<<(3*blockSize));
        if temp == 0 {
            continue;
        } else {
            *pDifStart = i*100 + 20 - firstShiftWithOneBit(temp, blockSize);
            if ((pInOut[i] & SFmask[iVar][0])) < ((pInOut[i] & SFmask[iVar][3])<<(3*blockSize)) {
                return 0;
            } else {
                return 3; }
        }
    }
    *pDifStart=0;
    return 0;

}

//It compares 1Q and 2Q and returns 1 if 1Q is smaller then 2Q ( comparison starts at highest bit) and visa versa
// DifStart contains the information about the first different bit in 1Q and 2Q
func minTemp1_fast(pInOut *word, iVar int32, nWords int32, pDifStart *int32) int32 {
    var i int32;
    var blockSize int32 = 1<<iVar;
    var temp word;
    for i=nWords - 1; i>=0; i-- {
        assert(2*blockSize < 64, "2*blockSize < 64");
        temp = ((pInOut[i] & SFmask[iVar][1])<<(blockSize)) ^ ((pInOut[i] & SFmask[iVar][2])<<(2*blockSize));
        if temp == 0 {
            continue;
        } else {
            *pDifStart = i*100 + 20 - firstShiftWithOneBit(temp, blockSize);
            if ((pInOut[i] & SFmask[iVar][1])<<(blockSize)) < ((pInOut[i] & SFmask[iVar][2])<<(2*blockSize)) {
                return 1;
            } else {
                return 2; }
        }
    }
    *pDifStart=0;
    return 1;
}

//It compares iQ and jQ and returns 0 if iQ is smaller then jQ ( comparison starts at highest bit) and 1 if jQ is
// DifStart contains the information about the first different bit in iQ and jQ
func minTemp2_fast(pInOut *word, iVar int32, iQ int32, jQ int32, nWords int32, pDifStart *int32) int32 {
    var i int32;
    var blockSize int32 = 1<<iVar;
    var temp word;
    for i=nWords - 1; i>=0; i-- {
        assert(jQ*blockSize < 64, "jQ*blockSize < 64");
        temp = ((pInOut[i] & SFmask[iVar][iQ])<<(iQ*blockSize)) ^ ((pInOut[i] & SFmask[iVar][jQ])<<(jQ*blockSize));
        if temp == 0 {
            continue;
        } else {
            *pDifStart = i*100 + 20 - firstShiftWithOneBit(temp, blockSize);
            if ((pInOut[i] & SFmask[iVar][iQ])<<(iQ*blockSize)) <= ((pInOut[i] & SFmask[iVar][jQ])<<(jQ*blockSize)) {
                return 0;
            } else {
                return 1; }
        }
    }
    *pDifStart=0;
    return 0;
}
// same as minTemp2_fast but this one has a start position
func minTemp3_fast(pInOut *word, iVar int32, start int32, finish int32, iQ int32, jQ int32, pDifStart *int32) int32 {
    var i int32;
    var blockSize int32 = 1<<iVar;
    var temp word;
    for i=start; i>=finish; i-- {
        assert(jQ*blockSize < 64, "jQ*blockSize < 64");
        temp = ((pInOut[i] & SFmask[iVar][iQ])<<(iQ*blockSize)) ^ ((pInOut[i] & SFmask[iVar][jQ])<<(jQ*blockSize));
        if temp == 0 {
            continue;
        } else {
            *pDifStart = i*100 + 20 - firstShiftWithOneBit(temp, blockSize);
            if ((pInOut[i] & SFmask[iVar][iQ])<<(iQ*blockSize)) <= ((pInOut[i] & SFmask[iVar][jQ])<<(jQ*blockSize)) {
                return 0;
            } else {
                return 1; }
        }
    }
    *pDifStart=0;
    return 0;
}

// It considers all swap and flip possibilities of iVar and iVar+1 and switches InOut to a minimal of them  
func minimalSwapAndFlipIVar_superFast_lessThen5(pInOut *word, iVar int32, nWords int32, pCanonPerm string, pCanonPhase *uint32) {
    var min1, min2, DifStart0, DifStart1, DifStartMin, DifStart4 int32;
    var M [2]int32;   
    M[0] = minTemp0_fast(pInOut, iVar, nWords, &DifStart0); // 0, 3
    M[1] = minTemp1_fast(pInOut, iVar, nWords, &DifStart1); // 1, 2
    min1 = minTemp2_fast(pInOut, iVar, M[0], M[1], nWords, &DifStartMin);
//     printf("\nDifStart0 = %d, DifStart1 = %d, DifStartMin = %d\n",DifStart0, DifStart1, DifStartMin);
//     printf("M[0] = %d, M[1] = %d, min1 = %d\n", M[0], M[1], min1);
    if DifStart0 != DifStart1 {
//         printf("if\n");
        if DifStartMin>=DifStart1 && DifStartMin>=DifStart0 {
            arrangeQuoters_superFast_lessThen5(pInOut, DifStartMin/100, M[min1], M[(min1+1)&1], 3 - M[(min1+1)&1], 3 - M[min1], iVar, nWords, pCanonPerm, pCanonPhase);
        } else if DifStart0 > DifStart1 {
            arrangeQuoters_superFast_lessThen5(pInOut,luckyMax(DifStartMin/100, DifStart0/100), M[0], M[1], 3 - M[1], 3 - M[0], iVar, nWords, pCanonPerm, pCanonPhase);
        } else {
            arrangeQuoters_superFast_lessThen5(pInOut,luckyMax(DifStartMin/100, DifStart1/100), M[1], M[0], 3 - M[0], 3 - M[1], iVar, nWords, pCanonPerm, pCanonPhase); }
    
    } else {
//         printf("else\n");
        if DifStartMin>=DifStart0 {
            arrangeQuoters_superFast_lessThen5(pInOut, DifStartMin/100, M[min1], M[(min1+1)&1], 3 - M[(min1+1)&1], 3 - M[min1], iVar, nWords, pCanonPerm, pCanonPhase);
        } else {
            min2 = minTemp3_fast(pInOut, iVar, DifStart0/100, DifStartMin/100, 3-M[0], 3-M[1], &DifStart4);  // no reuse DifStart1 because DifStart1 = DifStart1=0
//             printf("after minTemp3_fast min2 = %d, DifStart4 = %d\n", min2, DifStart4);
            if DifStart4>DifStartMin {
                arrangeQuoters_superFast_lessThen5(pInOut, DifStart0/100, M[(min2+1)&1], M[min2], 3 - M[min2], 3 - M[(min2+1)&1], iVar, nWords, pCanonPerm, pCanonPhase);
            } else {
                arrangeQuoters_superFast_lessThen5(pInOut, DifStart0/100, M[min1], M[(min1+1)&1], 3 - M[(min1+1)&1], 3 - M[min1], iVar, nWords, pCanonPerm, pCanonPhase); }
        }
    }
}

func minimalSwapAndFlipIVar_superFast_lessThen5_noEBFC(pInOut *word, iVar int32, nWords int32, pCanonPerm string, pCanonPhase *uint32) {
    var DifStart1 int32;
    if minTemp1_fast(pInOut, iVar, nWords, &DifStart1) == 2 {
        arrangeQuoters_superFast_lessThen5(pInOut, DifStart1/100, 0, 2, 1, 3, iVar, nWords, pCanonPerm, pCanonPhase); } 
}
////////////////////////////////////iVar = 5/////////////////////////////////////////////////////////////////////////////////////////////

// It rearranges InOut (swaps and flips through rearrangement of quoters)
// It updates Info at the end
func arrangeQuoters_superFast_iVar5(pInOut *uint32, temp *uint32, start int32, iQ int32, jQ int32, kQ int32, lQ int32, pCanonPerm string, pCanonPhase *uint32) {
    var i, blockSize, shiftSize int32;
    var tempPtr *uint32  = temp+start;
//     printf("in arrangeQuoters_superFast_iVar5\n");

    if iQ == 0 && jQ == 1 {
        return; } 
    blockSize = sizeof(unsigned);
    shiftSize = 4;
    for i=start-1;i>0;i-=shiftSize {       
        tempPtr -= 1;       
        memcpy(tempPtr, pInOut+i-iQ, (size_t)blockSize);
        tempPtr -= 1;
        memcpy(tempPtr, pInOut+i-jQ, (size_t)blockSize);
        tempPtr -= 1;
        memcpy(tempPtr, pInOut+i-kQ, (size_t)blockSize);
        tempPtr -= 1;
        memcpy(tempPtr, pInOut+i-lQ, (size_t)blockSize);        
    }   
    memcpy(pInOut, temp, start*sizeof(unsigned));
    updataInfo(iQ, jQ, 5, pCanonPerm, pCanonPhase);
}

//It compares 0Q and 3Q and returns 0 if 0Q is smaller then 3Q ( comparison starts at highest bit) and visa versa
// DifStart contains the information about the first different bit in 0Q and 3Q
func minTemp0_fast_iVar5(pInOut *uint32, nWords int32, pDifStart *int32) int32 {
    var i, temp int32;
//     printf("in minTemp0_fast_iVar5\n");
    for i=(nWords)*2 - 1; i>=0; i-=4 {
        temp = CompareWords(pInOut[i],pInOut[i-3]);
        if temp == 0 {
            continue;
        } else if temp == -1 {
            *pDifStart = i+1;
            return 0;
        
        } else {
            *pDifStart = i+1;
            return 3;
        }
    }
    *pDifStart=0;
    return 0;
}

//It compares 1Q and 2Q and returns 1 if 1Q is smaller then 2Q ( comparison starts at highest bit) and visa versa
// DifStart contains the information about the first different bit in 1Q and 2Q
func minTemp1_fast_iVar5(pInOut *uint32, nWords int32, pDifStart *int32) int32 {
    var i, temp int32;
//     printf("in minTemp1_fast_iVar5\n");
    for i=(nWords)*2 - 2; i>=0; i-=4 {
        temp = CompareWords(pInOut[i],pInOut[i-1]);
        if temp == 0 {
            continue;
        } else if temp == -1 {
            *pDifStart = i+2;
            return 1;
        
        } else {
            *pDifStart = i+2;
            return 2;
        }
    }
    *pDifStart=0;
    return 1;
}

//It compares iQ and jQ and returns 0 if iQ is smaller then jQ ( comparison starts at highest bit) and visa versa
// DifStart contains the information about the first different bit in iQ and jQ
func minTemp2_fast_iVar5(pInOut *uint32, iQ int32, jQ int32, nWords int32, pDifStart *int32) int32 {
    var i, temp int32;
//     printf("in minTemp2_fast_iVar5\n");

    for i=(nWords)*2 - 1; i>=0; i-=4 {
        temp = CompareWords(pInOut[i-iQ],pInOut[i-jQ]);
        if temp == 0 {
            continue;
        } else if temp == -1 {
            *pDifStart = i+1;
            return 0;
        
        } else {
            *pDifStart = i+1;
            return 1;
        }
    }
    *pDifStart=0;
    return 0;
}

// same as minTemp2_fast but this one has a start position
func minTemp3_fast_iVar5(pInOut *uint32, start int32, finish int32, iQ int32, jQ int32, pDifStart *int32) int32 {
    var i, temp int32;
//     printf("in minTemp3_fast_iVar5\n");

    for i=start-1; i>=finish; i-=4 {
        temp = CompareWords(pInOut[i-iQ],pInOut[i-jQ]);
        if temp == 0 {
            continue;
        } else if temp == -1 {
            *pDifStart = i+1;
            return 0;
        
        } else {
            *pDifStart = i+1;
            return 1;
        }
    }
    *pDifStart=0;
    return 0;
}

// It considers all swap and flip possibilities of iVar and iVar+1 and switches InOut to a minimal of them 
func minimalSwapAndFlipIVar_superFast_iVar5(pInOut *uint32, nWords int32, pCanonPerm string, pCanonPhase *uint32) {
    var min1, min2, DifStart0, DifStart1, DifStartMin int32;
    var M [2]int32;
    var temp [2048]uint32;
//     printf("in minimalSwapAndFlipIVar_superFast_iVar5\n");
    M[0] = minTemp0_fast_iVar5(pInOut, nWords, &DifStart0); // 0, 3
    M[1] = minTemp1_fast_iVar5(pInOut, nWords, &DifStart1); // 1, 2
    min1 = minTemp2_fast_iVar5(pInOut, M[0], M[1], nWords, &DifStartMin);
    if DifStart0 != DifStart1 {   
        if DifStartMin>=DifStart1 && DifStartMin>=DifStart0 {
            arrangeQuoters_superFast_iVar5(pInOut, temp, DifStartMin, M[min1], M[(min1+1)&1], 3 - M[(min1+1)&1], 3 - M[min1], pCanonPerm, pCanonPhase);
        } else if DifStart0 > DifStart1 {
            arrangeQuoters_superFast_iVar5(pInOut, temp, luckyMax(DifStartMin,DifStart0), M[0], M[1], 3 - M[1], 3 - M[0], pCanonPerm, pCanonPhase);
        } else {
            arrangeQuoters_superFast_iVar5(pInOut, temp, luckyMax(DifStartMin,DifStart1), M[1], M[0], 3 - M[0], 3 - M[1], pCanonPerm, pCanonPhase); }
    
    } else {
        if DifStartMin>=DifStart0 {
            arrangeQuoters_superFast_iVar5(pInOut, temp, DifStartMin, M[min1], M[(min1+1)&1], 3 - M[(min1+1)&1], 3 - M[min1], pCanonPerm, pCanonPhase);
        } else {
            min2 = minTemp3_fast_iVar5(pInOut, DifStart0, DifStartMin, 3-M[0], 3-M[1], &DifStart1);  // reuse DifStart1 because DifStart1 = DifStart1=0
            if DifStart1>DifStartMin {
                arrangeQuoters_superFast_iVar5(pInOut, temp, DifStart0, M[(min2+1)&1], M[min2], 3 - M[min2], 3 - M[(min2+1)&1], pCanonPerm, pCanonPhase);
            } else {
                arrangeQuoters_superFast_iVar5(pInOut, temp, DifStart0, M[min1], M[(min1+1)&1], 3 - M[(min1+1)&1], 3 - M[min1], pCanonPerm, pCanonPhase); }
        }
    }
}

func minimalSwapAndFlipIVar_superFast_iVar5_noEBFC(pInOut *uint32, nWords int32, pCanonPerm string, pCanonPhase *uint32) {
    var DifStart1 int32;
    var temp [2048]uint32;
    if minTemp1_fast_iVar5(pInOut, nWords, &DifStart1) == 2 {
        arrangeQuoters_superFast_iVar5(pInOut, temp, DifStart1, 0, 2, 1, 3, pCanonPerm, pCanonPhase); } 
}

////////////////////////////////////moreThen5/////////////////////////////////////////////////////////////////////////////////////////////

// It rearranges InOut (swaps and flips through rearrangement of quoters)
// It updates Info at the end
func arrangeQuoters_superFast_moreThen5(pInOut *word, temp *word, start int32, iQ int32, jQ int32, kQ int32, lQ int32, iVar int32, pCanonPerm string, pCanonPhase *uint32) {
    var i, wordBlock, blockSize, shiftSize int32;
    var tempPtr *word  = temp+start;
//     printf("in arrangeQuoters_superFast_moreThen5\n");

    if iQ == 0 && jQ == 1 {
        return; }
    wordBlock = (1<<(iVar-6));
    blockSize = wordBlock*sizeof(word);
    shiftSize = wordBlock*4;
    for i=start-wordBlock;i>0;i-=shiftSize {       
        tempPtr -= wordBlock;       
        memcpy(tempPtr, pInOut+i-iQ*wordBlock, (size_t)blockSize);
        tempPtr -= wordBlock;
        memcpy(tempPtr, pInOut+i-jQ*wordBlock, (size_t)blockSize);
        tempPtr -= wordBlock;
        memcpy(tempPtr, pInOut+i-kQ*wordBlock, (size_t)blockSize);
        tempPtr -= wordBlock;
        memcpy(tempPtr, pInOut+i-lQ*wordBlock, (size_t)blockSize);      
    }   
    memcpy(pInOut, temp, start*sizeof(word));
    updataInfo(iQ, jQ, iVar, pCanonPerm, pCanonPhase);
//    printf("out arrangeQuoters_superFast_moreThen5\n");

}

//It compares 0Q and 3Q and returns 0 if 0Q is smaller then 3Q ( comparison starts at highest bit) and visa versa
// DifStart contains the information about the first different bit in 0Q and 3Q
func minTemp0_fast_moreThen5(pInOut *word, iVar int32, nWords int32, pDifStart *int32) int32 {
    var i, j, temp int32;
    var wordBlock int32  = 1<<(iVar-6);
    var wordDif int32  = 3*wordBlock;
    var shiftBlock int32  = wordBlock*4;
//    printf("in minTemp0_fast_moreThen5\n");

    for i=nWords - 1; i>=0; i-=shiftBlock {
        for j=0;j<wordBlock;j++ {
            temp = CompareWords(pInOut[i-j],pInOut[i-j-wordDif]);
            if temp == 0 {
                continue;
            } else if temp == -1 {
                *pDifStart = i+1;
                return 0;
            
            } else {
                *pDifStart = i+1;
                return 3;
            }
        }
    }
    *pDifStart=0;
//    printf("out minTemp0_fast_moreThen5\n");

    return 0;
}

//It compares 1Q and 2Q and returns 1 if 1Q is smaller then 2Q ( comparison starts at highest bit) and visa versa
// DifStart contains the information about the first different bit in 1Q and 2Q
func minTemp1_fast_moreThen5(pInOut *word, iVar int32, nWords int32, pDifStart *int32) int32 {
    var i, j, temp int32;
    var wordBlock int32  = 1<<(iVar-6);
    var shiftBlock int32  = wordBlock*4;
//    printf("in minTemp1_fast_moreThen5\n");

    for i=nWords - wordBlock - 1; i>=0; i-=shiftBlock {
        for j=0;j<wordBlock;j++ {
            temp = CompareWords(pInOut[i-j],pInOut[i-j-wordBlock]);
            if temp == 0 {
                continue;
            } else if temp == -1 {
                *pDifStart = i+wordBlock+1;
                return 1;
            
            } else {
                *pDifStart = i+wordBlock+1;
                return 2;
            }
        }
    }
    *pDifStart=0;
//    printf("out minTemp1_fast_moreThen5\n");

    return 1;
}

//It compares iQ and jQ and returns 0 if iQ is smaller then jQ ( comparison starts at highest bit) and visa versa
// DifStart contains the information about the first different bit in iQ and jQ
func minTemp2_fast_moreThen5(pInOut *word, iVar int32, iQ int32, jQ int32, nWords int32, pDifStart *int32) int32 {
    var i, j, temp int32;
    var wordBlock int32  = 1<<(iVar-6);
    var shiftBlock int32  = wordBlock*4;
//    printf("in minTemp2_fast_moreThen5\n");

    for i=nWords - 1; i>=0; i-=shiftBlock {
        for j=0;j<wordBlock;j++ {
            temp = CompareWords(pInOut[i-j-iQ*wordBlock],pInOut[i-j-jQ*wordBlock]);
            if temp == 0 {
                continue;
            } else if temp == -1 {
                *pDifStart = i+1;
                return 0;
            
            } else {
                *pDifStart = i+1;
                return 1;
            }
        }
    }
    *pDifStart=0;
//    printf("out minTemp2_fast_moreThen5\n");
    
    return 0;
}

// same as minTemp2_fast but this one has a start position
func minTemp3_fast_moreThen5(pInOut *word, iVar int32, start int32, finish int32, iQ int32, jQ int32, pDifStart *int32) int32 {
    var i, j, temp int32;
    var wordBlock int32  = 1<<(iVar-6);
    var shiftBlock int32  = wordBlock*4;
//    printf("in minTemp3_fast_moreThen5\n");

    for i=start-1; i>=finish; i-=shiftBlock {
        for j=0;j<wordBlock;j++ {
            temp = CompareWords(pInOut[i-j-iQ*wordBlock],pInOut[i-j-jQ*wordBlock]);
            if temp == 0 {
                continue;
            } else if temp == -1 {
                *pDifStart = i+1;
                return 0;
            
            } else {
                *pDifStart = i+1;
                return 1;
            }
        }
    }
    *pDifStart=0;
//    printf("out minTemp3_fast_moreThen5\n");

    return 0;
}

// It considers all swap and flip possibilities of iVar and iVar+1 and switches InOut to a minimal of them 
func minimalSwapAndFlipIVar_superFast_moreThen5(pInOut *word, iVar int32, nWords int32, pCanonPerm string, pCanonPhase *uint32) {
    var min1, min2, DifStart0, DifStart1, DifStartMin int32;
    var M [2]int32;
    var temp [1024]word;
//    printf("in minimalSwapAndFlipIVar_superFast_moreThen5\n");
    M[0] = minTemp0_fast_moreThen5(pInOut, iVar, nWords, &DifStart0); // 0, 3
    M[1] = minTemp1_fast_moreThen5(pInOut, iVar, nWords, &DifStart1); // 1, 2
    min1 = minTemp2_fast_moreThen5(pInOut, iVar, M[0], M[1], nWords, &DifStartMin);
    if DifStart0 != DifStart1 {   
        if DifStartMin>=DifStart1 && DifStartMin>=DifStart0 {
            arrangeQuoters_superFast_moreThen5(pInOut, temp, DifStartMin, M[min1], M[(min1+1)&1], 3 - M[(min1+1)&1], 3 - M[min1], iVar, pCanonPerm, pCanonPhase);
        } else if DifStart0 > DifStart1 {
            arrangeQuoters_superFast_moreThen5(pInOut, temp, luckyMax(DifStartMin,DifStart0), M[0], M[1], 3 - M[1], 3 - M[0], iVar, pCanonPerm, pCanonPhase);
        } else {
            arrangeQuoters_superFast_moreThen5(pInOut, temp, luckyMax(DifStartMin,DifStart1), M[1], M[0], 3 - M[0], 3 - M[1], iVar, pCanonPerm, pCanonPhase); }
    
    } else {
        if DifStartMin>=DifStart0 {
            arrangeQuoters_superFast_moreThen5(pInOut, temp, DifStartMin, M[min1], M[(min1+1)&1], 3 - M[(min1+1)&1], 3 - M[min1], iVar, pCanonPerm, pCanonPhase);
        } else {
            min2 = minTemp3_fast_moreThen5(pInOut, iVar, DifStart0, DifStartMin, 3-M[0], 3-M[1], &DifStart1);  // reuse DifStart1 because DifStart1 = DifStart1=0
            if DifStart1>DifStartMin {
                arrangeQuoters_superFast_moreThen5(pInOut, temp, DifStart0, M[(min2+1)&1], M[min2], 3 - M[min2], 3 - M[(min2+1)&1], iVar, pCanonPerm, pCanonPhase);
            } else {
                arrangeQuoters_superFast_moreThen5(pInOut, temp, DifStart0, M[min1], M[(min1+1)&1], 3 - M[(min1+1)&1], 3 - M[min1], iVar, pCanonPerm, pCanonPhase); }
        }
    }
//    printf("out minimalSwapAndFlipIVar_superFast_moreThen5\n");

}

func minimalSwapAndFlipIVar_superFast_moreThen5_noEBFC(pInOut *word, iVar int32, nWords int32, pCanonPerm string, pCanonPhase *uint32) {
    var DifStart1 int32;
    var temp [1024]word;
    if minTemp1_fast_moreThen5(pInOut, iVar, nWords, &DifStart1) == 2 {
        arrangeQuoters_superFast_moreThen5(pInOut, temp, DifStart1, 0, 2, 1, 3, iVar, pCanonPerm, pCanonPhase); } 
}

/////////////////////////////////// for all /////////////////////////////////////////////////////////////////////////////////////////////
func minimalInitialFlip_fast_16Vars(pInOut *word, nVars int32, pCanonPhase *uint32) {
    var oneWord word =1;
    if (pInOut[Kit_TruthWordNum_64bit( nVars ) -1]>>63) & oneWord {
        Kit_TruthNot_64bit( pInOut, nVars );
        (* pCanonPhase) ^=(1<<nVars);       
    }

}

// this function finds minimal for all TIED(and tied only) iVars 
//it finds tied vars based on rearranged  Store info - group of tied vars has the same bit count in Store
func minimalSwapAndFlipIVar_superFast_all(pInOut *word, nVars int32, nWords int32, pStore *int32, pCanonPerm string, pCanonPhase *uint32) int32 {
    var i int32;
    var pDuplicate [1024]word;  
    var bitInfoTemp int32  = pStore[0];
    memcpy(pDuplicate,pInOut,nWords*sizeof(word));
//    printf("in minimalSwapAndFlipIVar_superFast_all\n");
    for i=0;i<5;i++ {
        if bitInfoTemp == pStore[i+1] {
            minimalSwapAndFlipIVar_superFast_lessThen5(pInOut, i, nWords, pCanonPerm, pCanonPhase);
        } else {
            bitInfoTemp = pStore[i+1];
            continue;
        }
    }
    if bitInfoTemp == pStore[i+1] {
        minimalSwapAndFlipIVar_superFast_iVar5((unsigned*) pInOut, nWords, pCanonPerm, pCanonPhase);
    } else {    
        bitInfoTemp = pStore[i+1]; }
    
    for i=6;i<nVars-1;i++ {
        if bitInfoTemp == pStore[i+1] {
            minimalSwapAndFlipIVar_superFast_moreThen5(pInOut, i, nWords, pCanonPerm, pCanonPhase);
        } else {
            bitInfoTemp = pStore[i+1];
            continue;
        }
    }
//    printf("out minimalSwapAndFlipIVar_superFast_all\n");

    if memcmp(pInOut,pDuplicate , nWords*sizeof(word)) == 0 {
        return 0;
    } else {
        return 1; }
}

func minimalSwapAndFlipIVar_superFast_all_noEBFC(pInOut *word, nVars int32, nWords int32, pStore *int32, pCanonPerm string, pCanonPhase *uint32) int32 {
    var i int32;
    var pDuplicate [1024]word;  
    var bitInfoTemp int32  = pStore[0];
    memcpy(pDuplicate,pInOut,nWords*sizeof(word));
    for i=0;i<5;i++ {
        if bitInfoTemp == pStore[i+1] {
            minimalSwapAndFlipIVar_superFast_lessThen5_noEBFC(pInOut, i, nWords, pCanonPerm, pCanonPhase);
        } else {
            bitInfoTemp = pStore[i+1];
            continue;
        }
    }
    if bitInfoTemp == pStore[i+1] {
        minimalSwapAndFlipIVar_superFast_iVar5_noEBFC((unsigned*) pInOut, nWords, pCanonPerm, pCanonPhase);
    } else {    
        bitInfoTemp = pStore[i+1]; }
    
    for i=6;i<nVars-1;i++ {
        if bitInfoTemp == pStore[i+1] {
            minimalSwapAndFlipIVar_superFast_moreThen5_noEBFC(pInOut, i, nWords, pCanonPerm, pCanonPhase);
        } else {
            bitInfoTemp = pStore[i+1];
            continue;
        }
    }
    if memcmp(pInOut,pDuplicate , nWords*sizeof(word)) == 0 {
        return 0;
    } else {
        return 1; }
}


// old version with out noEBFC
// void luckyCanonicizerS_F_first_16Vars(word* pInOut, int  nVars, int nWords, int * pStore, char * pCanonPerm, unsigned* pCanonPhase)
// {
//     while( minimalSwapAndFlipIVar_superFast_all(pInOut, nVars, nWords, pStore, pCanonPerm, pCanonPhase) != 0)
//         continue;
// }

func luckyCanonicizerS_F_first_16Vars1(pInOut *word, nVars int32, nWords int32, pStore *int32, pCanonPerm string, pCanonPhase *uint32) {
    if ((* pCanonPhase) >> (nVars+1)) & 1 {
        for  minimalSwapAndFlipIVar_superFast_all(pInOut, nVars, nWords, pStore, pCanonPerm, pCanonPhase) != 0 {
            continue;
        }
    } else {
        for  minimalSwapAndFlipIVar_superFast_all_noEBFC(pInOut, nVars, nWords, pStore, pCanonPerm, pCanonPhase) != 0 {
            continue;
        } }
}

func luckyCanonicizerS_F_first_16Vars11(pInOut *word, nVars int32, nWords int32, pStore *int32, pCanonPerm string, pCanonPhase *uint32) {
    var duplicate [1024]word;
    var pCanonPerm1 [16]byte;
    var uCanonPhase1 uint32;

    if (* pCanonPhase) >> (nVars+2) {  
        memcpy(duplicate, pInOut, sizeof(word)*nWords);
        Kit_TruthNot_64bit(duplicate, nVars);
        uCanonPhase1 = *pCanonPhase;
        uCanonPhase1 ^= (1 << nVars);
        memcpy(pCanonPerm1,pCanonPerm,sizeof(char)*16);
        luckyCanonicizerS_F_first_16Vars1(pInOut, nVars, nWords, pStore, pCanonPerm, pCanonPhase); 
        luckyCanonicizerS_F_first_16Vars1(duplicate, nVars, nWords, pStore, pCanonPerm1, &uCanonPhase1);
        if memCompare(pInOut, duplicate,nVars) == 1 {
            *pCanonPhase = uCanonPhase1;
            memcpy(pCanonPerm,pCanonPerm1,sizeof(char)*16);
            memcpy(pInOut, duplicate, sizeof(word)*nWords);
        }
    
    } else {
        luckyCanonicizerS_F_first_16Vars1(pInOut, nVars, nWords, pStore, pCanonPerm, pCanonPhase);
    }
}

func luckyCanonicizer_final_fast_16Vars(pInOut *word, nVars int32, nWords int32, pStore *int32, pCanonPerm string, pCanonPhase *uint32) {
    assert(nVars > 6 && nVars <= 16, "nVars > 6 && nVars <= 16");
    (* pCanonPhase) = Kit_TruthSemiCanonicize_Yasha1( pInOut, nVars, pCanonPerm, pStore );
    luckyCanonicizerS_F_first_16Vars1(pInOut, nVars, nWords, pStore, pCanonPerm, pCanonPhase ); 
}

func bitReverceOrder(x *word, nVars int32) {
    var i int32;
    for i= nVars-1;i>=0;i-- {
        Kit_TruthChangePhase_64bit( x, nVars, i );
    }
}


func luckyCanonicizer_final_fast_16Vars1(pInOut *word, nVars int32, nWords int32, pStore *int32, pCanonPerm string, pCanonPhase *uint32) {   
    assert(nVars > 6 && nVars <= 16, "nVars > 6 && nVars <= 16");
    (* pCanonPhase) = Kit_TruthSemiCanonicize_Yasha1( pInOut, nVars, pCanonPerm, pStore );
    luckyCanonicizerS_F_first_16Vars11(pInOut, nVars, nWords, pStore, pCanonPerm, pCanonPhase ); 
    bitReverceOrder(pInOut, nVars);
    (*pCanonPhase) ^= (1<<nVars) -1;
    luckyCanonicizerS_F_first_16Vars11(pInOut, nVars, nWords, pStore, pCanonPerm, pCanonPhase );
//     bitReverceOrder(pInOut, nVars);
//     (*pCanonPhase) ^= (1<<nVars) -1;
//     luckyCanonicizerS_F_first_16Vars11(pInOut, nVars, nWords, pStore, pCanonPerm, pCanonPhase );
}


// top-level procedure calling two special cases (nVars <= 6 and nVars <= 16)
func luckyCanonicizer_final_fast(pInOut *word, nVars int32, pCanonPerm string) uint32 {
    var nWords int32;
    var pStore [16]int32;
    var uCanonPhase uint32;
#ifdef LUCKY_VERIFY
    word temp[1024] = {0};
    word duplicate[1024] = {0};
    Kit_TruthCopy_64bit( duplicate, pInOut, nVars );
#endif
    if nVars <= 6 {
        pInOut[0] = luckyCanonicizer_final_fast_6Vars( pInOut[0], pStore, pCanonPerm, &uCanonPhase);
    } else if nVars <= 16 {
        if nVars <= 6 {
            nWords = 1;
        } else {
            nWords = (1 << (nVars - 6));
        }
        luckyCanonicizer_final_fast_16Vars( pInOut, nVars, nWords, pStore, pCanonPerm, &uCanonPhase );
    
    } else { assert(false, "0"); }
#ifdef LUCKY_VERIFY
    Kit_TruthCopy_64bit( temp, pInOut, nVars );
    assert( ! luckyCheck(temp, duplicate, nVars, pCanonPerm, uCanonPhase) );
#endif
        return uCanonPhase;
}

func luckyCanonicizer_final_fast1(pInOut *word, nVars int32, pCanonPerm string) uint32 {
    var nWords int32;
    var pStore [16]int32;
    var uCanonPhase uint32;
#ifdef LUCKY_VERIFY
    word temp[1024] = {0};
    word duplicate[1024] = {0};
    Kit_TruthCopy_64bit( duplicate, pInOut, nVars );
#endif
    if nVars <= 6 {
        pInOut[0] = luckyCanonicizer_final_fast_6Vars1( pInOut[0], pStore, pCanonPerm, &uCanonPhase);
    } else if nVars <= 16 {
        nWords = 1 << (nVars - 6);
        luckyCanonicizer_final_fast_16Vars1( pInOut, nVars, nWords, pStore, pCanonPerm, &uCanonPhase );
    
    } else { assert(false, "0"); }
#ifdef LUCKY_VERIFY
    Kit_TruthCopy_64bit( temp, pInOut, nVars );
    assert( ! luckyCheck(temp, duplicate, nVars, pCanonPerm, uCanonPhase) );
#endif
    return uCanonPhase;
}


//ABC_NAMESPACE_IMPL_END



