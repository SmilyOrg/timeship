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
          :class="{ selected: modelValue === snapshot.id }"
        >
          <td>
            <input
              type="checkbox"
              :id="snapshot.id"
              :checked="modelValue === snapshot.id"
              @change="select(($event.target as HTMLInputElement).checked ? snapshot : null)"
            ></input>
            <label :for="snapshot.id">
              {{ formatTimestamp(snapshot.timestamp) }}
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
import { useApi } from './api/api';
import { format } from 'date-fns';

interface Snapshot {
  id: string;
  timestamp: string;
  type: string;
}

defineProps<{
  modelValue?: string | null;
}>();

const emit = defineEmits<{
  'update:modelValue': [value: string | null];
}>();

const {
  data,
} = useApi("/storages/local/snapshots");

const select = (snapshot: Snapshot | null) => {
  emit('update:modelValue', snapshot?.id ?? null);
};

const snapshots = computed(() => {
  return data.value?.snapshots || [];
});

const formatTimestamp = (timestamp: string) => {
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