export class IvyEngine {
  async createSession() {
    throw new Error('createSession is not implemented');
  }

  async loadModel(_model) {
    throw new Error('loadModel is not implemented');
  }

  async getArg(_options = {}) {
    throw new Error('getArg is not implemented');
  }

  async getConcept(_options = {}) {
    throw new Error('getConcept is not implemented');
  }

  async getMenus() {
    throw new Error('getMenus is not implemented');
  }

  async check(_options = {}) {
    throw new Error('check is not implemented');
  }

  async runAction(_action) {
    throw new Error('runAction is not implemented');
  }

  async runArgAction(_action) {
    throw new Error('runArgAction is not implemented');
  }

  async getToggles() {
    throw new Error('getToggles is not implemented');
  }

  async setToggles(_toggles) {
    throw new Error('setToggles is not implemented');
  }

  async requestSession(_suffix, _options = {}) {
    throw new Error('requestSession is not implemented');
  }

  async fetchSession(_suffix, _options = {}) {
    throw new Error('fetchSession is not implemented');
  }

  subscribeEvents(_onEvent, _onError) {
    return () => {};
  }
}
