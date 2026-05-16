// Read-only Node-style fs/process/path hooks for Go's official js/wasm
// wasm_exec.js shim. The shim is Go-version specific and must stay vendored
// byte-for-byte; this file supplies only the browser-side host objects it
// intentionally consults through globalThis.fs, globalThis.process, and
// globalThis.path.
const textEncoder = new TextEncoder();
const textDecoder = new TextDecoder('utf-8', { fatal: false });

const S_IFDIR = 0o040000;
const S_IFREG = 0o100000;
const O_WRONLY = 0x0001;
const O_RDWR = 0x0002;
const O_CREAT = 0x0040;
const O_EXCL = 0x0080;
const O_TRUNC = 0x0200;
const O_APPEND = 0x0400;
const O_DIRECTORY = 0x10000;

class VNode {
  constructor(kind, name, parent = null, data = null) {
    this.kind = kind;
    this.name = name;
    this.parent = parent;
    this.children = new Map();
    this.data = data || new Uint8Array(0);
    this.ino = VNode.nextIno++;
  }

  get size() {
    return this.kind === 'file' ? this.data.length : 0;
  }
}

VNode.nextIno = 1;

class OpenFile {
  constructor(node) {
    this.node = node;
    this.pos = 0;
  }
}

function nodeError(code, path) {
  const err = new Error(code + (path ? ': ' + path : ''));
  err.code = code;
  err.path = path;
  return err;
}

function callbackValue(callback, value) {
  callback(null, value);
}

function callbackError(callback, code, path) {
  callback(nodeError(code, path));
}

function splitPath(path) {
  return String(path || '').split('/').filter(Boolean);
}

function normalizeAbsolute(path, cwd = '/') {
  path = String(path || '');
  if (path === '') {
    return '';
  }
  if (!path.startsWith('/')) {
    path = String(cwd || '/') + '/' + path;
  }
  const parts = [];
  for (const part of path.split('/')) {
    if (part === '' || part === '.') {
      continue;
    }
    if (part === '..') {
      if (parts.length > 0) {
        parts.pop();
      }
      continue;
    }
    parts.push(part);
  }
  return '/' + parts.join('/');
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
    const parts = splitPath(file.path || '');
    if (parts.length === 0) {
      continue;
    }
    const filename = parts.pop();
    let dir = root;
    for (const part of parts) {
      dir = childDir(dir, part);
    }
    dir.children.set(filename, new VNode('file', filename, dir, textEncoder.encode(String(file.data || ''))));
  }

  return root;
}

function makeStat(node) {
  const now = Date.now();
  const mode = node.kind === 'dir' ? S_IFDIR | 0o555 : S_IFREG | 0o444;
  return {
    dev: 1,
    ino: node.ino,
    mode,
    nlink: 1,
    uid: 0,
    gid: 0,
    rdev: 0,
    size: node.size,
    blksize: 4096,
    blocks: Math.ceil(node.size / 512),
    atimeMs: now,
    mtimeMs: now,
    ctimeMs: now,
    isDirectory() { return node.kind === 'dir'; },
    isFile() { return node.kind === 'file'; },
  };
}

