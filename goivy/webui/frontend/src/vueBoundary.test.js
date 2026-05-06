import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { describe, expect, it } from 'vitest';

const srcDir = path.dirname(fileURLToPath(import.meta.url));

function vueFiles(dir = srcDir) {
  const out = [];
  for (const entry of fs.readdirSync(dir, { withFileTypes: true })) {
    const full = path.join(dir, entry.name);
    if (entry.isDirectory()) {
      out.push(...vueFiles(full));
    } else if (entry.isFile() && entry.name.endsWith('.vue')) {
      out.push(full);
    }
  }
  return out;
}

function productionSourceFiles(dir = srcDir) {
  const out = [];
  for (const entry of fs.readdirSync(dir, { withFileTypes: true })) {
    const full = path.join(dir, entry.name);
    const rel = path.relative(srcDir, full);
    if (entry.isDirectory()) {
      if (rel === 'test') continue;
      out.push(...productionSourceFiles(full));
    } else if (
      entry.isFile()
      && (entry.name.endsWith('.js') || entry.name.endsWith('.vue'))
      && !entry.name.endsWith('.test.js')
    ) {
      out.push(full);
    }
  }
  return out;
}

describe('Vue production boundary', () => {
  it('keeps Vue components off the legacy IvyApp global and controller module', () => {
    const offenders = [];
    for (const file of vueFiles()) {
      const text = fs.readFileSync(file, 'utf8');
      if (/\b(?:window|globalThis\.window)\.ivyApp\b/.test(text) || /legacyAppController/.test(text)) {
        offenders.push(path.relative(srcDir, file));
      }
    }

    expect(offenders).toEqual([]);
  });

  it('keeps production legacy browser globals inside the temporary allowlist', () => {
    const allowed = new Set([
      'legacyAppController.js',
      'services/commandRegistry.js',
    ]);
    const legacyGlobal = /\b(?:window|globalThis\.window)\.(?:ivyApp|startIvyApp|IvyApp|IvyControls|IvyPersist|IvyGraph)\b/;
    const offenders = [];
    for (const file of productionSourceFiles()) {
      const text = fs.readFileSync(file, 'utf8');
      const rel = path.relative(srcDir, file);
      if (legacyGlobal.test(text) && !allowed.has(rel)) {
        offenders.push(rel);
      }
    }

    expect(offenders).toEqual([]);
  });
});
