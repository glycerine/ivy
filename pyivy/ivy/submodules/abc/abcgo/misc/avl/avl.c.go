package avl

import "fmt"

/*
 * Revision Control Information
 *
 * $Source: /vol/opua/opua2/sis/sis-1.2/common/src/sis/avl/RCS/avl.c,v $
 * $Author: sis $
 * $Revision: 1.3 $
 * $Date: 1994/07/15 23:00:40 $
 *
 */
/* LINTLIBRARY */


//#include <stdio.h>
//#include <stdlib.h>

//#include "avl.h"

//ABC_NAMESPACE_IMPL_START




// Original macro:
// #define HEIGHT(node) (node == NIL(avl_node) ? -1 : (node)->height)

// Original macro:
// #define BALANCE(node) (HEIGHT((node)->right) - HEIGHT((node)->left))


// Original macro:
// #define compute_height(node) {                \
//     int x=HEIGHT(node->left), y=HEIGHT(node->right);    \
//     (node)->height = MAX(x,y) + 1;            \
// }


// Original macro:
// #define COMPARE(key, nodekey, compare)             \
//     ((compare == avl_numcmp) ?                 \
//     (int) key - (int) nodekey :             \
//     (*compare)(key, nodekey))



const STACK_SIZE =    50

// static avl_node *new_node ();

// static avl_node *find_rightmost ();

// static void do_rebalance ();

static rotate_left ();
static rotate_right ();
// static int do_check_tree ();


avl_tree *
avl_init_table (compar)
 int (*compar) ();
{
    avl_tree *tree;

    // inlining of macro ALLOC
    tree = ((avl_tree *) malloc(sizeof(avl_tree) * (1)));
    // inlining of macro NIL
    tree.root = ((avl_node *) 0);
    tree.compar = compar;
    tree.num_entries = 0;
    return tree;
}



avl_lookup (tree, key, value_p)
 var tree *avl_tree;
 register char *key;
 var value_p *string;
{
    register avl_node *node;
    register int (*compare) () = tree.compar, diff;

    node = tree.root;
    // inlining of macro NIL
    while (node != ((avl_node *) 0)) {
        // inlining of macro COMPARE
        diff = ((compare == avl_numcmp) ?                     int32( key) - int32( node.key) :                 (*compare)(key, node.key));
        if diff == 0 {
            /* got a match */
            // inlining of macro NIL
            if (value_p != ((char * *) 0))
                 *value_p = node.value;
            return 1;
        }
        node = (diff < 0) ? node.left : node.right;
    }
    return 0;
}

avl_first (tree, key_p, value_p)
 var tree *avl_tree;
 var key_p *string;
 var value_p *string;
{
    register avl_node *node;

    if tree.root == 0 {
        return 0;                /* no entries */
    }
    else {
        /* walk down the tree; stop at leftmost leaf */
        for node = tree.root; node.left != 0; node = node.left {
        }
        // inlining of macro NIL
        if (key_p != ((char * *) 0))
             *key_p = node.key;
        // inlining of macro NIL
        if (value_p != ((char * *) 0))
             *value_p = node.value;
        return 1;
    }
}


avl_last (tree, key_p, value_p)
 var tree *avl_tree;
 var key_p *string;
 var value_p *string;
{
    register avl_node *node;

    if tree.root == 0 {
        return 0;                /* no entries */
    }
    else {
        /* walk down the tree; stop at rightmost leaf */
        for node = tree.root; node.right != 0; node = node.right {
        }
        // inlining of macro NIL
        if (key_p != ((char * *) 0))
             *key_p = node.key;
        // inlining of macro NIL
        if (value_p != ((char * *) 0))
             *value_p = node.value;
        return 1;
    }
}

avl_insert (tree, key, value)
 var tree *avl_tree;
 var key string;
 var value string;
{
    register avl_node **node_p, *node;
    register int stack_n = 0;
    register int (*compare) () = tree.compar;
    avl_node **stack_nodep[STACK_SIZE];
    int diff, status;

    node_p = &tree.root;

    /* walk down the tree (saving the path); stop at insertion point */
    status = 0;
    // inlining of macro NIL
    while ((node = *node_p) != ((avl_node *) 0)) {
        stack_nodep[stack_n++] = node_p;
        // inlining of macro COMPARE
        diff = ((compare == avl_numcmp) ?                     int32( key) - int32( node.key) :                 (*compare)(key, node.key));
        if (diff == 0)
            status = 1;
        node_p = (diff < 0) ? &node.left : &node.right;
    }

    /* insert the item and re-balance the tree */
    *node_p = new_node (key, value);
    do_rebalance (stack_nodep, stack_n);
    tree.num_entries++;
    tree.modified = 1;
    return status;
}



