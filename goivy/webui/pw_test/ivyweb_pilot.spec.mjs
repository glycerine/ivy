import { expect, test } from '@playwright/test';

async function openIvy(page) {
  const consoleErrors = [];
  page.on('console', (msg) => {
    if (msg.type() === 'error') {
      consoleErrors.push(msg.text());
    }
  });
  page.on('pageerror', (err) => {
    consoleErrors.push(err.message);
  });

  await page.goto('/');
  await expect(page).toHaveTitle(/ivy/i);
  await expect(page.locator('#menubar')).toBeVisible();
  await page.waitForFunction(() => window.ivyApp && window.ivyApp.api);
  return consoleErrors;
}

async function createSession(request) {
  const response = await request.post('/api/session/new');
  expect(response.ok()).toBe(true);
  const body = await response.json();
  expect(body.session_id).toBeTruthy();
  return body.session_id;
}

async function loadExampleIntoCurrentSession(page) {
  const ivyContent = `#lang ivy1.7
type client
type server
relation link(X:client, Y:server)
relation semaphore(X:server)
after init { semaphore(W) := true; link(X,Y) := false }
action connect(x:client,y:server) = { require semaphore(y); link(x,y) := true; semaphore(y) := false }
export connect
`;

  return page.evaluate(async (content) => {
    const sid = window.ivyApp.api.sessionId;
    const formData = new FormData();
    const blob = new Blob([content], { type: 'text/plain' });
    formData.append('file', blob, 'test.ivy');
    const loadResponse = await fetch(`/api/session/${sid}/load`, {
      method: 'POST',
      body: formData,
    });
    const loadBody = await loadResponse.json();
    const conceptResponse = await fetch(`/api/session/${sid}/concept`);
    const concept = await conceptResponse.json();
    if (window.ivyApp && window.ivyApp.conceptGraph) {
      window.ivyApp.conceptGraph.update(concept.elements, concept.positions);
    }
    return { loadBody, concept };
  }, ivyContent);
}

test('page loads', async ({ page }) => {
  const consoleErrors = await openIvy(page);

  await expect(page.locator('#arg-panel')).toBeVisible();
  await expect(page.locator('#concept-panel')).toBeVisible();
  expect(consoleErrors).toEqual([]);
});

test('title contains Ivy', async ({ page }) => {
  await openIvy(page);

  await expect(page).toHaveTitle(/ivy/i);
});

test('ARG and concept panels are present', async ({ page }) => {
  await openIvy(page);

  await expect(page.locator('#arg-panel, #arg-graph').first()).toBeVisible();
  await expect(page.locator('#concept-panel, #concept-graph').first()).toBeVisible();
});

test('menubar has expected top-level labels', async ({ page }) => {
  await openIvy(page);

  const menubar = page.locator('#menubar');
  await expect(menubar).toBeVisible();
  await expect(menubar).toContainText('File');
  await expect(menubar).toContainText('Mode');
  await expect(menubar).toContainText('Check');
});

test('status and info panels are present', async ({ page }) => {
  await openIvy(page);

  await expect(page.locator('#statusbar')).toBeVisible();
  await expect(page.locator('#info-panel')).toBeVisible();
});

test('browser fetch can create a session', async ({ page }) => {
  await openIvy(page);

  const sessionID = await page.evaluate(async () => {
    const response = await fetch('/api/session/new', { method: 'POST' });
    const body = await response.json();
    return body.session_id || '';
  });
  expect(sessionID).toBeTruthy();
});

test('browser fetch can get ARG JSON', async ({ page, request }) => {
  const sid = await createSession(request);
  await openIvy(page);

  const body = await page.evaluate(async (sessionID) => {
    const response = await fetch(`/api/session/${sessionID}/arg`);
    return response.json();
  }, sid);
  expect(body).toHaveProperty('elements');
});

test('browser fetch can get concept JSON', async ({ page, request }) => {
  const sid = await createSession(request);
  await openIvy(page);

  const body = await page.evaluate(async (sessionID) => {
    const response = await fetch(`/api/session/${sessionID}/concept`);
    return response.json();
  }, sid);
  expect(body).toBeTruthy();
});

test('mode select can change to abstract', async ({ page }) => {
  await openIvy(page);

  const mode = page.locator('#mode-select');
  await expect(mode).toBeVisible();
  await mode.selectOption('abstract');
  await expect(mode).toHaveValue('abstract');
});

test('graph health check passes', async ({ page }) => {
  await openIvy(page);

  await page.waitForFunction(
    () => window['__ivyGraphHealthy_arg-graph'] && window['__ivyGraphHealthy_concept-graph'],
    null,
    { timeout: 5_000 },
  );
  await expect(page.locator('#arg-graph')).toBeVisible();
  await expect(page.locator('#concept-graph')).toBeVisible();
  expect(await page.evaluate(() => window.__ivyInitError || '')).toBe('');
});

