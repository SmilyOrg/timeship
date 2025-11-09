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
  selected?: boolean;
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
}));

const snapshots = computed(() => {
  const apiSnapshots = data.value?.snapshots || [];
  const snapshots = apiSnapshots.map((s: ApiSnapshot) => ({
    id: s.id,
    title: formatTimestamp(s.timestamp),
    type: s.type,
    selected: props.modelValue === s.id,
  }));
  return [current.value, ...snapshots];
});

const formatTimestamp = (timestamp: string) => {
  if (!timestamp) return 'Current';
  const date = new Date(parseInt(timestamp, 10) * 1000);
  // return format(date, "HH:mm:ss dd LLL yyyy");
  return format(date, "dd LLL yyyy HH:mm:ss");
  // return format(date, "dd LLL yyyy");
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