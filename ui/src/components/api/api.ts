import { useQuery } from "@tanstack/vue-query";

// Uses useQuery for a generic API implementation
function useApi(endpoint: string) {
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

export { useApi };