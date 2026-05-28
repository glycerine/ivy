#
# Copyright (c) Microsoft Corporation. All Rights Reserved.
#
"""Small safety patches for PLY parser-table caching.

Python Ivy reloads its parser modules when a source file changes the Ivy
language version. PLY's default table writer truncates the shared parsetab file
before rewriting it, so an interrupted or concurrent process can leave a
syntactically invalid table behind. PLY also lets SyntaxError from such a table
escape instead of treating it as a cache miss.

These patches preserve PLY's behavior while making writes atomic and making a
corrupt generated table rebuildable.
"""

import os
import sys
import types
import uuid


def runtime_tabmodule(name):
    """Return a deliberately missing table-module name for in-memory parsers.

    PLY always tries to read tabmodule before building a parser. Passing one of
    these names with write_tables=False makes that read fail as a cache miss and
    then keeps the rebuilt tables in memory only.
    """
    return "_ivy_runtime_" + name


def install(yacc):
    """Install idempotent safety patches on a ply.yacc module."""
    if getattr(yacc, "_ivy_ply_patch_installed", False):
        return

    original_read_table = yacc.LRTable.read_table
    original_write_table = yacc.LRGeneratedTable.write_table

    def read_table_rebuilding_corrupt(self, module):
        try:
            return original_read_table(self, module)
        except SyntaxError as err:
            raise ImportError("corrupt PLY parser table %s: %s" % (module, err))

    def write_table_atomically(self, tabmodule, outputdir='', signature=''):
        if isinstance(tabmodule, types.ModuleType):
            return original_write_table(self, tabmodule, outputdir, signature)

        basemodulename = tabmodule.split('.')[-1]
        filename = os.path.join(outputdir, basemodulename) + '.py'
        tmp_basename = "%s_%s_%s_tmp" % (
            basemodulename,
            os.getpid(),
            uuid.uuid4().hex,
        )
        tmp_filename = os.path.join(outputdir, tmp_basename) + '.py'

        try:
            original_write_table(self, tmp_basename, outputdir, signature)
            replace = getattr(os, "replace", os.rename)
            replace(tmp_filename, filename)
        finally:
            if os.path.exists(tmp_filename):
                os.unlink(tmp_filename)

        if tabmodule in sys.modules:
            del sys.modules[tabmodule]

    yacc.LRTable.read_table = read_table_rebuilding_corrupt
    yacc.LRGeneratedTable.write_table = write_table_atomically
    yacc._ivy_ply_patch_installed = True
