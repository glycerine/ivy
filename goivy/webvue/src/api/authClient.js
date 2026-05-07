export function createAuthClient(fetchImpl = globalThis.fetch) {
  if (typeof fetchImpl !== 'function') {
    throw new Error('fetch implementation is required');
  }
  return {
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
