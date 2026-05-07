export function createAuthClient(fetchImpl = globalThis.fetch) {
  if (typeof fetchImpl !== 'function') {
    throw new Error('fetch implementation is required');
  }
  return {
    async currentSession() {
      const response = await fetchImpl('/auth/me', {
        method: 'GET',
        headers: {
          accept: 'application/json',
        },
      });
      if (!response.ok) {
        throw new Error(`auth session request failed with status ${response.status}`);
      }
      return response.json();
    },
    async requestEmailLogin(email) {
      const response = await fetchImpl('/auth/email/request', {
        method: 'POST',
        headers: {
          'content-type': 'application/json',
        },
        body: JSON.stringify({ email }),
      });
      if (!response.ok) {
        throw new Error(`email login request failed with status ${response.status}`);
      }
      return response.json();
    },
    async consumeEmailLogin(token) {
      const response = await fetchImpl('/auth/email/consume', {
        method: 'POST',
        headers: {
          'content-type': 'application/json',
        },
        body: JSON.stringify({ token }),
      });
      if (!response.ok) {
        throw new Error(`email login consume failed with status ${response.status}`);
      }
      return response.json();
    },
    async beginPasskeyRegistration() {
      const response = await fetchImpl('/auth/passkeys/register/options', {
        method: 'POST',
        headers: {
          accept: 'application/json',
        },
      });
      if (!response.ok) {
        throw new Error(`passkey registration options request failed with status ${response.status}`);
      }
      return response.json();
    },
    async finishPasskeyRegistration(credential) {
      const response = await fetchImpl('/auth/passkeys/register/finish', {
        method: 'POST',
        headers: {
          'content-type': 'application/json',
        },
        body: JSON.stringify(credential),
      });
      if (!response.ok) {
        throw new Error(`passkey registration finish request failed with status ${response.status}`);
      }
      return response.json();
    },
    async beginPasskeyLogin() {
      const response = await fetchImpl('/auth/passkeys/login/options', {
        method: 'POST',
        headers: {
          accept: 'application/json',
        },
      });
      if (!response.ok) {
        throw new Error(`passkey login options request failed with status ${response.status}`);
      }
      return response.json();
    },
    async finishPasskeyLogin(credential) {
      const response = await fetchImpl('/auth/passkeys/login/finish', {
        method: 'POST',
        headers: {
          'content-type': 'application/json',
        },
        body: JSON.stringify(credential),
      });
      if (!response.ok) {
        throw new Error(`passkey login finish request failed with status ${response.status}`);
      }
      return response.json();
    },
    async listUnverifiedEmails() {
      const response = await fetchImpl('/admin/api/unverified-emails', {
        method: 'GET',
        headers: {
          accept: 'application/json',
        },
      });
      if (!response.ok) {
        throw new Error(`admin unverified emails request failed with status ${response.status}`);
      }
      return response.json();
    },
  };
}
