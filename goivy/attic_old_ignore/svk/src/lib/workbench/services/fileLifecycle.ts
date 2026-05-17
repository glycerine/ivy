export function mergeDiskVersionIntoEditBuffer(baseContent: string, editorContent: string, diskContent: string) {
	if (editorContent === baseContent) return diskContent;
	if (diskContent === baseContent) return editorContent;
	return [
		'<<<<<<< EDIT BUFFER',
		editorContent.replace(/\s*$/, ''),
		'||||||| LAST SAVED',
		baseContent.replace(/\s*$/, ''),
		'=======',
		diskContent.replace(/\s*$/, ''),
		'>>>>>>> ON DISK',
		''
	].join('\n');
}

export function modelDownloadFilename(filename = '') {
	const trimmed = filename.trim();
	return trimmed || 'model.ivy';
}

export function analysisStateFilename(filename = '') {
	return modelDownloadFilename(filename).replace(/\.ivy$/i, '') + '.ivyweb.json';
}

export function downloadTextFile(filename: string, content: string, mimeType: string, doc = globalThis.document) {
	const blob = new Blob([content], { type: mimeType });
	const url = URL.createObjectURL(blob);
	const anchor = doc.createElement('a');
	anchor.href = url;
	anchor.download = filename;
	doc.body.appendChild(anchor);
	anchor.click();
	doc.body.removeChild(anchor);
	URL.revokeObjectURL(url);
}
