<script lang="ts">
	import type { EdgeDisplayClass, NodeLabelDisplayClass } from '$lib/types';
	import type { StateRelationRow } from '$lib/state/stateRelations.svelte';

	type Props = {
		stateLabel: string;
		rows: StateRelationRow[];
		onToggle?: (rowId: string, displayClass: string, checked: boolean) => void | Promise<void>;
	};

	let { stateLabel, rows, onToggle }: Props = $props();
	const columns = [
		['all_to_all', '+'],
		['edge_unknown', '?'],
		['none_to_none', '-'],
		['transitive', 'T']
	] as const satisfies readonly [EdgeDisplayClass, string][];

	function checked(row: StateRelationRow, key: EdgeDisplayClass | NodeLabelDisplayClass) {
		return row.columns.find((column) => column.key === key)?.checked ?? false;
	}
</script>

<section class="pane state-pane" aria-label="State relations">
	<div class="pane-title">
		<strong>State/relations</strong>
	</div>
	<div class="state-panel">
		<strong>State: {stateLabel}</strong>
		<table id="state-checkbox-table" aria-label="Relation toggles">
			<thead>
				<tr>
					{#each columns as [, label] (label)}
						<th class="chk-col">{label}</th>
					{/each}
					<th class="name-col"></th>
				</tr>
			</thead>
			<tbody id="state-checkbox-body">
				{#if rows.length > 0}
					{#each rows as row (row.name)}
						<tr>
							{#each columns as [key] (key)}
								<td>
									<input
										type="checkbox"
										checked={checked(row, key)}
										onchange={(event) => void onToggle?.(row.id, key, event.currentTarget.checked)}
									/>
								</td>
							{/each}
							<td class="name-col">{row.name}</td>
						</tr>
					{/each}
				{:else}
					<tr>
						<td colspan="5" class="state-relations-empty">No relations loaded</td>
					</tr>
				{/if}
			</tbody>
		</table>
	</div>
</section>
