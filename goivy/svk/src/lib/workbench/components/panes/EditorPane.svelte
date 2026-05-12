<script lang="ts">
	import type { ModelDocument } from '$lib/types';

	type Props = {
		activeModel: ModelDocument;
		editorText: string;
		saveState: 'idle' | 'saving' | 'saved';
		keymap: 'sublime' | 'emacs' | 'vim';
		onUpdateEditor: (text: string) => void;
		onSetKeymap: (keymap: 'sublime' | 'emacs' | 'vim') => void;
	};

	let { activeModel, editorText, saveState, keymap, onUpdateEditor, onSetKeymap }: Props = $props();
	const lineNumbers = $derived(editorText.split('\n').map((_, index) => index + 1));
	const editorLabel = $derived(editorTitle(activeModel.filename, activeModel.dirty, saveState));

	function editorTitle(filename: string, dirty: boolean, state: 'idle' | 'saving' | 'saved') {
		if (state === 'saving') {
			return `${filename} [saving...]`;
		}
		return dirty ? `** ${filename}` : `${filename} [saved]`;
	}
</script>

<section class="pane editor-pane" aria-label="Editor">
	<div class="pane-title">
		<strong>Editing: <span data-testid="editor-label">{editorLabel}</span></strong>
		<span data-testid="dirty-indicator">{saveState === 'saving' ? 'Saving' : activeModel.dirty ? 'Unsaved' : 'Saved'}</span>
	</div>
	<div class="editor-controls">
		<button type="button">x</button>
		<label><input type="radio" name="keymap" checked={keymap === 'sublime'} onchange={() => onSetKeymap('sublime')} /> Sublime</label>
		<label><input type="radio" name="keymap" checked={keymap === 'emacs'} onchange={() => onSetKeymap('emacs')} /> Emacs-ish</label>
		<label><input type="radio" name="keymap" checked={keymap === 'vim'} onchange={() => onSetKeymap('vim')} /> Vim</label>
		<a href="https://codemirror.net/5/doc/manual.html#keymaps" target="_blank" rel="noreferrer">keymap docs</a>
	</div>
	<div class="editor-wrap">
		<div class="line-gutter" aria-hidden="true">
			{#each lineNumbers as number (number)}
				<span>{number}</span>
			{/each}
		</div>
		<textarea
			id="model-editor"
			data-testid="model-editor"
			spellcheck="false"
			value={editorText}
			oninput={(event) => onUpdateEditor(event.currentTarget.value)}
		></textarea>
		{#if saveState === 'saving'}
			<div class="ivy-save-editor-sheen" data-testid="save-editor-sheen" aria-hidden="true"></div>
		{/if}
	</div>
</section>