export function installGoIvyNodeFS(options = {}) {
  const includeTree = options.includeTree || { files: [] };
  const includeRoot = normalizeAbsolute(options.includeRoot || includeTree.root || '/include');
  const cwdState = { value: normalizeAbsolute(options.cwd || '/') || '/' };
  const rootNode = buildIncludeRoot(includeTree);
  let nextFd = 10;
  const openFiles = new Map();
  const stdout = typeof options.stdout === 'function' ? options.stdout : () => {};
  const stderr = typeof options.stderr === 'function' ? options.stderr : () => {};
  const heapProfile = typeof options.heapProfile === 'function' ? options.heapProfile : () => {};

  function lookup(path) {
    const normalized = normalizeAbsolute(path, cwdState.value);
    if (normalized === '') {
      return { err: 'EINVAL' };
    }
    if (normalized === includeRoot) {
      return { node: rootNode, path: normalized };
    }
    const prefix = includeRoot.endsWith('/') ? includeRoot : includeRoot + '/';
    if (!normalized.startsWith(prefix)) {
      return { err: 'ENOENT', path: normalized };
    }
    const rel = normalized.slice(prefix.length);
    let node = rootNode;
    for (const part of splitPath(rel)) {
      if (node.kind !== 'dir') {
        return { err: 'ENOTDIR', path: normalized };
      }
      const child = node.children.get(part);
      if (!child) {
        return { err: 'ENOENT', path: normalized };
      }
      node = child;
    }
    return { node, path: normalized };
  }

  function writeBytes(fd, bytes) {
    const copy = new Uint8Array(bytes);
    if (fd === 1) {
      stdout(copy);
      return copy.length;
    }
    if (fd === 2) {
      stderr(copy);
      return copy.length;
    }
    if (fd === 4) {
      heapProfile(copy);
      return copy.length;
    }
    throw nodeError('EBADF');
  }

  globalThis.fs = {
    constants: {
      O_WRONLY,
      O_RDWR,
      O_CREAT,
      O_TRUNC,
      O_APPEND,
      O_EXCL,
      O_DIRECTORY,
    },
    writeSync(fd, buf) {
      return writeBytes(fd, buf);
    },
    write(fd, buf, offset, length, position, callback) {
      void position;
      try {
        const end = offset + length;
        callbackValue(callback, writeBytes(fd, buf.subarray(offset, end)));
      } catch (err) {
        callback(err);
      }
    },
    open(path, flags, mode, callback) {
      void mode;
      if ((flags & (O_WRONLY | O_RDWR | O_CREAT | O_TRUNC | O_APPEND | O_EXCL)) !== 0) {
        callbackError(callback, 'EROFS', path);
        return;
      }
      const found = lookup(path);
      if (found.err) {
        callbackError(callback, found.err, path);
        return;
      }
      if ((flags & O_DIRECTORY) !== 0 && found.node.kind !== 'dir') {
        callbackError(callback, 'ENOTDIR', path);
        return;
      }
      const fd = nextFd++;
      openFiles.set(fd, new OpenFile(found.node));
      callbackValue(callback, fd);
    },
    close(fd, callback) {
      if (fd >= 0 && fd <= 2) {
        callbackValue(callback, undefined);
        return;
      }
      if (!openFiles.delete(fd)) {
        callbackError(callback, 'EBADF');
        return;
      }
      callbackValue(callback, undefined);
    },
    fstat(fd, callback) {
      const open = openFiles.get(fd);
      if (!open) {
        callbackError(callback, 'EBADF');
        return;
      }
      callbackValue(callback, makeStat(open.node));
    },
    stat(path, callback) {
      const found = lookup(path);
      if (found.err) {
        callbackError(callback, found.err, path);
        return;
      }
      callbackValue(callback, makeStat(found.node));
    },
    lstat(path, callback) {
      this.stat(path, callback);
    },
    readdir(path, callback) {
      const found = lookup(path);
      if (found.err) {
        callbackError(callback, found.err, path);
        return;
      }
      if (found.node.kind !== 'dir') {
        callbackError(callback, 'ENOTDIR', path);
        return;
      }
      callbackValue(callback, Array.from(found.node.children.keys()).sort());
    },
    read(fd, buffer, offset, length, position, callback) {
      const open = openFiles.get(fd);
      if (!open) {
        callbackError(callback, 'EBADF');
        return;
      }
      if (open.node.kind !== 'file') {
        callbackError(callback, 'EISDIR');
        return;
      }
      const start = position === null || position === undefined ? open.pos : Number(position);
      if (!Number.isSafeInteger(start) || start < 0) {
        callbackError(callback, 'EINVAL');
        return;
      }
      const n = Math.min(length, Math.max(0, open.node.data.length - start));
      buffer.set(open.node.data.subarray(start, start + n), offset);
      if (position === null || position === undefined) {
        open.pos += n;
      }
      callbackValue(callback, n);
    },
    fsync(callbackFd, callback) {
      void callbackFd;
      callbackValue(callback, undefined);
    },
    readlink(path, callback) { callbackError(callback, 'EINVAL', path); },
    chmod(path, mode, callback) { void mode; callbackError(callback, 'EROFS', path); },
    chown(path, uid, gid, callback) { void uid; void gid; callbackError(callback, 'EROFS', path); },
    fchmod(fd, mode, callback) { void fd; void mode; callbackError(callback, 'EROFS'); },
    fchown(fd, uid, gid, callback) { void fd; void uid; void gid; callbackError(callback, 'EROFS'); },
    ftruncate(fd, length, callback) { void fd; void length; callbackError(callback, 'EROFS'); },
    link(path, link, callback) { void link; callbackError(callback, 'EROFS', path); },
    lchown(path, uid, gid, callback) { void uid; void gid; callbackError(callback, 'EROFS', path); },
    mkdir(path, perm, callback) { void perm; callbackError(callback, 'EROFS', path); },
    rename(from, to, callback) { void to; callbackError(callback, 'EROFS', from); },
    rmdir(path, callback) { callbackError(callback, 'EROFS', path); },
    symlink(path, link, callback) { void link; callbackError(callback, 'EROFS', path); },
    truncate(path, length, callback) { void length; callbackError(callback, 'EROFS', path); },
    unlink(path, callback) { callbackError(callback, 'EROFS', path); },
    utimes(path, atime, mtime, callback) { void atime; void mtime; callbackError(callback, 'EROFS', path); },
  };

  globalThis.process = {
    getuid() { return -1; },
    getgid() { return -1; },
    geteuid() { return -1; },
    getegid() { return -1; },
    getgroups() { return []; },
    pid: -1,
    ppid: -1,
    umask() { return 0; },
    cwd() { return cwdState.value; },
    chdir(path) { cwdState.value = normalizeAbsolute(path, cwdState.value) || '/'; },
  };

  globalThis.path = {
    resolve(...pathSegments) {
      let current = '';
      for (const segment of pathSegments) {
        const s = String(segment || '');
        if (s === '') {
          continue;
        }
        if (s.startsWith('/')) {
          current = s;
        } else {
          current = (current || cwdState.value) + '/' + s;
        }
      }
      return normalizeAbsolute(current || cwdState.value);
    },
  };

  return {
    includeRoot,
    decode(bytes) {
      return textDecoder.decode(bytes, { stream: true });
    },
  };
}
