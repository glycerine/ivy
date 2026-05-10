// Minimal WASI preview1 host for Go Ivy's TinyGo -target wasm artifact.
//
// TinyGo's wasm_exec.js provides a partial wasi_snapshot_preview1 object, but
// Go Ivy's full TinyGo build also imports wasi-libc file APIs such as
// fd_prestat_get, fd_read, and path_open. Keep this shim separate from
// goivyWasiP1.js: browser callers get the in-memory include tree only, while
// tinynode can opt into an additional Node-backed preopen for oracle testing.
const textEncoder = new TextEncoder();
const textDecoder = new TextDecoder('utf-8', { fatal: false });

const ERRNO_SUCCESS = 0;
const ERRNO_BADF = 8;
const ERRNO_EXIST = 20;
const ERRNO_FAULT = 21;
const ERRNO_INVAL = 28;
const ERRNO_ISDIR = 31;
const ERRNO_IO = 29;
const ERRNO_NAMETOOLONG = 37;
const ERRNO_NOENT = 44;
const ERRNO_NOSYS = 52;
const ERRNO_NOTDIR = 54;
const ERRNO_NOTEMPTY = 55;
const ERRNO_NOTSUP = 58;
const ERRNO_PERM = 63;
const ERRNO_ROFS = 69;
const ERRNO_NOTCAPABLE = 76;

export const GOIVY_WASI_ERRNO = Object.freeze({
  SUCCESS: ERRNO_SUCCESS,
  BADF: ERRNO_BADF,
  EXIST: ERRNO_EXIST,
  FAULT: ERRNO_FAULT,
  INVAL: ERRNO_INVAL,
  ISDIR: ERRNO_ISDIR,
  IO: ERRNO_IO,
  NAMETOOLONG: ERRNO_NAMETOOLONG,
  NOENT: ERRNO_NOENT,
  NOSYS: ERRNO_NOSYS,
  NOTDIR: ERRNO_NOTDIR,
  NOTEMPTY: ERRNO_NOTEMPTY,
  NOTSUP: ERRNO_NOTSUP,
  PERM: ERRNO_PERM,
  ROFS: ERRNO_ROFS,
  NOTCAPABLE: ERRNO_NOTCAPABLE,
});

const CLOCKID_REALTIME = 0;
const CLOCKID_MONOTONIC = 1;
const FILETYPE_CHARACTER_DEVICE = 2;
const FILETYPE_DIRECTORY = 3;
const FILETYPE_REGULAR_FILE = 4;
const WHENCE_SET = 0;
const WHENCE_CUR = 1;
const WHENCE_END = 2;
const OFLAGS_CREAT = 1 << 0;
const OFLAGS_DIRECTORY = 1 << 1;
const OFLAGS_EXCL = 1 << 2;
const OFLAGS_TRUNC = 1 << 3;
const FDFLAGS_APPEND = 1 << 0;
const EVENTTYPE_CLOCK = 0;
const EVENTTYPE_FD_READ = 1;
const EVENTTYPE_FD_WRITE = 2;
const SUBCLOCKFLAGS_SUBSCRIPTION_CLOCK_ABSTIME = 1 << 0;

const RIGHTS_FD_READ = 1n << 1n;
const RIGHTS_FD_SEEK = 1n << 2n;
const RIGHTS_FD_FDSTAT_SET_FLAGS = 1n << 3n;
const RIGHTS_FD_TELL = 1n << 5n;
const RIGHTS_FD_WRITE = 1n << 6n;
const RIGHTS_PATH_OPEN = 1n << 13n;
const RIGHTS_FD_READDIR = 1n << 14n;
const RIGHTS_PATH_FILESTAT_GET = 1n << 18n;
const RIGHTS_FD_FILESTAT_GET = 1n << 21n;
const RIGHTS_PATH_REMOVE_DIRECTORY = 1n << 25n;
const RIGHTS_PATH_UNLINK_FILE = 1n << 26n;

const RIGHTS_STDOUT = RIGHTS_FD_WRITE | RIGHTS_FD_FDSTAT_SET_FLAGS | RIGHTS_FD_FILESTAT_GET;
const RIGHTS_FILE_READONLY = RIGHTS_FD_READ | RIGHTS_FD_SEEK | RIGHTS_FD_TELL |
  RIGHTS_FD_FDSTAT_SET_FLAGS | RIGHTS_FD_FILESTAT_GET;
const RIGHTS_FILE_WRITEONLY = RIGHTS_FD_WRITE | RIGHTS_FD_SEEK | RIGHTS_FD_TELL |
  RIGHTS_FD_FDSTAT_SET_FLAGS | RIGHTS_FD_FILESTAT_GET;
const RIGHTS_FILE_READWRITE = RIGHTS_FILE_READONLY | RIGHTS_FILE_WRITEONLY;
const RIGHTS_DIR_READONLY = RIGHTS_FD_SEEK | RIGHTS_FD_FDSTAT_SET_FLAGS |
  RIGHTS_PATH_OPEN | RIGHTS_FD_READDIR | RIGHTS_PATH_FILESTAT_GET |
  RIGHTS_FD_FILESTAT_GET | RIGHTS_PATH_REMOVE_DIRECTORY | RIGHTS_PATH_UNLINK_FILE;

export class GoIvyTinyGoWasiProcExit extends Error {
  constructor(code) {
    super('exit with exit code ' + code);
    this.code = code | 0;
  }
}

class WasiTrap extends Error {
  constructor(syscall, message) {
    super(syscall + ': ' + message);
    this.syscall = syscall;
  }
}

class VNode {
  constructor(kind, name, parent = null, data = null) {
    this.kind = kind;
    this.name = name;
    this.parent = parent;
    this.ino = VNode.nextIno++;
    this.data = data;
    this.children = new Map();
  }

  get filetype() {
    return this.kind === 'dir' ? FILETYPE_DIRECTORY : FILETYPE_REGULAR_FILE;
  }

  get size() {
    return this.kind === 'file' ? this.data.byteLength : 0;
  }
}

VNode.nextIno = 1n;

class OpenFd {
  constructor(kind, options = {}) {
    this.kind = kind;
    this.node = options.node || null;
    this.pos = 0n;
    this.flags = options.flags || 0;
    this.preopenName = options.preopenName || '';
    this.write = options.write || null;
    this.stdinData = options.stdinData || new Uint8Array(0);
    this.stdinPos = 0;
    this.hostPath = options.hostPath || '';
    this.hostFd = options.hostFd ?? null;
    this.canRead = options.canRead !== false;
    this.canWrite = options.canWrite === true;
  }

