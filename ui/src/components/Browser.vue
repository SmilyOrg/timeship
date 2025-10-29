<template>
  <div>
    <vue-finder
      class="finder"
      id="vf"
      :request="request"
    ></vue-finder>
  </div>
</template>

<script setup lang="ts">

const request: {
  baseUrl: string;
  transformRequest?: (req: any) => any;
} = {
  baseUrl: 'http://localhost:8080',

  // Transform v1 API calls to v2 API format (mutates the request object)
  transformRequest: (req: any) => {
    // Only transform GET requests with q parameter (v1 API)
    if (req?.method === 'get' && req.params?.q) {
      const operation = req.params.q as string;
      const adapter = req.params.adapter || 'local';
      const path = req.params.path || '';

      // Remove adapter:// prefix from path if present
      let nodePath = path;
      const prefix = `${adapter}://`;
      if (nodePath.startsWith(prefix)) {
        nodePath = nodePath.slice(prefix.length);
      }

      const setUrlForNode = (base: string) => {
        req.url = nodePath ? `${base}/${nodePath}` : base;
      };

      const cleanParams = () => {
        delete req.params.q;
        delete req.params.adapter;
        delete req.params.path;
      };

      switch (operation) {
        case 'index': {
          const base = `${request.baseUrl}/storages/${adapter}/nodes`;
          setUrlForNode(base);
          cleanParams();
          break;
        }

        case 'download':
        case 'preview': {
          const base = `${request.baseUrl}/storages/${adapter}/nodes/${nodePath}`;
          req.url = base;
          cleanParams();
          break;
        }

        default:
          throw new Error(`Unsupported v1 operation: ${operation}`);
      }
    }

    return req;
  },
};

</script>

<style scoped>
.finder {
  width: 100%;
}
</style>