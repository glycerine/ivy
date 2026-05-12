<script lang="ts">
	import type { ModelDocument } from '$lib/types';

	type Props = {
		activeModel: ModelDocument;
		editorText: string;
		keymap: 'sublime' | 'emacs' | 'vim';
		onUpdateEditor: (text: string) => void;
		onSetKeymap: (keymap: 'sublime' | 'emacs' | 'vim') => void;
	};

	let { activeModel, editorText, keymap, onUpdateEditor, onSetKeymap }: Props = $props();
	const lineNumbers = $derived(editorText.split('\n').map((_, index) => index + 1));
</script>

<section class="pane editor-pane" aria-label="Editor">
	<div class="pane-title">
		<strong>Editing: {activeModel.filename} [{activeModel.dirty ? 'unsaved' : 'saved'}]</strong>
		<span data-testid="dirty-indicator">{activeModel.dirty ? 'Unsaved' : 'Saved'}</span>
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
	</div>
</section>