  get filetype() {
    if (this.kind === 'stdio') {
      return FILETYPE_CHARACTER_DEVICE;
    }
    if (this.kind === 'host-dir') {
      return FILETYPE_DIRECTORY;
    }
    if (this.kind === 'host-file') {
      return FILETYPE_REGULAR_FILE;
    }
    return this.node ? this.node.filetype : FILETYPE_CHARACTER_DEVICE;
  }

  get readable() {
    return this.kind === 'file' || this.kind === 'dir' || this.kind === 'stdin' ||
      (this.kind === 'host-file' && this.canRead) || this.kind === 'host-dir';
  }

  get writable() {
    return this.kind === 'stdio' || (this.kind === 'host-file' && this.canWrite);
  }

  rightsBase() {
    if (this.kind === 'stdio') {
      return RIGHTS_STDOUT;
    }
    if (this.kind === 'stdin') {
      return RIGHTS_FD_READ | RIGHTS_FD_FDSTAT_SET_FLAGS | RIGHTS_FD_FILESTAT_GET;
    }
    if (this.kind === 'host-dir') {
      return RIGHTS_DIR_READONLY;
    }
    if (this.kind === 'host-file') {
      if (this.canRead && this.canWrite) {
        return RIGHTS_FILE_READWRITE;
      }
      if (this.canWrite) {
        return RIGHTS_FILE_WRITEONLY;
      }
      return RIGHTS_FILE_READONLY;
    }
    if (this.node && this.node.kind === 'dir') {
      return RIGHTS_DIR_READONLY;
    }
    return RIGHTS_FILE_READONLY;
  }
}

function buildIncludeRoot(tree) {
  const root = new VNode('dir', '');

  function childDir(parent, name) {
    let child = parent.children.get(name);
    if (!child) {
      child = new VNode('dir', name, parent);
      parent.children.set(name, child);
    }
    if (child.kind !== 'dir') {
      throw new Error('include tree path component is already a file: ' + name);
    }
    return child;
  }

  for (const file of tree.files || []) {
    const parts = String(file.path || '').split('/').filter(Boolean);
    if (parts.length === 0) {
      continue;
    }
    const filename = parts.pop();
    let dir = root;
    for (const part of parts) {
      dir = childDir(dir, part);
    }
    const bytes = textEncoder.encode(String(file.data || ''));
    dir.children.set(filename, new VNode('file', filename, dir, bytes));
  }

  return root;
}

function sortedChildren(node) {
  return Array.from(node.children.entries()).sort(([a], [b]) => a.localeCompare(b));
}

function normalizePath(path) {
  if (path.indexOf('\0') >= 0) {
    return { ret: ERRNO_INVAL, parts: null, wantsDir: false };
  }
  if (path.startsWith('/')) {
    return { ret: ERRNO_PERM, parts: null, wantsDir: false };
  }
  const wantsDir = path.endsWith('/');
  const parts = [];
  for (const raw of path.split('/')) {
    if (raw === '' || raw === '.') {
      continue;
    }
    if (raw === '..') {
      if (parts.length === 0) {
        return { ret: ERRNO_PERM, parts: null, wantsDir };
      }
      parts.pop();
      continue;
    }
    parts.push(raw);
  }
  return { ret: ERRNO_SUCCESS, parts, wantsDir };
}

function lookup(root, path, allowMissingLeaf = false) {
  const normalized = normalizePath(path);
  if (normalized.ret !== ERRNO_SUCCESS) {
    return { ret: normalized.ret, node: null, parent: null, leaf: '' };
  }
  let node = root;
  for (let i = 0; i < normalized.parts.length; i += 1) {
    if (node.kind !== 'dir') {
      return { ret: ERRNO_NOTDIR, node: null, parent: null, leaf: '' };
    }
    const name = normalized.parts[i];
    const child = node.children.get(name);
    if (!child) {
      if (allowMissingLeaf && i === normalized.parts.length - 1) {
        return { ret: ERRNO_NOENT, node: null, parent: node, leaf: name };
      }
      return { ret: ERRNO_NOENT, node: null, parent: null, leaf: name };
    }
    node = child;
  }
  if (normalized.wantsDir && node.kind !== 'dir') {
    return { ret: ERRNO_NOTDIR, node: null, parent: null, leaf: '' };
  }
  return { ret: ERRNO_SUCCESS, node, parent: node.parent, leaf: node.name };
}

function errnoFromNodeError(error) {
  switch (error && error.code) {
    case 'EACCES':
    case 'EPERM':
      return ERRNO_PERM;
    case 'EBADF':
      return ERRNO_BADF;
    case 'EEXIST':
      return ERRNO_EXIST;
    case 'EISDIR':
      return ERRNO_ISDIR;
    case 'EINVAL':
      return ERRNO_INVAL;
    case 'ENOENT':
      return ERRNO_NOENT;
    case 'ENOTDIR':
      return ERRNO_NOTDIR;
    case 'ENOTEMPTY':
      return ERRNO_NOTEMPTY;
    case 'EROFS':
      return ERRNO_ROFS;
    default:
      return ERRNO_IO;
  }
}

function hostFiletype(stats) {
  if (stats && typeof stats.isDirectory === 'function' && stats.isDirectory()) {
    return FILETYPE_DIRECTORY;
  }
  return FILETYPE_REGULAR_FILE;
}

function hostFilestat(nodeFs, hostPath) {
  const stats = nodeFs.statSync(hostPath);
  return {
    filetype: hostFiletype(stats),
    ino: BigInt(stats.ino || 0),
    size: BigInt(stats.size || 0),
  };
}

function isHostPathInside(nodePath, root, candidate) {
  const rel = nodePath.relative(root, candidate);
  return rel === '' || (rel && !rel.startsWith('..') && !nodePath.isAbsolute(rel));
}

function resolveHostPath(openFd, wasiPath, nodePath) {
  const normalized = normalizePath(wasiPath);
  if (normalized.ret !== ERRNO_SUCCESS) {
    return { ret: normalized.ret, path: '' };
  }
  const resolved = nodePath.resolve(openFd.hostPath, ...normalized.parts);
  const root = nodePath.resolve(openFd.hostPath);
  if (!isHostPathInside(nodePath, root, resolved)) {
    return { ret: ERRNO_PERM, path: '' };
  }
  return { ret: ERRNO_SUCCESS, path: resolved, wantsDir: normalized.wantsDir };
}

function safeNumber(value, syscall, what) {
  const n = Number(value);
  if (!Number.isSafeInteger(n) || n < 0) {
    throw new WasiTrap(syscall, what + ' is not a safe non-negative number: ' + String(value));
  }
  return n;
}

