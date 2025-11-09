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
      <thead>
        <tr>
          <!-- <th>Name</th> -->
          <th>Snapshots</th>
          <th class="type">Type</th>
        </tr>
      </thead>
      <tbody
        @mousedown="startDrag"
        @touchstart="startDrag"
      >
        <tr
          v-for="snapshot in snapshots"
          :key="snapshot.id"
          :data-snapshot-id="snapshot.id"
          ref="rowRefs"
        >
          <td>
            <label>
              {{ snapshot.title }}
            </label>
          </td>
          <td class="type">
            <label>
              {{ snapshot.type }}
            </label>
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, onMounted, onUnmounted, watch, nextTick } from 'vue';
import { useApi, type Snapshot as ApiSnapshot } from './api/api';
import { format } from 'date-fns';

interface Snapshot {
  id: string;
  title: string;
  type?: string;
  selected?: boolean;
  timestamp?: string;
}

const props = defineProps<{
  modelValue?: string | null;
}>();

const emit = defineEmits<{
  'update:modelValue': [value: string | null];
}>();

const {
  data,
} = useApi("/storages/local/snapshots");

const tableRef = ref<HTMLElement | null>(null);
const rowRefs = ref<HTMLElement[]>([]);
const isDragging = ref(false);
const selectionTop = ref(-1);
const selectionHeight = ref(0);

const select = (snapshot: Snapshot | null) => {
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

const getSnapshotFromY = (clientY: number): Snapshot | null => {
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

const current = computed((): Snapshot => ({
  id: 'current',
  title: 'Current',
  selected: props.modelValue === null,
  timestamp: undefined,
}));

const snapshots = computed(() => {
  const apiSnapshots = data.value?.snapshots || [];
  const currentYear = new Date().getFullYear();
  
  // Parse all snapshots with their dates
  interface SnapshotWithDate {
    id: string;
    type: string;
    selected: boolean;
    timestamp: string;
    date: Date;
  }
  
  const snapshotsWithDates: SnapshotWithDate[] = apiSnapshots.map((s: ApiSnapshot) => ({
    id: s.id,
    type: s.type,
    selected: props.modelValue === s.id,
    timestamp: s.timestamp,
    date: new Date(parseInt(s.timestamp, 10) * 1000),
  }));
  
  // Group by date to find duplicates
  const dateGroups = new Map<string, number>();
  snapshotsWithDates.forEach((s: SnapshotWithDate) => {
    const dateKey = format(s.date, 'yyyy-MM-dd');
    dateGroups.set(dateKey, (dateGroups.get(dateKey) || 0) + 1);
  });
  
  // Group by date+time (including seconds) to find time duplicates
  const timeGroups = new Map<string, number>();
  snapshotsWithDates.forEach((s: SnapshotWithDate) => {
    const dateKey = format(s.date, 'yyyy-MM-dd');
    if (dateGroups.get(dateKey)! > 1) {
      const timeKey = format(s.date, 'yyyy-MM-dd HH:mm');
      timeGroups.set(timeKey, (timeGroups.get(timeKey) || 0) + 1);
    }
  });
  
  // Format each snapshot and calculate margin based on time differences
  const formattedSnapshots = snapshotsWithDates.map((s: SnapshotWithDate, index: number) => {
    const dateKey = format(s.date, 'yyyy-MM-dd');
    const timeKey = format(s.date, 'yyyy-MM-dd HH:mm');
    const hasDuplicateDate = dateGroups.get(dateKey)! > 1;
    const hasDuplicateTime = timeGroups.get(timeKey)! > 1;
    
    // Calculate margin bottom based on time difference to next snapshot
    let marginBottom = 0;
    if (index < snapshotsWithDates.length - 1) {
      const nextSnapshot = snapshotsWithDates[index + 1];
      if (nextSnapshot) {
        marginBottom = getSpaceFromTimeRange(nextSnapshot.date, s.date);
      }
    }
    
    return {
      id: s.id,
      title: formatTimestamp(s.date, currentYear, hasDuplicateDate, hasDuplicateTime),
      type: s.type,
      selected: s.selected,
      timestamp: s.timestamp,
      marginBottom,
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

th.type {
  visibility: hidden;
}

td.type {
  text-align: right;
  color: #888;
}

table td {
  text-wrap: nowrap;
}

</style>