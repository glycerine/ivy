import { expect, test } from '@playwright/test';

test('Feature: Login - unauthenticated /auth/me returns anonymous state', async ({ request }) => {
  const response = await request.get('/auth/me');
  expect(response.ok()).toBe(true);
  const body = await response.json();
  expect(body).toEqual({ authenticated: false });
});

test('Feature: Login - user can request an email magic link without sending real email', async ({ request }) => {
  const email = 'alice@example.test';
  const response = await request.post('/auth/email/request', {
    data: { email },
  });
  expect(response.ok()).toBe(true);
  expect(await response.json()).toEqual({ ok: true });

  const latest = await request.get(`/test/email/latest?email=${encodeURIComponent(email)}`);
  expect(latest.ok()).toBe(true);
  const message = await latest.json();
  expect(message.toEmail).toBe(email);
  expect(message.loginUrl).toContain('/auth/email/continue#token=');
});

test('Feature: Sign-up - email magic link verifies email and creates app session', async ({ page, request }) => {
  const email = 'signup@example.test';
  await request.post('/auth/email/request', {
    data: { email },
  });
  const latest = await request.get(`/test/email/latest?email=${encodeURIComponent(email)}`);
  const message = await latest.json();

  await page.goto(message.loginUrl);
  await expect(page).toHaveURL('http://127.0.0.1:18080/verified');

  const cookies = await page.context().cookies();
  const session = cookies.find((cookie) => cookie.name === 'ivy_webui_session');
  expect(session).toBeTruthy();
  expect(session.httpOnly).toBe(true);
  expect(session.expires).toBeGreaterThan(Date.now() / 1000 + 399 * 24 * 60 * 60);

  const response = await request.get('/auth/me', {
    headers: {
      cookie: `${session.name}=${session.value}`,
    },
  });
  expect(response.ok()).toBe(true);
  const body = await response.json();
  expect(body.authenticated).toBe(true);
  expect(body.user.email).toBe(email);
  expect(body.user.emailVerifiedAt).toBeTruthy();
  expect(body.accounts).toHaveLength(1);
  expect(body.teams).toHaveLength(1);
  expect(body.projects).toHaveLength(1);
  expect(body.roles[body.projects[0].id]).toBe('admin');
});

test('Feature: Sign-up - expired email magic link returns to sign-up in place', async ({ page, request }) => {
  const email = 'expired-link@example.test';
  await request.post('/auth/email/request', {
    data: { email },
  });
  const latest = await request.get(`/test/email/latest?email=${encodeURIComponent(email)}`);
  const message = await latest.json();
  const token = new URL(message.loginUrl).hash.slice('#token='.length);

  const first = await request.post('/auth/email/consume', { data: { token } });
  expect(first.ok()).toBe(true);

  await page.goto(message.loginUrl);
  await expect(page).toHaveURL('http://127.0.0.1:18080/');
  await expect(page.getByRole('alert')).toContainText('sign-in link is invalid or expired');
  await expect(page.getByRole('button', { name: 'Send sign-in link' })).toBeVisible();
});

test('Feature: Login - email magic links are single-use', async ({ request }) => {
  const email = 'single-use@example.test';
  await request.post('/auth/email/request', {
    data: { email },
  });
  const latest = await request.get(`/test/email/latest?email=${encodeURIComponent(email)}`);
  const message = await latest.json();
  const token = new URL(message.loginUrl).hash.slice('#token='.length);

  const first = await request.post('/auth/email/consume', { data: { token } });
  expect(first.ok()).toBe(true);

  const second = await request.post('/auth/email/consume', { data: { token } });
  expect(second.status()).toBe(400);
});

test('Feature: Login - user starts OIDC redirect for future OAuth providers', async ({ request }) => {
  const response = await request.get('/auth/login', { maxRedirects: 0 });
  expect(response.status()).toBe(302);
  const location = response.headers().location;
  expect(location).toBeTruthy();
  const current = new URL(location);
  expect(`${current.protocol}//${current.host}${current.pathname}`).toBe('http://127.0.0.1:18080/test-idp/login/oauth/authorize');
  expect(current.searchParams.get('response_type')).toBe('code');
  expect(current.searchParams.get('client_id')).toBe('ivy-control-local');
  expect(current.searchParams.get('redirect_uri')).toBe('http://127.0.0.1:18080/auth/callback');
  expect(current.searchParams.get('scope')).toContain('openid');
  expect(current.searchParams.get('state')).toBeTruthy();
  expect(current.searchParams.get('nonce')).toBeTruthy();
});

test('Feature: Login - successful OIDC callback creates app session and project view', async ({ page, request }) => {
  await page.goto('/auth/login');
  await expect(page).toHaveURL('http://127.0.0.1:18080/');

  const cookies = await page.context().cookies();
  const session = cookies.find((cookie) => cookie.name === 'ivy_webui_session');
  expect(session).toBeTruthy();
  expect(session.httpOnly).toBe(true);

  const response = await request.get('/auth/me', {
    headers: {
      cookie: `${session.name}=${session.value}`,
    },
  });
  expect(response.ok()).toBe(true);
  const body = await response.json();
  expect(body.authenticated).toBe(true);
  expect(body.user.email).toBe('tester@example.test');
  expect(body.accounts).toHaveLength(1);
  expect(body.teams).toHaveLength(1);
  expect(body.projects).toHaveLength(1);
  expect(body.projects[0].slug).toBe('client-server');
  expect(body.roles[body.projects[0].id]).toBe('admin');
});

test.fixme('Feature: Account recovery - user recovers access with a fresh email magic link', async () => {});
test.fixme('Feature: Billing information - account owner adds billing information', async () => {});
test.fixme('Feature: Alpha tester dashboard - admin invites tester and grants project access', async () => {});