function asU32(value) {
  return Number(value) >>> 0;
}

function u64(value) {
  return typeof value === 'bigint' ? value : BigInt(value);
}

function i64(value) {
  return typeof value === 'bigint' ? value : BigInt(value);
}

function checkedEnd(ptr, len, capacity, syscall, what) {
  ptr = asU32(ptr);
  len = asU32(len);
  const end = ptr + len;
  if (end < ptr || end > capacity) {
    throw new WasiTrap(syscall, what + ' out of bounds ptr=' + ptr + ' len=' + len + ' memory=' + capacity);
  }
  return end;
}

function readIovecs(view, ptr, len, syscall) {
  ptr = asU32(ptr);
  len = asU32(len);
  checkedEnd(ptr, len * 8, view.byteLength, syscall, 'iovec array');
  const out = [];
  for (let i = 0; i < len; i += 1) {
    const base = ptr + i * 8;
    out.push({
      ptr: view.getUint32(base, true),
      len: view.getUint32(base + 4, true),
    });
  }
  return out;
}

function writeFdstat(view, ptr, filetype, flags, rightsBase, rightsInheriting) {
  view.setUint8(ptr, filetype);
  view.setUint8(ptr + 1, 0);
  view.setUint16(ptr + 2, flags & 0xffff, true);
  view.setUint32(ptr + 4, 0, true);
  view.setBigUint64(ptr + 8, rightsBase, true);
  view.setBigUint64(ptr + 16, rightsInheriting, true);
}

function writeFilestat(view, ptr, node) {
  const size = node ? BigInt(node.size) : 0n;
  const ino = node ? node.ino : 0n;
  const filetype = node ? node.filetype : FILETYPE_CHARACTER_DEVICE;
  view.setBigUint64(ptr, 0n, true);
  view.setBigUint64(ptr + 8, ino, true);
  view.setUint8(ptr + 16, filetype);
  view.setUint8(ptr + 17, 0);
  view.setUint16(ptr + 18, 0, true);
  view.setUint32(ptr + 20, 0, true);
  view.setBigUint64(ptr + 24, 1n, true);
  view.setBigUint64(ptr + 32, size, true);
  view.setBigUint64(ptr + 40, 0n, true);
  view.setBigUint64(ptr + 48, 0n, true);
  view.setBigUint64(ptr + 56, 0n, true);
}

function writeDirent(view, ptr, nextCookie, ino, nameBytes, filetype) {
  view.setBigUint64(ptr, nextCookie, true);
  view.setBigUint64(ptr + 8, ino, true);
  view.setUint32(ptr + 16, nameBytes.byteLength, true);
  view.setUint8(ptr + 20, filetype);
  view.setUint8(ptr + 21, 0);
  view.setUint16(ptr + 22, 0, true);
}

function trapSafeWasiImport(name, fn, debug) {
  return function trapSafeWasiImportWrapper(...args) {
    try {
      return fn(...args);
    } catch (error) {
      if (error instanceof WasiTrap) {
        debug('[goldweb wasi] ' + error.message);
        return ERRNO_FAULT;
      }
      throw error;
    }
  };
}

function trapSafeWasiImports(imports, debug) {
  const wrapped = {};
  for (const [name, value] of Object.entries(imports)) {
    wrapped[name] = typeof value === 'function'
      ? trapSafeWasiImport(name, value, debug)
      : value;
  }
  return wrapped;
}

const TINYGO_WASI_IMPORTS = [
  'proc_exit',
  'fd_write',
  'fd_close',
  'fd_fdstat_get',
  'fd_fdstat_set_flags',
  'fd_filestat_get',
  'fd_prestat_get',
  'fd_prestat_dir_name',
  'fd_read',
  'fd_readdir',
  'fd_seek',
  'fd_tell',
  'path_open',
  'path_filestat_get',
  'path_create_directory',
  'path_remove_directory',
  'path_unlink_file',
  'path_rename',
  'random_get',
];

function tinyGoImportsOnly(imports) {
  const out = {};
  for (const name of TINYGO_WASI_IMPORTS) {
    out[name] = imports[name];
  }
  return out;
}

