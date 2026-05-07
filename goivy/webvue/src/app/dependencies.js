import { inject } from 'vue';
import { createAuthClient } from '../api/authClient.js';
import { createPasskeyBrowser } from '../api/passkeyBrowser.js';

export const authClientKey = Symbol('authClient');
export const passkeyBrowserKey = Symbol('passkeyBrowser');

export function useAuthClient() {
  return inject(authClientKey, createAuthClient());
}

export function usePasskeyBrowser() {
  return inject(passkeyBrowserKey, createPasskeyBrowser());
}
