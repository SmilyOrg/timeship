<template>
  <div class="snapshot-container">
    <div
      class="selection-indicator"
      :style="{
        top: `${selectionTop}px`,
        height: `${selectionHeight}px`,
        opacity: selectionTop >= 0 ? 1 : 0
      }"
    ></div>
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
  currentPath?: string;
}>();

const emit = defineEmits<{
  'update:modelValue': [value: string | null];
}>();

const { data } = useApi(ref("/storages/local/snapshots"));

// Build endpoints for checking snapshot availability
const nodeSnapshotEndpoints = computed(() => {
  const path = props.currentPath;
  const apiSnapshots = data.value?.snapshots || [];
  
  // If no path or at root, return empty array (no checks needed)
  if (!path || path === 'local://') {
    return [];
  }

  try {
    // Parse the path to extract storage and path components
    const url = new URL(path);
    const storage = url.protocol.replace(':', '');
    const urlPath = url.host + url.pathname;
    
    // Create endpoint for each snapshot
    return apiSnapshots.map((snapshot: ApiSnapshot) => 
      `/storages/${storage}/nodes/${urlPath}?snapshot=${snapshot.id}`
    );
  } catch (e) {
    console.error('Invalid path:', path, e);
    return [];
  }
});

// Use useApis to check all snapshots in parallel
const nodeSnapshots = useApis(nodeSnapshotEndpoints);

// Create a map of snapshot availability from the query results
const nodeSnapshotById = computed(() => {
  const map = new Map<string, Node | null>();
  const apiSnapshots = data.value?.snapshots || [];
  
  // If no path or at root, all snapshots are available
  if (!props.currentPath || props.currentPath === 'local://') {
    return map;
  }

  apiSnapshots.forEach((snapshot: ApiSnapshot, index: number) => {
    const result = nodeSnapshots.value[index];
    // If query is successful (no error), the path exists
    map.set(snapshot.id, result?.isSuccess && result.data as Node || null);
  });
  
  return map;
});


const tableRef = ref<HTMLElement | null>(null);
const isDragging = ref(false);
const selectionTop = ref(-1);
const selectionHeight = ref(0);

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
  
  const rows = tableRef.value.querySelectorAll('tbody tr');
  const selectedRow = rows[selectedIndex] as HTMLElement;
  if (!selectedRow) return;
  
  const tableRect = tableRef.value.getBoundingClientRect();
  const rowRect = selectedRow.getBoundingClientRect();
  
  selectionTop.value = rowRect.top - tableRect.top;
  selectionHeight.value = rowRect.height;
};

watch(() => props.modelValue, () => {
  nextTick(() => {
    updateSelectionPosition();
  });
});

watch(data, () => {
  nextTick(() => {
    updateSelectionPosition();
  });
});

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

    const node = nodeSnapshotById.value.get(s.id);
    const nextNode = nextSnapshot ? nodeSnapshotById.value.get(nextSnapshot.id) : null;

    const unmodified = node && nextNode && node.last_modified === nextNode.last_modified;
    
    return {
      id: s.id,
      title: formatTimestamp(s.date, currentYear, hasDuplicateDate, hasDuplicateTime),
      type: s.type,
      selected: s.selected,
      timestamp: s.timestamp,
      marginBottom,
      node,
      unmodified,
    };
  });
  
  return [current.value, ...formattedSnapshots];
});

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

</script>

<style scoped>

.snapshot-container {
  position: relative;
  max-height: 90vh;
  overflow-y: scroll;
}

.selection-indicator {
  position: absolute;
  left: 0;
  right: 0;
  background-color: #eef;
  border-left: 2px solid rgb(178, 178, 255);
  transition: top 0.15s ease-out, height 0.15s ease-out, opacity 0.15s ease-out;
  pointer-events: none;
  z-index: 0;
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

tr.not-found {
  opacity: 0.3;
}

tr.not-found td label {
  text-decoration: line-through;
}

tr.unmodified {
  opacity: 0.5;
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