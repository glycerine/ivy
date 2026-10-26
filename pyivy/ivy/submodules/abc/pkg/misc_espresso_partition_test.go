package abc

// Go-only graph-model and fuzz tests for src/misc/espresso/part.c.

import (
	"errors"
	"slices"
	"testing"
)

func mvcCloseBlockPartition(t testing.TB, result MvcBlockPartition) {
	t.Helper()
	if result.Left != nil {
		if err := result.Left.Close(); err != nil {
			t.Fatal(err)
		}
	}
	if result.Right != nil {
		if err := result.Right.Close(); err != nil {
			t.Fatal(err)
		}
	}
}

func mvcModelBlockPartition(columns int, rows []uint64) (bool, []int, []int, []int, []int) {
	if len(rows) == 0 {
		return false, nil, nil, nil, nil
	}
	rowSeen := make([]bool, len(rows))
	columnSeen := make([]bool, columns)
	rowSeen[0] = true
	changed := true
	for changed {
		changed = false
		for row, seen := range rowSeen {
			if !seen {
				continue
			}
			for column := 0; column < columns; column++ {
				if rows[row]>>uint(column)&1 != 0 && !columnSeen[column] {
					columnSeen[column] = true
					changed = true
				}
			}
		}
		for row, seen := range rowSeen {
			if seen {
				continue
			}
			for column := 0; column < columns; column++ {
				if columnSeen[column] && rows[row]>>uint(column)&1 != 0 {
					rowSeen[row] = true
					changed = true
					break
				}
			}
		}
	}
	leftRows, rightRows := make([]int, 0), make([]int, 0)
	for row, seen := range rowSeen {
		if seen {
			leftRows = append(leftRows, row)
		} else {
			rightRows = append(rightRows, row)
		}
	}
	leftColumns, rightColumns := make([]int, 0), make([]int, 0)
	allColumnsSeen := true
	for column := 0; column < columns; column++ {
		present := false
		for _, row := range rows {
			present = present || row>>uint(column)&1 != 0
		}
		if !present {
			continue
		}
		if columnSeen[column] {
			leftColumns = append(leftColumns, column)
		} else {
			rightColumns = append(rightColumns, column)
			allColumnsSeen = false
		}
	}
	partitioned := len(rightRows) != 0 && (!allColumnsSeen || len(leftColumns)+len(rightColumns) == 0)
	if !partitioned {
		return false, nil, nil, nil, nil
	}
	return true, leftRows, rightRows, leftColumns, rightColumns
}

func TestMvcBlockPartitionExactComponentsAndOwnership(t *testing.T) {
	rows := []uint64{0b0011, 0b0010, 0b1100, 0b1000}
	manager, cover := mvcMinimumCoverFixture(t, 4, rows)
	before := mvcDominanceMasks(t, cover)
	result, err := cover.BlockPartition(MvcBlockPartitionOptions{})
	if err != nil || !result.Partitioned || !slices.Equal(result.LeftRows, []int{0, 1}) || !slices.Equal(result.RightRows, []int{2, 3}) || !slices.Equal(result.LeftColumns, []int{0, 1}) || !slices.Equal(result.RightColumns, []int{2, 3}) {
		t.Fatal(result, err)
	}
	if !slices.Equal(mvcDominanceMasks(t, result.Left), rows[:2]) || !slices.Equal(mvcDominanceMasks(t, result.Right), rows[2:]) {
		t.Fatal(mvcDominanceMasks(t, result.Left), mvcDominanceMasks(t, result.Right))
	}
	result.Left.cubes[0].SetBit(3, true)
	if !slices.Equal(mvcDominanceMasks(t, cover), before) {
		t.Fatal("partition aliases input")
	}
	mvcCloseBlockPartition(t, result)
	mvcCloseMinimumFixture(t, manager, cover, nil)
}

func TestMvcBlockPartitionConnectedEmptyAndSourceEarlyExit(t *testing.T) {
	for _, test := range []struct {
		columns   int
		rows      []uint64
		partition bool
		left      []int
		right     []int
	}{
		{3, nil, false, nil, nil},
		{3, []uint64{0b011, 0b110}, false, nil, nil},
		{2, []uint64{0b11, 0}, false, nil, nil},
		{2, []uint64{0, 0b01}, true, []int{0}, []int{1}},
		{4, []uint64{0b0001, 0b0010, 0b0100, 0b1000}, true, []int{0}, []int{1, 2, 3}},
	} {
		manager, cover := mvcMinimumCoverFixture(t, test.columns, test.rows)
		result, err := cover.BlockPartition(MvcBlockPartitionOptions{})
		if err != nil || result.Partitioned != test.partition || !slices.Equal(result.LeftRows, test.left) || !slices.Equal(result.RightRows, test.right) {
			t.Fatal(test, result, err)
		}
		if !test.partition && (result.Left != nil || result.Right != nil) {
			t.Fatal("connected matrix allocated partitions")
		}
		mvcCloseBlockPartition(t, result)
		mvcCloseMinimumFixture(t, manager, cover, nil)
	}
}

