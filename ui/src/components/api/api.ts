import { useQueries, useQuery } from "@tanstack/vue-query";
import { computed, watch, type Ref } from "vue";

export interface Snapshot {
  id: string;
  timestamp: string;
  type: string;
}

export interface Node {
  basename: string;
  extension: string;
  file_size: number;
  last_modified: number;
  path: string;
  type: string;
}

// Uses useQuery for a generic API implementation
export function useApi(endpoint: string) {
  return useQuery({
    queryKey: [endpoint],
    queryFn: async () => {
      const response = await fetch(`http://localhost:8080${endpoint}`);
      if (!response.ok) {
        throw new Error('Request failed: ' + response.statusText);
      }
      return response.json();
    },
  });
}

export function useApis(endpoints: Ref<string[]>) {
  const queries = computed(() => {
    console.log("Endpoints:", endpoints.value);
    return endpoints.value.map(endpoint => ({
      queryKey: [endpoint],
      queryFn: async () => {
        const response = await fetch(`http://localhost:8080${endpoint}`);
        if (!response.ok) {
          throw new Error('Request failed: ' + response.statusText);
        }
        return response.json();
      },
    }));
  });
  watch(queries, (newQueries) => {
    console.log("Queries updated:", newQueries);
  });
  // const userQueries = useQueries({queries: queries})


  // const queries = computed(() => {
  //   return endpoints.value.map((endpoint) => ({
  //     queryKey: [endpoint],
  //     queryFn: async () => {
  //       const response = await fetch(`http://localhost:8080${endpoint}`);
  //       if (!response.ok) {
  //         throw new Error('Request failed: ' + response.statusText);
  //       }
  //       return response.json();
  //     },
  //   }));
  // });
  const data = useQueries({ queries });
  watch(data, (newData) => {
    console.log("Data updated:", newData);
  });
  return data;
}