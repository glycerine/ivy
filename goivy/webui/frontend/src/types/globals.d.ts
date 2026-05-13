export {};

declare global {
  interface Window {
    __IVY_ENGINE__?: any;
    __ivyDiagnostics?: any;
    __ivyInitError?: string;
    CodeMirror?: any;
    cytoscape?: any;
    cytoscapeDagre?: any;
    IvyGraph?: any;
    CONCEPT_STYLE?: any;
    ARG_STYLE?: any;
    PROOF_STYLE?: any;
    showOpenFilePicker?: any;
    showSaveFilePicker?: any;
  }

  interface EventTarget {
    [key: string]: any;
  }

  interface Node {
    [key: string]: any;
  }

  interface Element {
    [key: string]: any;
  }

  interface HTMLElement {
    [key: string]: any;
    __ivyCodeMirrorEditor?: any;
    _ivyContextSuppressed?: boolean;
    _ivyJobControlBound?: boolean;
    __ivyMenuItem?: any;
  }

  var CodeMirror: any;
}
