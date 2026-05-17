<script lang="ts">
	type Props = {
		loading?: boolean;
		message?: string;
	};

	let { loading = false, message = 'Loading...' }: Props = $props();
</script>

{#if loading}
	<div class="session-overlay-host" aria-live="polite">
		<div class="session-overlay-box">{message}</div>
	</div>
{/if}