func TestMvcBlockPartitionLimitsCancellationAndDepth(t *testing.T) {
	manager, cover := mvcMinimumCoverFixture(t, 3, []uint64{0b011, 0b110, 0b001})
	for _, options := range []MvcBlockPartitionOptions{
		{MaxRows: -1}, {MaxColumns: -1}, {MaxEntries: -1}, {MaxVisits: -1},
		{MaxRows: 2}, {MaxColumns: 2}, {MaxEntries: 4}, {MaxVisits: 1},
		{Stop: func() bool { return true }},
	} {
		result, err := cover.BlockPartition(options)
		if result.Left != nil || result.Right != nil || !(errors.Is(err, ErrDimension) || errors.Is(err, ErrLimit) || errors.Is(err, ErrCancelled)) {
			t.Fatal(options, result, err)
		}
	}
	mvcCloseMinimumFixture(t, manager, cover, nil)
	if result, err := (*MvcCover)(nil).BlockPartition(MvcBlockPartitionOptions{}); result.Left != nil || err == nil {
		t.Fatal(result, err)
	}

	manager = NewMvcManager()
	deep, err := NewMvcCover(manager, 10002)
	if err != nil {
		t.Fatal(err)
	}
	for row := 0; row < 10001; row++ {
		cube, _ := deep.NewCube()
		cube.SetBit(row, true)
		if row > 0 {
			cube.SetBit(row-1, true)
		}
		deep.AddTail(cube)
	}
	last, _ := deep.NewCube()
	last.SetBit(10001, true)
	deep.AddTail(last)
	result, err := deep.BlockPartition(MvcBlockPartitionOptions{MaxEntries: 30000, MaxVisits: 30000})
	if err != nil || !result.Partitioned || len(result.LeftRows) != 10001 || len(result.RightRows) != 1 {
		t.Fatal(len(result.LeftRows), len(result.RightRows), err)
	}
	mvcCloseBlockPartition(t, result)
	mvcCloseMinimumFixture(t, manager, deep, nil)
}

func FuzzMvcBlockPartition(f *testing.F) {
	f.Add(uint8(5), []byte{3, 6, 24, 16})
	f.Fuzz(func(t *testing.T, widthByte uint8, data []byte) {
		columns := 1 + int(widthByte%8)
		if len(data) > 12 {
			data = data[:12]
		}
		limit := byte((1 << uint(columns)) - 1)
		rows := make([]uint64, len(data))
		for index, value := range data {
			rows[index] = uint64(value & limit)
		}
		wantPartition, wantLeftRows, wantRightRows, wantLeftColumns, wantRightColumns := mvcModelBlockPartition(columns, rows)
		manager, cover := mvcMinimumCoverFixture(t, columns, rows)
		result, err := cover.BlockPartition(MvcBlockPartitionOptions{MaxEntries: 1000, MaxVisits: 1000})
		if err != nil || result.Partitioned != wantPartition {
			t.Fatal(result, wantPartition, err)
		}
		if wantPartition && (!slices.Equal(result.LeftRows, wantLeftRows) || !slices.Equal(result.RightRows, wantRightRows) || !slices.Equal(result.LeftColumns, wantLeftColumns) || !slices.Equal(result.RightColumns, wantRightColumns)) {
			t.Fatalf("got rows %v/%v columns %v/%v; want %v/%v %v/%v", result.LeftRows, result.RightRows, result.LeftColumns, result.RightColumns, wantLeftRows, wantRightRows, wantLeftColumns, wantRightColumns)
		}
		if wantPartition {
			left := make([]uint64, len(wantLeftRows))
			for index, row := range wantLeftRows {
				left[index] = rows[row]
			}
			right := make([]uint64, len(wantRightRows))
			for index, row := range wantRightRows {
				right[index] = rows[row]
			}
			if !slices.Equal(mvcDominanceMasks(t, result.Left), left) || !slices.Equal(mvcDominanceMasks(t, result.Right), right) {
				t.Fatal("partition data mismatch")
			}
		}
		mvcCloseBlockPartition(t, result)
		mvcCloseMinimumFixture(t, manager, cover, nil)
	})
}
