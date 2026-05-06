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
});
