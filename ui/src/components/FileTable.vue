<template>
  <div class="file-table-container">
    <table 
      ref="tableRef"
      class="file-table" 
      tabindex="0"
      :class="{ 'is-loading': loading }"
      @keydown="handleKeyDown"
    >
      <thead>
        <tr>
          <th class="col-icon"></th>
          <th class="col-name">Name</th>
          <th class="col-size">Size</th>
          <th class="col-date">Modified</th>
        </tr>
      </thead>
      <tbody v-if="!error && nodes.length > 0">
        <tr
          v-for="node in nodes"
          :key="node.path"
          :class="{ 'is-selected': selectedPaths.has(node.path) }"
          @click="handleClick($event, node)"
          @dblclick="handleDoubleClick(node)"
          @contextmenu="handleContextMenu($event, node)"
        >
          <td class="col-icon">{{ node.type === 'dir' ? '📁' : '📄' }}</td>
          <td class="col-name">{{ node.basename }}</td>
          <td class="col-size">{{ node.type === 'file' ? formatSize(node.file_size) : '' }}</td>
          <td class="col-date">{{ formatDate(node.last_modified) }}</td>
        </tr>
      </tbody>
    </table>

    <!-- Context Menu -->
    <div 
      v-if="contextMenu.visible"
      ref="contextMenuRef"
      class="context-menu ui vertical menu"
      :style="{ left: contextMenu.x + 'px', top: contextMenu.y + 'px' }"
    >
      <a 
        class="item"
        :href="downloadUrl"
        :download="contextMenu.node?.basename"
        @click="contextMenu.visible = false"
      >
        Download
      </a>
    </div>

    <div v-if="error" class="empty-state">
      <div class="empty-icon">{{ isNotFoundError ? '📂' : '⚠️' }}</div>
      <div class="empty-message">{{ errorMessage }}</div>
    </div>

    <div v-else-if="nodes.length === 0 && !loading" class="empty-state">
      <div class="empty-icon">📁</div>
      <div class="empty-message">Folder is empty</div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, computed, onMounted, onUnmounted } from 'vue';
import type { Node } from './api/api';
import { getNodeUrl } from './api/api';
import { format } from 'date-fns';

const props = withDefaults(defineProps<{
  nodes: Node[];
  loading?: boolean;
  error?: string | null;
  currentPath?: string;
  currentStorage?: string;
  snapshot?: string | null;
}>(), {
  loading: false,
  error: null,
  currentPath: '',
  currentStorage: 'local',
  snapshot: null,
});

const emit = defineEmits<{
  'update:selection': [paths: string[]];
  'navigate': [path: string];
}>();

// Computed error properties
const isNotFoundError = computed(() => props.error === 'Not found');
const errorMessage = computed(() => {
  if (isNotFoundError.value) {
    if (props.snapshot) {
      return 'The file/folder was not found in this snapshot';
    } else {
      return 'The file/folder was not found';
    }
  }
  return props.error || 'An error occurred';
});

// Compute download URL for context menu
const downloadUrl = computed(() => {
  const node = contextMenu.value.node;
  if (!node) return '';
  
  return getNodeUrl(props.currentStorage, node.path, {
    snapshot: props.snapshot,
    download: true
  });
});

// Selection state
const selectedPaths = ref<Set<string>>(new Set());
const lastSelectedIndex = ref<number>(-1);

// Table focus management
const tableRef = ref<HTMLTableElement | null>(null);

// Context menu state
const contextMenu = ref({
  visible: false,
  x: 0,
  y: 0,
  node: null as Node | null,
});

const contextMenuRef = ref<HTMLDivElement | null>(null);

// Emit selection changes
watch(selectedPaths, (newSelection) => {
  emit('update:selection', Array.from(newSelection));
}, { deep: true });

// Clear selection when nodes change
watch(() => props.nodes, () => {
  selectedPaths.value.clear();
  lastSelectedIndex.value = -1;
});

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
  return format(date, 'yyyy-MM-dd HH:mm');
}

// Handle single click
function handleClick(event: MouseEvent, node: Node) {
  const nodeIndex = props.nodes.findIndex(n => n.path === node.path);
  
  if (event.ctrlKey || event.metaKey) {
    // Ctrl+Click: Toggle selection
    if (selectedPaths.value.has(node.path)) {
      selectedPaths.value.delete(node.path);
    } else {
      selectedPaths.value.add(node.path);
    }
    lastSelectedIndex.value = nodeIndex;
  } else if (event.shiftKey && lastSelectedIndex.value !== -1) {
    // Shift+Click: Range select
    const start = Math.min(lastSelectedIndex.value, nodeIndex);
    const end = Math.max(lastSelectedIndex.value, nodeIndex);
    selectedPaths.value.clear();
    for (let i = start; i <= end; i++) {
      const nodeAtIndex = props.nodes[i];
      if (nodeAtIndex) {
        selectedPaths.value.add(nodeAtIndex.path);
      }
    }
  } else {
    // Regular click: Select single item
    selectedPaths.value.clear();
    selectedPaths.value.add(node.path);
    lastSelectedIndex.value = nodeIndex;
  }
}

