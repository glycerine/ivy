#!/usr/bin/env node

import fs from 'node:fs';
import path from 'node:path';
import { webcrypto } from 'node:crypto';
import { performance } from 'node:perf_hooks';
import { createRequire } from 'node:module';
import { pathToFileURL } from 'node:url';

const nodeProcess = globalThis.process;
const require = createRequire(import.meta.url);
const startedAt = performance.now();

function diag(message) {
  const elapsedMs = Math.round(performance.now() - startedAt);
  fs.writeSync(2, `[tinynode.mjs +${elapsedMs}ms] ${message}\n`);
}

function usage() {
  console.error('usage: node tinynode.mjs config.json');
}

function readJSON(file) {
  return JSON.parse(fs.readFileSync(file, 'utf8'));
}

function fileURL(file) {
  return pathToFileURL(path.resolve(file)).href;
}

function readIncludeTree(root) {
  const absRoot = path.resolve(root);
  const files = [];

  function walk(dir) {
    for (const entry of fs.readdirSync(dir, { withFileTypes: true })) {
      const full = path.join(dir, entry.name);
      if (entry.isDirectory()) {
        walk(full);
        continue;
      }
      if (!entry.isFile()) {
        continue;
      }
      files.push({
        path: path.relative(absRoot, full).split(path.sep).join('/'),
        data: fs.readFileSync(full, 'utf8'),
      });
    }
  }

  walk(absRoot);
  files.sort((a, b) => a.path.localeCompare(b.path));
  return { root: absRoot, files };
}

function loadZ3Glue(z3JSPath) {
  const source = fs.readFileSync(z3JSPath, 'utf8');
  const moduleObject = { exports: {} };
  return new Function(
    'module',
    'exports',
    'require',
    '__dirname',
    '__filename',
    source + '\nreturn module.exports;',
  )(moduleObject, moduleObject.exports, require, path.dirname(z3JSPath), z3JSPath);
}

function bytesToBuffer(bytes) {
  return Buffer.from(bytes.buffer, bytes.byteOffset, bytes.byteLength);
}

function writeAll(stream, bytes) {
  stream.write(bytesToBuffer(bytes));
}

function writeAllFD(fd, bytes) {
  fs.writeSync(fd, bytesToBuffer(bytes));
}

function writeLineFD(fd, text) {
  fs.writeSync(fd, String(text) + '\n');
}

async function flushWritable(stream) {
  await new Promise((resolve, reject) => {
    stream.write('', (err) => {
      if (err) {
        reject(err);
      } else {
        resolve();
      }
    });
  });
}

function installHostGlobals() {
  if (!globalThis.crypto) {
    globalThis.crypto = webcrypto;
  }
  if (!globalThis.performance) {
    globalThis.performance = performance;
  }
  if (!globalThis.TextEncoder) {
    globalThis.TextEncoder = TextEncoder;
  }
  if (!globalThis.TextDecoder) {
    globalThis.TextDecoder = TextDecoder;
  }
}

