import { describe, expect, it, vi } from 'vitest';
import { createAuthClient } from './authClient.js';

describe('authClient', () => {
  it('loads the current app session', async () => {
    const fetchImpl = vi.fn(async () => new Response(JSON.stringify({ authenticated: true }), { status: 200 }));
    const client = createAuthClient(fetchImpl);

    await expect(client.currentSession()).resolves.toEqual({ authenticated: true });

    expect(fetchImpl).toHaveBeenCalledWith('/auth/me', {
      method: 'GET',
      headers: {
        accept: 'application/json',
      },
    });
  });

  it('requests an email login link without exposing provider details to the caller', async () => {
    const fetchImpl = vi.fn(async () => new Response(JSON.stringify({ ok: true }), { status: 200 }));
    const client = createAuthClient(fetchImpl);

    await expect(client.requestEmailLogin('alice@example.test')).resolves.toEqual({ ok: true });

    expect(fetchImpl).toHaveBeenCalledWith('/auth/email/request', {
      method: 'POST',
      headers: {
        'content-type': 'application/json',
      },
      body: JSON.stringify({ email: 'alice@example.test' }),
    });
  });

  it('consumes an email login token', async () => {
    const fetchImpl = vi.fn(async () => new Response(JSON.stringify({ ok: true }), { status: 200 }));
    const client = createAuthClient(fetchImpl);

    await expect(client.consumeEmailLogin('token-123')).resolves.toEqual({ ok: true });

    expect(fetchImpl).toHaveBeenCalledWith('/auth/email/consume', {
      method: 'POST',
      headers: {
        'content-type': 'application/json',
      },
      body: JSON.stringify({ token: 'token-123' }),
    });
  });

  it('starts and finishes passkey registration', async () => {
    const fetchImpl = vi.fn(async (path) => {
      if (path.endsWith('/options')) {
        return new Response(JSON.stringify({ publicKey: { challenge: 'abc' } }), { status: 200 });
      }
      return new Response(JSON.stringify({ authenticated: true }), { status: 200 });
    });
    const client = createAuthClient(fetchImpl);

    await expect(client.beginPasskeyRegistration()).resolves.toEqual({ publicKey: { challenge: 'abc' } });
    await expect(client.finishPasskeyRegistration({ id: 'credential-1' })).resolves.toEqual({ authenticated: true });

    expect(fetchImpl).toHaveBeenNthCalledWith(1, '/auth/passkeys/register/options', {
      method: 'POST',
      headers: {
        accept: 'application/json',
      },
    });
    expect(fetchImpl).toHaveBeenNthCalledWith(2, '/auth/passkeys/register/finish', {
      method: 'POST',
      headers: {
        'content-type': 'application/json',
      },
      body: JSON.stringify({ id: 'credential-1' }),
    });
  });

  it('starts and finishes passkey login', async () => {
    const fetchImpl = vi.fn(async (path) => {
      if (path.endsWith('/options')) {
        return new Response(JSON.stringify({ publicKey: { challenge: 'abc' } }), { status: 200 });
      }
      return new Response(JSON.stringify({ authenticated: true }), { status: 200 });
    });
    const client = createAuthClient(fetchImpl);

    await expect(client.beginPasskeyLogin()).resolves.toEqual({ publicKey: { challenge: 'abc' } });
    await expect(client.finishPasskeyLogin({ id: 'credential-1' })).resolves.toEqual({ authenticated: true });

    expect(fetchImpl).toHaveBeenNthCalledWith(1, '/auth/passkeys/login/options', {
      method: 'POST',
      headers: {
        accept: 'application/json',
      },
    });
    expect(fetchImpl).toHaveBeenNthCalledWith(2, '/auth/passkeys/login/finish', {
      method: 'POST',
      headers: {
        'content-type': 'application/json',
      },
      body: JSON.stringify({ id: 'credential-1' }),
    });
  });

  it('loads admin unverified emails', async () => {
    const fetchImpl = vi.fn(async () => new Response(JSON.stringify({ emails: [] }), { status: 200 }));
    const client = createAuthClient(fetchImpl);

    await expect(client.listUnverifiedEmails()).resolves.toEqual({ emails: [] });

    expect(fetchImpl).toHaveBeenCalledWith('/admin/api/unverified-emails', {
      method: 'GET',
      headers: {
        accept: 'application/json',
      },
    });
  });
});
