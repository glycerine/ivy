package satoko

import (
	"fmt"
	"os"
)

//===--- watch_list.h -------------------------------------------------------===
//
//                     satoko: Satisfiability solver
//
// This file is distributed under the BSD 2-Clause License.
// See LICENSE for details.
//
//===------------------------------------------------------------------------===



//#include "utils/mem.h"
//#include "utils/misc.h"

//#include "misc/util/abc_global.h"
//ABC_NAMESPACE_HEADER_START

type watcher struct {
    cref uint32;
    blocker uint32;
}

type watch_list struct {
    cap uint32;
    size uint32;
    n_bin uint32;
    struct watcher *watchers;
}

   
type vec_wl_t struct {
    cap uint32;
    size uint32;
    struct watch_list *watch_lists;
}

//===------------------------------------------------------------------------===
// Watch list Macros
//===------------------------------------------------------------------------===
// Original macro:
// #define watch_list_foreach(vec, watch, lit) \
//     for (watch = watch_list_array(vec_wl_at(vec, lit)); \
//          watch < watch_list_array(vec_wl_at(vec, lit)) + watch_list_size(vec_wl_at(vec, lit)); \
//      watch++)


// Original macro:
// #define watch_list_foreach_bin(vec, watch, lit) \
//     for (watch = watch_list_array(vec_wl_at(vec, lit)); \
//          watch < watch_list_array(vec_wl_at(vec, lit)) + vec_wl_at(vec, lit)->n_bin; \
//      watch++)

//===------------------------------------------------------------------------===
// Watch list API
//===------------------------------------------------------------------------===
static inline void watch_list_free(struct watch_list *wl)
{
    if wl.watchers {
        // inlining of macro satoko_free
        for { free(wl.watchers); wl.watchers = nil; 
if !(false) { break }
}  }
}

static inline unsigned watch_list_size(struct watch_list *wl)
{
    return wl.size;
}

static inline void watch_list_shrink(struct watch_list *wl, unsigned size)
{
    assert(size <= wl.size, "size <= wl.size");
    wl.size = size;
}

static inline void watch_list_grow(struct watch_list *wl)
{
    var new_size uint32;
    if wl.cap < 4 {
        new_size = 4;
    } else {
        new_size = (wl.cap / 2) * 3;
    }
    struct watcher *watchers =
        // inlining of macro satoko_realloc
        ((struct watcher *) realloc(wl.watchers, (new_size) * sizeof(struct watcher)));
    if watchers == nil {
        fmt.Printf("Failed to realloc memory from %.1f MB to %.1f "
                       "MB.\n", 1.0 * wl.cap / (1 << 20), 1.0 * new_size / (1 << 20));
        fflush(os.Stdout);
        return;
    }
    wl.watchers = watchers;
    wl.cap = new_size;
}

static inline void watch_list_push(struct watch_list *wl, struct watcher w, unsigned is_bin)
{
    assert(wl, "wl");
    if wl.size == wl.cap {
        watch_list_grow(wl); }
    wl.watchers[wl.size++] = w;
    if is_bin && wl.size > wl.n_bin {
        // inlining of macro stk_swap
        { struct watcher t = wl.watchers[wl.n_bin]; wl.watchers[wl.n_bin] = wl.watchers[wl.size - 1]; wl.watchers[wl.size - 1] = t; };
        wl.n_bin++;
    }
}

static inline struct watcher *watch_list_array(struct watch_list *wl)
{
    return wl.watchers;
}

/* TODO: I still have mixed feelings if this change should be done, keeping the
 * old code commented after it. */