async function main() {
  if (nodeProcess.argv.length !== 3) {
    usage();
    return 2;
  }

  diag(`reading config ${nodeProcess.argv[2]}`);
  const cfg = readJSON(nodeProcess.argv[2]);
  diag(`reading spec ${cfg.specPath}`);
  const spec = fs.readFileSync(cfg.specPath, 'utf8');
  diag(`read spec bytes=${Buffer.byteLength(spec, 'utf8')}`);
  diag(`reading include tree ${cfg.includeDir}`);
  const includeTree = readIncludeTree(cfg.includeDir);
  const includeRoot = String(includeTree.root || cfg.includeDir);
  diag(`read include files=${includeTree.files.length} root=${includeRoot}`);

  installHostGlobals();
  diag('installed host globals');

  diag(`loading Z3 glue ${cfg.z3JS}`);
  const initZ3 = loadZ3Glue(cfg.z3JS);
  diag(`instantiating Z3 wasm ${cfg.z3Wasm}`);
  const z3 = await initZ3({
    print(text) {
      writeLineFD(1, text);
    },
    printErr(text) {
      writeLineFD(2, text);
    },
    locateFile(file) {
      if (file === 'z3-api.wasm') {
        diag(`found: ${file}, returning as cfg.z3Wasm ${cfg.z3Wasm}`);
        return cfg.z3Wasm;
      }
      diag(`did not find: ${file}, guessing ${path.dirname(cfg.z3JS)}/${file}`); // seen
      return path.join(path.dirname(cfg.z3JS), file);
    },
  });
  diag('Z3 wasm ready');

  diag(`importing smt bridge ${cfg.smtImports}`);
  const z3Imports = await import(fileURL(cfg.smtImports));
  diag(`importing node fs host ${cfg.nodeFS}`);
  const nodeFS = await import(fileURL(cfg.nodeFS));
  nodeFS.installGoIvyNodeFS({
    includeRoot,
    includeTree,
    stdout: (data) => writeAllFD(1, data),
    stderr: (data) => writeAllFD(2, data),
  });
  diag('installed Go wasm Node fs host');

  diag(`loading wasm_exec ${cfg.wasmExec}`);
  const wasmExecSource = fs.readFileSync(cfg.wasmExec, 'utf8');
  new Function(wasmExecSource + '\n//# sourceURL=' + cfg.wasmExec)();
  if (typeof globalThis.Go !== 'function') {
    throw new Error(cfg.wasmExec + ' did not expose globalThis.Go');
  }
  diag('wasm_exec loaded');

  let wasmMemory;
  let goRunError = null;
  const go = new globalThis.Go();
  go.argv = ['goivy_check_jswasm'];
  go.env = {
    GOIVY_INCLUDE: includeRoot,
    GOIVY_WASM_MEMORY_LIMIT: cfg.memoryLimit || '3GiB',
    GOIVY_WASM_GOGC: cfg.gogc || '50',
  };
  go.exit = (code) => {
    if (code !== 0) {
      nodeProcess.stderr.write('[go js/wasm] exit code ' + code + '\n');
    }
  };
  go.importObject.smt_z3 = z3Imports.createSmtZ3Imports({ z3, getGoMemory: () => wasmMemory });

  diag(`reading Go Ivy wasm ${cfg.goivyWasm}`);
  const wasmBytes = fs.readFileSync(cfg.goivyWasm);
  diag(`instantiating Go Ivy wasm bytes=${wasmBytes.length}`);
  const result = await WebAssembly.instantiate(wasmBytes, go.importObject);
  const instance = result.instance;
  wasmMemory = instance.exports.mem;
  if (!wasmMemory) {
    throw new Error('Go Ivy js/wasm does not export mem');
  }
  diag('Go Ivy wasm instantiated; starting Go runtime');

  go.run(instance).catch((error) => {
    goRunError = error;
  });
  diag('Go runtime started; waiting for goivyCheckReady');

  for (let i = 0; i < 2000 && !globalThis.goivyCheckReady; i += 1) {
    if (goRunError) {
      throw goRunError;
    }
    if (i > 0 && i % 100 === 0) {
      diag(`still waiting for goivyCheckReady after ${i * 5}ms`);
    }
    await new Promise((resolve) => setTimeout(resolve, 5));
  }
  if (goRunError) {
    throw goRunError;
  }
  if (typeof globalThis.goivyCheckRun !== 'function' || !globalThis.goivyCheckReady) {
    throw new Error('Go Ivy js/wasm did not publish goivyCheckRun');
  }
  diag('goivyCheckReady published');

  const baseParams = cfg.params || {};
  const isolates = Array.isArray(cfg.isolates) ? cfg.isolates : [];
  let code = 0;
  diag(`starting goivyCheckRun filename=${cfg.specPath} params=${JSON.stringify(baseParams)} isolates=${JSON.stringify(isolates)}`);

  if (isolates.length === 0) {
    diag('calling goivyCheckRun once');
    code = globalThis.goivyCheckRun(spec, JSON.stringify({
      filename: cfg.specPath,
      params: baseParams,
    })) | 0;
    diag(`goivyCheckRun returned code=${code}`);
  } else {
    for (const isolate of isolates) {
      diag(`calling goivyCheckRun isolate=${isolate}`);
      code = globalThis.goivyCheckRun(spec, JSON.stringify({
        filename: cfg.specPath,
        params: { ...baseParams, isolate },
      })) | 0;
      diag(`goivyCheckRun isolate=${isolate} returned code=${code}`);
      if (code !== 0) {
        break;
      }
    }
  }

  await flushWritable(nodeProcess.stdout);
  await flushWritable(nodeProcess.stderr);
  return code;
}

main().then((code) => {
  nodeProcess.exit(code | 0);
}).catch(async (error) => {
  nodeProcess.stderr.write((error && error.stack ? error.stack : String(error)) + '\n');
  try {
    await flushWritable(nodeProcess.stderr);
  } finally {
    nodeProcess.exit(1);
  }
});
