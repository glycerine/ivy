import type { AuthState, AuthUser } from '$lib/types';

export function createAuthState(initial: Partial<AuthState> = {}) {
	const state = $state<AuthState>({
		status: 'unknown',
		userId: null,
		...initial
	});

	let user = $state<AuthUser | null>(null);

	return {
		get current() {
			return state;
		},
		get user() {
			return user;
		},
		setAnonymous() {
			state.status = 'anonymous';
			state.userId = null;
			state.sessionExpiresAt = undefined;
			state.csrfToken = undefined;
			state.offlineUnlockedAt = undefined;
			user = null;
		},
		setAuthenticated(nextUser: AuthUser, sessionExpiresAt?: string, csrfToken?: string) {
			state.status = 'authenticated';
			state.userId = nextUser.id;
			state.sessionExpiresAt = sessionExpiresAt;
			state.csrfToken = csrfToken;
			state.offlineUnlockedAt = undefined;
			user = nextUser;
		},
		setOfflineAuthenticated(userId: string, offlineUnlockedAt: string) {
			state.status = 'offline-authenticated';
			state.userId = userId;
			state.sessionExpiresAt = undefined;
			state.csrfToken = undefined;
			state.offlineUnlockedAt = offlineUnlockedAt;
			user = null;
		},
		setExpired() {
			state.status = 'expired';
			state.userId = null;
			state.sessionExpiresAt = undefined;
			state.csrfToken = undefined;
			state.offlineUnlockedAt = undefined;
			user = null;
		}
	};
}
