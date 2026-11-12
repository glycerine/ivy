// Mapping conversion and sizing utilities from sclUtil.c.
package scl

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"

	abcmain "github.com/glycerine/abcgo/base/main"
	"github.com/glycerine/abcgo/map/mio"
)

func sclMappedNodes(network TimingNetwork) []int {
	var nodes []int
	for _, id := range network.SclObjectIDs() {
		if network.SclIsNode(id) && !network.SclIsBarrier(id) && len(network.SclFanins(id)) > 0 {
			nodes = append(nodes, id)
		}
	}
	return nodes
}
func ensureSclMapping(lib *SC_Lib, network TimingNetwork) error {
	mapping := network.SclMapping()
	if mapping == nil {
		return fmt.Errorf("SCL: missing network mapping")
	}
	if len(mapping.GateIDs) > 0 {
		if mapping.Lib != lib {
			return fmt.Errorf("SCL: assignment belongs to a different library")
		}
		for _, id := range sclMappedNodes(network) {
			if id >= len(mapping.GateIDs) || mapping.GateIDs[id] < 0 || int(mapping.GateIDs[id]) >= len(lib.Cells) {
				return fmt.Errorf("SCL: object %d has no valid cell", id)
			}
			cell := lib.Cells[mapping.GateIDs[id]]
			if cell.Inputs != len(network.SclFanins(id)) {
				return fmt.Errorf("SCL: object %d fanin count differs from cell %s", id, cell.Name)
			}
		}
		return nil
	}
	return Abc_SclMioGates2SclGates(lib, network)
}
func Abc_SclMioGates2SclGates(lib *SC_Lib, network TimingNetwork) error {
	mapping := network.SclMapping()
	if mapping == nil || lib == nil {
		return fmt.Errorf("SCL: missing library or mapping")
	}
	if len(mapping.GateIDs) > 0 {
		return ensureSclMapping(lib, network)
	}
	maxID := -1
	for _, id := range network.SclObjectIDs() {
		if id < 0 {
			return fmt.Errorf("SCL: negative object ID")
		}
		maxID = max(maxID, id)
	}
	gates := make([]int32, maxID+1)
	for i := range gates {
		gates[i] = -1
	}
	for _, id := range sclMappedNodes(network) {
		if id >= len(mapping.MioGates) || mapping.MioGates[id] == nil {
			return fmt.Errorf("SCL: object %d has no MIO gate", id)
		}
		name := mio.Mio_GateReadName(mapping.MioGates[id])
		cell := Abc_SclCellFind(lib, name)
		if cell < 0 {
			return fmt.Errorf("SCL: cannot find gate %s", name)
		}
		if lib.Cells[cell].Inputs != len(network.SclFanins(id)) {
			return fmt.Errorf("SCL: gate %s fanin mismatch", name)
		}
		gates[id] = int32(cell)
	}
	mapping.Lib, mapping.GateIDs = lib, gates
	return nil
}
func Abc_SclSclGates2MioGates(lib *SC_Lib, network TimingNetwork) error {
	if err := ensureSclMapping(lib, network); err != nil {
		return err
	}
	mapping := network.SclMapping()
	if mapping.Mio == nil {
		return fmt.Errorf("SCL: MIO library unavailable")
	}
	gates := make([]*mio.Mio_Gate_t, len(mapping.GateIDs))
	copy(gates, mapping.MioGates)
	for _, id := range sclMappedNodes(network) {
		cell := lib.Cells[mapping.GateIDs[id]]
		gate := mio.Mio_LibraryReadGateByName(mapping.Mio, cell.Name, "")
		if gate == nil {
			return fmt.Errorf("SCL: MIO gate %s unavailable", cell.Name)
		}
		gates[id] = gate
	}
	mapping.MioGates = gates
	return nil
}

// GateCopyNetwork supplies original pCopy correspondence without guessing IDs.
type GateCopyNetwork interface {
	TimingNetwork
	SclCopyID(int) (int, bool)
}

