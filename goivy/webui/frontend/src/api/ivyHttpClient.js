export class IvyHttpClient {
  constructor({ baseURL = '', fetchImpl = globalThis.fetch } = {}) {
    if (typeof fetchImpl !== 'function') {
      throw new Error('IvyHttpClient requires a fetch implementation');
    }
    this.baseURL = baseURL;
    this.fetchImpl = fetchImpl;
  }

  async request(path, options = {}) {
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

    const contentType = response.headers?.get?.('content-type') || '';
    if (contentType.includes('application/json')) {
      return response.json();
    }
    return response.text();
  }
}
