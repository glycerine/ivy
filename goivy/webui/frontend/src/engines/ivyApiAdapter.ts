const DEFAULT_CAPABILITIES = Object.freeze({
  offline: false,
  persistentJobs: false,
  cancelJob: false,
  eventStream: false,
  parallelJobs: false,
});

export class IvyApiAdapter {
  [key: string]: any;
  kind: string;
  capabilities: Record<string, boolean>;
  onConnectionLost: null | (() => void);
  _unsubscribe: null | (() => void);

  constructor() {
    this.kind = this.kind || 'ivy-api';
    this.capabilities = this.capabilities || { ...DEFAULT_CAPABILITIES };
    this.onConnectionLost = null;
    this._unsubscribe = null;
  }

  async createSession(): Promise<any> {
    this._notImplemented('createSession');
  }

  async loadModel(_model: any, _requestOptions: RequestInit = {}): Promise<any> {
    this._notImplemented('loadModel');
  }

  async runCommand(_intent: any, _requestOptions: RequestInit = {}): Promise<any> {
    this._notImplemented('runCommand');
  }

  async getSnapshot(_request: any = {}): Promise<any> {
    this._notImplemented('getSnapshot');
  }

  subscribe(_onEvent: any, _onError: any): any {
    this._notImplemented('subscribe');
  }

  async cancelJob(_jobId: any): Promise<any> {
    this._notImplemented('cancelJob');
  }

  async saveSession(): Promise<any> {
    this._notImplemented('saveSession');
  }

  async loadFile(file, options = {}, requestOptions: RequestInit = {}) {
    return this.loadModel({ file, ...(options || {}) }, requestOptions);
  }

  async reloadContent(content, filename, options = {}, requestOptions: RequestInit = {}) {
    return this.loadModel({ content, filename: filename || 'model.ivy', ...(options || {}) }, requestOptions);
  }

  async getARG(options = {}) {
    const snapshot = await this.getSnapshot({ arg: options || {} });
    return snapshot && snapshot.arg;
  }

  async getCTIARG(options = {}) {
    const snapshot = await this.getSnapshot({ ctiArg: options || {} });
    return snapshot && snapshot.ctiArg;
  }

  async getMenus(options = {}) {
    const snapshot = await this.getSnapshot({ menus: options || {} });
    return snapshot && snapshot.menus;
  }

  async getConceptGraph(nodeId, sheetId) {
    const snapshot = await this.getSnapshot({ concept: { nodeId, sheetId } });
    return snapshot && snapshot.concept;
  }

  async getProofGraph() {
    const snapshot = await this.getSnapshot({ proof: true });
    return snapshot && snapshot.proof;
  }

  async executeAction(action, args) {
    return this.runCommand({ commandId: action, args: args || {} });
  }

  async splitConcept(concept, splitBy) {
    return this.runCommand({ commandId: 'concept.split', args: { concept, splitBy } });
  }

  async supposeEmpty(concept) {
    return this.runCommand({ commandId: 'concept.empty', args: { concept } });
  }

  async removeConcept(concept) {
    return this.runCommand({ commandId: 'concept.remove', args: { concept } });
  }

  async undo() {
    return this.runCommand({ commandId: 'concept.undo', args: {} });
  }

  async materializeNode(concept) {
    return this.runCommand({ commandId: 'concept.materializeNode', args: { concept } });
  }

  async materializeEdge(relation, source, target, positive) {
    return this.runCommand({
      commandId: 'concept.materializeEdge',
      args: { relation, source, target, positive },
    });
  }

  async addProjection(name, concept) {
    return this.runCommand({ commandId: 'concept.projection', args: { name, concept } });
  }

  async runCheck(mode, options = {}, requestOptions: RequestInit = {}) {
    return this.runCommand(
      { commandId: `check.${mode || 'pdr'}`, args: { ...options, mode: mode || 'pdr' } },
      requestOptions,
    );
  }

  async resetDomain() {
    return this.runCommand({ commandId: 'concept.reset', args: {} });
  }

  async diagramDomain() {
    return this.runCommand({ commandId: 'concept.diagram', args: {} });
  }

  async getToggles() {
    const snapshot = await this.getSnapshot({ toggles: true });
    return snapshot && snapshot.toggles;
  }

  async setToggles(toggles) {
    return this.runCommand({ commandId: 'toggles.set', args: toggles || {} });
  }

  async argNodeAction(nodeId, action, args) {
    return this.runCommand({
      commandId: action,
      target: { kind: 'arg', nodeId },
      args: args || {},
    });
  }

  async proofGoalAction(goalId, action) {
    return this.runCommand({
      commandId: 'proof.action',
      target: { kind: 'proof', goalId },
      args: { action },
    });
  }

  connectEvents(onEvent) {
    this.disconnectEvents();
    this._unsubscribe = this.subscribe(onEvent, () => {
      if (this.onConnectionLost) this.onConnectionLost();
    });
    return this._unsubscribe;
  }

  disconnectEvents() {
    if (this._unsubscribe) {
      this._unsubscribe();
      this._unsubscribe = null;
    }
  }

  _notImplemented(method: string): never {
    throw new Error(`${this.constructor.name}.${method} is not implemented`);
  }
}