test('undo button path does not kill the page', async ({ page }) => {
  await openIvy(page);

  await page.locator('#btn-undo').click();
  await expect(page).toHaveTitle(/ivy/i);
});

test('Cytoscape is loaded', async ({ page }) => {
  await openIvy(page);

  expect(await page.evaluate(() => typeof cytoscape)).toBe('function');
});

test('ARG graph initializes a Cytoscape container', async ({ page }) => {
  await openIvy(page);

  const hasCy = await page.evaluate(() => {
    const el = document.getElementById('arg-graph');
    return !!el && (el.querySelector('canvas') !== null || el.children.length > 0);
  });
  expect(hasCy).toBe(true);
});

test('context menu starts hidden', async ({ page }) => {
  await openIvy(page);

  await expect(page.locator('#context-menu')).toHaveCSS('display', 'none');
});

test('right-click context menu path does not crash', async ({ page }) => {
  await openIvy(page);

  await expect(page.locator('#context-menu')).toHaveCount(1);
  await page.locator('#arg-graph').dispatchEvent('contextmenu', {
    bubbles: true,
    clientX: 50,
    clientY: 50,
  });
  expect(await page.evaluate(() => window.__ivyInitError || '')).toBe('');
});

test('divider exists', async ({ page }) => {
  await openIvy(page);

  await expect(page.locator('#divider')).toBeVisible();
});

test('check button reports a visible status', async ({ page }) => {
  await openIvy(page);

  await page.locator('#btn-check').click();
  await expect(page.locator('#statusbar')).not.toHaveText('');
});

test('SSE connection receives check events', async ({ page, request }) => {
  const sid = await createSession(request);
  await openIvy(page);

  await page.evaluate((sessionID) => {
    window._sseEvents = [];
    window._es = new EventSource(`/api/session/${sessionID}/events`);
    window._es.onmessage = (event) => {
      window._sseEvents.push(event.data);
    };
  }, sid);

  const response = await request.post(`/api/session/${sid}/check`);
  expect(response.ok()).toBe(true);
  await page.waitForFunction(() => window._sseEvents.length > 0, null, { timeout: 5_000 });
  await page.evaluate(() => {
    if (window._es) window._es.close();
  });
});

test('invalid session returns 404', async ({ page }) => {
  await openIvy(page);

  const status = await page.evaluate(async () => {
    const response = await fetch('/api/session/nonexistent/arg');
    return response.status;
  });
  expect(status).toBe(404);
});

test('static CSS is loaded', async ({ page }) => {
  await openIvy(page);

  const bg = await page.locator('#menubar').evaluate((el) => getComputedStyle(el).backgroundColor);
  expect(bg).toBeTruthy();
  expect(bg).toContain('45, 45, 45');
});

test('concept graph right-click menu path does not crash', async ({ page }) => {
  await openIvy(page);
  await loadExampleIntoCurrentSession(page);

  const nodeCount = await page.evaluate(() => {
    if (window.ivyApp && window.ivyApp.conceptGraph && window.ivyApp.conceptGraph.cy) {
      return window.ivyApp.conceptGraph.cy.nodes().length;
    }
    return 0;
  });

  if (nodeCount > 0) {
    const menuVisible = await page.evaluate(() => {
      const cy = window.ivyApp.conceptGraph.cy;
      const node = cy.nodes()[0];
      const pos = node.renderedPosition();
      node.emit('cxttap', { renderedPosition: pos });
      const cm = document.getElementById('context-menu');
      return !!cm && cm.style.display !== 'none';
    });
    expect(menuVisible).toBe(true);
    await expect(page.locator('#context-menu')).toContainText(/Splatter|Materialize|Remove/);
  }

  expect(await page.evaluate(() => window.__ivyInitError || '')).toBe('');
});

test('static JS globals are loaded', async ({ page }) => {
  await openIvy(page);

  const globals = await page.evaluate(() => ({
    api: typeof IvyAPI,
    app: typeof IvyApp,
  }));
  expect([globals.api, globals.app]).toContain('function');
});

test('panel resize path does not crash', async ({ page }) => {
  await openIvy(page);

  const before = await page.locator('#arg-panel').boundingBox();
  test.skip(!before || before.width <= 0, 'arg panel has no measurable width');

  const divider = await page.locator('#divider').boundingBox();
  test.skip(!divider, 'divider has no measurable box');

  const startX = divider.x + divider.width / 2;
  const startY = divider.y + divider.height / 2;
  await page.mouse.move(startX, startY);
  await page.mouse.down();
  await page.mouse.move(startX + 100, startY);
  await page.mouse.up();

  const after = await page.locator('#arg-panel').boundingBox();
  expect(after?.width || 0).toBeGreaterThan(0);
  expect(await page.evaluate(() => window.__ivyInitError || '')).toBe('');
});
