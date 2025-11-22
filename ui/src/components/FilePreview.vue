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
      
      <!-- Preview Placeholder -->
      <div class="ui bottom attached placeholder segment">
        <div class="ui icon header">
          <i class="search icon"></i>
          Detailed preview coming soon...
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import type { Node } from './api/api';
import { getNodeUrl } from './api/api';
import { format } from 'date-fns';

const props = defineProps<{
  fileInfo?: Node | null;
  loading?: boolean;
  error?: string | null;
  snapshot?: string | null;
}>();

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
</style>
