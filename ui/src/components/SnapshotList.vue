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
        <tr v-for="snapshot in snapshots" :key="snapshot.id">
          <td>{{ formatTimestamp(snapshot.timestamp) }}</td>
          <!-- <td>{{ snapshot.name }}</td> -->
          <td class="type">{{ snapshot.type }}</td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<script setup lang="ts">
import { computed, watch } from 'vue';
import { useApi } from './api/api';
import { format } from 'date-fns';

const {
  isPending,
  isFetching,
  isError,
  data,
  error
} = useApi("/storages/local/snapshots");

const snapshots = computed(() => {
  return data.value?.snapshots || [];
});

const formatTimestamp = (timestamp: string) => {
  const date = new Date(parseInt(timestamp, 10) * 1000);
  // return format(date, "HH:mm:ss dd LLL yyyy");
  return format(date, "dd LLL yyyy");
};

</script>

<style scoped>

table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.7em;
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