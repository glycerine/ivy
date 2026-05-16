export function populateConstraintFacts(app, conceptData, {
  doc = globalThis.document,
} = {}) {
  const facts = conceptData && Array.isArray(conceptData.facts) ? conceptData.facts : [];
  const info = doc && (
    doc.querySelector('.sheet-content.active .info-panel [id^="info-content"]') ||
    doc.getElementById('info-content')
  );
  if (!info) return;
  if (facts.length === 0) {
    const kind = info.getAttribute('data-ivy-details-kind') || '';
    const text = (info.textContent || '').trim();
    if (kind === 'selection' || (text && text !== 'Select a node or edge to see details' && kind !== 'constraints')) {
      return;
    }
    info.innerHTML = '';
    info.textContent = 'Select a node or edge to see details';
    info.setAttribute('data-ivy-details-kind', 'placeholder');
    return;
  }

  info.innerHTML = '';
  info.setAttribute('data-ivy-details-kind', 'constraints');
  const title = doc.createElement('div');
  title.className = 'constraint-facts-title';
  title.textContent = 'Constraints:';
  info.appendChild(title);

  facts.forEach((fact, offset) => {
    const index = typeof fact.index === 'number' ? fact.index : offset;
    const row = doc.createElement('button');
    row.type = 'button';
    row.className = 'constraint-fact';
    row.setAttribute('data-constraint-fact', String(index));
    row.setAttribute('aria-pressed', fact.selected ? 'true' : 'false');
    row.textContent = fact.text || '';
    if (!fact.selected) {
      row.classList.add('inactive');
    }
    row.addEventListener('click', async () => {
      const selected = row.classList.contains('inactive');
      row.classList.toggle('inactive', !selected);
      row.setAttribute('aria-pressed', selected ? 'true' : 'false');
      try {
        await app.api.executeAction('set_fact_selection', {
          index,
          selected,
        });
      } catch (err) {
        row.classList.toggle('inactive', selected);
        row.setAttribute('aria-pressed', selected ? 'false' : 'true');
        app.controls.setStatus(`Fact selection failed: ${err.message}`, 'error');
      }
    });
    info.appendChild(row);
  });
}
