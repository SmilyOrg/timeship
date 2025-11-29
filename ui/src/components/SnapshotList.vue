<template>
  <div class="snapshot-container">
    <div class="controls">
      <button 
        v-if="node?.type === 'file'"
        :class="['toggle-btn', { active: !hideUnmodified }]"
        @click="hideUnmodified = !hideUnmodified"
        :title="hideUnmodified ? 'Show unmodified' : 'Hide unmodified'"
      >
        <i class="clock outline icon"></i>
      </button>
      <button 
        :class="['toggle-btn', { active: !hideNotFound }]"
        @click="hideNotFound = !hideNotFound"
        :title="hideNotFound ? 'Show not found' : 'Hide not found'"
      >
        <i class="ban icon"></i>
      </button>
    </div>
    <div class="table-wrapper" ref="tableWrapperRef">
      <table ref="tableRef">
      <!-- <thead>
        <tr>
          <th>Time</th>
          <th class="type">Type</th>
        </tr>
      </thead> -->
      <tbody
        @mousedown="startDrag"
        @touchstart="startDrag"
        :style="{
          '--selection-top': `${selectionTop}px`,
          '--selection-height': `${selectionHeight}px`,
          '--selection-opacity': selectionTop >= 0 ? 1 : 0
        }"
      >
        <tr
          v-for="snapshot in snapshots"
          :key="snapshot.id"
          :data-snapshot-id="snapshot.id"
          :class="{ 'not-found': snapshot.node === null, 'unmodified': snapshot.unmodified }"
        >
          <td>
            <label>
              {{ snapshot.title }}
            </label>
          </td>
          <!--<td class="type">
            <label>
              {{ snapshot.type }}
            </label>
          </td>-->
        </tr>
      </tbody>
    </table>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, onMounted, onUnmounted, watch, nextTick } from 'vue';
import { useApi, useApis, type Snapshot as ApiSnapshot, type Node } from './api/api';
import { format } from 'date-fns';

interface SnapshotWithDate extends ApiSnapshot {
  date: Date;
  selected: boolean;
}

interface FormattedSnapshot {
  id: string;
  title: string;
  type: string;
  selected: boolean;
  timestamp?: string;
  marginBottom?: number;
  node?: Node | null;
  unmodified?: boolean;
}

const props = defineProps<{
  modelValue?: string | null;
  currentStorage?: string;
  currentPath?: string;
  node?: Node;
}>();

const emit = defineEmits<{
  'update:modelValue': [value: string | null];
}>();

// Build the snapshots endpoint based on the current path
const snapshotsEndpoint = computed(() => {
  const storage = props.currentStorage || 'local';
  const path = props.currentPath;
  return `/storages/${storage}/snapshots${path ? `/${path}` : ''}`;
});

const { data } = useApi(snapshotsEndpoint);

const isRootPath = computed(() => {
  return !props.currentPath;
});

// Build endpoints for checking snapshot availability
const nodeSnapshotEndpoints = computed(() => {
  const apiSnapshots = data.value?.snapshots || [];
  
  // If no path or at root, return empty array (no checks needed)
  if (isRootPath.value) {
    return apiSnapshots;
  }

  try {
    const storage = props.currentStorage;
    const urlPath = props.currentPath;
    
    // Create endpoint for each snapshot
    return apiSnapshots.map((snapshot: ApiSnapshot) => 
      `/storages/${storage}/nodes/${urlPath}?snapshot=${snapshot.id}`
    );
  } catch (e) {
    console.error('Invalid path:', props.currentPath, e);
    return [];
  }
});

// Use useApis to check all snapshots in parallel
const nodeSnapshots = useApis(nodeSnapshotEndpoints);