export function createGoIvyTinyGoWasiP1(options) {
  const args = options.args || [];
  const env = options.env || [];
  const includeRootName = String(options.includeRoot || (options.includeTree && options.includeTree.root) || 'include');
  const root = buildIncludeRoot(options.includeTree || { files: [] });
  const stdinBytes = options.stdin || new Uint8Array(0);
  const stdout = options.stdout || (() => {});
  const stderr = options.stderr || (() => {});
  const debug = options.debug || (() => {});
  const nodeFs = options.nodeFilesystem || null;
  const nodePath = options.nodePath || null;
  const hostPreopenPath = nodeFs && nodePath
    ? nodePath.resolve(String(options.hostPreopenPath || '/'))
    : '';
  const hostPreopenName = String(options.hostPreopenName || hostPreopenPath || '/');
  const procExit = options.procExit || ((code) => {
    throw new GoIvyTinyGoWasiProcExit(code);
  });

  let instance = null;
  const fds = [
    new OpenFd('stdin', { stdinData: stdinBytes }),
    new OpenFd('stdio', { write: stdout }),
    new OpenFd('stdio', { write: stderr }),
    new OpenFd('dir', { node: root, preopenName: includeRootName }),
  ];
  if (nodeFs && nodePath) {
    fds.push(new OpenFd('host-dir', { hostPath: hostPreopenPath, preopenName: hostPreopenName }));
  }

  function memory() {
    const mem = instance && instance.exports && instance.exports.memory;
    if (!mem) {
      throw new Error('WASI memory is not initialized');
    }
    return mem;
  }

  function view() {
    return new DataView(memory().buffer);
  }

  function bytes() {
    return new Uint8Array(memory().buffer);
  }

  function fd(fd) {
    return fds[fd] || null;
  }

  function readString(ptr, len, syscall) {
    const mem = bytes();
    checkedEnd(ptr, len, mem.byteLength, syscall, 'string');
    return textDecoder.decode(mem.subarray(asU32(ptr), asU32(ptr) + asU32(len)));
  }

  function writeBytes(ptr, data, syscall) {
    const mem = bytes();
    checkedEnd(ptr, data.byteLength, mem.byteLength, syscall, 'write');
    mem.set(data, asU32(ptr));
  }

  function writeU32(ptr, value, syscall) {
    const v = view();
    checkedEnd(ptr, 4, v.byteLength, syscall, 'u32 result');
    v.setUint32(asU32(ptr), value >>> 0, true);
  }

  function writeU64(ptr, value, syscall) {
    const v = view();
    checkedEnd(ptr, 8, v.byteLength, syscall, 'u64 result');
    v.setBigUint64(asU32(ptr), u64(value), true);
  }

  function pushFd(openFd) {
    for (let i = 4; i < fds.length; i += 1) {
      if (!fds[i]) {
        fds[i] = openFd;
        return i;
      }
    }
    fds.push(openFd);
    return fds.length - 1;
  }

  function hostOpenFlags(oflags, rightsBase, fdFlags, existed) {
    const rights = u64(rightsBase);
    const wantsRead = (rights & RIGHTS_FD_READ) !== 0n;
    const wantsWrite = (rights & RIGHTS_FD_WRITE) !== 0n ||
      (oflags & (OFLAGS_CREAT | OFLAGS_TRUNC)) !== 0 ||
      (fdFlags & FDFLAGS_APPEND) !== 0;
    const create = (oflags & OFLAGS_CREAT) !== 0;
    const excl = (oflags & OFLAGS_EXCL) !== 0;
    const trunc = (oflags & OFLAGS_TRUNC) !== 0;
    const append = (fdFlags & FDFLAGS_APPEND) !== 0;

    if (!wantsWrite) {
      return { flags: 'r', canRead: true, canWrite: false, truncateAfterOpen: false };
    }
    if (!create && !existed) {
      return { ret: ERRNO_NOENT };
    }
    if (create && excl && existed) {
      return { ret: ERRNO_EXIST };
    }
    if (append) {
      return { flags: wantsRead ? (excl ? 'ax+' : 'a+') : (excl ? 'ax' : 'a'), canRead: wantsRead, canWrite: true, truncateAfterOpen: false };
    }
    if (trunc) {
      if (create) {
        return { flags: wantsRead ? (excl ? 'wx+' : 'w+') : (excl ? 'wx' : 'w'), canRead: wantsRead, canWrite: true, truncateAfterOpen: false };
      }
      return { flags: 'r+', canRead: true, canWrite: true, truncateAfterOpen: true };
    }
    if (create && !existed) {
      return { flags: wantsRead ? (excl ? 'wx+' : 'w+') : (excl ? 'wx' : 'w'), canRead: wantsRead, canWrite: true, truncateAfterOpen: false };
    }
    return { flags: wantsRead ? 'r+' : 'r+', canRead: true, canWrite: true, truncateAfterOpen: false };
  }

  function pathOpenHost(f, pathName, oflags, rightsBase, fdFlags, openedFdPtr) {
    if (!nodeFs || !nodePath) {
      return ERRNO_BADF;
    }
    if ((oflags & OFLAGS_DIRECTORY) !== 0 && (oflags & OFLAGS_CREAT) !== 0) {
      return ERRNO_INVAL;
    }
    const resolved = resolveHostPath(f, pathName, nodePath);
    if (resolved.ret !== ERRNO_SUCCESS) {
      return resolved.ret;
    }

    let stats = null;
    let existed = false;
    try {
      stats = nodeFs.statSync(resolved.path);
      existed = true;
    } catch (error) {
      const ret = errnoFromNodeError(error);
      if (ret !== ERRNO_NOENT) {
        return ret;
      }
    }

    const wantsDirectory = (oflags & OFLAGS_DIRECTORY) !== 0 || resolved.wantsDir;
    if (wantsDirectory) {
      if (!existed) {
        return ERRNO_NOENT;
      }
      if (!stats.isDirectory()) {
        return ERRNO_NOTDIR;
      }
      checkedEnd(openedFdPtr, 4, view().byteLength, 'path_open', 'opened fd result');
      const newFd = pushFd(new OpenFd('host-dir', { hostPath: resolved.path, flags: fdFlags >>> 0 }));
      writeU32(openedFdPtr, newFd, 'path_open');
      return ERRNO_SUCCESS;
    }

    if (existed && stats.isDirectory()) {
      return ERRNO_ISDIR;
    }

    const openPlan = hostOpenFlags(oflags, rightsBase, fdFlags, existed);
    if (openPlan.ret) {
      return openPlan.ret;
    }

    let hostFd;
    try {
      hostFd = nodeFs.openSync(resolved.path, openPlan.flags, 0o666);
      if (openPlan.truncateAfterOpen) {
        nodeFs.ftruncateSync(hostFd, 0);
      }
    } catch (error) {
      return errnoFromNodeError(error);
    }

    checkedEnd(openedFdPtr, 4, view().byteLength, 'path_open', 'opened fd result');
    const opened = new OpenFd('host-file', {
      hostPath: resolved.path,
      hostFd,
      flags: fdFlags >>> 0,
      canRead: openPlan.canRead,
      canWrite: openPlan.canWrite,
    });
    if ((fdFlags & FDFLAGS_APPEND) !== 0) {
      try {
        opened.pos = BigInt(nodeFs.fstatSync(hostFd).size || 0);
      } catch (error) {
        nodeFs.closeSync(hostFd);
        return errnoFromNodeError(error);
      }
    }
    const newFd = pushFd(opened);
    writeU32(openedFdPtr, newFd, 'path_open');
    return ERRNO_SUCCESS;
  }

  function hostRelativePathStat(f, pathName, syscall) {
    if (!nodeFs || !nodePath) {
      return { ret: ERRNO_BADF, info: null };
    }
    const resolved = resolveHostPath(f, pathName, nodePath);
    if (resolved.ret !== ERRNO_SUCCESS) {
      return { ret: resolved.ret, info: null };
    }
    try {
      return { ret: ERRNO_SUCCESS, info: hostFilestat(nodeFs, resolved.path), path: resolved.path };
    } catch (error) {
      debug('[goldweb wasi] ' + syscall + ' failed path=' + resolved.path + ' error=' + (error && error.message ? error.message : String(error)));
      return { ret: errnoFromNodeError(error), info: null };
    }
  }

  function unsupported(name, argsLike) {
    debug('[goldweb wasi] unsupported ' + name + '(' + Array.from(argsLike).join(', ') + ')');
    return ERRNO_NOSYS;
  }

  const wasiImport = {
    args_sizes_get(argcPtr, argvBufSizePtr) {
      const encoded = args.map((arg) => textEncoder.encode(String(arg)));
      writeU32(argcPtr, encoded.length, 'args_sizes_get');
      writeU32(argvBufSizePtr, encoded.reduce((n, arg) => n + arg.byteLength + 1, 0), 'args_sizes_get');
      return ERRNO_SUCCESS;
    },

    args_get(argvPtr, argvBufPtr) {
      const v = view();
      const mem = bytes();
      let argv = asU32(argvPtr);
      let cursor = asU32(argvBufPtr);
      checkedEnd(argv, args.length * 4, mem.byteLength, 'args_get', 'argv');
      for (const arg of args) {
        const data = textEncoder.encode(String(arg));
        checkedEnd(cursor, data.byteLength + 1, mem.byteLength, 'args_get', 'argv buffer');
        v.setUint32(argv, cursor, true);
        mem.set(data, cursor);
        mem[cursor + data.byteLength] = 0;
        argv += 4;
        cursor += data.byteLength + 1;
      }
      return ERRNO_SUCCESS;
    },

    environ_sizes_get(countPtr, sizePtr) {
      const encoded = env.map((entry) => textEncoder.encode(String(entry)));
      writeU32(countPtr, encoded.length, 'environ_sizes_get');
      writeU32(sizePtr, encoded.reduce((n, entry) => n + entry.byteLength + 1, 0), 'environ_sizes_get');
      return ERRNO_SUCCESS;
    },

    environ_get(environPtr, environBufPtr) {
      const v = view();
      const mem = bytes();
      let envp = asU32(environPtr);
      let cursor = asU32(environBufPtr);
      checkedEnd(envp, env.length * 4, mem.byteLength, 'environ_get', 'environ');
      for (const entry of env) {
        const data = textEncoder.encode(String(entry));
        checkedEnd(cursor, data.byteLength + 1, mem.byteLength, 'environ_get', 'environ buffer');
        v.setUint32(envp, cursor, true);
        mem.set(data, cursor);
        mem[cursor + data.byteLength] = 0;
        envp += 4;
        cursor += data.byteLength + 1;
      }
      return ERRNO_SUCCESS;
    },

    clock_time_get(id, precision, timePtr) {
      void precision;
      if (id === CLOCKID_REALTIME) {
        writeU64(timePtr, BigInt(Date.now()) * 1000000n, 'clock_time_get');
        return ERRNO_SUCCESS;
      }
      if (id === CLOCKID_MONOTONIC) {
        writeU64(timePtr, BigInt(Math.round(performance.now() * 1000000)), 'clock_time_get');
        return ERRNO_SUCCESS;
      }
      return ERRNO_INVAL;
    },

    fd_close(fdnum) {
      const f = fd(fdnum);
      if (!f) {
        return ERRNO_BADF;
      }
      if (fdnum >= 4) {
        if (f.kind === 'host-file' && f.hostFd !== null && nodeFs) {
          try {
            nodeFs.closeSync(f.hostFd);
          } catch (error) {
            fds[fdnum] = null;
            return errnoFromNodeError(error);
          }
        }
        fds[fdnum] = null;
      }
      return ERRNO_SUCCESS;
    },

    fd_fdstat_get(fdnum, statPtr) {
      const f = fd(fdnum);
      if (!f) {
        return ERRNO_BADF;
      }
      const v = view();
      checkedEnd(statPtr, 24, v.byteLength, 'fd_fdstat_get', 'fdstat');
      writeFdstat(v, asU32(statPtr), f.filetype, f.flags, f.rightsBase(), f.rightsBase());
      return ERRNO_SUCCESS;
    },

    fd_fdstat_set_flags(fdnum, flags) {
      const f = fd(fdnum);
      if (!f) {
        return ERRNO_BADF;
      }
      f.flags = flags >>> 0;
      return ERRNO_SUCCESS;
    },

    fd_filestat_get(fdnum, statPtr) {
      const f = fd(fdnum);
      if (!f) {
        return ERRNO_BADF;
      }
      const v = view();
      checkedEnd(statPtr, 64, v.byteLength, 'fd_filestat_get', 'filestat');
      if (f.kind === 'host-file' || f.kind === 'host-dir') {
        if (!nodeFs) {
          return ERRNO_BADF;
        }
        try {
          writeFilestat(v, asU32(statPtr), hostFilestat(nodeFs, f.hostPath));
          return ERRNO_SUCCESS;
        } catch (error) {
          return errnoFromNodeError(error);
        }
      }
      writeFilestat(v, asU32(statPtr), f.node);
      return ERRNO_SUCCESS;
    },

    fd_prestat_get(fdnum, prestatPtr) {
      const f = fd(fdnum);
      if (!f) {
        return ERRNO_BADF;
      }
      if (!f.preopenName) {
        return ERRNO_NOTDIR;
      }
      const v = view();
      const nameBytes = textEncoder.encode(f.preopenName);
      checkedEnd(prestatPtr, 8, v.byteLength, 'fd_prestat_get', 'prestat');
      v.setUint32(asU32(prestatPtr), 0, true);
      v.setUint32(asU32(prestatPtr) + 4, nameBytes.byteLength, true);
      return ERRNO_SUCCESS;
    },

    fd_prestat_dir_name(fdnum, pathPtr, pathLen) {
      const f = fd(fdnum);
      if (!f) {
        return ERRNO_BADF;
      }
      if (!f.preopenName) {
        return ERRNO_NOTDIR;
      }
      const nameBytes = textEncoder.encode(f.preopenName);
      if (nameBytes.byteLength < pathLen) {
        return ERRNO_NAMETOOLONG;
      }
      writeBytes(pathPtr, nameBytes.subarray(0, asU32(pathLen)), 'fd_prestat_dir_name');
      return ERRNO_SUCCESS;
    },

    fd_read(fdnum, iovsPtr, iovsLen, nreadPtr) {
      const f = fd(fdnum);
      if (!f) {
        return ERRNO_BADF;
      }
      if (f.kind !== 'file' && f.kind !== 'stdin' && f.kind !== 'host-file') {
        writeU32(nreadPtr, 0, 'fd_read');
        return ERRNO_BADF;
      }
      const v = view();
      const mem = bytes();
      const iovecs = readIovecs(v, iovsPtr, iovsLen, 'fd_read');
      let total = 0;
      for (const iovec of iovecs) {
        checkedEnd(iovec.ptr, iovec.len, mem.byteLength, 'fd_read', 'iovec buffer');
        let src;
        if (f.kind === 'stdin') {
          src = f.stdinData.subarray(f.stdinPos, f.stdinPos + iovec.len);
          f.stdinPos += src.byteLength;
        } else if (f.kind === 'host-file') {
          if (!nodeFs || !f.canRead) {
            writeU32(nreadPtr, total, 'fd_read');
            return ERRNO_BADF;
          }
          let n;
          try {
            n = nodeFs.readSync(f.hostFd, mem, iovec.ptr, iovec.len, safeNumber(f.pos, 'fd_read', 'file offset'));
          } catch (error) {
            writeU32(nreadPtr, total, 'fd_read');
            return total > 0 ? ERRNO_SUCCESS : errnoFromNodeError(error);
          }
          f.pos += BigInt(n);
          total += n;
          if (n !== iovec.len) {
            break;
          }
          continue;
        } else {
          const start = Number(f.pos);
          src = f.node.data.subarray(start, start + iovec.len);
          f.pos += BigInt(src.byteLength);
        }
        mem.set(src, iovec.ptr);
        total += src.byteLength;
        if (src.byteLength !== iovec.len) {
          break;
        }
      }
      writeU32(nreadPtr, total, 'fd_read');
      return ERRNO_SUCCESS;
    },

    fd_readdir(fdnum, bufPtr, bufLen, cookie, bufusedPtr) {
      const f = fd(fdnum);
      if (!f) {
        return ERRNO_BADF;
      }
      if (f.kind === 'host-dir') {
        if (!nodeFs) {
          writeU32(bufusedPtr, 0, 'fd_readdir');
          return ERRNO_BADF;
        }
        if (asU32(bufLen) < 24) {
          return ERRNO_INVAL;
        }
        const mem = bytes();
        const v = view();
        bufPtr = asU32(bufPtr);
        bufLen = asU32(bufLen);
        checkedEnd(bufPtr, bufLen, mem.byteLength, 'fd_readdir', 'dirent buffer');

        let entries;
        try {
          entries = [
            { name: '.', ino: 1n, type: FILETYPE_DIRECTORY },
            { name: '..', ino: 1n, type: FILETYPE_DIRECTORY },
          ].concat(nodeFs.readdirSync(f.hostPath, { withFileTypes: true })
            .sort((a, b) => a.name.localeCompare(b.name))
            .map((entry, index) => ({
              name: entry.name,
              ino: BigInt(index + 3),
              type: entry.isDirectory() ? FILETYPE_DIRECTORY : FILETYPE_REGULAR_FILE,
            })));
        } catch (error) {
          writeU32(bufusedPtr, 0, 'fd_readdir');
          return errnoFromNodeError(error);
        }

        let used = 0;
        let index = Number(cookie);
        if (!Number.isSafeInteger(index) || index < 0 || index > entries.length) {
          return ERRNO_NOENT;
        }
        while (index < entries.length && used < bufLen) {
          const entry = entries[index];
          const nameBytes = textEncoder.encode(entry.name);
          const recordLen = 24 + nameBytes.byteLength;
          const remaining = bufLen - used;
          if (recordLen > remaining) {
            if (remaining >= 24) {
              writeDirent(v, bufPtr + used, BigInt(index + 1), entry.ino, nameBytes, entry.type);
            }
            used = bufLen;
            break;
          }
          writeDirent(v, bufPtr + used, BigInt(index + 1), entry.ino, nameBytes, entry.type);
          mem.set(nameBytes, bufPtr + used + 24);
          used += recordLen;
          index += 1;
        }
        writeU32(bufusedPtr, used, 'fd_readdir');
        return ERRNO_SUCCESS;
      }
      if (!f.node || f.node.kind !== 'dir') {
        writeU32(bufusedPtr, 0, 'fd_readdir');
        return ERRNO_BADF;
      }
      if (asU32(bufLen) < 24) {
        return ERRNO_INVAL;
      }
      const mem = bytes();
      const v = view();
      bufPtr = asU32(bufPtr);
      bufLen = asU32(bufLen);
      checkedEnd(bufPtr, bufLen, mem.byteLength, 'fd_readdir', 'dirent buffer');

      const dot = { name: '.', ino: f.node.ino, type: FILETYPE_DIRECTORY };
      const dotdot = { name: '..', ino: f.node.parent ? f.node.parent.ino : f.node.ino, type: FILETYPE_DIRECTORY };
      const entries = [dot, dotdot].concat(sortedChildren(f.node).map(([name, child]) => ({
        name,
        ino: child.ino,
        type: child.filetype,
      })));

      let used = 0;
      let index = Number(cookie);
      if (!Number.isSafeInteger(index) || index < 0 || index > entries.length) {
        return ERRNO_NOENT;
      }
      while (index < entries.length && used < bufLen) {
        const entry = entries[index];
        const nameBytes = textEncoder.encode(entry.name);
        const recordLen = 24 + nameBytes.byteLength;
        const remaining = bufLen - used;
        if (recordLen > remaining) {
          if (remaining >= 24) {
            writeDirent(v, bufPtr + used, BigInt(index + 1), entry.ino, nameBytes, entry.type);
          }
          used = bufLen;
          break;
        }
        writeDirent(v, bufPtr + used, BigInt(index + 1), entry.ino, nameBytes, entry.type);
        mem.set(nameBytes, bufPtr + used + 24);
        used += recordLen;
        index += 1;
      }
      writeU32(bufusedPtr, used, 'fd_readdir');
      return ERRNO_SUCCESS;
    },

    fd_seek(fdnum, offset, whence, newOffsetPtr) {
      const f = fd(fdnum);
      if (!f) {
        writeU64(newOffsetPtr, 0n, 'fd_seek');
        return ERRNO_BADF;
      }
      if (f.kind === 'dir' || f.kind === 'host-dir') {
        return ERRNO_ISDIR;
      }
      if (f.kind !== 'file' && f.kind !== 'stdin' && f.kind !== 'host-file') {
        writeU64(newOffsetPtr, 0n, 'fd_seek');
        return ERRNO_BADF;
      }
      let next;
      if (whence === WHENCE_SET) {
        next = i64(offset);
      } else if (whence === WHENCE_CUR) {
        next = f.pos + i64(offset);
      } else if (whence === WHENCE_END && f.kind === 'file') {
        next = BigInt(f.node.data.byteLength) + i64(offset);
      } else if (whence === WHENCE_END && f.kind === 'host-file') {
        try {
          next = BigInt(nodeFs.fstatSync(f.hostFd).size || 0) + i64(offset);
        } catch (error) {
          writeU64(newOffsetPtr, f.pos, 'fd_seek');
          return errnoFromNodeError(error);
        }
      } else {
        writeU64(newOffsetPtr, f.pos, 'fd_seek');
        return ERRNO_INVAL;
      }
      if (next < 0n) {
        writeU64(newOffsetPtr, f.pos, 'fd_seek');
        return ERRNO_INVAL;
      }
      f.pos = next;
      writeU64(newOffsetPtr, next, 'fd_seek');
      return ERRNO_SUCCESS;
    },

    fd_write(fdnum, iovsPtr, iovsLen, nwrittenPtr) {
      const f = fd(fdnum);
      if (!f || !f.writable) {
        writeU32(nwrittenPtr, 0, 'fd_write');
        return ERRNO_BADF;
      }
      const v = view();
      const mem = bytes();
      const iovecs = readIovecs(v, iovsPtr, iovsLen, 'fd_write');
      let total = 0;
      for (const iovec of iovecs) {
        checkedEnd(iovec.ptr, iovec.len, mem.byteLength, 'fd_write', 'iovec buffer');
        if (f.kind === 'host-file') {
          try {
            const position = (f.flags & FDFLAGS_APPEND) !== 0 ? null : safeNumber(f.pos, 'fd_write', 'file offset');
            const n = nodeFs.writeSync(f.hostFd, mem, iovec.ptr, iovec.len, position);
            total += n;
            if (position === null) {
              f.pos = BigInt(nodeFs.fstatSync(f.hostFd).size || 0);
            } else {
              f.pos += BigInt(n);
            }
            if (n !== iovec.len) {
              break;
            }
          } catch (error) {
            writeU32(nwrittenPtr, total, 'fd_write');
            return total > 0 ? ERRNO_SUCCESS : errnoFromNodeError(error);
          }
        } else {
          const data = mem.slice(iovec.ptr, iovec.ptr + iovec.len);
          f.write(data);
          total += data.byteLength;
        }
      }
      writeU32(nwrittenPtr, total, 'fd_write');
      return ERRNO_SUCCESS;
    },

    path_filestat_get(fdnum, flags, pathPtr, pathLen, statPtr) {
      void flags;
      const f = fd(fdnum);
      if (!f) {
        return ERRNO_BADF;
      }
      const path = readString(pathPtr, pathLen, 'path_filestat_get');
      if (f.kind === 'host-dir') {
        const result = hostRelativePathStat(f, path, 'path_filestat_get');
        if (result.ret !== ERRNO_SUCCESS) {
          return result.ret;
        }
        const v = view();
        checkedEnd(statPtr, 64, v.byteLength, 'path_filestat_get', 'filestat');
        writeFilestat(v, asU32(statPtr), result.info);
        return ERRNO_SUCCESS;
      }
      if (!f.node || f.node.kind !== 'dir') {
        return ERRNO_BADF;
      }
      const result = lookup(f.node, path);
      if (result.ret !== ERRNO_SUCCESS) {
        return result.ret;
      }
      const v = view();
      checkedEnd(statPtr, 64, v.byteLength, 'path_filestat_get', 'filestat');
      writeFilestat(v, asU32(statPtr), result.node);
      return ERRNO_SUCCESS;
    },

    path_open(fdnum, dirflags, pathPtr, pathLen, oflags, rightsBase, rightsInheriting, fdFlags, openedFdPtr) {
      void dirflags;
      void rightsBase;
      void rightsInheriting;
      const f = fd(fdnum);
      if (!f) {
        return ERRNO_BADF;
      }
      const path = readString(pathPtr, pathLen, 'path_open');
      if (asU32(pathLen) === 0) {
        return ERRNO_INVAL;
      }
      if (f.kind === 'host-dir') {
        return pathOpenHost(f, path, oflags, rightsBase, fdFlags, openedFdPtr);
      }
      if (!f.node || f.node.kind !== 'dir') {
        return ERRNO_BADF;
      }
      if ((oflags & OFLAGS_DIRECTORY) !== 0 && (oflags & OFLAGS_CREAT) !== 0) {
        return ERRNO_INVAL;
      }
      const result = lookup(f.node, path, (oflags & OFLAGS_CREAT) !== 0);
      if (result.ret === ERRNO_NOENT && (oflags & OFLAGS_CREAT) !== 0) {
        return ERRNO_ROFS;
      }
      if (result.ret !== ERRNO_SUCCESS) {
        return result.ret;
      }
      if ((oflags & OFLAGS_EXCL) !== 0) {
        return ERRNO_EXIST;
      }
      if ((oflags & OFLAGS_TRUNC) !== 0) {
        return ERRNO_ROFS;
      }
      if ((oflags & OFLAGS_DIRECTORY) !== 0 && result.node.kind !== 'dir') {
        return ERRNO_NOTDIR;
      }
      if ((oflags & OFLAGS_DIRECTORY) === 0 && result.node.kind === 'dir') {
        return ERRNO_ISDIR;
      }
      checkedEnd(openedFdPtr, 4, view().byteLength, 'path_open', 'opened fd result');
      const opened = new OpenFd(result.node.kind, { node: result.node, flags: fdFlags >>> 0 });
      if ((fdFlags & FDFLAGS_APPEND) !== 0 && result.node.kind === 'file') {
        opened.pos = BigInt(result.node.data.byteLength);
      }
      const newFd = pushFd(opened);
      writeU32(openedFdPtr, newFd, 'path_open');
      return ERRNO_SUCCESS;
    },

    path_remove_directory(fdnum, pathPtr, pathLen) {
      const f = fd(fdnum);
      if (!f) {
        return ERRNO_BADF;
      }
      const path = readString(pathPtr, pathLen, 'path_remove_directory');
      if (f.kind === 'host-dir') {
        const resolved = resolveHostPath(f, path, nodePath);
        if (resolved.ret !== ERRNO_SUCCESS) {
          return resolved.ret;
        }
        try {
          nodeFs.rmdirSync(resolved.path);
          return ERRNO_SUCCESS;
        } catch (error) {
          return errnoFromNodeError(error);
        }
      }
      if (!f.node || f.node.kind !== 'dir') {
        return ERRNO_BADF;
      }
      const result = lookup(f.node, path);
      if (result.ret !== ERRNO_SUCCESS) {
        return result.ret;
      }
      if (result.node.kind !== 'dir') {
        return ERRNO_NOTDIR;
      }
      if (result.node.children.size !== 0) {
        return ERRNO_NOTEMPTY;
      }
      return ERRNO_ROFS;
    },

    path_unlink_file(fdnum, pathPtr, pathLen) {
      const f = fd(fdnum);
      if (!f) {
        return ERRNO_BADF;
      }
      const path = readString(pathPtr, pathLen, 'path_unlink_file');
      if (f.kind === 'host-dir') {
        const resolved = resolveHostPath(f, path, nodePath);
        if (resolved.ret !== ERRNO_SUCCESS) {
          return resolved.ret;
        }
        try {
          const stats = nodeFs.statSync(resolved.path);
          if (stats.isDirectory()) {
            return ERRNO_ISDIR;
          }
          nodeFs.unlinkSync(resolved.path);
          return ERRNO_SUCCESS;
        } catch (error) {
          return errnoFromNodeError(error);
        }
      }
      if (!f.node || f.node.kind !== 'dir') {
        return ERRNO_BADF;
      }
      const result = lookup(f.node, path);
      if (result.ret !== ERRNO_SUCCESS) {
        return result.ret;
      }
      if (result.node.kind === 'dir') {
        return ERRNO_ISDIR;
      }
      return ERRNO_ROFS;
    },

    poll_oneoff(inPtr, outPtr, nsubscriptions, neventsPtr) {
      const n = asU32(nsubscriptions);
      if (n === 0) {
        return ERRNO_INVAL;
      }
      const v = view();
      checkedEnd(inPtr, n * 48, v.byteLength, 'poll_oneoff', 'subscriptions');
      checkedEnd(outPtr, 32, v.byteLength, 'poll_oneoff', 'event');
      checkedEnd(neventsPtr, 4, v.byteLength, 'poll_oneoff', 'nevents');

      for (let i = 0; i < n; i += 1) {
        const sub = asU32(inPtr) + i * 48;
        const userdata = v.getBigUint64(sub, true);
        const eventtype = v.getUint8(sub + 8);
        if (eventtype === EVENTTYPE_CLOCK) {
          v.setBigUint64(outPtr, userdata, true);
          v.setUint16(outPtr + 8, ERRNO_SUCCESS, true);
          v.setUint8(outPtr + 10, EVENTTYPE_CLOCK);
          v.setUint8(outPtr + 11, 0);
          v.setUint32(outPtr + 12, 0, true);
          writeU32(neventsPtr, 1, 'poll_oneoff');
          return ERRNO_SUCCESS;
        }
        if (eventtype === EVENTTYPE_FD_READ || eventtype === EVENTTYPE_FD_WRITE) {
          continue;
        }
      }
      writeU32(neventsPtr, 0, 'poll_oneoff');
      return ERRNO_SUCCESS;
    },

    proc_exit(code) {
      return procExit(code);
    },

    random_get(bufPtr, bufLen) {
      const mem = bytes();
      bufPtr = asU32(bufPtr);
      bufLen = asU32(bufLen);
      checkedEnd(bufPtr, bufLen, mem.byteLength, 'random_get', 'random buffer');
      const out = mem.subarray(bufPtr, bufPtr + bufLen);
      if (globalThis.crypto && typeof globalThis.crypto.getRandomValues === 'function') {
        for (let offset = 0; offset < out.byteLength; offset += 65536) {
          globalThis.crypto.getRandomValues(out.subarray(offset, Math.min(offset + 65536, out.byteLength)));
        }
      } else {
        for (let i = 0; i < out.byteLength; i += 1) {
          out[i] = Math.floor(Math.random() * 256);
        }
      }
      return ERRNO_SUCCESS;
    },

    sched_yield() {
      return ERRNO_SUCCESS;
    },

    fd_advise() { return unsupported('fd_advise', arguments); },
    fd_allocate() { return unsupported('fd_allocate', arguments); },
    fd_datasync() { return unsupported('fd_datasync', arguments); },
    fd_fdstat_set_rights() { return unsupported('fd_fdstat_set_rights', arguments); },
    fd_filestat_set_size() { return unsupported('fd_filestat_set_size', arguments); },
    fd_filestat_set_times() { return unsupported('fd_filestat_set_times', arguments); },
    fd_pread() { return unsupported('fd_pread', arguments); },
    fd_pwrite() { return unsupported('fd_pwrite', arguments); },
    fd_sync() { return ERRNO_SUCCESS; },
    fd_tell(fdnum, offsetPtr) {
      const f = fd(fdnum);
      if (!f) {
        writeU64(offsetPtr, 0n, 'fd_tell');
        return ERRNO_BADF;
      }
      writeU64(offsetPtr, f.pos || 0n, 'fd_tell');
      return ERRNO_SUCCESS;
    },
    path_create_directory(fdnum, pathPtr, pathLen) {
      const f = fd(fdnum);
      if (!f) {
        return ERRNO_BADF;
      }
      if (f.kind !== 'host-dir') {
        return ERRNO_ROFS;
      }
      const path = readString(pathPtr, pathLen, 'path_create_directory');
      const resolved = resolveHostPath(f, path, nodePath);
      if (resolved.ret !== ERRNO_SUCCESS) {
        return resolved.ret;
      }
      try {
        nodeFs.mkdirSync(resolved.path);
        return ERRNO_SUCCESS;
      } catch (error) {
        return errnoFromNodeError(error);
      }
    },
    path_filestat_set_times() { return ERRNO_ROFS; },
    path_link() { return ERRNO_ROFS; },
    path_readlink() { return ERRNO_INVAL; },
    path_rename(oldFdnum, oldPathPtr, oldPathLen, newFdnum, newPathPtr, newPathLen) {
      const oldFd = fd(oldFdnum);
      const newFd = fd(newFdnum);
      if (!oldFd || !newFd) {
        return ERRNO_BADF;
      }
      if (oldFd.kind !== 'host-dir' || newFd.kind !== 'host-dir') {
        return ERRNO_ROFS;
      }
      const oldPath = readString(oldPathPtr, oldPathLen, 'path_rename');
      const newPath = readString(newPathPtr, newPathLen, 'path_rename');
      const oldResolved = resolveHostPath(oldFd, oldPath, nodePath);
      if (oldResolved.ret !== ERRNO_SUCCESS) {
        return oldResolved.ret;
      }
      const newResolved = resolveHostPath(newFd, newPath, nodePath);
      if (newResolved.ret !== ERRNO_SUCCESS) {
        return newResolved.ret;
      }
      try {
        nodeFs.renameSync(oldResolved.path, newResolved.path);
        return ERRNO_SUCCESS;
      } catch (error) {
        return errnoFromNodeError(error);
      }
    },
    path_symlink() { return ERRNO_ROFS; },
    proc_raise(sig) {
      throw new Error('WASI proc_raise signal ' + sig);
    },
    sock_accept() { return ERRNO_NOSYS; },
    sock_recv() { return ERRNO_NOSYS; },
    sock_send() { return ERRNO_NOSYS; },
    sock_shutdown() { return ERRNO_NOSYS; },
  };

  return {
    wasiImport: trapSafeWasiImports(tinyGoImportsOnly(wasiImport), debug),
    setInstance(wasmInstance) {
      instance = wasmInstance;
    },
    initialize(wasmInstance) {
      instance = wasmInstance;
    },
    start(wasmInstance) {
      instance = wasmInstance;
      try {
        instance.exports._start();
        return 0;
      } catch (error) {
        if (error instanceof GoIvyTinyGoWasiProcExit) {
          return error.code;
        }
        throw error;
      }
    },
  };
}
