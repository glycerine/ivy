<script lang="ts">
	import type { CheckResult } from '$lib/types';

	type Props = {
		latestCheck: CheckResult | null;
		sessionLabel: string;
		message: string;
		level?: '' | 'info' | 'success' | 'warning' | 'error';
	};

	let { latestCheck, sessionLabel, message, level = '' }: Props = $props();
	const effectiveLevel = $derived(latestCheck ? checkLevel(latestCheck.result) : level);

	function checkLevel(result: CheckResult['result']) {
		if (result === 'pass') {
			return 'success';
		}
		return 'error';
	}
</script>

<footer
	class="statusbar"
	class:info={effectiveLevel === 'info'}
	class:success={effectiveLevel === 'success'}
	class:warning={effectiveLevel === 'warning'}
	class:error={effectiveLevel === 'error'}
	data-testid="status-strip"
>
	<span>
		{#if latestCheck}
			Check {latestCheck.result.toUpperCase()} ({latestCheck.mode}) [Z3: {latestCheck.z3Contacted ? 'yes' : 'no'}] - {latestCheck.message}
		{:else}
			{message}
		{/if}
	</span>
	<span>Session: {sessionLabel}</span>
</footer>
