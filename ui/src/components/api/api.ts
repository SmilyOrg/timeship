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
export function useApi(endpoint: Ref<string>) {
  return useQuery({
    queryKey: [endpoint],
    queryFn: async () => {
      const response = await fetch(`http://localhost:8080${endpoint.value}`);
      if (!response.ok) {
        if (response.status === 404) {
          throw new Error('Not found');
        }
        throw new Error('Request failed: ' + response.statusText);
      }
      return response.json();
    },
    retry: (failureCount, error) => {
      // Don't retry on 404 errors
      if (error.message === 'Not found') {
        return false;
      }
      // Retry other errors up to 3 times
      return failureCount < 3;
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
          if (response.status === 404) {
            throw new Error('Not found');
          }
          throw new Error('Request failed: ' + response.statusText);
        }
        return response.json();
      },
      retry: (failureCount: number, error: Error) => {
        // Don't retry on 404 errors
        if (error.message === 'Not found') {
          return false;
        }
        // Retry other errors up to 3 times
        return failureCount < 3;
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