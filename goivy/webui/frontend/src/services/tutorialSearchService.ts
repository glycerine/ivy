import MiniSearch from 'minisearch';

export type TutorialSearchDocument = {
  id: string;
  title: string;
  url: string;
  body: string;
};

type TutorialSearchOptions = {
  doc?: Document;
  fetchImpl?: typeof fetch;
  navigateTo?: (url: string) => void;
  setStatus?: (message: string, kind?: string) => void;
  endpoint?: string;
};

type TutorialSearchResult = TutorialSearchDocument & {
  score?: number;
};

export function setupTutorialSearch({
  doc = document,
  fetchImpl = globalThis.fetch ? globalThis.fetch.bind(globalThis) : undefined,
  navigateTo = () => undefined,
  setStatus = () => undefined,
  endpoint = '/api/tutorial/search-docs',
}: TutorialSearchOptions = {}) {
  const input = doc.getElementById('tutorial-search') as HTMLInputElement | null;
  const resultsBox = doc.getElementById('tutorial-search-results') as HTMLElement | null;
  let miniSearch: MiniSearch<TutorialSearchDocument> | null = null;
  let loadPromise: Promise<MiniSearch<TutorialSearchDocument> | null> | null = null;
  let lastResults: TutorialSearchResult[] = [];

  function clear() {
    if (input) input.value = '';
    lastResults = [];
    renderResults('', []);
  }

  function hide() {
    if (!resultsBox) return;
    resultsBox.hidden = true;
    resultsBox.replaceChildren();
  }

  function ready() {
    return loadPromise || Promise.resolve(miniSearch);
  }

  async function ensureIndex() {
    if (miniSearch) return miniSearch;
    if (loadPromise) return loadPromise;
    if (!fetchImpl) return null;
    loadPromise = fetchImpl(endpoint)
      .then(async (resp: any) => {
        if (!resp || !resp.ok) {
          throw new Error('tutorial search index unavailable');
        }
        const payload = await resp.json();
        const documents = Array.isArray(payload?.documents) ? payload.documents : [];
        const index = new MiniSearch<TutorialSearchDocument>({
          fields: ['title', 'body'],
          storeFields: ['title', 'url', 'body'],
          searchOptions: {
            boost: { title: 2 },
            fuzzy: 0.35,
            prefix: true,
          },
        });
        index.addAll(documents.filter(isTutorialSearchDocument));
        miniSearch = index;
        return index;
      })
      .catch((err) => {
        setStatus('Tutorial search failed: ' + (err?.message || String(err)), 'error');
        return null;
      });
    return loadPromise;
  }

  async function search(query: string) {
    const trimmed = query.trim();
    if (trimmed.length < 2) {
      renderResults(trimmed, []);
      return [];
    }
    const index = await ensureIndex();
    if (!index) {
      renderResults(trimmed, []);
      return [];
    }
    const found = index.search(trimmed)
      .slice(0, 8)
      .map(searchResultToTutorialDocument)
      .filter(isTutorialSearchDocument) as TutorialSearchResult[];
    lastResults = found;
    renderResults(trimmed, found);
    return found;
  }

  function renderResults(query: string, results: TutorialSearchResult[]) {
    if (!resultsBox) return;
    resultsBox.replaceChildren();
    if (!query) {
      resultsBox.hidden = true;
      return;
    }
    resultsBox.hidden = false;
    if (results.length === 0) {
      const empty = doc.createElement('div');
      empty.className = 'tutorial-search-empty';
      empty.textContent = 'No help pages found';
      resultsBox.appendChild(empty);
      return;
    }
    for (const result of results) {
      const button = doc.createElement('button');
      button.type = 'button';
      button.className = 'tutorial-search-result';
      button.addEventListener('click', () => {
        navigateTo(result.url);
        clear();
      });

      const title = doc.createElement('span');
      title.className = 'tutorial-search-result-title';
      title.textContent = result.title || result.url;
      button.appendChild(title);

      const url = doc.createElement('span');
      url.className = 'tutorial-search-result-url';
      url.textContent = result.url;
      button.appendChild(url);

      const snippet = doc.createElement('span');
      snippet.className = 'tutorial-search-result-snippet';
      snippet.textContent = tutorialSearchSnippet(result.body || '', query);
      button.appendChild(snippet);

      resultsBox.appendChild(button);
    }
  }

  if (input && resultsBox) {
    input.addEventListener('focus', () => {
      void ensureIndex();
    });
    input.addEventListener('input', () => {
      void search(input.value);
    });
    input.addEventListener('keydown', (event) => {
      if (event.key === 'Escape') {
        event.preventDefault();
        clear();
        return;
      }
      if (event.key === 'Enter' && lastResults.length > 0) {
        event.preventDefault();
        navigateTo(lastResults[0].url);
        clear();
      }
    });
    doc.addEventListener('click', (event) => {
      const target = event.target as Node | null;
      if (!target || input.contains(target) || resultsBox.contains(target)) return;
      hide();
    });
  }

  return { clear, hide, ready, search };
}

function isTutorialSearchDocument(value: any): value is TutorialSearchDocument {
  return !!value &&
    typeof value.id === 'string' &&
    typeof value.title === 'string' &&
    typeof value.url === 'string' &&
    typeof value.body === 'string';
}

function searchResultToTutorialDocument(value: any): TutorialSearchResult {
  return {
    id: String(value?.id || value?.url || ''),
    title: String(value?.title || ''),
    url: String(value?.url || ''),
    body: String(value?.body || ''),
    score: typeof value?.score === 'number' ? value.score : undefined,
  };
}

export function tutorialSearchSnippet(body: string, query: string) {
  const compact = body.replace(/\s+/g, ' ').trim();
  if (compact.length <= 160) return compact;
  const terms = query.toLowerCase().split(/\s+/).filter(Boolean);
  const lower = compact.toLowerCase();
  let start = -1;
  for (const term of terms) {
    const idx = lower.indexOf(term);
    if (idx >= 0 && (start < 0 || idx < start)) start = idx;
  }
  if (start < 0) start = 0;
  start = Math.max(0, start - 45);
  let snippet = compact.slice(start, start + 160);
  if (start > 0) snippet = '...' + snippet;
  if (start + 160 < compact.length) snippet += '...';
  return snippet;
}
