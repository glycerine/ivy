package gia

/**CFile****************************************************************

  FileName    [giaMem.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [AIG package.]

  Synopsis    [Memory managers.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - April 28, 2007.]

  Revision    [$Id: giaMem.c,v 1.00 2007/04/28 00:00:00 alanmi Exp $]

***********************************************************************/

import "github.com/glycerine/abcgo/misc/mem"

// GIA and the shared memory package use the same byte-arena algorithms with
// different chunk policies. Keep the original entry points without duplicating
// pools, cached allocation sizes, free lists or pointer arithmetic.
type Gia_MmFixed_t = mem.Mem_Fixed_t
type Gia_MmFlex_t = mem.Mem_Flex_t
type Gia_MmStep_t = mem.Mem_Step_t

func Gia_MmFixedStart(size, entriesMax int32) *Gia_MmFixed_t {
	entries := entriesMax / 8
	if entries < 8 {
		entries = 8
	}
	return mem.Mem_FixedStartWithChunk(size, entries)
}
func Gia_MmFixedStop(p *Gia_MmFixed_t, verbose int32)         { mem.Mem_FixedStop(p, verbose) }
func Gia_MmFixedEntryFetch(p *Gia_MmFixed_t) []byte           { return mem.Mem_FixedEntryFetch(p) }
func Gia_MmFixedEntryRecycle(p *Gia_MmFixed_t, entry []byte)  { mem.Mem_FixedEntryRecycle(p, entry) }
func Gia_MmFixedRestart(p *Gia_MmFixed_t)                     { mem.Mem_FixedRestart(p) }
func Gia_MmFixedReadMemUsage(p *Gia_MmFixed_t) int32          { return mem.Mem_FixedReadMemUsage(p) }
func Gia_MmFixedReadMaxEntriesUsed(p *Gia_MmFixed_t) int32    { return mem.Mem_FixedReadMaxEntriesUsed(p) }
func Gia_MmFlexStart() *Gia_MmFlex_t                          { return mem.Mem_FlexStartWithChunk(1 << 18) }
func Gia_MmFlexStop(p *Gia_MmFlex_t, verbose int32)           { mem.Mem_FlexStop(p, verbose) }
func Gia_MmFlexEntryFetch(p *Gia_MmFlex_t, size int32) []byte { return mem.Mem_FlexEntryFetch(p, size) }
func Gia_MmFlexRestart(p *Gia_MmFlex_t)                       { mem.Mem_FlexRestart(p) }
func Gia_MmFlexReadMemUsage(p *Gia_MmFlex_t) int32            { return mem.Mem_FlexReadMemUsage(p) }
func Gia_MmStepStart(steps int32) *Gia_MmStep_t               { return mem.Mem_StepStartWithChunk(steps, 1024) }
func Gia_MmStepStop(p *Gia_MmStep_t, verbose int32)           { mem.Mem_StepStop(p, verbose) }
func Gia_MmStepEntryFetch(p *Gia_MmStep_t, size int32) []byte { return mem.Mem_StepEntryFetch(p, size) }
func Gia_MmStepEntryRecycle(p *Gia_MmStep_t, entry []byte, size int32) {
	mem.Mem_StepEntryRecycle(p, entry, size)
}
func Gia_MmStepReadMemUsage(p *Gia_MmStep_t) int32 { return mem.Mem_StepReadMemUsage(p) }