avl_find_or_add (tree, key, slot_p)
 var tree *avl_tree;
 var key string;
 var slot_p **string;
{
    register avl_node **node_p, *node;
    register int stack_n = 0;
    register int (*compare) () = tree.compar;
    avl_node **stack_nodep[STACK_SIZE];
    int diff;

    node_p = &tree.root;

    /* walk down the tree (saving the path); stop at insertion point */
    // inlining of macro NIL
    while ((node = *node_p) != ((avl_node *) 0)) {
        stack_nodep[stack_n++] = node_p;
        // inlining of macro COMPARE
        diff = ((compare == avl_numcmp) ?                     int32( key) - int32( node.key) :                 (*compare)(key, node.key));
        if diff == 0 {
            if (slot_p != 0)
                *slot_p = &node.value;
            return 1;            /* found */
        }
        node_p = (diff < 0) ? &node.left : &node.right;
    }

    /* insert the item and re-balance the tree */
    // inlining of macro NIL
    *node_p = new_node (key, ((char *) 0));
    if (slot_p != 0)
        *slot_p = &(*node_p).value;
    do_rebalance (stack_nodep, stack_n);
    tree.num_entries++;
    tree.modified = 1;
    return 0;                    /* not already in tree */
}

avl_delete (tree, key_p, value_p)
 var tree *avl_tree;
 var key_p *string;
 var value_p *string;
{
    register avl_node **node_p, *node, *rightmost;
    register int stack_n = 0;
    char *key = *key_p;
    int (*compare) () = tree.compar;
diff;
    avl_node **stack_nodep[STACK_SIZE];

    node_p = &tree.root;

    /* Walk down the tree saving the path; return if not found */
    // inlining of macro NIL
    while ((node = *node_p) != ((avl_node *) 0)) {
        // inlining of macro COMPARE
        diff = ((compare == avl_numcmp) ?                     int32( key) - int32( node.key) :                 (*compare)(key, node.key));
        if (diff == 0)
            goto delete_item;
        stack_nodep[stack_n++] = node_p;
        node_p = (diff < 0) ? &node.left : &node.right;
    }
    return 0;                    /* not found */

    /* prepare to delete node and replace it with rightmost of left tree */
  delete_item:
    *key_p = node.key;
    if (value_p != 0)
        *value_p = node.value;
    // inlining of macro NIL
    if node.left == ((avl_node *) 0) {
        *node_p = node.right;
    }
    else {
        rightmost = find_rightmost (&node.left);
        rightmost.left = node.left;
        rightmost.right = node.right;
        rightmost.height = -2;    /* mark bogus height for do_rebal */
        *node_p = rightmost;
        stack_nodep[stack_n++] = node_p;
    }
    // inlining of macro FREE
    free((char *) (node));

    /* work our way back up, re-balancing the tree */
    do_rebalance (stack_nodep, stack_n);
    tree.num_entries--;
    tree.modified = 1;
    return 1;
}

static void
avl_record_gen_forward (node, gen)
 var node *avl_node;
 var gen *avl_generator;
{
    // inlining of macro NIL
    if node != ((avl_node *) 0) {
        avl_record_gen_forward (node.left, gen);
        gen.nodelist[gen.count++] = node;
        avl_record_gen_forward (node.right, gen);
    }
}


static void
avl_record_gen_backward (node, gen)
 var node *avl_node;
 var gen *avl_generator;
{
    // inlining of macro NIL
    if node != ((avl_node *) 0) {
        avl_record_gen_backward (node.right, gen);
        gen.nodelist[gen.count++] = node;
        avl_record_gen_backward (node.left, gen);
    }
}


avl_generator *
avl_init_gen (tree, dir)
 var tree *avl_tree;
 var dir int32;
{
    avl_generator *gen;

    /* what a hack */
    // inlining of macro ALLOC
    gen = ((avl_generator *) malloc(sizeof(avl_generator) * (1)));
    gen.tree = tree;
    // inlining of macro ALLOC
    gen.nodelist = ((avl_node * *) malloc(sizeof(avl_node *) * (avl_count (tree))));
    gen.count = 0;
    if dir == AVL_FORWARD {
        avl_record_gen_forward (tree.root, gen);
    }
    else {
        avl_record_gen_backward (tree.root, gen);
    }
    gen.count = 0;

    /* catch any attempt to modify the tree while we generate */
    tree.modified = 0;
    return gen;
}


avl_gen (gen, key_p, value_p)
 var gen *avl_generator;
 var key_p *string;
 var value_p *string;
{
    avl_node *node;

    if gen.count == gen.tree.num_entries {
        return 0;
    }
    else {
        node = gen.nodelist[gen.count++];
        // inlining of macro NIL
        if (key_p != ((char * *) 0))
             *key_p = node.key;
        // inlining of macro NIL
        if (value_p != ((char * *) 0))
             *value_p = node.value;
        return 1;
    }
}


