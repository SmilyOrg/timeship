<template>
  <div class="browser">
    <snapshot-list
      :model-value="selectedSnapshot"
      :current-path="path"
      :node="node"
      @update:model-value="onSnapshotChange">
    </snapshot-list>
    <div class="explorer">
      <breadcrumbs
        :items="breadcrumbItems"
        @navigate="onPathChange($event.href || '')"
      ></breadcrumbs>
      <file-table
        v-if="!isViewingFile"
        class="file-table"
        :nodes="nodes"
        :loading="isLoading"
        :error="error?.message"
        :current-path="path"
        :snapshot="selectedSnapshot"
        @navigate="onPathChange($event)"
        @update:selection="selectedFiles = $event"
      ></file-table>
      <file-preview
        v-else
        class="file-preview"
        :file-info="currentFileInfo"
        :loading="isLoading"
        :error="error?.message"
        :snapshot="selectedSnapshot"
      ></file-preview>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue';
import { useRouter } from 'vue-router';
import SnapshotList from './SnapshotList.vue';
import FileTable from './FileTable.vue';
import FilePreview from './FilePreview.vue';
import Breadcrumbs from './Breadcrumbs.vue';
import { useApi } from './api/api';
import type { Node } from './api/api';

// Props from router
const props = defineProps<{
  storage: string;
  path: string | string[];
  snapshot: string | null;
}>();

const router = useRouter();

const selectedFiles = ref<string[]>([]);

// Derive full path from props
const path = computed(() => {
  const pathArray = Array.isArray(props.path) ? props.path : props.path ? [props.path] : [];
  const pathStr = pathArray.join('/');
  return `${props.storage}://${pathStr}`;
});

// Derive snapshot from props
const selectedSnapshot = computed(() => props.snapshot);

// Compute breadcrumb items from path
const breadcrumbItems = computed(() => {
  return [
    { text: 'Storage', href: `${props.storage}://`, active: path.value === `${props.storage}://` },
    ...path.value.split('/').slice(2).map((part, index, arr) => {
      const isActive = index === arr.length - 1;
      return {
        text: part,
        href: `${props.storage}://${arr.slice(0, index + 1).join('/')}`,
        active: isActive,
      };
    }),
  ];
});

// Parse path to extract storage and path components
const parsedPath = computed(() => {
  const pathArray = Array.isArray(props.path) ? props.path : props.path ? [props.path] : [];
  const pathStr = pathArray.join('/');
  return { storage: props.storage, path: pathStr };
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

// Fetch data from API
const { data: node, error, isLoading } = useApi(apiEndpoint);

// Extract nodes from API response
const nodes = computed(() => {
  return node.value?.files || [];
});

// Determine if we're viewing a file (API returns a single file object instead of a files array)
const isViewingFile = computed(() => {
  return node.value && !node.value.files;
});

// Get current file info when viewing a file
const currentFileInfo = computed<Node | null>(() => {
  if (isViewingFile.value && node.value) {
    // When viewing a file, the API returns the file object directly
    return node.value as Node;
  }
  return null;
});

const onPathChange = (newPath: string) => {
  if (path.value === newPath) {
    return;
  }
  
  // Parse the new path to extract storage and path
  try {
    const url = new URL(newPath);
    const storage = url.protocol.replace(':', '');
    const pathStr = decodeURIComponent(url.host) + decodeURIComponent(url.pathname);
    
    // Split path into segments to avoid URL encoding issues
    const pathSegments = pathStr ? pathStr.split('/').filter(s => s) : [];
    
    // Update the route
    router.push({
      name: 'browse',
      params: {
        storage,
        path: pathSegments.length > 0 ? pathSegments : undefined
      },
      query: selectedSnapshot.value ? { snapshot: selectedSnapshot.value } : {}
    });
  } catch (e) {
    console.error('Invalid path:', newPath, e);
  }
};

const onSnapshotChange = (newSnapshot: string | null) => {
  const { storage, path: urlPath } = parsedPath.value;
  
  // Split path into segments to avoid URL encoding issues
  const pathSegments = urlPath ? urlPath.split('/').filter(s => s) : [];
  
  router.push({
    name: 'browse',
    params: {
      storage,
      path: pathSegments.length > 0 ? pathSegments : undefined
    },
    query: newSnapshot ? { snapshot: newSnapshot } : {}
  });
};

</script>

<style scoped>
.browser {
  display: flex;
  flex-direction: row;
  gap: 16px;
}

.explorer {
  flex-grow: 1;
  display: flex;
  flex-direction: column;
  gap: 8px;
}
</style>