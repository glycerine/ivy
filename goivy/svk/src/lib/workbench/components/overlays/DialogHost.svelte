<script lang="ts">
	import type { DialogRequest } from '$lib/workbench/state';

	type Props = {
		dialog?: DialogRequest | null;
		onResolve?: (value: unknown) => void;
	};

	let { dialog = null, onResolve }: Props = $props();
</script>

{#if dialog}
	<div class="dialog-overlay" role="presentation">
		<div class="dialog-box" role="dialog" aria-modal="true" aria-labelledby="dialog-title">
			<h2 id="dialog-title">{dialog.title}</h2>
			<p>{dialog.message}</p>
			{#if dialog.kind === 'entry'}
				<input aria-label={dialog.title} value={dialog.value} />
			{:else if dialog.kind === 'list'}
				<select aria-label={dialog.title}>
					{#each dialog.choices as choice (choice)}
						<option>{choice}</option>
					{/each}
				</select>
			{/if}
			<div class="dialog-buttons">
				{#if dialog.kind === 'confirm'}
					<button type="button" onclick={() => onResolve?.(false)}>Cancel</button>
				{/if}
				<button type="button" onclick={() => onResolve?.(true)}>OK</button>
			</div>
		</div>
	</div>
{/if}
