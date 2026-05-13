export function populateConstraintFacts(app, conceptData, {
  doc = globalThis.document,
} = {}) {
  const facts = conceptData && Array.isArray(conceptData.facts) ? conceptData.facts : [];
  const info = doc && doc.getElementById('info-content');
  if (!info) return;
  info.innerHTML = '';
  if (facts.length === 0) {
    info.textContent = 'Select a node or edge to see details';
    return;
  }

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
