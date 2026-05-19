import { expect, test } from '@playwright/test';
import { readFileSync } from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const webuiDir = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..');
const ivyRoot = path.resolve(webuiDir, '../..');
const ordLivePath = path.join(ivyRoot, 'ivy-lang-examples', 'doc', 'examples', 'apple', 'ord_live.ivy');
const ordLiveContent = readFileSync(ordLivePath, 'utf8');
const clientServerIvyContent = `#lang ivy1.7

type client
type server

relation link(X:client, Y:server)
relation semaphore(X:server)

after init {
    link(X,Y) := false;
    semaphore(Y) := true
}

action connect(x:client, y:server) = {
    require semaphore(y);
    link(x,y) := true;
    semaphore(y) := false
}

action disconnect(x:client, y:server) = {
    require link(x,y);
    link(x,y) := false;
    semaphore(y) := true
}

export connect
export disconnect

conjecture link(X,Y) -> ~semaphore(Y)
conjecture ~link(X,Y) | ~link(X,Z) | Y = Z
`;

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
  await page.waitForFunction(() => window.__ivyDiagnostics && window.__ivyDiagnostics.runtime() && window.__ivyDiagnostics.runtime().api);
  return consoleErrors;
}

async function openIvyWithSavedSession(page, savedState) {
  await page.addInitScript((state) => {
    localStorage.clear();
    localStorage.setItem('ivy_last_session', state.sessionId);
    localStorage.setItem('ivy_sessions', JSON.stringify([state.sessionId]));
    localStorage.setItem(`ivy_sess_${state.sessionId}`, JSON.stringify(state));
  }, savedState);
  return openIvy(page);
}

async function createSession(request) {
  const response = await request.post('/api/session/new');
  expect(response.ok()).toBe(true);
  const body = await response.json();
  expect(body.session_id).toBeTruthy();
  return body.session_id;
}

