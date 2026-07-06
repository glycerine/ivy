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
const clientServerWithIndividualsContent = clientServerIvyContent.replace(
  'type server\n\nrelation link',
  'type server\n\nindividual c0 : client\nindividual c1 : client\nindividual s0 : server\n\nrelation link',
);
const labeledSaveInvariantContent = `#lang ivy1.7
type node
relation p(X:node)
relation q(X:node)
after init { p(X) := true; q(X) := true }
invariant [drop_p] p(X) -> q(X)
invariant [keep_q] q(X) -> q(X)
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
  await page.waitForFunction(() => window.__ivyDiagnostics.runtime().api.sessionId);
  return consoleErrors;
}

async function openIvyAt(page, target) {
  const consoleErrors = [];
  page.on('console', (msg) => {
    if (msg.type() === 'error') {
      consoleErrors.push(msg.text());
    }
  });
  page.on('pageerror', (err) => {
    consoleErrors.push(err.message);
  });

  await page.goto(target);
  await expect(page).toHaveTitle(/ivy/i);
  await expect(page.locator('#menubar')).toBeVisible();
  await page.waitForFunction(() => window.__ivyDiagnostics && window.__ivyDiagnostics.runtime() && window.__ivyDiagnostics.runtime().api);
  await page.waitForFunction(() => window.__ivyDiagnostics.runtime().api.sessionId);
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
  return loadContentIntoCurrentSession(page, clientServerIvyContent, 'test.ivy');
}

async function loadContentIntoCurrentSession(page, content, filename) {
  return page.evaluate(async ({ content, filename }) => {
    const app = window.__ivyDiagnostics.runtime();
    const file = new File([content], filename, { type: 'text/plain' });
    const loadBody = await app.api.loadFile(file, { isolate: '' });
    const arg = await app.api.getARG();
    const concept = await app.api.getConceptGraph();
    if (app.applyArgSnapshot) {
      app.applyArgSnapshot('sheet-1', arg);
    } else if (app.argGraph) {
      app.argGraph.update(arg.elements, arg.positions);
    }
    if (app.applyConceptSnapshot) {
      app.applyConceptSnapshot('sheet-1', concept);
    } else if (app.conceptGraph) {
      app.conceptGraph.update(concept.elements, concept.positions);
    }
    return { loadBody, arg, concept };
  }, { content, filename });
}

async function selectArgStateForConcept(page, stateId = 'state_0') {
  await page.waitForFunction(() => {
    const app = window.__ivyDiagnostics.runtime();
    const sheet = app.sheets && app.sheets['sheet-1'];
    const graph = (sheet && sheet.argGraph) || app.argGraph;
    return graph && graph.cy;
  });
  await page.evaluate((stateId) => {
    const app = window.__ivyDiagnostics.runtime();
    if (app.switchSheet) app.switchSheet('sheet-1');
    const sheet = app.sheets && app.sheets['sheet-1'];
    const graph = (sheet && sheet.argGraph) || app.argGraph;
    const nodes = graph.cy.nodes().toArray();
    const node = nodes.find((n) => n.data('obj') === stateId || n.id() === stateId);
    if (!node) {
      throw new Error(`ARG state ${stateId} not found; saw ${nodes.map((n) => `${n.id()}:${n.data('obj')}`).join(', ')}`);
    }
    node.emit('tap', { target: node });
  }, stateId);
  await page.waitForFunction((stateId) => {
    const app = window.__ivyDiagnostics.runtime();
    const sheet = app.sheets && app.sheets[app.activeSheetId || 'sheet-1'];
    const graph = (sheet && sheet.conceptGraph) || app.conceptGraph;
    return sheet && sheet.selectedArgNode === stateId && graph && graph.cy.nodes().length > 0;
  }, stateId);
}

async function openConceptNodeContextMenu(page, labelFragment) {
  await page.evaluate((labelFragment) => {
    const app = window.__ivyDiagnostics.runtime();
    const sheet = app.sheets && app.sheets[app.activeSheetId || 'sheet-1'];
    const graph = (sheet && sheet.conceptGraph) || app.conceptGraph;
    const nodes = graph.cy.nodes().toArray();
    const candidates = nodes.filter((n) => !(n.hasClass && n.hasClass('subgraph_box')));
    const node = candidates.find((n) => String(n.data('obj') || '') === labelFragment) || candidates.find((n) => {
      const haystack = `${n.data('obj') || ''}\n${n.data('label') || ''}`;
      return haystack.includes(labelFragment);
    });
    if (!node) {
      throw new Error(`concept node containing ${labelFragment} not found; saw ${nodes.map((n) => `${n.data('obj')}:${n.data('label')}`).join(', ')}`);
    }
    const pos = node.renderedPosition ? node.renderedPosition() : { x: 24, y: 24 };
    app.onConceptNodeRightClick(node.data(), pos);
  }, labelFragment);
  await expect(page.locator('#context-menu')).toBeVisible();
}

async function openConceptEdgeContextMenu(page, labelFragment) {
  await page.evaluate((labelFragment) => {
    const app = window.__ivyDiagnostics.runtime();
    const sheet = app.sheets && app.sheets[app.activeSheetId || 'sheet-1'];
    const graph = (sheet && sheet.conceptGraph) || app.conceptGraph;
    const edges = graph.cy.edges().toArray();
    const edge = edges.find((e) => {
      const haystack = `${e.data('obj') || ''}\n${e.data('label') || ''}\n${e.data('short_info') || ''}`;
      return haystack.includes(labelFragment);
    });
    if (!edge) {
      throw new Error(`concept edge containing ${labelFragment} not found; saw ${edges.map((e) => `${e.data('obj')}:${e.data('label')}:${e.data('short_info')}`).join(', ')}`);
    }
    let pos = edge.renderedMidpoint ? edge.renderedMidpoint() : null;
    if (!pos) {
      const source = edge.source && edge.source();
      const target = edge.target && edge.target();
      const sp = source && source.renderedPosition ? source.renderedPosition() : { x: 24, y: 24 };
      const tp = target && target.renderedPosition ? target.renderedPosition() : sp;
      pos = { x: (sp.x + tp.x) / 2, y: (sp.y + tp.y) / 2 };
    }
    app.onConceptEdgeRightClick(edge.data(), pos);
  }, labelFragment);
  await expect(page.locator('#context-menu')).toBeVisible();
}

async function conceptContextActionState(page) {
  return page.evaluate(() => {
    const app = window.__ivyDiagnostics.runtime();
    const sheetId = app.activeSheetId || 'sheet-1';
    const sheet = app.uiDataModel && app.uiDataModel.sheets && app.uiDataModel.sheets[sheetId];
    const runtimeSheet = app.sheets && app.sheets[sheetId];
    const graph = (runtimeSheet && runtimeSheet.conceptGraph) || app.conceptGraph;
    const graphStack = sheet && sheet.concept && sheet.concept.graphStack;
    const conceptNames = Object.keys((sheet && sheet.concept && sheet.concept.domain && sheet.concept.domain.concepts) || {});
    const factTexts = ((sheet && sheet.concept && sheet.concept.facts) || []).map((fact) => String(fact.text || ''));
    const labels = graph.cy.nodes().map((n) => String(n.data('label') || ''));
    const objects = graph.cy.nodes().map((n) => String(n.data('obj') || ''));
    const classesByObj = {};
    graph.cy.nodes().forEach((n) => {
      classesByObj[String(n.data('obj') || '')] = String(n.classes ? n.classes() : '');
    });
    const linkRow = Array.from(document.querySelectorAll('#state-checkbox-body tr'))
      .find((row) => row.textContent && row.textContent.includes('link'));
    const linkUnknown = linkRow ? linkRow.querySelectorAll('input[type="checkbox"]')[1] : null;
    const values = labels.concat(objects);
    return {
      labels,
      objects,
      conceptNames,
      factTexts,
      values,
      classesByObj,
      graphStack: graphStack ? {
        canUndo: graphStack.canUndo,
        canRedo: graphStack.canRedo,
        undoDepth: graphStack.undoDepth,
        redoDepth: graphStack.redoDepth,
      } : null,
      linkUnknownChecked: !!(linkUnknown && linkUnknown.checked),
    };
  });
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

test('active event sheets hide analysis workflow menus', async ({ page }) => {
  await openIvy(page);

  await page.locator('#ui-mode-select').selectOption('reachability');
  await expect(page.locator('#mode-select')).toBeVisible();
  await expect(page.locator('#btn-check')).toBeVisible();

  await page.evaluate(() => {
    window.__ivyDiagnostics.runtime().openEventTraceSheet('Trace', {
      sheet_id: 'events-1',
      events: [
        { text: 'root(a)', address: '0', children: [] },
      ],
      patterns: ['root(a)'],
    }, 'events-1');
  });

  await expect(page.locator('.sheet-tab[data-sheet="events-1"]')).toHaveClass(/active/);
  await expect(page.locator('body')).toHaveAttribute('data-active-sheet-type', 'events');
  await expect(page.locator('#mode-select')).toBeHidden();
  await expect(page.locator('#btn-check')).toBeHidden();
  await expect(page.locator('#btn-show-reachable')).toBeHidden();

  await page.locator('.sheet-tab[data-sheet="sheet-1"]').click();
  await expect(page.locator('body')).toHaveAttribute('data-active-sheet-type', 'analysis');
  await expect(page.locator('#mode-select')).toBeVisible();
  await expect(page.locator('#btn-check')).toBeVisible();
});

test('CTI relations-to-minimize field is sent to check and minimize actions', async ({ page }) => {
  await openIvy(page);

  const input = page.locator('#cti-relations-to-minimize');
  await expect(input).toBeVisible();
  await input.fill('q');

  await page.evaluate(async () => {
    const app = window.__ivyDiagnostics.runtime();
    app.getMode = () => 'induction';
    app._persistedFileContent = '';
    app.api.runCheck = async (mode, options) => {
      window._ctiCheckOptions = { mode, options };
      return { result: 'pass', mode, message: 'ok' };
    };
    app.api.getARG = async () => null;
    app.api.getConceptGraph = async () => null;
    app.showCheckResult = () => {};
    app._autoCheckUsedRelations = async () => {};
    await app.runCheck();
  });

  expect(await page.evaluate(() => window._ctiCheckOptions)).toEqual({
    mode: 'induction',
    options: { relations_to_minimize: 'q' },
  });

  await page.evaluate(async () => {
    const app = window.__ivyDiagnostics.runtime();
    app.api.executeAction = async (action, args) => {
      window._ctiMinimizeCall = { action, args };
      return { message: 'minimized' };
    };
    app.refreshConceptGraph = async () => {};
    await app.ctiConceptAction('cti_minimize');
  });

  expect(await page.evaluate(() => window._ctiMinimizeCall)).toEqual({
    action: 'cti_minimize',
    args: { sheet_id: 'sheet-1', relations_to_minimize: 'q' },
  });
});

test('CTI minimize shows BMC bound core details and omits unselected facts', async ({ page }) => {
  await openIvy(page);

  const minimizeRun = page.evaluate(async () => {
    const app = window.__ivyDiagnostics.runtime();
    app.activeSheetId = 'sheet-1';
    app.api.executeAction = async (action, args) => {
      window._ctiMinimizeDetailsCall = { action, args };
      return {
        message: 'Conjecture minimized using BMC bound 4; kept 1 of 2 selected facts.',
        bound: 4,
        input_facts: ['p(X)', 'q(X)'],
        core_facts: ['p(X)'],
        removed_facts: ['q(X)'],
        conjecture: '~p(X)',
        concept: {
          sheet_id: 'sheet-1',
          facts: [{ text: 'r(X)' }],
          elements: [],
        },
      };
    };
    await app.ctiConceptAction('cti_minimize');
    return {
      call: window._ctiMinimizeDetailsCall,
      status: document.querySelector('#statusbar')?.textContent || '',
    };
  });

  const dialogText = page.locator('[data-ivy-dialog-text]');
  await expect(page.locator('[data-ivy-dialog]')).toBeVisible();
  await expect(page.locator('[data-ivy-dialog]')).toContainText('Conjecture minimized using BMC bound 4');
  await expect(dialogText).toHaveValue(/BMC bound: 4/);
  await expect(dialogText).toHaveValue(/Selected facts:\n- p\(X\)\n- q\(X\)/);
  await expect(dialogText).toHaveValue(/Core facts kept:\n- p\(X\)/);
  await expect(dialogText).toHaveValue(/Removed facts:\n- q\(X\)/);
  await expect(dialogText).toHaveValue(/Resulting conjecture:\n~p\(X\)/);
  expect(await dialogText.inputValue()).not.toContain('r(X)');
  await page.getByRole('button', { name: 'OK' }).click();

  await expect(minimizeRun).resolves.toEqual({
    call: {
      action: 'cti_minimize',
      args: { sheet_id: 'sheet-1', relations_to_minimize: 'relations to minimize' },
    },
    status: expect.stringContaining('Conjecture minimized using BMC bound 4'),
  });
});

test('CTI sufficient and relative induction checks show selected conjecture dialogs', async ({ page }) => {
  await openIvy(page);

  const scenarios = [
    {
      action: 'cti_check_sufficient',
      result: 'insufficient',
      message: '(1) does not imply (2) at the next time.',
      title: 'CTI check sufficient',
    },
    {
      action: 'cti_check_inductive',
      result: 'inductive',
      message: '(1) is relatively inductive.',
      title: 'CTI relative induction',
    },
  ];

  for (const scenario of scenarios) {
    const checkRun = page.evaluate(async (scenario) => {
      const app = window.__ivyDiagnostics.runtime();
      app.activeSheetId = 'sheet-1';
      app.api.executeAction = async (action, args) => {
        window._ctiCheckDetailsCall = { action, args };
        return {
          ok: scenario.result === 'inductive',
          check: action === 'cti_check_sufficient' ? 'sufficient' : 'relative_induction',
          result: scenario.result,
          message: scenario.message,
          selected_conjecture: 'forall X. p(X)',
          target_conjecture: 'forall X. q(X)',
          concept: { sheet_id: 'sheet-1', elements: [] },
        };
      };
      await app.ctiConceptAction(scenario.action);
      return {
        call: window._ctiCheckDetailsCall,
        status: document.querySelector('#statusbar')?.textContent || '',
      };
    }, scenario);

    const dialogText = page.locator('[data-ivy-dialog-text]');
    await expect(page.locator('[data-ivy-dialog]')).toBeVisible();
    await expect(page.locator('[data-ivy-dialog]')).toContainText(scenario.title);
    await expect(page.locator('[data-ivy-dialog]')).toContainText(scenario.message);
    await expect(dialogText).toHaveValue(new RegExp(`Result: ${scenario.result}`));
    await expect(dialogText).toHaveValue(/Selected conjecture:\nforall X\. p\(X\)/);
    await expect(dialogText).toHaveValue(/Target conjecture:\nforall X\. q\(X\)/);
    await page.getByRole('button', { name: 'OK' }).click();

    await expect(checkRun).resolves.toEqual({
      call: {
        action: scenario.action,
        args: { sheet_id: 'sheet-1', relations_to_minimize: 'relations to minimize' },
      },
      status: expect.stringContaining(scenario.message),
    });
  }
});

test('CTI diagram sends CTI mode and labels the pre-state', async ({ page }) => {
  await openIvy(page);

  const result = await page.evaluate(async () => {
    const app = window.__ivyDiagnostics.runtime();
    app.activeSheetId = 'sheet-1';
    app.setUIMode('cti');
    app.api.executeAction = async (action, args) => {
      window._ctiDiagramCall = { action, args };
      return {
        status: 'diagrammed',
        message: 'Diagram complete.',
        concept: {
          sheet_id: 'sheet-1',
          elements: [],
          cti_state_label: 'CTI pre-state 0',
        },
      };
    };
    app.refreshConceptGraph = async () => {
      window._ctiDiagramRefreshed = true;
    };
    await app.diagramCurrentState();
    return {
      call: window._ctiDiagramCall,
      label: document.querySelector('#state-label')?.textContent || '',
      status: document.querySelector('#statusbar')?.textContent || '',
      refreshed: !!window._ctiDiagramRefreshed,
    };
  });

  expect(result).toEqual({
    call: { action: 'diagram', args: { sheet_id: 'sheet-1', ui_mode: 'cti' } },
    label: 'State: CTI pre-state 0',
    status: expect.stringContaining('Diagram complete'),
    refreshed: false,
  });
});

test('CTI strengthen confirms the exact conjecture before appending', async ({ page }) => {
  await openIvy(page);

  const firstRun = page.evaluate(async () => {
    const app = window.__ivyDiagnostics.runtime();
    app.activeSheetId = 'sheet-1';
    window._ctiStrengthenCalls = [];
    app.api.executeAction = async (action, args) => {
      window._ctiStrengthenCalls.push({ action, args });
      if (action === 'cti_strengthen_preview') {
        return { conjecture: 'forall X. p(X)' };
      }
      return { message: 'Invariant strengthened', concept: { sheet_id: 'sheet-1', elements: [] } };
    };
    return app.ctiConceptAction('cti_strengthen');
  });

  await expect(page.locator('[data-ivy-dialog]')).toBeVisible();
  await expect(page.locator('[data-ivy-dialog]')).toContainText('Add this conjecture as an invariant?');
  await expect(page.locator('[data-ivy-dialog-text]')).toHaveValue('forall X. p(X)');
  await page.getByRole('button', { name: 'Cancel' }).click();
  await expect(firstRun).resolves.toBeNull();
  const strengthenArgs = { sheet_id: 'sheet-1', relations_to_minimize: 'relations to minimize' };
  expect(await page.evaluate(() => window._ctiStrengthenCalls)).toEqual([
    { action: 'cti_strengthen_preview', args: strengthenArgs },
  ]);

  const secondRun = page.evaluate(async () => {
    const app = window.__ivyDiagnostics.runtime();
    window._ctiStrengthenCalls = [];
    return app.ctiConceptAction('cti_strengthen');
  });
  await expect(page.locator('[data-ivy-dialog]')).toBeVisible();
  await expect(page.locator('[data-ivy-dialog-text]')).toHaveValue('forall X. p(X)');
  await page.getByRole('button', { name: 'Strengthen' }).click();
  await expect(secondRun).resolves.toMatchObject({ message: 'Invariant strengthened' });
  expect(await page.evaluate(() => window._ctiStrengthenCalls)).toEqual([
    { action: 'cti_strengthen_preview', args: strengthenArgs },
    { action: 'cti_strengthen', args: strengthenArgs },
  ]);
});

test('CTI save invariant writes kept dropped and new sections through the browser picker', async ({ page }) => {
  await openIvy(page);
  await loadContentIntoCurrentSession(page, labeledSaveInvariantContent, 'cti_save.ivy');

  const result = await page.evaluate(async () => {
    const app = window.__ivyDiagnostics.runtime();
    await app.api.executeAction('weaken', { indices: [0] });
    await app.api.executeAction('cti_strengthen', { sheet_id: 'sheet-1' });
    window._savedInvariant = {
      options: null,
      content: '',
      closed: false,
    };
    window.showSaveFilePicker = async (options) => {
      window._savedInvariant.options = options;
      return {
        name: 'invariant.ivy',
        async createWritable() {
          return {
            async write(text) {
              window._savedInvariant.content = String(text);
            },
            async close() {
              window._savedInvariant.closed = true;
            },
          };
        },
      };
    };
    await app.saveInvariant();
    return {
      ...window._savedInvariant,
      status: document.querySelector('#statusbar')?.textContent || '',
    };
  });

  expect(result.options).toEqual({
    suggestedName: 'invariant.ivy',
    types: [{ description: 'Ivy invariant files', accept: { 'text/plain': ['.ivy'] } }],
  });
  expect(result.closed).toBe(true);
  expect(result.status).toContain('Invariant saved: invariant.ivy');
  expect(result.content).toContain('# This file was generated by ivy.');
  expect(result.content).toContain('# original conjectures kept');
  expect(result.content).toContain('invariant [keep_q] q(X) -> q(X)');
  expect(result.content).toContain('# original conjectures dropped');
  expect(result.content).toContain('# invariant [drop_p] p(X) -> q(X)');
  expect(result.content).toContain('# new conjectures');
  expect(result.content).toMatch(/^invariant .*true/m);
  expect(result.content.indexOf('# original conjectures kept')).toBeLessThan(result.content.indexOf('# original conjectures dropped'));
  expect(result.content.indexOf('# original conjectures dropped')).toBeLessThan(result.content.indexOf('# new conjectures'));
});

test('CTI bounded check View opens the counterexample trace sheet', async ({ page }) => {
  await openIvy(page);

  const bmcRun = page.evaluate(async () => {
    const app = window.__ivyDiagnostics.runtime();
    app.activeSheetId = 'sheet-1';
    app.currentBound = 3;
    app.api.executeAction = async (action, args) => {
      window._ctiBmcCall = { action, args };
      return {
        found: true,
        result: 'fail',
        view: 'trace',
        message: 'BMC with bound 0 found a counter-example to:\n(~true)',
        conjecture: '(~true)',
        trace_sheet_id: 'trace-cti',
        trace_label: 'CTI counterexample',
        trace_arg: {
          elements: [
            { group: 'nodes', data: { id: 'state_0', obj: 'state_0', label: '0' } },
            { group: 'nodes', data: { id: 'state_1', obj: 'state_1', label: '1' } },
            { group: 'edges', data: { id: 'trace_edge', source: 'state_0', target: 'state_1', label: 'trace' } },
          ],
        },
      };
    };
    await app.ctiBoundedCheck();
    const sheet = app.uiDataModel && app.uiDataModel.sheets && app.uiDataModel.sheets[app.activeSheetId];
    return {
      call: window._ctiBmcCall,
      activeSheet: app.activeSheetId,
      uiMode: document.body.getAttribute('data-ui-mode'),
      reachabilityOnly: !!(sheet && sheet.reachabilityOnly),
      activeTabText: document.querySelector('.sheet-tab.active')?.textContent || '',
    };
  });

  await expect(page.locator('[data-ivy-dialog]')).toBeVisible();
  await page.locator('[data-ivy-dialog-int]').fill('0');
  await page.getByRole('button', { name: 'OK' }).click();

  await expect(page.locator('[data-ivy-dialog]')).toBeVisible();
  await expect(page.locator('[data-ivy-dialog]')).toContainText('BMC with bound 0 found a counter-example to:');
  await expect(page.locator('[data-ivy-dialog-text]')).toHaveValue('(~true)');
  await page.getByRole('button', { name: 'View' }).click();

  await expect(bmcRun).resolves.toEqual({
    call: { action: 'cti_bounded_check', args: { sheet_id: 'sheet-1', bound: 0 } },
    activeSheet: 'trace-cti',
    uiMode: 'reachability',
    reachabilityOnly: true,
    activeTabText: expect.stringContaining('CTI counterexample'),
  });
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

test('reach action shows eliminated conjectures dialog', async ({ page }) => {
  await openIvy(page);

  const reachPromise = page.evaluate(async () => {
    const app = window.__ivyDiagnostics.runtime();
    app.activeSheetId = 'sheet-1';
    app.api.executeAction = async (action, args) => {
      window._reachActionCall = { action, args };
      return {
        reachable: true,
        eliminated_conjectures_message: 'The following conjectures have been eliminated:',
        eliminated_conjectures: ['p', 'q'],
      };
    };
    app.refreshConceptGraph = async () => undefined;
    await app.reachStep();
    return {
      call: window._reachActionCall,
      status: document.querySelector('#statusbar')?.textContent || '',
    };
  });

  await expect(page.locator('[data-ivy-dialog]')).toBeVisible();
  await expect(page.locator('[data-ivy-dialog]')).toContainText('The following conjectures have been eliminated:');
  await expect(page.locator('[data-ivy-dialog-list]')).toContainText('p');
  await expect(page.locator('[data-ivy-dialog-list]')).toContainText('q');
  await expect(page.getByRole('button', { name: 'Cancel' })).toHaveCount(0);
  await page.locator('[data-ivy-dialog-list]').selectOption('p');
  await page.getByRole('button', { name: 'OK' }).click();

  await expect(reachPromise).resolves.toEqual({
    call: { action: 'reach', args: { sheet_id: 'sheet-1' } },
    status: expect.stringContaining('Reach complete'),
  });
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

test('concept node Splatter context action preserves relation toggles and graph-stack undo redo', async ({ page }) => {
  await openIvy(page);
  await page.evaluate(async () => {
    const app = window.__ivyDiagnostics.runtime();
    await app._switchJobSubmissionBackend('remote');
  });
  await page.waitForFunction(() => {
    const app = window.__ivyDiagnostics.runtime();
    return app.api && app.api.kind === 'hosted-go' && app.api.sessionId;
  });
  await page.locator('#ui-mode-select').selectOption('reachability');
  await loadContentIntoCurrentSession(page, clientServerWithIndividualsContent, 'client_server_with_individuals.ivy');

  const linkUnknown = page.locator('#state-checkbox-body tr', { hasText: 'link' }).locator('input[type="checkbox"]').nth(1);
  await expect(linkUnknown).toBeVisible();
  await linkUnknown.check();
  await expect(linkUnknown).toBeChecked();

  await openConceptNodeContextMenu(page, 'client');
  await page.locator('.context-menu-item', { hasText: 'Splatter' }).click();
  await expect(page.locator('#statusbar')).toContainText('Splattered: client', { timeout: 10_000 });

  await page.waitForFunction(() => {
    const app = window.__ivyDiagnostics.runtime();
    const values = app.conceptGraph.cy.nodes().flatMap((n) => [String(n.data('obj') || ''), String(n.data('label') || '')]);
    return values.some((value) => value.includes('c0')) && values.some((value) => value.includes('c1'));
  });
  const splattered = await conceptContextActionState(page);
  expect(splattered.values.some((value) => value.includes('c0'))).toBe(true);
  expect(splattered.values.some((value) => value.includes('c1'))).toBe(true);
  expect(splattered.linkUnknownChecked).toBe(true);
  expect(splattered.graphStack).toMatchObject({ canUndo: true, canRedo: false });
  expect(splattered.graphStack.undoDepth).toBeGreaterThan(0);

  await page.evaluate(async () => {
    await window.__ivyDiagnostics.runtime().doUndo();
  });
  await page.waitForFunction(() => {
    const app = window.__ivyDiagnostics.runtime();
    const sheet = app.uiDataModel.sheets[app.activeSheetId || 'sheet-1'];
    return sheet.concept.graphStack.canRedo;
  });
  const undone = await conceptContextActionState(page);
  expect(undone.values.some((value) => value.includes('c0'))).toBe(false);
  expect(undone.values.some((value) => value.includes('c1'))).toBe(false);
  expect(undone.linkUnknownChecked).toBe(true);
  expect(undone.graphStack).toMatchObject({ canUndo: false, canRedo: true });

  await page.evaluate(async () => {
    await window.__ivyDiagnostics.runtime().doRedo();
  });
  await page.waitForFunction(() => {
    const app = window.__ivyDiagnostics.runtime();
    const values = app.conceptGraph.cy.nodes().flatMap((n) => [String(n.data('obj') || ''), String(n.data('label') || '')]);
    return values.some((value) => value.includes('c0')) && values.some((value) => value.includes('c1'));
  });
  const redone = await conceptContextActionState(page);
  expect(redone.linkUnknownChecked).toBe(true);
  expect(redone.graphStack).toMatchObject({ canUndo: true, canRedo: false });
});

test('concept node Empty context action preserves relation toggles and graph-stack undo redo', async ({ page }) => {
  await openIvy(page);
  await page.evaluate(async () => {
    const app = window.__ivyDiagnostics.runtime();
    await app._switchJobSubmissionBackend('remote');
  });
  await page.waitForFunction(() => {
    const app = window.__ivyDiagnostics.runtime();
    return app.api && app.api.kind === 'hosted-go' && app.api.sessionId;
  });
  await page.locator('#ui-mode-select').selectOption('reachability');
  await loadContentIntoCurrentSession(page, clientServerWithIndividualsContent, 'client_server_with_individuals.ivy');

  const linkUnknown = page.locator('#state-checkbox-body tr', { hasText: 'link' }).locator('input[type="checkbox"]').nth(1);
  await expect(linkUnknown).toBeVisible();
  await linkUnknown.check();
  await expect(linkUnknown).toBeChecked();

  await openConceptNodeContextMenu(page, 'client');
  await page.locator('.context-menu-item', { hasText: 'Empty' }).click();
  await expect(page.locator('#statusbar')).toContainText('Suppose empty applied', { timeout: 10_000 });

  await page.waitForFunction(() => {
    const app = window.__ivyDiagnostics.runtime();
    const classes = app.conceptGraph.cy.nodes().toArray()
      .filter((n) => String(n.data('obj') || '') === 'client')
      .map((n) => String(n.classes ? n.classes() : ''))
      .join(' ');
    return classes.includes('non_existing');
  });
  const emptied = await conceptContextActionState(page);
  expect(emptied.classesByObj.client).toContain('non_existing');
  expect(emptied.linkUnknownChecked).toBe(true);
  expect(emptied.graphStack).toMatchObject({ canUndo: true, canRedo: false });

  await page.evaluate(async () => {
    await window.__ivyDiagnostics.runtime().doUndo();
  });
  await page.waitForFunction(() => {
    const app = window.__ivyDiagnostics.runtime();
    const sheet = app.uiDataModel.sheets[app.activeSheetId || 'sheet-1'];
    return sheet.concept.graphStack.canRedo;
  });
  const undone = await conceptContextActionState(page);
  expect(undone.classesByObj.client || '').not.toContain('non_existing');
  expect(undone.linkUnknownChecked).toBe(true);
  expect(undone.graphStack).toMatchObject({ canRedo: true });

  await page.evaluate(async () => {
    await window.__ivyDiagnostics.runtime().doRedo();
  });
  await page.waitForFunction(() => {
    const app = window.__ivyDiagnostics.runtime();
    const classes = app.conceptGraph.cy.nodes().toArray()
      .filter((n) => String(n.data('obj') || '') === 'client')
      .map((n) => String(n.classes ? n.classes() : ''))
      .join(' ');
    return classes.includes('non_existing');
  });
  const redone = await conceptContextActionState(page);
  expect(redone.classesByObj.client).toContain('non_existing');
  expect(redone.linkUnknownChecked).toBe(true);
  expect(redone.graphStack).toMatchObject({ canUndo: true, canRedo: false });
});

test('concept node Materialize context action creates a fresh witness with graph-stack undo redo', async ({ page }) => {
  await openIvy(page);
  await page.evaluate(async () => {
    const app = window.__ivyDiagnostics.runtime();
    await app._switchJobSubmissionBackend('remote');
  });
  await page.waitForFunction(() => {
    const app = window.__ivyDiagnostics.runtime();
    return app.api && app.api.kind === 'hosted-go' && app.api.sessionId;
  });
  await page.locator('#ui-mode-select').selectOption('reachability');
  await loadContentIntoCurrentSession(page, clientServerWithIndividualsContent, 'client_server_with_individuals.ivy');

  const linkUnknown = page.locator('#state-checkbox-body tr', { hasText: 'link' }).locator('input[type="checkbox"]').nth(1);
  await expect(linkUnknown).toBeVisible();
  await linkUnknown.check();
  await expect(linkUnknown).toBeChecked();

  await openConceptNodeContextMenu(page, 'client');
  await page.locator('.context-menu-item').filter({ hasText: /^Materialize$/ }).click();
  await expect(page.locator('#statusbar')).toContainText('Node materialized', { timeout: 10_000 });

  await page.waitForFunction(() => {
    const app = window.__ivyDiagnostics.runtime();
    const sheet = app.uiDataModel.sheets[app.activeSheetId || 'sheet-1'];
    return !!(sheet && sheet.concept && sheet.concept.domain && sheet.concept.domain.concepts['=__c0']);
  });
  const materialized = await conceptContextActionState(page);
  expect(materialized.conceptNames).toContain('=__c0');
  expect(materialized.linkUnknownChecked).toBe(true);
  expect(materialized.graphStack).toMatchObject({ canUndo: true, canRedo: false });

  await page.evaluate(async () => {
    await window.__ivyDiagnostics.runtime().doUndo();
  });
  await page.waitForFunction(() => {
    const app = window.__ivyDiagnostics.runtime();
    const sheet = app.uiDataModel.sheets[app.activeSheetId || 'sheet-1'];
    return sheet.concept.graphStack.canRedo;
  });
  const undone = await conceptContextActionState(page);
  expect(undone.conceptNames).not.toContain('=__c0');
  expect(undone.linkUnknownChecked).toBe(true);

  await page.evaluate(async () => {
    await window.__ivyDiagnostics.runtime().doRedo();
  });
  await page.waitForFunction(() => {
    const app = window.__ivyDiagnostics.runtime();
    const sheet = app.uiDataModel.sheets[app.activeSheetId || 'sheet-1'];
    return !!(sheet && sheet.concept && sheet.concept.domain && sheet.concept.domain.concepts['=__c0']);
  });
  const redone = await conceptContextActionState(page);
  expect(redone.conceptNames).toContain('=__c0');
  expect(redone.linkUnknownChecked).toBe(true);
  expect(redone.graphStack).toMatchObject({ canUndo: true, canRedo: false });
});

test('concept edge Materialize context action records positive fact with graph-stack undo redo', async ({ page }) => {
  await openIvy(page);
  await page.evaluate(async () => {
    const app = window.__ivyDiagnostics.runtime();
    await app._switchJobSubmissionBackend('remote');
  });
  await page.waitForFunction(() => {
    const app = window.__ivyDiagnostics.runtime();
    return app.api && app.api.kind === 'hosted-go' && app.api.sessionId;
  });
  await page.locator('#ui-mode-select').selectOption('reachability');
  await loadContentIntoCurrentSession(page, clientServerWithIndividualsContent, 'client_server_with_individuals.ivy');

  const linkUnknown = page.locator('#state-checkbox-body tr', { hasText: 'link' }).locator('input[type="checkbox"]').nth(1);
  await expect(linkUnknown).toBeVisible();
  await linkUnknown.check();
  await expect(linkUnknown).toBeChecked();
  await page.waitForFunction(() => {
    const app = window.__ivyDiagnostics.runtime();
    return app.conceptGraph.cy.edges().toArray().some((edge) => {
      const haystack = `${edge.data('obj') || ''}\n${edge.data('label') || ''}\n${edge.data('short_info') || ''}`;
      return haystack.includes('link');
    });
  });

  await openConceptEdgeContextMenu(page, 'link');
  await page.locator('.context-menu-item').filter({ hasText: /^Materialize$/ }).click();
  await expect(page.locator('#statusbar')).toContainText('Edge materialized (+)', { timeout: 10_000 });
  await page.waitForFunction(() => {
    const app = window.__ivyDiagnostics.runtime();
    const sheet = app.uiDataModel.sheets[app.activeSheetId || 'sheet-1'];
    return ((sheet && sheet.concept && sheet.concept.facts) || [])
      .some((fact) => String(fact.text || '').includes('link'));
  });
  const materialized = await conceptContextActionState(page);
  expect(materialized.factTexts.some((text) => text.includes('link'))).toBe(true);
  expect(materialized.linkUnknownChecked).toBe(true);
  expect(materialized.graphStack).toMatchObject({ canUndo: true, canRedo: false });

  await page.evaluate(async () => {
    await window.__ivyDiagnostics.runtime().doUndo();
  });
  await page.waitForFunction(() => {
    const app = window.__ivyDiagnostics.runtime();
    const sheet = app.uiDataModel.sheets[app.activeSheetId || 'sheet-1'];
    return sheet.concept.graphStack.canRedo;
  });
  const undone = await conceptContextActionState(page);
  expect(undone.factTexts.some((text) => text.includes('link'))).toBe(false);
  expect(undone.linkUnknownChecked).toBe(true);

  await page.evaluate(async () => {
    await window.__ivyDiagnostics.runtime().doRedo();
  });
  await page.waitForFunction(() => {
    const app = window.__ivyDiagnostics.runtime();
    const sheet = app.uiDataModel.sheets[app.activeSheetId || 'sheet-1'];
    return ((sheet && sheet.concept && sheet.concept.facts) || [])
      .some((fact) => String(fact.text || '').includes('link'));
  });
  const redone = await conceptContextActionState(page);
  expect(redone.factTexts.some((text) => text.includes('link'))).toBe(true);
  expect(redone.linkUnknownChecked).toBe(true);
  expect(redone.graphStack).toMatchObject({ canUndo: true, canRedo: false });
});

test('concept edge Dematerialize context action records negative fact with graph-stack undo redo', async ({ page }) => {
  await openIvy(page);
  await page.evaluate(async () => {
    const app = window.__ivyDiagnostics.runtime();
    await app._switchJobSubmissionBackend('remote');
  });
  await page.waitForFunction(() => {
    const app = window.__ivyDiagnostics.runtime();
    return app.api && app.api.kind === 'hosted-go' && app.api.sessionId;
  });
  await page.locator('#ui-mode-select').selectOption('reachability');
  await loadContentIntoCurrentSession(page, clientServerWithIndividualsContent, 'client_server_with_individuals.ivy');

  const linkUnknown = page.locator('#state-checkbox-body tr', { hasText: 'link' }).locator('input[type="checkbox"]').nth(1);
  await expect(linkUnknown).toBeVisible();
  await linkUnknown.check();
  await expect(linkUnknown).toBeChecked();
  await page.waitForFunction(() => {
    const app = window.__ivyDiagnostics.runtime();
    return app.conceptGraph.cy.edges().toArray().some((edge) => {
      const haystack = `${edge.data('obj') || ''}\n${edge.data('label') || ''}\n${edge.data('short_info') || ''}`;
      return haystack.includes('link');
    });
  });

  await openConceptEdgeContextMenu(page, 'link');
  await page.locator('.context-menu-item').filter({ hasText: /^Dematerialize$/ }).click();
  await expect(page.locator('#statusbar')).toContainText('Edge materialized', { timeout: 10_000 });
  await page.waitForFunction(() => {
    const app = window.__ivyDiagnostics.runtime();
    const sheet = app.uiDataModel.sheets[app.activeSheetId || 'sheet-1'];
    return ((sheet && sheet.concept && sheet.concept.facts) || [])
      .some((fact) => {
        const text = String(fact.text || '');
        return text.includes('link') && (text.includes('~') || text.toLowerCase().includes('not'));
      });
  });
  const dematerialized = await conceptContextActionState(page);
  expect(dematerialized.factTexts.some((text) => text.includes('link') && (text.includes('~') || text.toLowerCase().includes('not')))).toBe(true);
  expect(dematerialized.linkUnknownChecked).toBe(true);
  expect(dematerialized.graphStack).toMatchObject({ canUndo: true, canRedo: false });

  await page.evaluate(async () => {
    await window.__ivyDiagnostics.runtime().doUndo();
  });
  await page.waitForFunction(() => {
    const app = window.__ivyDiagnostics.runtime();
    const sheet = app.uiDataModel.sheets[app.activeSheetId || 'sheet-1'];
    return sheet.concept.graphStack.canRedo;
  });
  const undone = await conceptContextActionState(page);
  expect(undone.factTexts.some((text) => text.includes('link') && (text.includes('~') || text.toLowerCase().includes('not')))).toBe(false);
  expect(undone.linkUnknownChecked).toBe(true);

  await page.evaluate(async () => {
    await window.__ivyDiagnostics.runtime().doRedo();
  });
  await page.waitForFunction(() => {
    const app = window.__ivyDiagnostics.runtime();
    const sheet = app.uiDataModel.sheets[app.activeSheetId || 'sheet-1'];
    return ((sheet && sheet.concept && sheet.concept.facts) || [])
      .some((fact) => {
        const text = String(fact.text || '');
        return text.includes('link') && (text.includes('~') || text.toLowerCase().includes('not'));
      });
  });
  const redone = await conceptContextActionState(page);
  expect(redone.factTexts.some((text) => text.includes('link') && (text.includes('~') || text.toLowerCase().includes('not')))).toBe(true);
  expect(redone.linkUnknownChecked).toBe(true);
  expect(redone.graphStack).toMatchObject({ canUndo: true, canRedo: false });
});

test('concept node Materialize edge action prompts for selected source relation', async ({ page }) => {
  await openIvy(page);
  await page.evaluate(async () => {
    const app = window.__ivyDiagnostics.runtime();
    await app._switchJobSubmissionBackend('remote');
  });
  await page.waitForFunction(() => {
    const app = window.__ivyDiagnostics.runtime();
    return app.api && app.api.kind === 'hosted-go' && app.api.sessionId;
  });
  await page.locator('#ui-mode-select').selectOption('reachability');
  await loadContentIntoCurrentSession(page, clientServerWithIndividualsContent, 'client_server_with_individuals.ivy');

  const linkUnknown = page.locator('#state-checkbox-body tr', { hasText: 'link' }).locator('input[type="checkbox"]').nth(1);
  await expect(linkUnknown).toBeVisible();
  await linkUnknown.check();
  await expect(linkUnknown).toBeChecked();

  await openConceptNodeContextMenu(page, 'client');
  await page.locator('.context-menu-item').filter({ hasText: /^Select$/ }).click();
  await expect(page.locator('#statusbar')).toContainText('Selected: client');

  await openConceptNodeContextMenu(page, 'server');
  await page.locator('.context-menu-item').filter({ hasText: /^Materialize edge$/ }).click();
  await expect(page.locator('[data-ivy-dialog]')).toBeVisible();
  await expect(page.locator('[data-ivy-dialog]')).toContainText('Materialize edge');
  await page.locator('[data-ivy-dialog-list]').selectOption('link');
  await page.getByRole('button', { name: 'OK' }).click();
  await expect(page.locator('#statusbar')).toContainText('Edge materialized (+)', { timeout: 10_000 });
  await page.waitForFunction(() => {
    const app = window.__ivyDiagnostics.runtime();
    const sheet = app.uiDataModel.sheets[app.activeSheetId || 'sheet-1'];
    return ((sheet && sheet.concept && sheet.concept.facts) || [])
      .some((fact) => String(fact.text || '').includes('link'));
  });
  const materialized = await conceptContextActionState(page);
  expect(materialized.factTexts.some((text) => text.includes('link'))).toBe(true);
  expect(materialized.linkUnknownChecked).toBe(true);
  expect(materialized.graphStack).toMatchObject({ canUndo: true, canRedo: false });

  await page.evaluate(async () => {
    await window.__ivyDiagnostics.runtime().doUndo();
  });
  await page.waitForFunction(() => {
    const app = window.__ivyDiagnostics.runtime();
    const sheet = app.uiDataModel.sheets[app.activeSheetId || 'sheet-1'];
    return sheet.concept.graphStack.canRedo;
  });
  const undone = await conceptContextActionState(page);
  expect(undone.factTexts.some((text) => text.includes('link'))).toBe(false);
  expect(undone.linkUnknownChecked).toBe(true);

  await page.evaluate(async () => {
    await window.__ivyDiagnostics.runtime().doRedo();
  });
  await page.waitForFunction(() => {
    const app = window.__ivyDiagnostics.runtime();
    const sheet = app.uiDataModel.sheets[app.activeSheetId || 'sheet-1'];
    return ((sheet && sheet.concept && sheet.concept.facts) || [])
      .some((fact) => String(fact.text || '').includes('link'));
  });
  const redone = await conceptContextActionState(page);
  expect(redone.factTexts.some((text) => text.includes('link'))).toBe(true);
  expect(redone.linkUnknownChecked).toBe(true);
  expect(redone.graphStack).toMatchObject({ canUndo: true, canRedo: false });
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

test('event tree selection, keyboard traversal, and filtered selection preservation', async ({ page }) => {
  await openIvy(page);

  await page.evaluate(() => {
    window.__ivyDiagnostics.runtime().openEventTraceSheet('Trace', {
      sheet_id: 'events-keyboard',
      events: [
        { text: 'root(a)', address: '0', subs: [{ text: 'child(a)', address: '0/0' }] },
        { text: 'done', address: '1' },
      ],
      patterns: ['child(a)'],
    }, 'events-keyboard');
  });

  const tree = page.locator('#events-keyboard .event-tree');
  const root = page.locator('#events-keyboard [data-event-address="0"]');
  const child = page.locator('#events-keyboard [data-event-address="0/0"]');
  const done = page.locator('#events-keyboard [data-event-address="1"]');

  await expect(tree).toHaveAttribute('role', 'tree');
  await expect(root).toHaveAttribute('role', 'treeitem');
  await expect(root).toHaveAttribute('tabindex', '0');
  await expect(root).toHaveAttribute('aria-expanded', 'false');
  await expect(root).toHaveAttribute('aria-selected', 'false');

  await root.click();
  await expect(root).toHaveClass(/selected/);
  await expect(root).toHaveAttribute('aria-selected', 'true');
  await expect(root).toBeFocused();

  await root.press('ArrowRight');
  await expect(root).toHaveAttribute('aria-expanded', 'true');
  await expect(child).toContainText('child(a)');

  await root.press('ArrowRight');
  await expect(child).toHaveClass(/selected/);
  await expect(child).toHaveAttribute('aria-selected', 'true');
  await expect(child).toBeFocused();

  await child.press('ArrowDown');
  await expect(done).toHaveClass(/selected/);
  await done.press('ArrowUp');
  await expect(child).toHaveClass(/selected/);

  await child.press('ArrowLeft');
  await expect(root).toHaveClass(/selected/);
  await root.press('ArrowLeft');
  await expect(root).toHaveAttribute('aria-expanded', 'false');
  await expect(child).toHaveCount(0);

  await root.press('ArrowRight');
  await root.press('ArrowRight');
  await expect(child).toHaveClass(/selected/);

  await page.evaluate(async () => {
    const app = window.__ivyDiagnostics.runtime();
    app.api.executeAction = async (actionName) => {
      if (actionName !== 'events_filter') throw new Error(actionName);
      return {
        sheet_id: 'events-filtered-keyboard',
        label: 'Filtered',
        events: [{ text: 'child(a)', address: '0/0' }],
        patterns: ['child(a)'],
      };
    };
    await app.filterEventTrace('child(a)');
  });

  const filteredChild = page.locator('#events-filtered-keyboard [data-event-address="0/0"]');
  await expect(filteredChild).toHaveClass(/selected/);
  await expect(filteredChild).toHaveAttribute('aria-selected', 'true');
  const filteredSelection = await page.evaluate(() => {
    const app = window.__ivyDiagnostics.runtime();
    return app.sheets['events-filtered-keyboard'].selectedEventAddress;
  });
  expect(filteredSelection).toBe('0/0');
});

test('event pattern load and save dialogs keep backend state authoritative', async ({ page }) => {
  await openIvy(page);

  await page.evaluate(() => {
    window.__eventPatternCalls = [];
    window.__eventPatternWrites = [];
    window.__eventPatternPickerOptions = null;
    window.showSaveFilePicker = async (options) => {
      window.__eventPatternPickerOptions = options;
      return {
        async createWritable() {
          return {
            async write(content) {
              window.__eventPatternWrites.push(String(content));
            },
            async close() {
              window.__eventPatternClosed = true;
            },
          };
        },
      };
    };

    const app = window.__ivyDiagnostics.runtime();
    app.openEventTraceSheet('Trace', {
      sheet_id: 'events-patterns',
      events: [{ text: 'root(a)', address: '0', subs: [{ text: 'child(a)', address: '0/0' }] }],
      patterns: ['root(a)', 'child(a)'],
    }, 'events-patterns');
    app.api.executeAction = async (actionName, args) => {
      window.__eventPatternCalls.push({ actionName, args });
      if (actionName === 'events_load_patterns') {
        if (args.patterns.includes('broken(')) {
          throw new Error('malformed pattern');
        }
        const loaded = String(args.patterns || '').split(/\r?\n/).map((line) => line.trim()).filter(Boolean);
        return { patterns: ['root(a)', 'child(a)'].concat(loaded) };
      }
      if (actionName === 'events_save_patterns') {
        return { content: 'root(a)\nchild(a)\nchild(a)\ndone\n' };
      }
      throw new Error(actionName);
    };
  });

  const list = page.locator('#events-patterns .event-pattern-list');
  await list.selectOption('child(a)');

  await page.locator('#events-patterns .event-pattern-load').click();
  await page.locator('[data-ivy-dialog-text]').fill('broken(');
  await page.locator('[data-ivy-dialog] [data-ivy-dialog-button]', { hasText: 'Load' }).click();
  await expect(page.locator('#statusbar')).toHaveClass(/error/);
  await expect(page.locator('#statusbar .status-message')).toContainText('Load patterns failed: malformed pattern');
  let patternState = await page.evaluate(() => {
    const select = document.querySelector('#events-patterns .event-pattern-list');
    return {
      options: Array.from(select.options).map((option) => option.textContent),
      selected: select.value,
      selectedIndex: select.selectedIndex,
      patterns: window.__ivyDiagnostics.runtime().sheets['events-patterns'].patterns,
    };
  });
  expect(patternState.options).toEqual(['root(a)', 'child(a)']);
  expect(patternState.patterns).toEqual(['root(a)', 'child(a)']);
  expect(patternState.selected).toBe('child(a)');
  expect(patternState.selectedIndex).toBe(1);

  await page.locator('#events-patterns .event-pattern-load').click();
  await page.locator('[data-ivy-dialog-text]').fill('child(a)\r\ndone\n\n');
  await page.locator('[data-ivy-dialog] [data-ivy-dialog-button]', { hasText: 'Load' }).click();
  await expect(page.locator('#events-patterns .event-pattern-list option')).toHaveCount(4);
  patternState = await page.evaluate(() => {
    const select = document.querySelector('#events-patterns .event-pattern-list');
    const loadCalls = window.__eventPatternCalls.filter((call) => call.actionName === 'events_load_patterns');
    return {
      options: Array.from(select.options).map((option) => option.textContent),
      selected: select.value,
      selectedIndex: select.selectedIndex,
      patterns: window.__ivyDiagnostics.runtime().sheets['events-patterns'].patterns,
      lastLoadText: loadCalls[loadCalls.length - 1].args.patterns,
    };
  });
  expect(patternState.options).toEqual(['root(a)', 'child(a)', 'child(a)', 'done']);
  expect(patternState.patterns).toEqual(['root(a)', 'child(a)', 'child(a)', 'done']);
  expect(patternState.selected).toBe('child(a)');
  expect(patternState.selectedIndex).toBe(1);
  expect(patternState.lastLoadText).toMatch(/^child\(a\)\r?\ndone\n\n$/);

  await page.locator('#events-patterns .event-pattern-save').click();
  await page.waitForFunction(() => window.__eventPatternWrites.length === 1);
  const saved = await page.evaluate(() => ({
    content: window.__eventPatternWrites[0],
    closed: !!window.__eventPatternClosed,
    suggestedName: window.__eventPatternPickerOptions && window.__eventPatternPickerOptions.suggestedName,
    accept: window.__eventPatternPickerOptions && window.__eventPatternPickerOptions.types[0].accept['text/plain'],
  }));
  expect(saved).toEqual({
    content: 'root(a)\nchild(a)\nchild(a)\ndone\n',
    closed: true,
    suggestedName: 'event_patterns.pats',
    accept: ['.pats'],
  });
});

test('analysis session history navigation restores graph snapshots and step info', async ({ page }) => {
  await openIvy(page);

  await page.evaluate(() => {
    const app = window.__ivyDiagnostics.runtime();
    app._recordAnalysisHistoryStep('sheet-1', {
      arg: {
        elements: [{ group: 'nodes', data: { id: 'state_0', label: 'init', obj: 'state_0' } }],
        positions: { state_0: { x: 0, y: 0 } },
      },
      concept: {
        elements: [{ group: 'nodes', data: { id: 'concept_init', label: 'initial facts', obj: 'initial facts' } }],
        positions: { concept_init: { x: 0, y: 0 } },
      },
      transition: { label: 'init' },
      step_info: { tactic: 'init', msg: 'Initial analysis state' },
    });
    app._recordAnalysisHistoryStep('sheet-1', {
      arg: {
        elements: [{ group: 'nodes', data: { id: 'state_1', label: 'after connect', obj: 'state_1' } }],
        positions: { state_1: { x: 0, y: 0 } },
      },
      concept: {
        elements: [{ group: 'nodes', data: { id: 'concept_connect', label: 'link(c,s)', obj: 'link(c,s)' } }],
        positions: { concept_connect: { x: 0, y: 0 } },
      },
      transition: { label: 'connect' },
      step_info: { tactic: 'pdr_step', msg: 'Advanced through connect', transition: 'connect' },
    });
  });

  const history = page.locator('#sheet-1 [data-analysis-history]');
  await expect(history).toBeVisible();
  await expect(page.locator('#sheet-1 [data-analysis-history-status]')).toContainText('Step 2 of 2');
  await expect(page.locator('#sheet-1 [data-analysis-step-info]')).toContainText('Tactic: pdr_step');
  await expect(page.locator('#sheet-1 [data-analysis-step-info]')).toContainText('Transition: connect');

  let labels = await page.evaluate(() => {
    const app = window.__ivyDiagnostics.runtime();
    return {
      arg: app.sheets['sheet-1'].argGraph.cy.nodes().map((node) => node.data('label')),
      concept: app.sheets['sheet-1'].conceptGraph.cy.nodes().map((node) => node.data('label')),
    };
  });
  expect(labels.arg).toEqual(['after connect']);
  expect(labels.concept).toEqual(['link(c,s)']);

  await page.locator('#sheet-1 [data-analysis-history-action="prev"]').click();
  await expect(page.locator('#sheet-1 [data-analysis-history-status]')).toContainText('Step 1 of 2');
  await expect(page.locator('#sheet-1 [data-analysis-step-info]')).toContainText('Tactic: init');
  await expect(page.locator('#sheet-1 [data-analysis-step-info]')).toContainText('Transition: init');
  labels = await page.evaluate(() => {
    const app = window.__ivyDiagnostics.runtime();
    return {
      arg: app.sheets['sheet-1'].argGraph.cy.nodes().map((node) => node.data('label')),
      concept: app.sheets['sheet-1'].conceptGraph.cy.nodes().map((node) => node.data('label')),
    };
  });
  expect(labels.arg).toEqual(['init']);
  expect(labels.concept).toEqual(['initial facts']);

  await expect(page.locator('#sheet-1 [data-analysis-history-action="first"]')).toBeDisabled();
  await page.locator('#sheet-1 [data-analysis-history-action="next"]').click();
  await expect(page.locator('#sheet-1 [data-analysis-history-status]')).toContainText('Step 2 of 2');
  await expect(page.locator('#sheet-1 [data-analysis-history-action="next"]')).toBeDisabled();
  await page.locator('#sheet-1 [data-analysis-history-action="first"]').click();
  await expect(page.locator('#sheet-1 [data-analysis-history-status]')).toContainText('Step 1 of 2');
  await page.locator('#sheet-1 [data-analysis-history-action="last"]').click();
  await expect(page.locator('#sheet-1 [data-analysis-history-status]')).toContainText('Step 2 of 2');
});

test('proof goal and CRG interactions update selected goal concept and transition views', async ({ page }) => {
  await openIvy(page);

  await page.evaluate(async () => {
    const app = window.__ivyDiagnostics.runtime();
    app.api.getProofGraph = async () => ({
      elements: [
        { group: 'nodes', data: { id: 'goal_0', label: 'safety', obj: 'goal_0', info: 'goal info' }, classes: 'proof_goal' },
      ],
      positions: { goal_0: { x: 0, y: 0 } },
    });
    app.api.proofGoalAction = async (goalId, actionName) => {
      if (goalId !== 'goal_0' || actionName !== 'view') throw new Error(`${goalId}:${actionName}`);
      return {
        status: 'ok',
        goal: goalId,
        action: actionName,
        info: 'forall X. safe(X)',
        concept: {
          elements: [{ group: 'nodes', data: { id: 'goal_concept', label: 'goal facts', obj: 'goal facts' } }],
          positions: { goal_concept: { x: 0, y: 0 } },
        },
        transition: { label: 'proof goal safety', detail: 'viewed from proof goal' },
        crg: {
          nodes: [{
            id: 'crg_1',
            label: 'crg after connect',
            transition: { label: 'connect transition', detail: 'from crg node' },
            concept: {
              elements: [{ group: 'nodes', data: { id: 'crg_concept', label: 'crg facts', obj: 'crg facts' } }],
              positions: { crg_concept: { x: 0, y: 0 } },
            },
          }],
        },
      };
    };
    await app._refreshProofGoalPane('sheet-1');
  });

  await expect(page.locator('#sheet-1 [data-proof-goal-pane]')).toBeVisible();
  await page.evaluate(async () => {
    const app = window.__ivyDiagnostics.runtime();
    const node = app.sheets['sheet-1'].proofGraph.cy.getElementById('goal_0');
    node.emit('tap', { target: node });
    await new Promise((resolve) => setTimeout(resolve, 0));
  });

  await expect(page.locator('#sheet-1 [data-selected-proof-goal]')).toContainText('goal_0');
  await expect(page.locator('#sheet-1 [data-proof-goal-info]')).toContainText('forall X. safe(X)');
  await expect(page.locator('#sheet-1 [data-transition-view]')).toContainText('proof goal safety');
  await expect(page.locator('#sheet-1 [data-crg-node="crg_1"]')).toContainText('crg after connect');
  let state = await page.evaluate(() => {
    const app = window.__ivyDiagnostics.runtime();
    return {
      selectedGoal: app.sheets['sheet-1'].selectedProofGoal,
      conceptLabels: app.sheets['sheet-1'].conceptGraph.cy.nodes().map((node) => node.data('label')),
    };
  });
  expect(state.selectedGoal).toBe('goal_0');
  expect(state.conceptLabels).toEqual(['goal facts']);

  await page.locator('#sheet-1 [data-crg-node="crg_1"]').click();
  await expect(page.locator('#sheet-1 [data-transition-view]')).toContainText('connect transition');
  state = await page.evaluate(() => {
    const app = window.__ivyDiagnostics.runtime();
    return {
      selectedCrgNode: app.sheets['sheet-1'].selectedCrgNode,
      conceptLabels: app.sheets['sheet-1'].conceptGraph.cy.nodes().map((node) => node.data('label')),
    };
  });
  expect(state.selectedCrgNode).toBe('crg_1');
  expect(state.conceptLabels).toEqual(['crg facts']);
});

test('event viewer launch mode opens an iev trace without a model file', async ({ page }) => {
  const params = new URLSearchParams({
    'event-viewer': '1',
    'event-trace': 'root(a){child(b)}; done',
    'event-filename': 'trace.iev',
  });
  await openIvyAt(page, `/?${params.toString()}`);

  await expect(page.locator('body')).toHaveAttribute('data-event-viewer-only', 'true');
  await expect(page.locator('.sheet-tab[data-sheet="sheet-1"]')).toBeHidden();
  await expect(page.locator('.sheet-tab.active')).toContainText('trace.iev');
  await expect(page.locator('[data-event-address="0"]')).toContainText('root(a)');
  await page.locator('[data-event-toggle="0"]').click();
  await expect(page.locator('[data-event-address="0/0"]')).toContainText('child(b)');
  await expect(page.locator('[data-event-address="1"]')).toContainText('done');

  const state = await page.evaluate(() => {
    const app = window.__ivyDiagnostics.runtime();
    return {
      activeSheetId: app.activeSheetId,
      fileName: app._persistedFileName || '',
      fileContent: app._persistedFileContent || '',
      editorContent: app.cmEditor && app.cmEditor.getValue ? app.cmEditor.getValue() : '',
      loadedFileText: document.querySelector('#loaded-file')?.textContent || '',
    };
  });
  expect(state.activeSheetId).toMatch(/^sht\d+$/);
  expect(state.fileName).toBe('');
  expect(state.fileContent).toBe('');
  expect(state.editorContent).toBe('');
  expect(state.loadedFileText).not.toContain('.ivy');
});

test('event trace sheets expose compact and detailed serialized trace text', async ({ page }) => {
  await openIvy(page);

  await page.evaluate(() => {
    window.__ivyDiagnostics.runtime().openEventTraceSheet('Trace', {
      sheet_id: 'events-trace-text',
      events: [
        { text: 'connect(c0)', kind: 'env_call', line: 'model.ivy:12', state_key: 's0', state: { pc: 'idle' }, subs: [
          { text: 'helper(c0)', kind: 'call', line: 'model.ivy:8', state: { owner: 'c0' } },
        ] },
        { text: 'connect(c0)', kind: 'env_call', state_key: 's0', state: { pc: 'idle' } },
      ],
      trace_text: '> connect(c0)\n  < helper(c0)\n> connect(c0)\n',
      trace_text_detailed: 'model.ivy:12\nconnect(c0)\n{\n    model.ivy:8\n    helper(c0)\n}\n[\n    pc = idle\n]\n\n--- the following repeats infinitely ---\n',
      patterns: [],
    }, 'events-trace-text');
  });

  const tracePane = page.locator('#events-trace-text [data-event-trace-text]');
  await expect(tracePane).toBeVisible();
  await expect(tracePane).toContainText('> connect(c0)');
  await expect(tracePane).toContainText('< helper(c0)');
  await expect(tracePane).not.toContainText('model.ivy:12');

  await page.locator('#events-trace-text .event-trace-detailed-toggle').check();
  await expect(tracePane).toContainText('model.ivy:12');
  await expect(tracePane).toContainText('pc = idle');
  await expect(tracePane).toContainText('repeats infinitely');
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

test('failed ARG node safety result can open its trace ARG in a sheet', async ({ page }) => {
  await openIvy(page);
  await page.waitForFunction(() => window.__ivyDiagnostics.runtime().argGraph);

  await page.evaluate(() => {
    const app = window.__ivyDiagnostics.runtime();
    window._nodeSafetyCalls = [];
    app.api.argNodeAction = async (node, action, args) => {
      window._nodeSafetyCalls.push({ node, action, args });
      return {
        status: 'ok',
        result: 'fail',
        safe: false,
        message: 'The node is unsafe: View error trace?',
        arg: {
          elements: [
            { group: 'nodes', data: { id: 'state_0', obj: 'state_0', label: '0' } },
          ],
        },
        trace_arg: {
          elements: [
            { group: 'nodes', data: { id: 'state_0', obj: 'state_0', label: '0' } },
            { group: 'nodes', data: { id: 'state_1', obj: 'state_1', label: '1' }, classes: 'state marked_state' },
            { group: 'edges', data: { id: 'trace_edge', source: 'state_0', target: 'state_1', label: 'trace' } },
          ],
        },
        trace_sheet_id: 'sheet-9',
      };
    };
    app.argGraph.update([
      { group: 'nodes', data: { id: 'state_0', obj: 'state_0', label: '0' } },
    ]);
    app.onArgNodeRightClick(app.argGraph.cy.getElementById('state_0').data(), { x: 16, y: 16 }, 'sheet-1');
  });

  await page.locator('.context-menu-item', { hasText: 'Check safety' }).click();
  await page.waitForFunction(() => window._nodeSafetyCalls && window._nodeSafetyCalls.length === 1);
  await expect(page.locator('[data-check-view-trace]')).toBeVisible();
  await page.locator('[data-check-view-trace]').click();

  const result = await page.evaluate(() => {
    const app = window.__ivyDiagnostics.runtime();
    const sheet = app.sheets[app.activeSheetId];
    return {
      activeSheet: app.activeSheetId,
      labels: sheet.argGraph.cy.nodes().map((n) => n.data('label')),
      marked: sheet.argGraph.cy.nodes('.marked_state').map((n) => n.data('label')),
      edgeLabels: sheet.argGraph.cy.edges().map((e) => e.data('label')),
    };
  });

  expect(result.activeSheet).toBe('sheet-9');
  expect(result.labels).toEqual(['0', '1']);
  expect(result.marked).toEqual(['1']);
  expect(result.edgeLabels).toEqual(['trace']);

  const infoPanelIds = await page.evaluate(() => Array.from(document.querySelectorAll('.info-panel')).map((panel) => panel.id));
  expect(infoPanelIds).toContain('info-panel');
  expect(infoPanelIds).toContain('info-panel-9');

  const traceDetails = page.locator('#sheet-9 .info-panel');
  const traceDetailsHeader = page.locator('#sheet-9 .info-header');
  const beforeDetailsBox = await traceDetails.boundingBox();
  const headerBox = await traceDetailsHeader.boundingBox();
  test.skip(!beforeDetailsBox || !headerBox, 'trace details pane has no measurable box');

  await page.mouse.move(headerBox.x + headerBox.width / 2, headerBox.y + headerBox.height / 2);
  await page.mouse.down();
  await page.mouse.move(headerBox.x + headerBox.width / 2, headerBox.y + headerBox.height / 2 + 40);
  await page.mouse.up();

  const afterDetailsBox = await traceDetails.boundingBox();
  expect(afterDetailsBox?.height || 0).toBeLessThan(beforeDetailsBox.height);
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