void
avl_free_gen (gen)
 var gen *avl_generator;
{
    // inlining of macro FREE
    free((char *) (gen.nodelist));
    // inlining of macro FREE
    free((char *) (gen));
}

static avl_node *
find_rightmost (node_p)
 register avl_node **node_p;
{
    register avl_node *node;
    register int stack_n = 0;
    avl_node **stack_nodep[STACK_SIZE];

    node = *node_p;
    // inlining of macro NIL
    while (node.right != ((avl_node *) 0)) {
        stack_nodep[stack_n++] = node_p;
        node_p = &node.right;
        node = *node_p;
    }
    *node_p = node.left;

    do_rebalance (stack_nodep, stack_n);
    return node;
}


static void
do_rebalance (stack_nodep, stack_n)
 register avl_node ***stack_nodep;
 register int stack_n;
{
    register avl_node **node_p, *node;
    register int hl, hr;
    int height;

    /* work our way back up, re-balancing the tree */
    while (--stack_n >= 0) {
        node_p = stack_nodep[stack_n];
        node = *node_p;
        // inlining of macro HEIGHT
        // inlining of macro NIL
        hl = (node.left == ((avl_node *) 0) ? -1 : (node.left).height);    /* watch for NIL */
        // inlining of macro HEIGHT
        // inlining of macro NIL
        hr = (node.right == ((avl_node *) 0) ? -1 : (node.right).height);    /* watch for NIL */
        if (hr - hl) < -1 {
            rotate_right (node_p);
        }
        else if (hr - hl) > 1 {
            rotate_left (node_p);
        }
        else {
            // inlining of macro MAX
            height = ((hl) > (hr) ? (hl) : (hr)) + 1;
            if (height == node.height)
                break;
            node.height = height;
        }
    }
}

static
rotate_left (node_p)
 register avl_node **node_p;
{
    register avl_node *old_root = *node_p, *new_root, *new_right;

    // inlining of macro BALANCE
    // inlining of macro HEIGHT
    // inlining of macro NIL
    if (((old_root.right).right == ((avl_node *) 0) ? -1 : ((old_root.right).right).height) - ((old_root.right).left == ((avl_node *) 0) ? -1 : ((old_root.right).left).height)) >= 0 {
        *node_p = new_root = old_root.right;
        old_root.right = new_root.left;
        new_root.left = old_root;
    }
    else {
        new_right = old_root.right;
        *node_p = new_root = new_right.left;
        old_root.right = new_root.left;
        new_right.left = new_root.right;
        new_root.right = new_right;
        new_root.left = old_root;
        // inlining of macro compute_height
        // inlining of macro HEIGHT
        // inlining of macro NIL
        // inlining of macro MAX
        {                    int x=(new_right.left == ((avl_node *) 0) ? -1 : (new_right.left).height), y=(new_right.right == ((avl_node *) 0) ? -1 : (new_right.right).height);        (new_right).height = ((x) > (y) ? (x) : (y)) + 1;            };
    }
    // inlining of macro compute_height
    // inlining of macro HEIGHT
    // inlining of macro NIL
    // inlining of macro MAX
    {                    int x=(old_root.left == ((avl_node *) 0) ? -1 : (old_root.left).height), y=(old_root.right == ((avl_node *) 0) ? -1 : (old_root.right).height);        (old_root).height = ((x) > (y) ? (x) : (y)) + 1;            };
    // inlining of macro compute_height
    // inlining of macro HEIGHT
    // inlining of macro NIL
    // inlining of macro MAX
    {                    int x=(new_root.left == ((avl_node *) 0) ? -1 : (new_root.left).height), y=(new_root.right == ((avl_node *) 0) ? -1 : (new_root.right).height);        (new_root).height = ((x) > (y) ? (x) : (y)) + 1;            };
}


static
rotate_right (node_p)
 var node_p **avl_node;
{
    register avl_node *old_root = *node_p, *new_root, *new_left;

    // inlining of macro BALANCE
    // inlining of macro HEIGHT
    // inlining of macro NIL
    if (((old_root.left).right == ((avl_node *) 0) ? -1 : ((old_root.left).right).height) - ((old_root.left).left == ((avl_node *) 0) ? -1 : ((old_root.left).left).height)) <= 0 {
        *node_p = new_root = old_root.left;
        old_root.left = new_root.right;
        new_root.right = old_root;
    }
    else {
        new_left = old_root.left;
        *node_p = new_root = new_left.right;
        old_root.left = new_root.right;
        new_left.right = new_root.left;
        new_root.left = new_left;
        new_root.right = old_root;
        // inlining of macro compute_height
        // inlining of macro HEIGHT
        // inlining of macro NIL
        // inlining of macro MAX
        {                    int x=(new_left.left == ((avl_node *) 0) ? -1 : (new_left.left).height), y=(new_left.right == ((avl_node *) 0) ? -1 : (new_left.right).height);        (new_left).height = ((x) > (y) ? (x) : (y)) + 1;            };
    }
    // inlining of macro compute_height
    // inlining of macro HEIGHT
    // inlining of macro NIL
    // inlining of macro MAX
    {                    int x=(old_root.left == ((avl_node *) 0) ? -1 : (old_root.left).height), y=(old_root.right == ((avl_node *) 0) ? -1 : (old_root.right).height);        (old_root).height = ((x) > (y) ? (x) : (y)) + 1;            };
    // inlining of macro compute_height
    // inlining of macro HEIGHT
    // inlining of macro NIL
    // inlining of macro MAX
    {                    int x=(new_root.left == ((avl_node *) 0) ? -1 : (new_root.left).height), y=(new_root.right == ((avl_node *) 0) ? -1 : (new_root.right).height);        (new_root).height = ((x) > (y) ? (x) : (y)) + 1;            };
}

