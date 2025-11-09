<template>
  <div>
    <table>
      <thead>
        <tr>
          <!-- <th>Name</th> -->
          <th>Snapshots</th>
          <th class="type">Type</th>
        </tr>
      </thead>
      <tbody>
        <tr
          v-for="snapshot in snapshots"
          :key="snapshot.id"
          :class="{ selected: snapshot.selected }"
        >
          <td>
            <input
              type="checkbox"
              :id="snapshot.id"
              :checked="snapshot.selected"
              @change="select(($event.target as HTMLInputElement).checked ? snapshot : null)"
            ></input>
            <label :for="snapshot.id">
              {{ snapshot.title }}
            </label>
          </td>
          <td class="type">
            <label :for="snapshot.id">
              {{ snapshot.type }}
            </label>
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue';
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

const select = (snapshot: Snapshot | null) => {
  if (!snapshot?.id || snapshot?.id === 'current') {
    emit('update:modelValue', null);
    return;
  }
  emit('update:modelValue', snapshot.id);
};

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

input {
  display: none;
}

table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.7em;
  border: 0;
  border-spacing: 0;
}

td {
  margin: 0;
  padding: 0px;
  border: 0;
  line-height: 1;
}

tr.selected td {
  background-color: #eef !important;
}

tr {
  border: 0;
  padding: 0;
  line-height: 1;
  border-left: 2px solid transparent;
}

tr.selected {
  /* background-color: #eef !important; */
  border-left: 2px solid rgb(178, 178, 255);
}

td label {
  cursor: pointer;
  display: block;
  padding: 8px;
  width: 100%;
  height: 100%;
  margin: 0;
  line-height: 1.5;
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