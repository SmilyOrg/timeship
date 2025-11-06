<template>
  <div class="browser">
    <snapshot-list
      v-model="selectedSnapshot">
    </snapshot-list>
    <vue-finder
      :key="'vf-' + selectedSnapshot"
      class="finder"
      id="vf"
      ref="vf"
      :driver="driver"
    ></vue-finder>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue';
import SnapshotList from './SnapshotList.vue';
import { TimeshipRemoteDriver } from './api/TimeshipRemoteDriver';
// import * as VF from 'vuefinder';
import { App } from 'vuefinder/'

const selectedSnapshot = ref<string | null>(null);
const vf = ref<InstanceType<typeof import('vuefinder').VueFinder> | null>(null);

const driver = new TimeshipRemoteDriver({
  baseURL: 'http://localhost:8080',
  adapter: 'local',
});

     
// function useVueFinderRefresh(vfId) {
//   console.log(VF);
//   const refresh = () => {
//     const app = VF.useApp(vfId);
//     const currentPath = app.fs.path.get()?.path;
//     app.adapter.open(currentPath); // This triggers the full refresh cycle
//   };

//   return { refresh };
// }

// const { refresh } = useVueFinderRefresh('vf');
// console.log(refresh);


window.driver = driver; // For debugging
watch(vf, newVf => {
  if (newVf) {
    window.vf = newVf; // For debugging
  }
});

// Watch for snapshot changes and update the driver
watch(selectedSnapshot, (newSnapshot) => {
  driver.setSnapshot(newSnapshot);
  // Refresh VueFinder when snapshot changes
  // if (vf.value) {
  //   refresh();
  // }
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