const SAVE_DIALOG_SPECS = {
  model: {
    description: 'Ivy model files',
    mimeType: 'text/plain',
    extensions: ['.ivy'],
  },
  analysisState: {
    description: 'IvyWeb analysis state files',
    mimeType: 'application/json',
    extensions: ['.ivyweb.json', '.json'],
  },
  invariant: {
    description: 'Ivy invariant files',
    mimeType: 'text/plain',
    extensions: ['.ivy'],
  },
  abstraction: {
    description: 'Ivy abstraction files',
    mimeType: 'text/plain',
    extensions: ['.ivy'],
  },
  dot: {
    description: 'Graphviz DOT files',
    mimeType: 'text/vnd.graphviz',
    extensions: ['.dot'],
  },
  eventPatterns: {
    description: 'Ivy event pattern files',
    mimeType: 'text/plain',
    extensions: ['.pats'],
  },
};

export function saveDialogSpec(kind) {
  const spec = SAVE_DIALOG_SPECS[kind];
  if (!spec) throw new Error(`unknown save dialog kind: ${kind}`);
  return {
    description: spec.description,
    mimeType: spec.mimeType,
    extensions: spec.extensions.slice(),
  };
}

export function saveMimeType(kind) {
  return saveDialogSpec(kind).mimeType;
}

export function savePickerOptions(kind, suggestedName) {
  const spec = saveDialogSpec(kind);
  return {
    suggestedName,
    types: [{
      description: spec.description,
      accept: { [spec.mimeType]: spec.extensions },
    }],
  };
}

export function ivyStem(filename, fallback = 'model') {
  const name = String(filename || fallback || 'model');
  return name.replace(/\.ivy$/i, '') || fallback || 'model';
}

export function analysisStateSuggestedName(filename) {
  return `${ivyStem(filename, 'ivy_analysis')}.ivyweb.json`;
}

export function invariantSuggestedName(filename) {
  return `${ivyStem(filename, 'model')}_invariant.ivy`;
}

export function abstractionSuggestedName(filename) {
  return `${ivyStem(filename, 'abstraction')}_abstraction.ivy`;
}

export function eventPatternsSuggestedName() {
  return 'event_patterns.pats';
}
