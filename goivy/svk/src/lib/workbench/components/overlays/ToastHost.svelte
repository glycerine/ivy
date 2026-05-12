<script lang="ts">
	import type { Toast } from '$lib/workbench/state';

	type Props = {
		toasts?: Toast[];
		onDismiss?: (id: string) => void;
	};

	let { toasts = [], onDismiss }: Props = $props();
</script>

{#if toasts.length > 0}
	<div class="toast-host" aria-live="polite">
		{#each toasts as toast (toast.id)}
			<div class={`toast ${toast.level}`}>
				<span>{toast.message}</span>
				<button type="button" aria-label={`Dismiss ${toast.message}`} onclick={() => onDismiss?.(toast.id)}>x</button>
			</div>
		{/each}
	</div>
{/if}