static void
avl_walk_forward (node, func)
 var node *avl_node;
 void (*func) ();
{
    // inlining of macro NIL
    if node != ((avl_node *) 0) {
        avl_walk_forward (node.left, func);
        (*func) (node.key, node.value);
        avl_walk_forward (node.right, func);
    }
}


static void
avl_walk_backward (node, func)
 var node *avl_node;
 void (*func) ();
{
    // inlining of macro NIL
    if node != ((avl_node *) 0) {
        avl_walk_backward (node.right, func);
        (*func) (node.key, node.value);
        avl_walk_backward (node.left, func);
    }
}


void
avl_foreach (tree, func, direction)
 var tree *avl_tree;
 void (*func) ();
 var direction int32;
{
    if direction == AVL_FORWARD {
        avl_walk_forward (tree.root, func);
    }
    else {
        avl_walk_backward (tree.root, func);
    }
}


static void
free_entry (node, key_free, value_free)
 var node *avl_node;
 void (*key_free) ();
 void (*value_free) ();
{
    // inlining of macro NIL
    if node != ((avl_node *) 0) {
        free_entry (node.left, key_free, value_free);
        free_entry (node.right, key_free, value_free);
        if (key_free != 0)
            (*key_free) (node.key);
        if (value_free != 0)
            (*value_free) (node.value);
        // inlining of macro FREE
        free((char *) (node));
    }
}


void
avl_free_table (tree, key_free, value_free)
 var tree *avl_tree;
 void (*key_free) ();
 void (*value_free) ();
{
    free_entry (tree.root, key_free, value_free);
    // inlining of macro FREE
    free((char *) (tree));
}


int
avl_count (tree)
 var tree *avl_tree;
{
    return tree.num_entries;
}

static avl_node *
new_node (key, value)
 var key string;
 var value string;
{
    register avl_node *new;

    // inlining of macro ALLOC
    new = ((avl_node *) malloc(sizeof(avl_node) * (1)));
    new.key = key;
    new.value = value;
    new.height = 0;
    // inlining of macro NIL
    new.left = new.right = ((avl_node *) 0);
    return new;
}


int
avl_numcmp (x, y)
 var x, y string;
{
    return int32( x) - int32( y);
}

int
avl_check_tree (tree)
 var tree *avl_tree;
{
    int error = 0;
    (void) do_check_tree (tree.root, tree.compar, &error);
    return error;
}


static int
do_check_tree (node, compar, error)
 var node *avl_node;
 int (*compar) ();
 var error *int32;
{
    int l_height, r_height, comp_height, bal;

    // inlining of macro NIL
    if node == ((avl_node *) 0) {
        return -1;
    }

    r_height = do_check_tree (node.right, compar, error);
    l_height = do_check_tree (node.left, compar, error);

    // inlining of macro MAX
    comp_height = ((l_height) > (r_height) ? (l_height) : (r_height)) + 1;
    bal = r_height - l_height;

    if comp_height != node.height {
        (void) fmt.Printf("Bad height for 0x%08x: computed=%d stored=%d\n", node, comp_height, node.height);
        ++*error;
    }

    if bal > 1 || bal < -1 {
        (void) fmt.Printf("Out of balance at node 0x%08x, balance = %d\n", node, bal);
        ++*error;
    }

    // inlining of macro NIL
    if node.left != ((avl_node *) 0) &&
        (*compar) (node.left.key, node.key) > 0 {
        (void) fmt.Printf("Bad ordering between 0x%08x and 0x%08x", node, node.left);
        ++*error;
    }

    // inlining of macro NIL
    if node.right != ((avl_node *) 0) &&
        (*compar) (node.key, node.right.key) > 0 {
        (void) fmt.Printf("Bad ordering between 0x%08x and 0x%08x", node, node.right);
        ++*error;
    }

    return comp_height;
}
//ABC_NAMESPACE_IMPL_END

