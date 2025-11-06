<template>
  <div class="browser">
    <snapshot-list
      v-model="selectedSnapshot">
    </snapshot-list>
    <vue-finder
      class="finder"
      id="vf"
      :driver="driver"
    ></vue-finder>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import SnapshotList from './SnapshotList.vue';
import { TimeshipRemoteDriver } from './api/TimeshipRemoteDriver';

const selectedSnapshot = ref<string | null>(null);

const driver = new TimeshipRemoteDriver({
  baseURL: 'http://localhost:8080',
  adapter: 'local',
});

</script>

<style scoped>
.browser {
  display: flex;
  flex-direction: row;
  gap: 16px;
  width: 100%;
  height: 100%;
}

.finder {
  width: 100%;
}
</style>