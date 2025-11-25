<template>
  <div class="file-preview">
    <div v-if="error" class="ui center aligned basic segment">
      <i class="huge warning circle icon"></i>
      <div class="ui header">{{ error }}</div>
    </div>
    
    <div v-else-if="fileInfo">
      <!-- Compact File Header -->
      <div class="ui top attached segment file-header">
        <div class="header-content">
          <div class="icon-column">
            <i class="huge file outline icon"></i>
          </div>
          <div class="info-column ui">
            <h3 class="ui header" style="margin-bottom: 0.25em;">
              {{ fileInfo.basename }}
            </h3>
            <div class="ui small text">
              <span v-if="fileInfo.mime_type" class="ui label line">{{ fileInfo.mime_type }}</span>
              {{ formatSize(fileInfo.file_size) }}
              <span v-if="fileInfo.last_modified" class="line"> · {{ formatDate(fileInfo.last_modified) }}</span>
            </div>
          </div>
          <div class="action-column">
            <a 
              class="ui primary button"
              :href="downloadUrl"
              :download="fileInfo.basename"
            >
              <i class="download icon"></i>
              Download
            </a>
          </div>
        </div>
      </div>
      
      <!-- Preview Content -->
      <div v-if="isTextFile" class="text-file ui bottom attached segment" style="padding: 0;">
        <div v-if="loadingContent" class="ui active loader"></div>
        <div v-else-if="contentError" class="ui center aligned basic segment">
          <i class="large warning circle icon"></i>
          <div class="ui header">{{ contentError }}</div>
        </div>
        <TextViewer 
          v-else
          :content="fileContent"
          :filename="fileInfo.basename"
        />
      </div>
      
      <!-- Preview Placeholder for non-text files -->
      <div v-else class="ui bottom attached placeholder segment">
        <div class="ui icon header">
          <i class="search icon"></i>
          Preview not available.
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import type { Node } from './api/api';
import { getNodeUrl } from './api/api';
import { format } from 'date-fns';
import TextViewer from './TextViewer.vue';

const props = defineProps<{
  fileInfo?: Node | null;
  loading?: boolean;
  error?: string | null;
  snapshot?: string | null;
}>();

const fileContent = ref<string>('');
const loadingContent = ref(false);
const contentError = ref<string | null>(null);

// Format file size
function formatSize(bytes: number): string {
  if (bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return Math.round(bytes / Math.pow(k, i) * 100) / 100 + ' ' + sizes[i];
}

// Format date
function formatDate(timestamp: number): string {
  const date = new Date(timestamp * 1000);
  return format(date, 'MMM d, yyyy');
}

// Compute download URL
const downloadUrl = computed(() => {
  if (!props.fileInfo) return '';
  
  return getNodeUrl(props.fileInfo.path, {
    snapshot: props.snapshot,
    download: true
  });
});

// Check if the file is a text file based on mime type
const isTextFile = computed(() => {
  if (!props.fileInfo?.mime_type) return false;
  
  const mimeType = props.fileInfo.mime_type.toLowerCase();
  
  // Common text mime types
  return (
    mimeType.startsWith('text/') ||
    mimeType === 'application/json' ||
    mimeType === 'application/javascript' ||
    mimeType === 'application/typescript' ||
    mimeType === 'application/xml' ||
    mimeType === 'application/x-yaml' ||
    mimeType === 'application/x-sh' ||
    mimeType.includes('yaml') ||
    mimeType.includes('xml')
  );
});

// Fetch file content for text files
async function fetchFileContent() {
  if (!props.fileInfo || !isTextFile.value) {
    fileContent.value = '';
    return;
  }
  
  loadingContent.value = true;
  contentError.value = null;
  
  try {
    const url = getNodeUrl(props.fileInfo.path, {
      snapshot: props.snapshot
    });
    
    const response = await fetch(url);
    if (!response.ok) {
      throw new Error(`Failed to fetch file: ${response.statusText}`);
    }
    
    const text = await response.text();
    fileContent.value = text;
  } catch (err) {
    contentError.value = err instanceof Error ? err.message : 'Failed to load file content';
    fileContent.value = '';
  } finally {
    loadingContent.value = false;
  }
}

// Watch for changes in file info and fetch content
watch(() => [props.fileInfo, props.snapshot, isTextFile.value], () => {
  if (isTextFile.value) {
    fetchFileContent();
  } else {
    fileContent.value = '';
  }
}, { immediate: true });
</script>

<style scoped>
.header-content {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.icon-column {
  flex: 0 0 auto;
  display: flex;
  align-items: center;
  justify-content: center;
  min-width: 60px;
}

.info-column {
  flex: 1 1 auto;
  min-width: 0;
}

.action-column {
  flex: 0 0 auto;
}

.text-file {
  height: 400px;
}
</style>
