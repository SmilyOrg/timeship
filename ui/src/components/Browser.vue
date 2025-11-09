<template>
  <div class="browser">
    <snapshot-list
      v-model="selectedSnapshot">
    </snapshot-list>
    <div class="finder">
      <vue-finder
        :key="'vf-' + selectedSnapshot"
        class="finder"
        id="vf"
        ref="vf"
        :driver="driver"
        @path-change="onPathChange($event)"
      ></vue-finder>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import SnapshotList from './SnapshotList.vue';
import { TimeshipRemoteDriver } from './api/TimeshipRemoteDriver';
import { useLocalStorage } from '@vueuse/core';

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


const path = ref("");

const configStr = useLocalStorage("vuefinder_config_vf");
const config = computed(() => {
  try {
    return JSON.parse(configStr.value || '{}');
  } catch (e) {
    console.error('Failed to parse localStorage for vuefinder_config_vf:', e);
    return {};
  }
});

const patchConfig = (newConfig: Record<string, any>) => {
  const mergedConfig = { ...config.value, ...newConfig };
  configStr.value = JSON.stringify(mergedConfig);
};

watch(config, () => {
  console.log('Local storage updated:', config.value);
}, { immediate: true });


const initialPath = computed(() => {
  return config.value?.path || "local://";
});

watch(initialPath, (newPath) => {
  console.log('Initial path updated:', newPath);
}, { immediate: true });

const onPathChange = (newPath: string) => {
  if (path.value === newPath) {
    return;
  }
  console.log('Path changed to:', newPath);
  path.value = newPath;
  patchConfig({ path: newPath, initialPath: newPath });
};

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
}

.finder {
  flex-grow: 1;
}
</style>