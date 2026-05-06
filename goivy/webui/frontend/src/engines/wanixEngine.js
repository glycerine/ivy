import { IvyEngine } from './ivyEngine.js';

export class WanixEngine extends IvyEngine {
  constructor() {
    super();
    this.kind = 'wanix';
  }

  unavailable() {
    throw new Error('Wanix Ivy engine is not implemented yet');
  }

  async createSession() {
    this.unavailable();
  }

  async loadModel() {
    this.unavailable();
  }

  async getArg() {
    this.unavailable();
  }

  async getConcept() {
    this.unavailable();
  }

  async getMenus() {
    this.unavailable();
  }

  async check() {
    this.unavailable();
  }

  async runAction() {
    this.unavailable();
  }

  async runArgAction() {
    this.unavailable();
  }

  async getToggles() {
    this.unavailable();
  }

  async setToggles() {
    this.unavailable();
  }

  async requestSession() {
    this.unavailable();
  }

  async fetchSession() {
    this.unavailable();
  }
}