async function loadExampleIntoCurrentSession(page) {
  return page.evaluate(async (content) => {
    const sid = window.__ivyDiagnostics.runtime().api.sessionId;
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
    if (window.__ivyDiagnostics.runtime() && window.__ivyDiagnostics.runtime().conceptGraph) {
      window.__ivyDiagnostics.runtime().conceptGraph.update(concept.elements, concept.positions);
    }
    return { loadBody, concept };
  }, clientServerIvyContent);
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
  await expect(page.locator('#ui-mode-select')).toBeVisible();
  await expect(page.locator('#ui-mode-select')).toHaveValue('cti');
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

test('workflow mode selector switches CTI and reachability controls', async ({ page }) => {
  await openIvy(page);

  await expect(page.locator('#mode-select')).toBeHidden();
  await expect(page.locator('#btn-check')).toBeHidden();
  await expect(page.locator('#btn-show-reachable')).toBeHidden();
  await expect(page.locator('#btn-undo')).toBeHidden();
  await expect(page.locator('#btn-reset-domain')).toBeHidden();
  await expect(page.locator('#btn-diagram-domain')).toBeHidden();
  await expect(page.locator('[data-dropdown="conj-menu"]')).toBeVisible();
  await expect(page.locator('[data-dropdown="reach-action-menu"]')).toBeHidden();

  await page.locator('#ui-mode-select').selectOption('reachability');

  await expect(page.locator('#mode-select')).toBeVisible();
  await expect(page.locator('#btn-check')).toBeVisible();
  await expect(page.locator('#btn-show-reachable')).toBeVisible();
  await expect(page.locator('[data-dropdown="conj-menu"]')).toBeHidden();
  await expect(page.locator('[data-dropdown="reach-action-menu"]')).toBeVisible();
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

test('concept graph stays fitted after tutorial hide/show', async ({ page }) => {
  await openIvy(page);
  await loadExampleIntoCurrentSession(page);
  await page.waitForFunction(() => window.__ivyDiagnostics.runtime().conceptGraph.cy.nodes().length > 0);

  await page.locator('#btn-toggle-tutorial').click();
  await expect(page.locator('#tutorial-container')).toBeHidden();
  await page.locator('#btn-toggle-tutorial').click();
  await expect(page.locator('#tutorial-container')).toBeVisible();
  await page.waitForTimeout(180);

  const box = await page.evaluate(() => {
    const cy = window.__ivyDiagnostics.runtime().conceptGraph.cy;
    const nodes = cy.nodes().filter((node) => node.visible());
    const bounds = nodes.renderedBoundingBox({ includeLabels: false, includeOverlays: false });
    return {
      count: nodes.length,
      width: cy.width(),
      height: cy.height(),
      x1: bounds.x1,
      y1: bounds.y1,
      x2: bounds.x2,
      y2: bounds.y2,
    };
  });

  expect(box.count).toBeGreaterThan(0);
  expect(box.x1).toBeGreaterThanOrEqual(-4);
  expect(box.y1).toBeGreaterThanOrEqual(-4);
  expect(box.x2).toBeLessThanOrEqual(box.width + 4);
  expect(box.y2).toBeLessThanOrEqual(box.height + 4);
});

test('conjecture undo menu path does not kill the page', async ({ page }) => {
  await openIvy(page);

  await page.locator('[data-dropdown="conj-menu"]').click();
  await expect(page.locator('#conj-cti-bounded-check')).toBeVisible();
  await page.locator('#conj-undo').click();
  await expect(page).toHaveTitle(/ivy/i);
});

test('panel dropdown menus render above the details pane', async ({ page }) => {
  await openIvy(page);

  await page.evaluate(() => {
    const app = window.__ivyDiagnostics.runtime();
    const sheetLeft = document.querySelector('.sheet-left');
    const sheetMain = sheetLeft?.querySelector('.sheet-main');
    const infoPanel = sheetLeft?.querySelector('.info-panel');
    const minMainHeight = app._detailsResizerMinimumMainHeight(sheetLeft);
    if (sheetMain) {
      sheetMain.style.flex = `0 0 ${minMainHeight}px`;
      sheetMain.style.minHeight = `${minMainHeight}px`;
    }
    if (infoPanel) {
      infoPanel.style.flex = '1 1 auto';
      infoPanel.style.height = '';
    }
  });

  await page.locator('[data-dropdown="conj-menu"]').click();
  await expect(page.locator('#conj-menu')).toBeVisible();

  const result = await page.evaluate(() => {
    const menu = document.getElementById('conj-menu');
    const infoPanel = document.getElementById('info-panel');
    const menuBox = menu.getBoundingClientRect();
    const infoBox = infoPanel.getBoundingClientRect();
    const x = menuBox.left + 24;
    const y = Math.min(menuBox.bottom - 6, Math.max(infoBox.top + 8, menuBox.top + 8));
    const top = document.elementFromPoint(x, y);
    return {
      menuBottom: menuBox.bottom,
      detailsTop: infoBox.top,
      insideMenu: !!top?.closest('#conj-menu'),
      insideDetails: !!top?.closest('#info-panel'),
    };
  });

  expect(result.menuBottom).toBeGreaterThan(result.detailsTop);
  expect(result.insideMenu).toBe(true);
  expect(result.insideDetails).toBe(false);
});

test('shared action runner reports backend errors', async ({ page }) => {
  await openIvy(page);

  const result = await page.evaluate(async () => {
    return window.__ivyDiagnostics.runtime().runAction('definitely_not_a_real_action');
  });
  expect(result.ok).toBe(false);
  await expect(page.locator('#statusbar')).toContainText('Action failed');
  await expect(page.locator('#statusbar')).toHaveClass(/error/);
});

test('dialog primitives accept integer and list selections', async ({ page }) => {
  await openIvy(page);

  const intPromise = page.evaluate(async () => {
    return window.__ivyDiagnostics.runtime().integerDialog('Bound', 'choose bound', 1, { min: 1, max: 9 });
  });
  await page.locator('[data-ivy-dialog-int]').fill('4');
  await page.getByRole('button', { name: 'OK' }).click();
  await expect(intPromise).resolves.toBe(4);

  const listPromise = page.evaluate(async () => {
    return window.__ivyDiagnostics.runtime().listboxDialog('Pick', 'choose one', ['alpha', 'beta']);
  });
  await page.locator('[data-ivy-dialog-list]').selectOption('beta');
  await page.getByRole('button', { name: 'OK' }).click();
  await expect(listPromise).resolves.toBe('beta');
});

test('Go-supplied concept menu descriptor renders and dispatches', async ({ page }) => {
  await openIvy(page);

  const menuRoot = page.locator('[data-dynamic-menu-region="concept"]');
  await expect(menuRoot).toBeVisible();
  await menuRoot.locator('.panel-menu', { hasText: 'Action' }).click();
  await menuRoot.locator('[data-menu-action="undo"]').click();

  await expect(page.locator('#statusbar')).toContainText('Done: undo');
});

test('clicking ARG nodes reloads concept graph for the selected state', async ({ page }) => {
  await openIvy(page);

  await page.evaluate(() => {
    const app = window.__ivyDiagnostics.runtime();
    app.api.getConceptGraph = async (nodeId) => ({
      selected_node: nodeId,
      state_label: nodeId === 'state_0' ? '0' : '1',
      elements: [
        {
          group: 'nodes',
          data: {
            id: 'concept-node',
            obj: 'concept-' + nodeId,
            label: 'concept ' + nodeId,
            short_info: nodeId,
            long_info: nodeId,
            shape: 'octagon',
            width: 120,
            height: 60,
          },
          classes: 'node_unknown',
        },
      ],
    });
    app.argGraph.update([
      {
        group: 'nodes',
        data: {
          id: 'arg-state-0',
          obj: 'state_0',
          label: '0',
          short_info: 'state 0',
          long_info: 'state 0',
          shape: 'ellipse',
          width: 60,
          height: 50,
        },
        classes: 'state',
      },
      {
        group: 'nodes',
        data: {
          id: 'arg-state-1',
          obj: 'state_1',
          label: '1',
          short_info: 'state 1',
          long_info: 'state 1',
          shape: 'ellipse',
          width: 60,
          height: 50,
        },
        classes: 'state',
      },
    ]);
  });

  await page.evaluate(() => {
    const node = window.__ivyDiagnostics.runtime().argGraph.cy.nodes().toArray().find((n) => n.data('obj') === 'state_0');
    node.emit('tap', { target: node });
  });
  await page.waitForFunction(() => {
    const nodes = window.__ivyDiagnostics.runtime().conceptGraph.cy.nodes();
    return nodes.length > 0 && nodes[0].data('label') === 'concept state_0';
  });
  await expect(page.locator('#state-label')).toContainText('State: 0');

  await page.evaluate(() => {
    const node = window.__ivyDiagnostics.runtime().argGraph.cy.nodes().toArray().find((n) => n.data('obj') === 'state_1');
    node.emit('tap', { target: node });
  });
  await page.waitForFunction(() => {
    const nodes = window.__ivyDiagnostics.runtime().conceptGraph.cy.nodes();
    return nodes.length > 0 && nodes[0].data('label') === 'concept state_1';
  });
  await expect(page.locator('#state-label')).toContainText('State: 1');
});

test('backend relation toggle controls concept edge rendering and survives refresh', async ({ page }) => {
  await openIvy(page);
  const loaded = await loadExampleIntoCurrentSession(page);
  await page.evaluate((concept) => {
    window.__ivyDiagnostics.runtime().populateStateCheckboxes(concept);
  }, loaded.concept);

  expect(await page.evaluate(() => window.__ivyDiagnostics.runtime().conceptGraph.cy.edges().length)).toBe(0);

  const linkRow = page.locator('#state-checkbox-body tr', { hasText: 'link' });
  await expect(linkRow).toBeVisible();
  await linkRow.locator('input[type="checkbox"]').nth(1).check();

  await page.waitForFunction(() => {
    return window.__ivyDiagnostics.runtime().conceptGraph.cy.edges().toArray().some((e) => e.data('obj') === 'link');
  });

  await page.evaluate(async () => {
    const concept = await window.__ivyDiagnostics.runtime().api.getConceptGraph();
    window.__ivyDiagnostics.runtime().conceptGraph.update(concept.elements, concept.positions);
    window.__ivyDiagnostics.runtime().populateStateCheckboxes(concept);
  });

  await expect(page.locator('#state-checkbox-body tr', { hasText: 'link' }).locator('input[type="checkbox"]').nth(1)).toBeChecked();
  expect(await page.evaluate(() => window.__ivyDiagnostics.runtime().conceptGraph.cy.edges().toArray().filter((e) => e.data('obj') === 'link').length)).toBeGreaterThan(0);
});

test('ord_live load keeps state relations visible five seconds after load completes', async ({ page }) => {
  test.setTimeout(60_000);
  await openIvy(page);

  await page.locator('#ui-mode-select').selectOption('reachability');
  await expect(page.locator('#statusbar')).toContainText('Ready');
  await page.evaluate(async (content) => {
    const file = new File([content], 'client_server_example.ivy', { type: 'text/plain' });
    await window.__ivyDiagnostics.runtime().loadFile(file);
  }, clientServerIvyContent);
  await expect(page.locator('#statusbar')).toContainText('Loaded: client_server_example.ivy', { timeout: 20_000 });
  await page.locator('#btn-show-reachable').click();
  await expect(page.locator('#statusbar')).toContainText('Reachable states opened', { timeout: 20_000 });
  await page.waitForFunction(() => window.__ivyDiagnostics.runtime().activeSheetId !== 'sheet-1');

  await page.evaluate(async (content) => {
    const file = new File([content], 'ord_live.ivy', { type: 'text/plain' });
    await window.__ivyDiagnostics.runtime().loadFile(file);
  }, ordLiveContent);

  await expect(page.locator('#statusbar')).toContainText('Loaded: ord_live.ivy', { timeout: 45_000 });
  await expect(page.locator('#state-panel')).toBeVisible();
  await page.waitForFunction(() => document.querySelectorAll('#state-checkbox-body tr').length > 1);

  await page.waitForTimeout(5_000);

  await expect(page.locator('#statusbar')).toContainText('Loaded: ord_live.ivy');
  await expect(page.locator('#state-panel')).toBeVisible();
  await expect(page.locator('#state-checkbox-body tr').first()).not.toContainText('No relations loaded');
  expect(await page.locator('#state-checkbox-body tr').count()).toBeGreaterThan(1);
});

test('constraint facts render below the graph and toggle through backend action', async ({ page }) => {
  await openIvy(page);

  await page.evaluate(() => {
    window._factActions = [];
    window.__ivyDiagnostics.runtime().api.executeAction = async (action, args) => {
      window._factActions.push({ action, args });
      return { status: 'ok' };
    };
    window.__ivyDiagnostics.runtime().populateStateCheckboxes({
      relations: [],
      facts: [
        { index: 0, text: 'link(a,b)', selected: true },
        { index: 1, text: 'semaphore(b)', selected: true },
      ],
      toggles: { edges: {}, labels: {} },
    });
  });

  await expect(page.locator('[data-constraint-fact="0"]')).toBeVisible();
  await expect(page.locator('[data-constraint-fact="0"]')).toHaveText('link(a,b)');
  await page.locator('[data-constraint-fact="0"]').click();
  await expect(page.locator('[data-constraint-fact="0"]')).toHaveClass(/inactive/);
  expect(await page.evaluate(() => window._factActions)).toEqual([
    { action: 'set_fact_selection', args: { index: 0, selected: false } },
  ]);
});

test('View Source edge action loads source text and highlights the backend line', async ({ page }) => {
  await openIvy(page);

  const result = await page.evaluate(async () => {
    window.__ivyDiagnostics.runtime().api.argNodeAction = async () => {
      window.__ivyDiagnostics.runtime()._testArgNodeActionResult = {
        status: 'ok',
        file: 'sample.ivy',
        lineno: 2,
        source: 'line1\naction go = {}\nline3\n',
      };
      return window.__ivyDiagnostics.runtime()._testArgNodeActionResult;
    };
    const originalScroll = window.__ivyDiagnostics.runtime().scrollEditorToLine.bind(window.__ivyDiagnostics.runtime());
    window.__ivyDiagnostics.runtime().scrollEditorToLine = (lineno) => {
      window.__ivyDiagnostics.runtime()._testScrollLine = lineno;
      return originalScroll(lineno);
    };
    const originalSetEditor = window.__ivyDiagnostics.runtime().setEditorContent.bind(window.__ivyDiagnostics.runtime());
    window.__ivyDiagnostics.runtime().setEditorContent = (source) => {
      window.__ivyDiagnostics.runtime()._testSetEditorSource = source;
      return originalSetEditor(source);
    };
    await window.__ivyDiagnostics.runtime().executeArgEdgeAction({ source_obj: 'state_0', target_obj: 'state_1' }, 'view_source');
    return {
      value: window.__ivyDiagnostics.runtime().cmEditor.getValue(),
      highlightedLine: window.__ivyDiagnostics.runtime()._highlightedEditorLine,
      scrollLine: window.__ivyDiagnostics.runtime()._testScrollLine,
      setEditorSource: window.__ivyDiagnostics.runtime()._testSetEditorSource,
      apiResult: window.__ivyDiagnostics.runtime()._testArgNodeActionResult,
      status: document.getElementById('statusbar').textContent,
      details: document.getElementById('info-content').textContent,
    };
  });

  expect(result.value).toBe('line1\naction go = {}\nline3\n');
  expect(result.setEditorSource).toBe('line1\naction go = {}\nline3\n');
  expect(result.apiResult.lineno).toBe(2);
  expect(result.status).toContain('Done: view_source');
  expect(result.scrollLine).toBe(2);
  expect(result.highlightedLine).toBe(2);
  expect(result.details).toContain('sample.ivy line 2');
});

test('sheet graph instances are owned independently when switching tabs', async ({ page }) => {
  await openIvy(page);

  const result = await page.evaluate(() => {
    const app = window.__ivyDiagnostics.runtime();
    const mainElement = {
      group: 'nodes',
      data: { id: 'main-node', obj: 'main', label: 'main' },
    };
    const secondElement = {
      group: 'nodes',
      data: { id: 'second-node', obj: 'second', label: 'second' },
    };
    const updatedMainElement = {
      group: 'nodes',
      data: { id: 'main-node-2', obj: 'main2', label: 'main2' },
    };

    app.argGraph.update([mainElement]);
    const sheetId = app.addSheet('Sheet 2');
    const firstGraph = app.sheets && app.sheets['sheet-1'] && app.sheets['sheet-1'].argGraph;
    const secondGraph = app.sheets && app.sheets[sheetId] && app.sheets[sheetId].argGraph;
    const activeAfterAdd = app.argGraph;

    if (app.argGraph) {
      app.argGraph.update([secondElement]);
    }
    app.switchSheet('sheet-1');
    if (app.argGraph) {
      app.argGraph.update([updatedMainElement]);
    }

    return {
      distinct: !!firstGraph && !!secondGraph && firstGraph !== secondGraph && activeAfterAdd === secondGraph,
      activeSheet: app.activeSheetId || '',
      firstLabels: firstGraph ? firstGraph.cy.nodes().map((n) => n.data('label')) : [],
      secondLabels: secondGraph ? secondGraph.cy.nodes().map((n) => n.data('label')) : [],
    };
  });

  expect(result.distinct).toBe(true);
  expect(result.activeSheet).toBe('sheet-1');
  expect(result.firstLabels).toEqual(['main2']);
  expect(result.secondLabels).toEqual(['second']);
});

test('event trace sheets render through the controller', async ({ page }) => {
  await openIvy(page);

  await page.evaluate(() => {
    window.__ivyDiagnostics.runtime().openEventTraceSheet('Trace', {
      sheet_id: 'events-1',
      events: [
        { text: 'root(a)', address: '0', subs: [{ text: 'child(a)', address: '0/0' }] },
      ],
      patterns: ['root(a)'],
    }, 'events-1');
  });

  await expect(page.locator('.sheet-tab[data-sheet="events-1"]')).toContainText('Trace');
  await expect(page.locator('#events-1 [data-event-address="0"]')).toContainText('root(a)');
  await expect(page.locator('#events-1 [data-event-address="0/0"]')).toHaveCount(0);

  await page.locator('#events-1 [data-event-toggle="0"]').click();
  await expect(page.locator('#events-1 [data-event-address="0/0"]')).toContainText('child(a)');

  await page.locator('#events-1 .event-pattern-list').selectOption('root(a)');
  const selectedPattern = await page.evaluate(() => window.__ivyDiagnostics.runtime().selectedEventPattern('events-1'));
  expect(selectedPattern).toBe('root(a)');
});

test('Step in opens a backend-owned sheet whose node clicks load that sheet concept graph', async ({ page }) => {
  await openIvy(page);

  const result = await page.evaluate(async () => {
    const app = window.__ivyDiagnostics.runtime();
    window._stepInCalls = { arg: [], concept: [] };
    app.api.argNodeAction = async (node, action, args) => {
      window._stepInCalls.arg.push({ node, action, args });
      return {
        status: 'ok',
        decomposed: true,
        sheet_id: 'sheet-2',
        sub_arg: {
          elements: [
            { group: 'nodes', data: { id: 'state_0', obj: 'state_0', label: '0' } },
            { group: 'nodes', data: { id: 'state_1', obj: 'state_1', label: '1' } },
            { group: 'edges', data: { id: 'step-edge', source: 'state_0', target: 'state_1', label: 'sub' } },
          ],
        },
      };
    };
    app.api.getConceptGraph = async (node, sheet) => {
      window._stepInCalls.concept.push({ node, sheet });
      return {
        selected_node: node,
        sheet_id: sheet,
        elements: [
          { group: 'nodes', data: { id: 'concept-' + sheet + '-' + node, label: sheet + ':' + node } },
        ],
        toggles: { edges: {}, labels: {} },
        facts: [],
      };
    };

    await app.executeArgEdgeAction({
      source_obj: 'state_0',
      target_obj: 'state_1',
      label: 'call ext',
    }, 'decompose', 'sheet-1');

    const sheet = app.sheets['sheet-2'];
    const node = sheet.argGraph.cy.getElementById('state_1');
    node.emit('tap', { target: node });
    await new Promise((resolve) => setTimeout(resolve, 0));

    return {
      activeSheet: app.activeSheetId,
      selected: sheet.selectedArgNode,
      argCalls: window._stepInCalls.arg,
      conceptCalls: window._stepInCalls.concept,
      conceptLabels: sheet.conceptGraph.cy.nodes().map((n) => n.data('label')),
      rootConceptLabels: app.sheets['sheet-1'].conceptGraph.cy.nodes().map((n) => n.data('label')),
    };
  });

  expect(result.activeSheet).toBe('sheet-2');
  expect(result.selected).toBe('state_1');
  expect(result.argCalls).toEqual([
    { node: 'state_0', action: 'decompose', args: { target: 'state_1', sheet_id: 'sheet-1' } },
  ]);
  expect(result.conceptCalls).toEqual([
    { node: 'state_1', sheet: 'sheet-2' },
  ]);
  expect(result.conceptLabels).toEqual(['sheet-2:state_1']);
  expect(result.rootConceptLabels).toEqual([]);
});

test('ARG node execute actions are rendered from backend descriptors and dispatch action args', async ({ page }) => {
  await openIvy(page);

  await page.evaluate(() => {
    const app = window.__ivyDiagnostics.runtime();
    window._executeActionCalls = [];
    app.api.argNodeAction = async (node, action, args) => {
      window._executeActionCalls.push({ node, action, args });
      return {
        status: 'ok',
        arg: {
          elements: [
            { group: 'nodes', data: { id: 'state_0', obj: 'state_0', label: '0' } },
            { group: 'nodes', data: { id: 'state_1', obj: 'state_1', label: '1' } },
            { group: 'edges', data: { id: 'edge_0_1', source: 'state_0', target: 'state_1', label: 'ext:connect' } },
          ],
        },
      };
    };
    app.argGraph.update([
      {
        group: 'nodes',
        data: {
          id: 'state_0',
          obj: 'state_0',
          label: '0',
          actions: [
            { label: 'Execute action:', action: '' },
            { label: '---', action: '' },
            { label: 'ext:connect', action: 'execute_action', args: { action_name: 'ext:connect', action_label: 'ext:connect' } },
          ],
        },
      },
    ]);
    app.onArgNodeRightClick(app.argGraph.cy.getElementById('state_0').data(), { x: 12, y: 12 }, 'sheet-1');
  });

  await expect(page.locator('.context-menu-header', { hasText: 'Execute action:' })).toBeVisible();
  await page.locator('.context-menu-item', { hasText: 'ext:connect' }).click();
  const calls = await page.evaluate(() => window._executeActionCalls);
  expect(calls).toEqual([
    { node: 'state_0', action: 'execute_action', args: { action_name: 'ext:connect', action_label: 'ext:connect', sheet_id: 'sheet-1' } },
  ]);
});

test('failed check result can open its trace ARG in a sheet', async ({ page }) => {
  await openIvy(page);

  await page.evaluate(() => {
    window.__ivyDiagnostics.runtime().showCheckResult({
      result: 'fail',
      z3_contacted: true,
      message: 'The node is unsafe: View error trace?',
      trace_arg: {
        elements: [
          { group: 'nodes', data: { id: 'state_0', obj: 'state_0', label: '0' } },
          { group: 'nodes', data: { id: 'state_1', obj: 'state_1', label: '1' } },
          { group: 'edges', data: { id: 'trace_edge', source: 'state_0', target: 'state_1', label: 'trace' } },
        ],
      },
    });
  });

  await expect(page.locator('[data-check-view-trace]')).toBeVisible();
  await page.locator('[data-check-view-trace]').click();
  const result = await page.evaluate(() => {
    const sheet = window.__ivyDiagnostics.runtime().sheets[window.__ivyDiagnostics.runtime().activeSheetId];
    return {
      activeSheet: window.__ivyDiagnostics.runtime().activeSheetId,
      labels: sheet.argGraph.cy.nodes().map((n) => n.data('label')),
      edgeLabels: sheet.argGraph.cy.edges().map((e) => e.data('label')),
    };
  });

  expect(result.activeSheet).toBe('trace-1');
  expect(result.labels).toEqual(['0', '1']);
  expect(result.edgeLabels).toEqual(['trace']);
});

test('Show Reachable command opens a reachable-state ARG sheet', async ({ page }) => {
  await openIvy(page);

  await page.evaluate(() => {
    window.__ivyDiagnostics.runtime().api.executeAction = async (action, args) => {
      window._showReachableCall = { action, args };
      return {
        status: 'ok',
        sheet_id: 'sheet-2',
        arg: {
          elements: [
            { group: 'nodes', data: { id: 'state_0', obj: 'state_0', label: '0' } },
          ],
        },
      };
    };
  });

  await page.evaluate(async () => {
    await window.__ivyDiagnostics.runtime().showReachableStates();
  });
  const result = await page.evaluate(() => {
    const sheet = window.__ivyDiagnostics.runtime().sheets[window.__ivyDiagnostics.runtime().activeSheetId];
    return {
      call: window._showReachableCall,
      activeSheet: window.__ivyDiagnostics.runtime().activeSheetId,
      labels: sheet.argGraph.cy.nodes().map((n) => n.data('label')),
      status: document.getElementById('statusbar').textContent,
    };
  });

  expect(result.call).toEqual({ action: 'show_reachable', args: {} });
  expect(result.activeSheet).toBe('sheet-2');
  expect(result.labels).toEqual(['0']);
  expect(result.status).toContain('Reachable states opened');
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

test('controller context menu stays inside the viewport', async ({ page }) => {
  await openIvy(page);

  await page.evaluate(() => {
    window.__ivyDiagnostics.runtime().controls.showContextMenu(window.innerWidth - 4, window.innerHeight - 4, [
      { name: 'Near edge action', id: 'near-edge-action', callback: () => {} },
    ]);
  });
  const menu = page.locator('#context-menu');
  await expect(menu).toBeVisible();
  await expect(menu.locator('[data-action-id="near-edge-action"]')).toBeVisible();

  const box = await menu.boundingBox();
  const viewport = page.viewportSize();
  expect(box.x).toBeGreaterThanOrEqual(0);
  expect(box.y).toBeGreaterThanOrEqual(0);
  expect(box.x + box.width).toBeLessThanOrEqual(viewport.width + 1);
  expect(box.y + box.height).toBeLessThanOrEqual(viewport.height + 1);
});

test('divider exists', async ({ page }) => {
  await openIvy(page);

  await expect(page.locator('#divider')).toBeVisible();
});

test('check command reports a visible status', async ({ page }) => {
  await openIvy(page);

  await page.evaluate(async () => {
    await window.__ivyDiagnostics.runtime().runCheck();
  });
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
    if (window.__ivyDiagnostics.runtime() && window.__ivyDiagnostics.runtime().conceptGraph && window.__ivyDiagnostics.runtime().conceptGraph.cy) {
      return window.__ivyDiagnostics.runtime().conceptGraph.cy.nodes().length;
    }
    return 0;
  });

  if (nodeCount > 0) {
    const menuVisible = await page.evaluate(() => {
      const cy = window.__ivyDiagnostics.runtime().conceptGraph.cy;
      const node = cy.nodes()[0];
      const pos = node.renderedPosition();
      node.emit('cxttap', { renderedPosition: pos });
      const cm = document.getElementById('context-menu');
      return !!cm && cm.style.display !== 'none';
    });
    expect(menuVisible).toBe(true);
    await expect(page.locator('#context-menu')).toContainText('Projections...');
    await expect(page.locator('#context-menu')).not.toContainText(/Splatter|Materialize|Remove/);
  }

  expect(await page.evaluate(() => window.__ivyInitError || '')).toBe('');
});

test('diagnostics bridge is loaded', async ({ page }) => {
  await openIvy(page);

  const globals = await page.evaluate(() => ({
    diagnostics: typeof window.__ivyDiagnostics,
    runtime: typeof window.__ivyDiagnostics.runtime,
    api: typeof window.__ivyDiagnostics.runtime().api,
  }));
  expect(globals).toEqual({
    diagnostics: 'object',
    runtime: 'function',
    api: 'object',
  });
});

test('Vite bundle owns the Ivy runtime script path', async ({ page }) => {
  await openIvy(page);

  const legacyScripts = await page.evaluate(() => {
    const documentScripts = Array.from(document.scripts)
      .map((script) => script.src || '')
      .filter((src) => src.includes('/static/js/ivyweb_'));
    const resourceScripts = performance.getEntriesByType('resource')
      .map((entry) => entry.name || '')
      .filter((name) => name.includes('/static/js/ivyweb_'));
    const dynamicMarkers = document.querySelectorAll('[data-ivy-legacy-script]').length;
    return { documentScripts, resourceScripts, dynamicMarkers };
  });

  expect(legacyScripts.documentScripts).toEqual([]);
  expect(legacyScripts.resourceScripts).toEqual([]);
  expect(legacyScripts.dynamicMarkers).toBe(0);
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
