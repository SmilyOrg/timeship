import { useQueries, useQuery } from "@tanstack/vue-query";
import { computed, watch, type Ref } from "vue";
import { API_BASE_URL } from "../../config";

console.log("API_BASE_URL:", API_BASE_URL);

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

/**
 * Parse a node path into storage and path components
 * Example: "local://path/to/file" -> { storage: "local", path: "path/to/file" }
 */
function parsePath(fullPath: string): { storage: string; path: string } {
  // Parse the path to extract storage and path components
  const url = new URL(fullPath);
  const storage = url.protocol.replace(':', '');
  const path = url.host + url.pathname;
  return { storage, path };
}

/**
 * Generate a URL for a node (file or directory)
 * @param nodePath - The full path of the node (e.g., "/storage1/path/to/file")
 * @param options - Optional parameters
 * @param options.snapshot - Snapshot ID to include in the URL
 * @param options.download - Whether to add the download parameter
 * @returns The full URL to access the node
 */
export function getNodeUrl(
  nodePath: string,
  options?: { snapshot?: string | null; download?: boolean }
): string {
  const { storage, path: urlPath } = parsePath(nodePath);
  let url = `${API_BASE_URL}/storages/${storage}/nodes/${urlPath}`;
  
  const params = new URLSearchParams();
  if (options?.snapshot) {
    params.set('snapshot', options.snapshot);
  }
  if (options?.download) {
    params.set('download', 'true');
  }
  
  const queryString = params.toString();
  return queryString ? `${url}?${queryString}` : url;
}

// Uses useQuery for a generic API implementation
export function useApi(endpoint: Ref<string>) {
  return useQuery({
    queryKey: [endpoint],
    queryFn: async () => {
      const response = await fetch(`${API_BASE_URL}${endpoint.value}`);
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
        const response = await fetch(`${API_BASE_URL}${endpoint}`);
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