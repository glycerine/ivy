import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const scriptCache = new Map();
const helperDir = path.dirname(fileURLToPath(import.meta.url));
const staticJsDir = path.resolve(helperDir, '../../static/js');

function readStaticScript(name) {
  if (!scriptCache.has(name)) {
    scriptCache.set(name, fs.readFileSync(path.join(staticJsDir, name), 'utf8'));
  }
  return scriptCache.get(name);
}

export function loadIvyPersist() {
  const source = readStaticScript('ivyweb_persist.js');
  const factory = new window.Function(
    'window',
    'document',
    'localStorage',
    'indexedDB',
    'Blob',
    'File',
    'FileReader',
    'URL',
    `${source}\nreturn IvyPersist;`,
  );
  return factory(
    window,
    document,
    window.localStorage,
    window.indexedDB,
    window.Blob,
    window.File,
    window.FileReader,
    window.URL,
  );
}

export function loadIvyControls() {
  const source = readStaticScript('ivyweb_controls.js');
  const factory = new window.Function(
    'window',
    'document',
    `${source}\nreturn IvyControls;`,
  );
  return factory(window, document);
}

export function loadIvyApp(deps = {}) {
  const source = readStaticScript('ivyweb_app.js');
  const factory = new window.Function(
    'IvyAPI',
    'IvyControls',
    'IvyGraph',
    'IvyPersist',
    'CodeMirror',
    'ARG_STYLE',
    'CONCEPT_STYLE',
    'window',
    'document',
    'Blob',
    'File',
    'FileReader',
    'URL',
    `${source}\nreturn IvyApp;`,
  );
  return factory(
    deps.IvyAPI,
    deps.IvyControls,
    deps.IvyGraph,
    deps.IvyPersist,
    deps.CodeMirror,
    deps.ARG_STYLE || {},
    deps.CONCEPT_STYLE || {},
    window,
    document,
    window.Blob,
    window.File,
    window.FileReader,
    window.URL,
  );
}
