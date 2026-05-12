import { describe, expect, it } from 'vitest';
import { routeWorkbenchCommand } from './commandRouting';

describe('routeWorkbenchCommand', () => {
	it('maps legacy webui command names onto neutral engine intents', () => {
		expect.hasAssertions();
		expect(routeWorkbenchCommand('checkInduction')).toBe('check.induction');
		expect(routeWorkbenchCommand('boundedCheck')).toBe('check.bounded');
		expect(routeWorkbenchCommand('showReachableStates')).toBe('concept.action');
		expect(routeWorkbenchCommand('addRelationFromString')).toBe('concept.add-relation');
	});

	it('leaves already-neutral commands unchanged', () => {
		expect.hasAssertions();
		expect(routeWorkbenchCommand('check.induction')).toBe('check.induction');
		expect(routeWorkbenchCommand('file.download')).toBe('file.download');
	});
});