static inline void watch_list_remove(struct watch_list *wl, unsigned cref, unsigned is_bin)
{
    struct watcher *watchers = watch_list_array(wl);
    var i uint32;
    if is_bin {
        for i = 0; watchers[i].cref != cref; i++ {}
        assert(i < watch_list_size(wl), "i < watch_list_size(wl)");
        wl.n_bin--;
        memmove((wl.watchers + i), (wl.watchers + i + 1),
                        (wl.size - i - 1) * sizeof(struct watcher));
    } else {
        for i = wl.n_bin; watchers[i].cref != cref; i++ {}
        assert(i < watch_list_size(wl), "i < watch_list_size(wl)");
        // inlining of macro stk_swap
        { struct watcher t = wl.watchers[i]; wl.watchers[i] = wl.watchers[wl.size - 1]; wl.watchers[wl.size - 1] = t; };
    }
    wl.size -= 1;
}

/*
static inline void watch_list_remove(struct watch_list *wl, unsigned cref, unsigned is_bin)
{
    struct watcher *watchers = watch_list_array(wl);
    unsigned i;
    if (is_bin) {
        for (i = 0; watchers[i].cref != cref; i++);
        wl->n_bin--;
    } else
        for (i = wl->n_bin; watchers[i].cref != cref; i++);
    assert(i < watch_list_size(wl));
    memmove((wl->watchers + i), (wl->watchers + i + 1),
            (wl->size - i - 1) * sizeof(struct watcher));
    wl->size -= 1;
}
*/

func vec_wl_alloc(cap uint32) *vec_wl_t {
    // inlining of macro satoko_alloc
    var vec_wl *vec_wl_t  = ((vec_wl_t *) malloc((1) * sizeof(vec_wl_t)));

    if cap == 0 {
        vec_wl.cap = 4;
    } else {
        vec_wl.cap = cap; }
    vec_wl.size = 0;
    // inlining of macro satoko_calloc
    vec_wl.watch_lists = ((struct watch_list *) calloc((sizeof(struct watch_list) * vec_wl.cap), sizeof(struct watch_list)));
    return vec_wl;
}

func vec_wl_free(vec_wl *vec_wl_t) {
    var i uint32;
    for i = 0; i < vec_wl.cap; i++ {
        watch_list_free(vec_wl.watch_lists + i);
    }
    // inlining of macro satoko_free
    for { free(vec_wl.watch_lists); vec_wl.watch_lists = nil; 
if !(false) { break }
} 
    // inlining of macro satoko_free
    for { free(vec_wl); vec_wl = nil; 
if !(false) { break }
} 
}

func vec_wl_clean(vec_wl *vec_wl_t) {
    var i uint32;
    for i = 0; i < vec_wl.size; i++ {
        vec_wl.watch_lists[i].size = 0;
        vec_wl.watch_lists[i].n_bin = 0;
    }
    vec_wl.size = 0;
}

func vec_wl_push(vec_wl *vec_wl_t) {
    if vec_wl.size == vec_wl.cap {
        var new_size uint32;
        if vec_wl.cap < 4 {
            new_size = vec_wl.cap * 2;
        } else {
            new_size = (vec_wl.cap / 2) * 3;
        }

        // inlining of macro satoko_realloc
        vec_wl.watch_lists = ((struct watch_list *) realloc(vec_wl.watch_lists, (new_size) * sizeof(struct watch_list)));
        memset(vec_wl.watch_lists + vec_wl.cap, 0,
                       sizeof(struct watch_list) * (new_size - vec_wl.cap));
        if vec_wl.watch_lists == nil {
            fmt.Printf("failed to realloc memory from %.1f mb to %.1f "
                               "mb.\n", 1.0 * vec_wl.cap / (1 << 20), 1.0 * new_size / (1 << 20));
            fflush(os.Stdout);
        }
        vec_wl.cap = new_size;
    }
    vec_wl.size++;
}

func vec_wl_at(vec_wl *vec_wl_t, idx uint32) *struct watch_list {
    assert(idx < vec_wl.cap, "idx < vec_wl.cap");
    assert(idx < vec_wl.size, "idx < vec_wl.size");
    return vec_wl.watch_lists + idx;
}

//ABC_NAMESPACE_HEADER_END

