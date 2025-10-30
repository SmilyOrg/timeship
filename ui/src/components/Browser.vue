<template>
  <div class="browser">
    <snapshot-list>
    </snapshot-list>
    <vue-finder
      class="finder"
      id="vf"
      :request="request"
    ></vue-finder>
  </div>
</template>

<script setup lang="ts">
import SnapshotList from './SnapshotList.vue';


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

      const cleanParams = () => {
        delete req.params.q;
        delete req.params.adapter;
        delete req.params.path;
      };

      switch (operation) {
        case 'index': {
          const base = `${request.baseUrl}/storages/${adapter}/nodes`;
          req.url = nodePath ? `${base}/${nodePath}` : base;
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
.browser {
  display: flex;
  flex-direction: row;
  gap: 16px;
  width: 100%;
  height: 100%;
}

.finder {
  width: 100%;
}
</style>