// Handle double click
function handleDoubleClick(node: Node) {
  emit('navigate', node.path);
}

// Handle context menu
function handleContextMenu(event: MouseEvent, node: Node) {
  event.preventDefault();
  
  // Only show context menu for files
  if (node.type !== 'file') {
    return;
  }
  
  // Select the node if not already selected
  if (!selectedPaths.value.has(node.path)) {
    selectedPaths.value.clear();
    selectedPaths.value.add(node.path);
    const nodeIndex = props.nodes.findIndex(n => n.path === node.path);
    lastSelectedIndex.value = nodeIndex;
  }
  
  contextMenu.value = {
    visible: true,
    x: event.clientX,
    y: event.clientY,
    node,
  };
}

// Close context menu when clicking outside
function handleClickOutside(event: MouseEvent) {
  if (contextMenu.value.visible && contextMenuRef.value && !contextMenuRef.value.contains(event.target as HTMLElement)) {
    contextMenu.value.visible = false;
  }
}

// Lifecycle hooks
onMounted(() => {
  document.addEventListener('click', handleClickOutside);
});

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside);
});

// Handle keyboard navigation
function handleKeyDown(event: KeyboardEvent) {
  if (props.nodes.length === 0) return;

  if (event.key === 'ArrowDown' || event.key === 'ArrowUp') {
    event.preventDefault();
    
    const direction = event.key === 'ArrowDown' ? 1 : -1;
    
    // Find current index
    let currentIndex = lastSelectedIndex.value;
    if (currentIndex === -1 && selectedPaths.value.size > 0) {
      // Find first selected item
      const firstSelected = Array.from(selectedPaths.value)[0];
      currentIndex = props.nodes.findIndex(n => n.path === firstSelected);
    }
    
    // Calculate new index
    const newIndex = Math.max(0, Math.min(props.nodes.length - 1, currentIndex + direction));
    if (newIndex === currentIndex && currentIndex !== -1) return;
    
    const newNode = props.nodes[newIndex];
    
    if (event.shiftKey && lastSelectedIndex.value !== -1) {
      // Shift+Arrow: Range select
      const start = Math.min(lastSelectedIndex.value, newIndex);
      const end = Math.max(lastSelectedIndex.value, newIndex);
      selectedPaths.value.clear();
      for (let i = start; i <= end; i++) {
        const nodeAtIndex = props.nodes[i];
        if (nodeAtIndex) {
          selectedPaths.value.add(nodeAtIndex.path);
        }
      }
    } else if (event.ctrlKey || event.metaKey) {
      // Ctrl+Arrow: Move focus and add to selection
      if (newNode) {
        selectedPaths.value.add(newNode.path);
      }
      lastSelectedIndex.value = newIndex;
    } else {
      // Regular arrow: Move selection
      selectedPaths.value.clear();
      if (newNode) {
        selectedPaths.value.add(newNode.path);
      }
      lastSelectedIndex.value = newIndex;
    }
    
    // Scroll into view
    const rows = tableRef.value?.querySelectorAll('tbody tr');
    if (rows && rows[newIndex]) {
      rows[newIndex].scrollIntoView({ block: 'nearest', behavior: 'smooth' });
    }
  } else if (event.key === 'Enter' && selectedPaths.value.size === 1) {
    // Enter key: Navigate into folder
    const selectedPath = Array.from(selectedPaths.value)[0];
    const node = props.nodes.find(n => n.path === selectedPath);
    if (node && node.type === 'dir') {
      emit('navigate', node.path);
    }
  }
}
</script>

<style scoped>
.file-table-container {
  position: relative;
  overflow: auto;
}

.file-table {
  width: 100%;
  border-collapse: collapse;
  outline: none;
}

.file-table.is-loading {
  opacity: 0.6;
  animation: pulse 1.5s ease-in-out infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 0.6; }
  50% { opacity: 0.3; }
}

thead {
  position: sticky;
  top: 0;
  background: white;
  z-index: 1;
}

th {
  text-align: left;
  padding: 8px;
  border-bottom: 2px solid #e0e0e0;
  font-weight: 600;
}

td {
  padding: 8px;
  border-bottom: 1px solid #f0f0f0;
  background-color: unset;
}

tr {
  cursor: pointer;
  user-select: none;
}

tr:hover {
  background: #f5f5f5;
}

tr.is-selected {
  background: #e3f2fd;
}

tr.is-selected:hover {
  background: #bbdefb;
}

.col-icon {
  width: 40px;
  text-align: center;
}

.col-size {
  width: 100px;
  text-align: right;
}

.col-date {
  width: 180px;
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 64px 32px;
  text-align: center;
}

.empty-icon {
  font-size: 64px;
  height: 64px;
}

.empty-message {
  font-size: 16px;
  color: #666;
}

.context-menu {
  position: fixed;
  z-index: 1000;
  min-width: 120px;
}

</style>