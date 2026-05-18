import { expect, test } from '@playwright/test';
import fs from 'node:fs/promises';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const webuiDir = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..');
const repoRoot = path.resolve(webuiDir, '../..');
const clientServerPath = path.join(repoRoot, 'ivy-lang-examples/doc/examples/client_server_example.ivy');

async function openFreshIvy(page) {
  await page.goto('/');
  await page.evaluate(() => {
    localStorage.clear();
    sessionStorage.clear();
  });
  await page.reload();
  await expect(page.locator('#menubar')).toBeVisible();
  await page.waitForFunction(() => (
    window.__ivyDiagnostics &&
    window.__ivyDiagnostics.runtime() &&
    window.__ivyDiagnostics.runtime().api &&
    window.__ivyDiagnostics.runtime().api.sessionId
  ), null, { timeout: 60_000 });
}

test('browser wasm induction accepts strengthened client/server invariant', async ({ page }) => {
  test.setTimeout(120_000);
  const base = await fs.readFile(clientServerPath, 'utf8');
  const strengthened = `${base}

private {
    invariant ~(link(X,Y) & semaphore(Y))
}
`;

  await openFreshIvy(page);

  const result = await page.evaluate(async (content) => {
    const app = window.__ivyDiagnostics.runtime();
    if (app.jobSubmissionMode !== 'browser' && typeof app._switchJobSubmissionBackend === 'function') {
      await app._switchJobSubmissionBackend('browser');
    }
    if (app.cmEditor && typeof app.cmEditor.setValue === 'function') {
      app.cmEditor.setValue(content);
    }
    app._persistedFileName = 'client_server_example.ivy';
    app._persistedFileContent = content;
    const load = await app.api.reloadContent(content, 'client_server_example.ivy', {
      isolate: app.activeIsolate || '',
    });
    if (typeof app.setIsolates === 'function') {
      app.setIsolates(load && load.isolates ? load.isolates : [], load && load.isolate ? load.isolate : '');
    }
    return app.api.runCheck('induction');
  }, strengthened);

  expect(result).toMatchObject({
    status: 'ok',
    result: 'pass',
    mode: 'induction',
  });
});
