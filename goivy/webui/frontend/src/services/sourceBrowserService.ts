type SourceBrowserPayload = {
  source?: string;
  content?: string;
  file?: string;
  filename?: string;
  lineno?: number;
  line?: number;
};

let raiseSeq = 0;

function sourceBrowserHost(doc: Document) {
  return doc.querySelector('.sheet-content.active') || doc.body;
}

function ensureSourceBrowser(doc: Document) {
  let browser = doc.querySelector('[data-source-browser]') as HTMLElement | null;
  if (browser) return browser;

  browser = doc.createElement('section');
  browser.className = 'source-browser panel';
  browser.setAttribute('data-source-browser', 'true');
  browser.setAttribute('aria-label', 'Source browser');
  browser.innerHTML = [
    '<div class="panel-header source-browser-header">',
    '  <span class="panel-title" data-source-browser-title>Source</span>',
    '</div>',
    '<pre class="source-browser-content" data-source-browser-content></pre>',
  ].join('');
  sourceBrowserHost(doc).appendChild(browser);
  return browser;
}

function renderSourceLines(pre: HTMLElement, doc: Document, lines: string[], highlightLine: number) {
  pre.textContent = '';
  lines.forEach((line, index) => {
    const lineNo = index + 1;
    const row = doc.createElement('span');
    row.className = 'source-browser-line';
    row.setAttribute('data-source-line', String(lineNo));
    row.textContent = line;
    if (lineNo === highlightLine) {
      row.classList.add('source-browser-line-highlight');
    }
    pre.appendChild(row);
    if (index < lines.length - 1) pre.appendChild(doc.createTextNode('\n'));
  });
}

export function openSourceBrowser(app: any, payload: SourceBrowserPayload = {}, {
  doc = globalThis.document,
}: { doc?: Document } = {}) {
  if (!doc) return null;
  const filename = payload.file || payload.filename || 'source';
  const content = payload.source ?? payload.content ?? '';
  const highlightLine = Math.max(0, Number(payload.lineno || payload.line || 0) || 0);
  const lines = String(content).split('\n');
  const browser = ensureSourceBrowser(doc);
  const title = browser.querySelector('[data-source-browser-title]');
  const pre = browser.querySelector('[data-source-browser-content]') as HTMLElement | null;

  browser.setAttribute('data-source-file', filename);
  browser.setAttribute('data-source-line', String(highlightLine));
  raiseSeq += 1;
  browser.setAttribute('data-raise-seq', String(raiseSeq));
  browser.classList.add('source-browser-raised');
  if (browser.parentElement) browser.parentElement.appendChild(browser);
  if (title) {
    title.textContent = `${filename}${highlightLine ? `:${highlightLine}` : ''}`;
  }
  if (pre) {
    renderSourceLines(pre, doc, lines, highlightLine);
    const highlighted = pre.querySelector('.source-browser-line-highlight') as HTMLElement | null;
    if (highlighted && typeof highlighted.scrollIntoView === 'function') {
      highlighted.scrollIntoView({ block: 'center' });
    }
  }

  const state = {
    filename,
    content,
    highlightLine,
    lines,
  };
  if (app) app.sourceBrowserState = state;
  return state;
}
