package utils

//===--- sort.h -------------------------------------------------------------===
//
//                     satoko: Satisfiability solver
//
// This file is distributed under the BSD 2-Clause License.
// See LICENSE for details.
//
//===------------------------------------------------------------------------===



//#include "misc/util/abc_global.h"
//ABC_NAMESPACE_HEADER_START

func select_sort(data *any, size uint32, comp_fn func(any, any) int32) {
    var i, j, i_best uint32;
    var temp any;

    for i = 0; i < (size - 1); i++ {
        i_best = i;
        for j = i + 1; j < size; j++ {
            if comp_fn(data[j], data[i_best]) {
                i_best = j; }
        }
        temp = data[i];
        data[i] = data[i_best];
        data[i_best] = temp;
    }
}

func satoko_sort(data *any, size uint32, comp_fn func(any, any) int32) {
    if size <= 15 {
        select_sort(data, size, comp_fn);
    } else {
        var pivot any  = data[size / 2];
        var temp any;
        var j uint32  = size;
        var i int32  = -1;

        for {
            for {
                i++;
            
if !(comp_fn(data[i], pivot)) { break }
} 
            for {
                j--;
            
if !(comp_fn(pivot, data[j])) { break }
} 

            if uint32( i) >= j {
                break; }

            temp = data[i];
            data[i] = data[j];
            data[j] = temp;
        }
        satoko_sort(data, uint32( i), comp_fn);
        satoko_sort(data + i, (size - uint32( i)), comp_fn);
    }
}

//ABC_NAMESPACE_HEADER_END