// Create a map of snapshot availability from the query results
const nodeSnapshotById = computed(() => {
  const map = new Map<string, Node | null | undefined>();
  const apiSnapshots = data.value?.snapshots || [];
  
  // If no path or at root, all snapshots are available
  if (isRootPath.value) {
    return map;
  }

  apiSnapshots.forEach((snapshot: ApiSnapshot, index: number) => {
    const result = nodeSnapshots.value[index];
    // If query is successful (no error), the path exists
    if (!result) {
      map.set(snapshot.id, undefined);
      return;
    }
    if (result.isSuccess) {
      map.set(snapshot.id, result.data as Node);
      return;
    }
    if (result.isError) {
      map.set(snapshot.id, null);
      return;
    }
    if (result.isLoading) {
      map.set(snapshot.id, undefined);
      return;
    }
  });
  
  return map;
});


const tableRef = ref<HTMLElement | null>(null);
const tableWrapperRef = ref<HTMLElement | null>(null);
const isDragging = ref(false);
const selectionTop = ref(-1);
const selectionHeight = ref(0);
const hideUnmodified = ref(true);
const hideNotFound = ref(true);

const select = (snapshot: FormattedSnapshot | null) => {
  if (!snapshot?.id || snapshot?.id === 'current') {
    emit('update:modelValue', null);
    return;
  }
  emit('update:modelValue', snapshot.id);
};

const getSpaceFromTimeRange = (a: Date, b: Date): number => {
  const diffMs = b.getTime() - a.getTime();
  const seconds = diffMs / 1000;
  
  const spaceMin = 0;
  const spaceMax = 40;
  const multiply = 0.01;
  const power = 0.62;
  const space = Math.max(Math.min(Math.pow(multiply * seconds, power), spaceMax), spaceMin);
  
  return space;
};

const getSnapshotFromY = (clientY: number): FormattedSnapshot | null => {
  if (!tableRef.value) return null;
  
  const rows = tableRef.value.querySelectorAll('tbody tr');
  let closestRow: Element | null = null;
  let closestDistance = Infinity;
  
  rows.forEach((row) => {
    const rect = row.getBoundingClientRect();
    const rowCenter = rect.top + rect.height / 2;
    const distance = Math.abs(clientY - rowCenter);
    
    if (distance < closestDistance) {
      closestDistance = distance;
      closestRow = row;
    }
  });
  
  if (!closestRow) return null;
  
  const snapshotId = (closestRow as HTMLElement).dataset.snapshotId;
  return snapshots.value.find(s => s.id === snapshotId) || null;
};

const handleDrag = (clientY: number) => {
  if (!isDragging.value) return;
  
  const snapshot = getSnapshotFromY(clientY);
  if (snapshot) {
    select(snapshot);
  }
};

const startDrag = (event: MouseEvent | TouchEvent) => {
  isDragging.value = true;
  const clientY = 'touches' in event ? event.touches[0]?.clientY : event.clientY;
  if (clientY !== undefined) {
    handleDrag(clientY);
  }
  event.preventDefault();
};

const onMouseMove = (event: MouseEvent) => {
  handleDrag(event.clientY);
};

const onTouchMove = (event: TouchEvent) => {
  if (event.touches.length > 0 && event.touches[0]) {
    handleDrag(event.touches[0].clientY);
  }
};

const stopDrag = () => {
  isDragging.value = false;
};

const updateSelectionPosition = () => {
  if (!tableRef.value) return;
  
  const selectedIndex = snapshots.value.findIndex(s => s.selected);
  if (selectedIndex === -1) {
    selectionTop.value = -1;
    return;
  }
  
  const tbody = tableRef.value.querySelector('tbody');
  if (!tbody) return;
  
  const rows = tbody.querySelectorAll('tr');
  const selectedRow = rows[selectedIndex] as HTMLElement;
  if (!selectedRow) return;
  
  const tbodyRect = tbody.getBoundingClientRect();
  const rowRect = selectedRow.getBoundingClientRect();
  
  selectionTop.value = rowRect.top - tbodyRect.top;
  selectionHeight.value = rowRect.height;
};