func Abc_SclTransferGates(old, new TimingNetwork) error {
	correspondence, ok := old.(GateCopyNetwork)
	if !ok {
		return fmt.Errorf("SCL: copied-object correspondence unavailable")
	}
	mapping := old.SclMapping()
	source := new.SclMapping()
	if mapping == nil || source == nil {
		return fmt.Errorf("SCL: missing mapping")
	}
	staged := append([]*mio.Mio_Gate_t(nil), mapping.MioGates...)
	for _, id := range sclMappedNodes(old) {
		copyID, exists := correspondence.SclCopyID(id)
		if !exists {
			continue
		}
		if copyID < 0 || copyID >= len(source.MioGates) || source.MioGates[copyID] == nil {
			return fmt.Errorf("SCL: copied object %d has no gate", copyID)
		}
		for len(staged) <= id {
			staged = append(staged, nil)
		}
		staged[id] = source.MioGates[copyID]
	}
	mapping.MioGates = staged
	return nil
}
func Abc_SclFindMaxAreaCell(representative *SC_Cell) *SC_Cell {
	best := representative
	seen := make(map[*SC_Cell]bool)
	for cell := representative; cell != nil && !seen[cell]; cell = cell.Next {
		seen[cell] = true
		if cell.Area > best.Area {
			best = cell
		}
	}
	return best
}
func Abc_SclFindMinAreas(lib *SC_Lib, useMax int32) []int32 {
	result := make([]int32, len(lib.Cells))
	for i := range result {
		result[i] = -1
	}
	for _, representative := range lib.CellClasses {
		best := representative
		if useMax != 0 {
			best = Abc_SclFindMaxAreaCell(representative)
		}
		seen := make(map[*SC_Cell]bool)
		for cell := representative; cell != nil && !seen[cell]; cell = cell.Next {
			seen[cell] = true
			if cell.ID >= 0 && cell.ID < len(result) {
				result[cell.ID] = int32(best.ID)
			}
		}
	}
	// A singleton without linked class metadata still has a valid prototype.
	for i, cell := range lib.Cells {
		if result[i] < 0 && cell != nil {
			result[i] = int32(i)
		}
	}
	return result
}
func Abc_SclMinsizePerform(lib *SC_Lib, network TimingNetwork, useMax, verbose int32) error {
	if err := ensureSclMapping(lib, network); err != nil {
		return err
	}
	prototypes := Abc_SclFindMinAreas(lib, useMax)
	mapping := network.SclMapping()
	old := append([]int32(nil), mapping.GateIDs...)
	for _, id := range sclMappedNodes(network) {
		mapping.GateIDs[id] = prototypes[mapping.GateIDs[id]]
	}
	if mapping.Mio != nil {
		if err := Abc_SclSclGates2MioGates(lib, network); err != nil {
			mapping.GateIDs = old
			return err
		}
	}
	return nil
}
func Abc_SclCountMinSize(lib *SC_Lib, network TimingNetwork, useMax int32) int {
	prototypes := Abc_SclFindMinAreas(lib, useMax)
	mapping := network.SclMapping()
	count := 0
	for _, id := range sclMappedNodes(network) {
		if id >= len(mapping.GateIDs) {
			continue
		}
		gate := mapping.GateIDs[id]
		if gate >= 0 && int(gate) < len(prototypes) && gate == prototypes[gate] {
			count++
		}
	}
	return count
}
func Abc_SclGetBufInvCount(network TimingNetwork) int {
	count := 0
	for _, id := range sclMappedNodes(network) {
		if len(network.SclFanins(id)) == 1 {
			count++
		}
	}
	return count
}
func Abc_SclGetTotalArea(network TimingNetwork) float32 {
	mapping := network.SclMapping()
	var area float64
	for _, id := range sclMappedNodes(network) {
		if id < len(mapping.GateIDs) && mapping.GateIDs[id] >= 0 {
			area += float64(mapping.Lib.Cells[mapping.GateIDs[id]].Area)
		}
	}
	return float32(area)
}
func Abc_SclGetAverageSize(network TimingNetwork) float32 {
	mapping := network.SclMapping()
	var sum float32
	count := 0
	for _, id := range sclMappedNodes(network) {
		cell := mapping.Lib.Cells[mapping.GateIDs[id]]
		if cell.Gates > 0 {
			sum += 100 * float32(cell.Order) / float32(cell.Gates)
		}
		count++
	}
	if count == 0 {
		return 0
	}
	return sum / float32(count)
}
func Abc_SclManPrintGateSizes(lib *SC_Lib, network TimingNetwork, gates []int32) {
	counts := make(map[int]int)
	areas := make(map[int]float32)
	var total float32
	for _, id := range sclMappedNodes(network) {
		cell := lib.Cells[gates[id]]
		counts[cell.Order]++
		areas[cell.Order] += cell.Area
		total += cell.Area
	}
	fmt.Printf("Total gates = %d. Total area = %.1f\n", len(sclMappedNodes(network)), total)
	for order := 0; order < len(lib.Cells); order++ {
		if counts[order] > 0 {
			fmt.Printf("Cell size = %d. Count = %d Area = %.1f\n", order, counts[order], areas[order])
		}
	}
}
func Abc_SclPrintGateSizes(lib *SC_Lib, network TimingNetwork) error {
	if err := ensureSclMapping(lib, network); err != nil {
		return err
	}
	Abc_SclManPrintGateSizes(lib, network, network.SclMapping().GateIDs)
	return nil
}
func Abc_SclExtractBarBufs(network TimingNetwork) ([]int, error) {
	mapping := network.SclMapping()
	if mapping == nil || mapping.Mio == nil {
		return nil, fmt.Errorf("SCL: mapped library unavailable")
	}
	buffer := mio.Mio_LibraryReadBuf(mapping.Mio)
	if buffer == nil {
		return nil, fmt.Errorf("SCL: mapped buffer cell unavailable")
	}
	var ids []int
	for _, id := range network.SclObjectIDs() {
		if network.SclIsBarrier(id) {
			for len(mapping.MioGates) <= id {
				mapping.MioGates = append(mapping.MioGates, nil)
			}
			if mapping.MioGates[id] != nil {
				return nil, fmt.Errorf("SCL: barrier %d already has a function", id)
			}
			ids = append(ids, id)
		}
	}
	for _, id := range ids {
		mapping.MioGates[id] = buffer
	}
	return ids, nil
}
func Abc_SclInsertBarBufs(network TimingNetwork, ids []int) error {
	mapping := network.SclMapping()
	for _, id := range ids {
		if !network.SclIsBarrier(id) || id < 0 || id >= len(mapping.MioGates) {
			return fmt.Errorf("SCL: invalid barrier ID %d", id)
		}
	}
	for _, id := range ids {
		mapping.MioGates[id] = nil
	}
	return nil
}
func Abc_SclReadTimingConstr(frame *abcmain.Frame, filename string, verbose int32) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	var driving *string
	var load *float32
	line := 0
	for scanner.Scan() {
		line++
		text, _, _ := strings.Cut(scanner.Text(), "#")
		fields := strings.Fields(text)
		if len(fields) == 0 {
			continue
		}
		if len(fields) != 2 {
			return fmt.Errorf("%s:%d: expected command and value", filename, line)
		}
		switch fields[0] {
		case "set_driving_cell":
			value := fields[1]
			driving = &value
		case "set_load":
			number, err := strconv.ParseFloat(fields[1], 32)
			if err != nil || number < 0 || math.IsNaN(number) || math.IsInf(number, 0) {
				return fmt.Errorf("%s:%d: invalid output load", filename, line)
			}
			value := float32(number)
			load = &value
		default:
			return fmt.Errorf("%s:%d: unknown constraint %s", filename, line, fields[0])
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	if driving != nil {
		frame.SetStringValue("pDrivingCell", *driving)
	}
	if load != nil {
		frame.MaxLoad = *load
	}
	return nil
}
