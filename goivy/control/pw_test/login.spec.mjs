import { expect, test } from '@playwright/test';

test('Feature: Login - unauthenticated /auth/me returns anonymous state', async ({ request }) => {
  const response = await request.get('/auth/me');
  expect(response.ok()).toBe(true);
  const body = await response.json();
  expect(body).toEqual({ authenticated: false });
});

test('Feature: Login - user starts login through Casdoor redirect', async ({ request }) => {
  const response = await request.get('/auth/login', { maxRedirects: 0 });
  expect(response.status()).toBe(302);
  const location = response.headers().location;
  expect(location).toBeTruthy();
  const current = new URL(location);
  expect(`${current.protocol}//${current.host}${current.pathname}`).toBe('http://127.0.0.1:18082/login/oauth/authorize');
  expect(current.searchParams.get('response_type')).toBe('code');
  expect(current.searchParams.get('client_id')).toBe('ivy-control-local');
  expect(current.searchParams.get('redirect_uri')).toBe('http://127.0.0.1:18080/auth/callback');
  expect(current.searchParams.get('scope')).toContain('openid');
  expect(current.searchParams.get('state')).toBeTruthy();
  expect(current.searchParams.get('nonce')).toBeTruthy();
});

test.fixme('Feature: Sign-up - new user signs up and verifies email', async () => {});
test.fixme('Feature: Account recovery - user resets password from recovery email', async () => {});
test.fixme('Feature: Billing information - account owner adds billing information', async () => {});
test.fixme('Feature: Alpha tester dashboard - admin invites tester and grants project access', async () => {});
