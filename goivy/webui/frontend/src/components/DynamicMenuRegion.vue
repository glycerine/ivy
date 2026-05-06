<script setup>
import { computed } from 'vue';
import { useDropdownStore } from '../stores/dropdownStore.js';
import { useMenuDescriptorStore } from '../stores/menuDescriptorStore.js';

const props = defineProps({
  region: {
    type: String,
    required: true,
  },
});

const menuStore = useMenuDescriptorStore();
const dropdownStore = useDropdownStore();
const menus = computed(() => menuStore.menusFor(props.region));

function closeStaticDropdowns() {
  dropdownStore.closeAll();
}

function toggleMenu(index) {
  closeStaticDropdowns();
  menuStore.toggle(props.region, index);
}

function runItem(item) {
  menuStore.runItem(props.region, item);
}
</script>

<template>
  <div v-if="menus.length" class="dynamic-menu-root" :data-dynamic-menu-region="region">
    <div
      v-for="(menu, index) in menus"
      :key="menu.key"
      class="dropdown"
      :class="{ open: menuStore.isOpen(region, index) }"
    >
      <span
        class="panel-menu"
        :data-dropdown="menu.contentId"
        @click.stop.prevent="toggleMenu(index)"
      >
        {{ menu.label }}
      </span>
      <div :id="menu.contentId" class="dropdown-content">
        <template v-for="item in menu.items" :key="item.key">
          <div v-if="item.type === 'separator'" class="dropdown-sep"></div>
          <a
            v-else
            href="#"
            :class="{ disabled: item.enabled === false }"
            :aria-disabled="item.enabled === false ? 'true' : undefined"
            :data-menu-action="item.action || ''"
            :data-menu-dispatch="item.dispatch || ''"
            @click.stop.prevent="runItem(item)"
          >
            {{ item.label }}
          </a>
        </template>
      </div>
    </div>
  </div>
</template>