const formatTimestamp = (date: Date, currentYear: number, hasDuplicateDate: boolean, hasDuplicateTime: boolean) => {
  const year = date.getFullYear();
  const day = date.getDate();
  const month = format(date, 'MMM');
  
  // Format day with leading space for single digits
  const dayStr = day < 10 ? `0${day}` : `${day}`;
  
  // Build the date part
  let result = `${dayStr} ${month}`;
  
  // Add year if not current year
  if (year !== currentYear) {
    result += ` ${year}`;
  }
  
  // Add time if there are multiple snapshots on the same date
  if (hasDuplicateDate) {
    const time = hasDuplicateTime 
      ? format(date, 'HH:mm:ss')  // Include seconds to disambiguate
      : format(date, 'HH:mm');
    result += ` • ${time}`;
  }
  
  return result;
};

onMounted(() => {
  document.addEventListener('mousemove', onMouseMove);
  document.addEventListener('mouseup', stopDrag);
  document.addEventListener('touchmove', onTouchMove);
  document.addEventListener('touchend', stopDrag);
  
  nextTick(() => {
    updateSelectionPosition();
  });
});

onUnmounted(() => {
  document.removeEventListener('mousemove', onMouseMove);
  document.removeEventListener('mouseup', stopDrag);
  document.removeEventListener('touchmove', onTouchMove);
  document.removeEventListener('touchend', stopDrag);
});

const current = computed((): FormattedSnapshot => ({
  id: 'current',
  title: 'Current',
  type: '',
  selected: props.modelValue === null,
  timestamp: undefined,
}));

const snapshots = computed(() => {
  const apiSnapshots = data.value?.snapshots || [];
  const currentYear = new Date().getFullYear();
  
  // Parse all snapshots with their dates
  const snapshotsWithDates: SnapshotWithDate[] = apiSnapshots.map((s: ApiSnapshot) => ({
    ...s,
    selected: props.modelValue === s.id,
    date: new Date(parseInt(s.timestamp, 10) * 1000),
  }));
  
  // Format each snapshot and calculate margin based on time differences
  const formattedSnapshots: FormattedSnapshot[] = snapshotsWithDates.map((s, index) => {
    const dayTimestamp = Math.floor(new Date(s.date).setHours(0, 0, 0, 0) / 86400000);
    const minuteTimestamp = Math.floor(s.date.getTime() / 60000);
    
    // Check prev/next snapshots to determine if there are duplicates on the same day
    const prevSnapshot = index > 0 ? snapshotsWithDates[index - 1] : null;
    const nextSnapshot = index < snapshotsWithDates.length - 1 ? snapshotsWithDates[index + 1] : null;
    
    const prevDayTimestamp = prevSnapshot ? Math.floor(new Date(prevSnapshot.date).setHours(0, 0, 0, 0) / 86400000) : null;
    const nextDayTimestamp = nextSnapshot ? Math.floor(new Date(nextSnapshot.date).setHours(0, 0, 0, 0) / 86400000) : null;
    
    const hasDuplicateDate = dayTimestamp === prevDayTimestamp || dayTimestamp === nextDayTimestamp;
    
    // Check if there are duplicates at the same minute (only if there's a duplicate date)
    let hasDuplicateTime = false;
    if (hasDuplicateDate) {
      const prevMinuteTimestamp = prevSnapshot ? Math.floor(prevSnapshot.date.getTime() / 60000) : null;
      const nextMinuteTimestamp = nextSnapshot ? Math.floor(nextSnapshot.date.getTime() / 60000) : null;
      hasDuplicateTime = minuteTimestamp === prevMinuteTimestamp || minuteTimestamp === nextMinuteTimestamp;
    }
    
    // Calculate margin bottom based on time difference to next snapshot
    let marginBottom = 0;
    if (nextSnapshot) {
      marginBottom = getSpaceFromTimeRange(nextSnapshot.date, s.date);
    }

    const snapNode = nodeSnapshotById.value.get(s.id);
    const nextSnapNode = nextSnapshot ? nodeSnapshotById.value.get(nextSnapshot.id) : null;

    const unmodified = snapNode && nextSnapNode &&
      snapNode.type === "file" &&
      snapNode.last_modified === nextSnapNode.last_modified;
    
    return {
      id: s.id,
      title: formatTimestamp(s.date, currentYear, hasDuplicateDate, hasDuplicateTime),
      type: s.type,
      selected: s.selected,
      timestamp: s.timestamp,
      marginBottom,
      node: snapNode,
      unmodified: unmodified == null ? undefined : unmodified,
    };
  });
  
  // Filter out unmodified snapshots if hideUnmodified is true
  const filtered = isRootPath.value ? formattedSnapshots : formattedSnapshots.filter(s => {
    return s.selected ||
      (!hideUnmodified.value || s.unmodified !== true) &&
      (!hideNotFound.value || s.node !== null);
  });
  
  return [current.value, ...filtered];
});

