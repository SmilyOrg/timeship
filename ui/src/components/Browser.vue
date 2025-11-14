<template>
  <div class="browser">
    <snapshot-list
      v-model="selectedSnapshot">
    </snapshot-list>
    <file-table
      class="finder"
      :nodes="nodes"
      :loading="isLoading"
      :error="error?.message"
      @navigate="onPathChange($event)"
      @update:selection="selectedFiles = $event"
    ></file-table>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import SnapshotList from './SnapshotList.vue';
import FileTable from './FileTable.vue';
import { useApi } from './api/api';

const selectedSnapshot = ref<string | null>(null);
const selectedFiles = ref<string[]>([]);

const path = ref("local://");

// Parse path to extract storage and path components
const parsedPath = computed(() => {
  try {
    const url = new URL(path.value);
    const storage = url.protocol.replace(':', '');
    const urlPath = url.host + url.pathname;
    console.log('Parsed path:', url, path.value, '->', { storage, path: urlPath });
    return { storage, path: urlPath };
  } catch (e) {
    console.error('Invalid path:', path.value, e);
    return { storage: 'local', path: '' };
  }
});

// Build API endpoint
const apiEndpoint = computed(() => {
  const { storage, path: urlPath } = parsedPath.value;
  const base = `/storages/${storage}/nodes${urlPath ? '/' + urlPath : ''}`;
  const params = new URLSearchParams();
  if (selectedSnapshot.value) {
    params.set('snapshot', selectedSnapshot.value);
  }
  return params.toString() ? `${base}?${params.toString()}` : base;
});

watch(apiEndpoint, (newEndpoint) => {
  console.log('API Endpoint updated:', newEndpoint);
}, { immediate: true });

// Fetch data from API
const { data, error, isLoading } = useApi(apiEndpoint);

// Extract nodes from API response
const nodes = computed(() => {
  return data.value?.files || [];
});

const onPathChange = (newPath: string) => {
  if (path.value === newPath) {
    return;
  }
  console.log('Path changed to:', newPath);
  path.value = newPath;
};

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