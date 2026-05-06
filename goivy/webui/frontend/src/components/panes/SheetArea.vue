<template>
  <div id="sheet-area">
    <div id="tab-bar">
      <button class="sheet-tab active" data-sheet="sheet-1">Sheet 1</button>
    </div>

    <div id="sheet-1" class="sheet-content active">
      <div class="sheet-columns">
        <div class="sheet-left">
          <div class="sheet-main">
            <ArgPane />
            <div id="divider" class="divider"></div>
            <ConceptPane />
          </div>

          <DetailsPane />
        </div>

        <div id="divider2" class="divider"></div>
        <StateRelationsPane />
      </div>
    </div>
  </div>
</template>

<script setup>
import ArgPane from './ArgPane.vue';
import ConceptPane from './ConceptPane.vue';
import DetailsPane from './DetailsPane.vue';
import StateRelationsPane from './StateRelationsPane.vue';
</script>
