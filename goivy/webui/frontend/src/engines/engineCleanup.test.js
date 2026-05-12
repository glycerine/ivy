import { describe, expect, it } from 'vitest';
import * as engines from './index.js';
import { IvyApiAdapter } from './ivyApiAdapter.js';

describe('engine cleanup', () => {
  it('exports only the canonical adapter contract and real hosted implementation', () => {
    expect(engines.IvyApiAdapter).toBe(IvyApiAdapter);
    expect(engines.HostedGoIvyApiAdapter).toBeTypeOf('function');
    expect('IvyEngine' in engines).toBe(false);
    expect('HostedGoEngine' in engines).toBe(false);
    expect('WanixEngine' in engines).toBe(false);
  });

  it('keeps raw endpoint helpers off the public API contract', () => {
    const api = new IvyApiAdapter();

    expect(api.requestSession).toBeUndefined();
    expect(api.fetchSession).toBeUndefined();
    expect(api.subscribeEvents).toBeUndefined();
  });
});
