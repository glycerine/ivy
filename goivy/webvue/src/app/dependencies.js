import { inject } from 'vue';
import { createAuthClient } from '../api/authClient.js';

export const authClientKey = Symbol('authClient');

export function useAuthClient() {
  return inject(authClientKey, createAuthClient());
}