watch(() => props.modelValue, () => {
  nextTick(() => {
    updateSelectionPosition();
  });
});

watch(snapshots, () => {
  nextTick(() => {
    updateSelectionPosition();
  });
});

</script>

<style scoped>

.snapshot-container {
  display: flex;
  flex-direction: column;
  max-height: 90vh;
  width: 140px;
}

.controls {
  position: sticky;
  top: 0;
  background-color: white;
  padding: 4px 8px;
  border-bottom: 1px solid #e5e5e5;
  z-index: 10;
  display: flex;
  gap: 4px;
  justify-content: flex-end;
  flex-shrink: 0;
  height: 36px;
  align-items: center;
}

.table-wrapper {
  position: relative;
  overflow-y: scroll;
  flex: 1;
}

.toggle-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
  padding: 0;
  border: 1px solid transparent;
  border-radius: 5px;
  background: transparent;
  opacity: 0.4;
  color: var(--color-text-dark);
  cursor: pointer;
  transition: background-color 0.1s ease;
}

.toggle-btn i {
  pointer-events: none;
  margin: 1px 1px 0 0;
}

.toggle-btn:hover {
  background-color: rgba(90, 93, 94, 0.1);
}

.toggle-btn.active {
  /* background-color: rgba(90, 93, 94, 0.31); */
  opacity: 1;
}

.toggle-btn.active:hover {
  background-color: rgba(90, 93, 94, 0.41);
}

table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.7em;
  border: 0;
  border-spacing: 0;
  user-select: none;
  position: relative;
  z-index: 1;
  background: transparent;
}

thead {
  position: sticky;
  top: 0;
  background-color: white;
  z-index: 2;
}

tbody {
  cursor: grab;
  position: relative;
}

tbody::before {
  content: '';
  position: absolute;
  left: 0;
  right: 0;
  top: var(--selection-top, 0);
  height: var(--selection-height, 0);
  background-color: #eef;
  border-left: 2px solid rgb(178, 178, 255);
  opacity: var(--selection-opacity, 0);
  transition: top 0.15s ease-out, height 0.15s ease-out, opacity 0.15s ease-out;
  pointer-events: none;
  z-index: -1;
}

tbody:active {
  cursor: grabbing;
}

td {
  margin: 0;
  padding: 0px;
  border: 0;
  line-height: 1;
  background: transparent;
}

tr {
  border: 0;
  padding: 0;
  line-height: 1;
}

tr.not-found, tr.unmodified {
  opacity: 0.5;
}

tr.not-found td label {
  text-decoration: line-through;
}

td label {
  cursor: inherit;
  display: block;
  padding: 8px;
  width: 100%;
  height: 100%;
  margin: 0;
  line-height: 1.5;
  pointer-events: none;
  font-size: 1.2em;
  font-family: monospace;
  font-variant-numeric: tabular-nums;
}

/* th.type {
  visibility: hidden;
} */

td.type, th.type {
  text-align: right;
  color: #888;
}

td.type label {
  color: inherit;
}

table td {
  text-wrap: nowrap;
}

</style>