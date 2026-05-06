import { defineStore } from 'pinia';

function slug(value, fallback) {
  const text = String(value || fallback || 'menu').toLowerCase().replace(/[^a-z0-9]+/g, '-').replace(/^-|-$/g, '');
  return text || String(fallback || 'menu');
}

function normalizeItem(item, index) {
  if (item && item.type === 'separator') {
    return {
      ...item,
      type: 'separator',
      key: `sep-${index}`,
    };
  }
  return {
    ...(item || {}),
    type: (item && item.type) || 'item',
    key: `item-${index}-${(item && (item.action || item.label)) || ''}`,
    label: (item && (item.label || item.action)) || '',
  };
}

function normalizeMenu(region, menu, index) {
  const label = (menu && menu.label) || 'Menu';
  return {
    ...(menu || {}),
    key: `${region}-${index}-${slug(label, 'menu')}`,
    label,
    contentId: `dynamic-${region}-${index}-${slug(label, 'menu')}`,
    items: Array.isArray(menu && menu.items) ? menu.items.map(normalizeItem) : [],
  };
}

export const useMenuDescriptorStore = defineStore('menuDescriptor', {
  state: () => ({
    regions: {
      arg: [],
      concept: [],
    },
    dispatchers: {},
    openKey: '',
  }),
  getters: {
    menusFor: (state) => (region) => state.regions[region] || [],
    isOpen: (state) => (region, index) => state.openKey === `${region}:${index}`,
  },
  actions: {
    setRegion(region, menus = [], dispatcher = null) {
      this.regions[region] = Array.isArray(menus)
        ? menus.map((menu, index) => normalizeMenu(region, menu, index))
        : [];
      if (typeof dispatcher === 'function') {
        this.dispatchers[region] = dispatcher;
      } else {
        delete this.dispatchers[region];
      }
      this.openKey = '';
    },
    toggle(region, index) {
      const key = `${region}:${index}`;
      this.openKey = this.openKey === key ? '' : key;
    },
    closeAll() {
      this.openKey = '';
    },
    runItem(region, item) {
      if (!item || item.enabled === false) {
        return Promise.resolve({ ok: false, error: 'disabled action' });
      }
      this.closeAll();
      const dispatcher = this.dispatchers[region];
      if (typeof dispatcher === 'function') {
        return dispatcher(item);
      }
      return Promise.resolve({ ok: false, error: 'no dispatcher' });
    },
  },
});
