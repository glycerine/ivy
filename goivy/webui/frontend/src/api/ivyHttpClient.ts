export class IvyHttpClient {
  baseURL: string;
  fetchImpl: typeof fetch;

  constructor({
    baseURL = '',
    fetchImpl = typeof globalThis.fetch === 'function' ? globalThis.fetch.bind(globalThis) : undefined,
  }: { baseURL?: string; fetchImpl?: typeof fetch } = {}) {
    if (typeof fetchImpl !== 'function') {
      throw new Error('IvyHttpClient requires a fetch implementation');
    }
    this.baseURL = baseURL;
    this.fetchImpl = fetchImpl;
  }

  async request(path: string, options: RequestInit = {}): Promise<any> {
    const response = await this.fetch(path, options);
    const contentType = response.headers?.get?.('content-type') || '';
    if (contentType.includes('application/json')) {
      return response.json();
    }
    return response.text();
  }

  async fetch(path: string, options: RequestInit = {}): Promise<Response> {
    const response = await this.fetchImpl(this.baseURL + path, options);
    if (!response.ok) {
      let text = '';
      try {
        text = await response.text();
      } catch (_err) {
        text = '';
      }
      throw new Error(`API error ${response.status}: ${text || response.statusText}`);
    }
    return response;
  }
}
