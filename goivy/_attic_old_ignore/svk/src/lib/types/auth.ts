import type { Id } from './ids';

export type AuthStatus =
	| 'unknown'
	| 'anonymous'
	| 'authenticated'
	| 'offline-authenticated'
	| 'expired';

export type AuthUser = {
	id: Id;
	primaryEmail: string;
	displayName: string;
	avatarUrl?: string;
	createdAt: string;
};

export type AuthState = {
	status: AuthStatus;
	userId: Id | null;
	sessionExpiresAt?: string;
	csrfToken?: string;
	offlineUnlockedAt?: string;
};
