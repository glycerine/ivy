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
