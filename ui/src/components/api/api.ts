import { useQueries, useQuery } from "@tanstack/vue-query";
import { computed, type Ref } from "vue";
import { API_BASE_URL } from "../../config";

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
  path: string;  // Path relative to storage root (no longer includes storage prefix)
  type: string;
  mime_type?: string;
}

/**
 * Generate a URL for a node (file or directory)
 * @param storage - The storage identifier
 * @param nodePath - The path of the node relative to storage root
 * @param options - Optional parameters
 * @param options.snapshot - Snapshot ID to include in the URL
 * @param options.download - Whether to add the download parameter
 * @returns The full URL to access the node
 */
export function getNodeUrl(
  storage: string,
  nodePath: string,
  options?: { snapshot?: string | null; download?: boolean }
): string {
  let url = `${API_BASE_URL}/storages/${storage}/nodes/${nodePath}`;
  
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
      const response = await fetch(`${API_BASE_URL}${endpoint.value}`, {
        headers: {
          'Accept': 'application/json',
        },
      });
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
    return endpoints.value.map(endpoint => ({
      queryKey: [endpoint],
      queryFn: async () => {
        const response = await fetch(`${API_BASE_URL}${endpoint}`, {
          headers: {
            'Accept': 'application/json',
          },
        });
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
  
  return useQueries({ queries });
}