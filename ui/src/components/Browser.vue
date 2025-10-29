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

// const request = "http://localhost:8080/";

const request = {
  baseUrl: "http://localhost:8080",
  
  // Transform v1 API calls to v2 API format
  transformRequest: (req: any) => {
    console.log("Original request:", req);
    
    // Only transform GET requests with q parameter (v1 API)
    if (req.method === 'get' && req.params?.q) {
      const operation = req.params.q;
      const adapter = req.params.adapter || 'local'; // default to 'local' if not specified
      const path = req.params.path || ''; // default to empty (root)
      
      // Store operation for response transformation
      // req._v1Operation = operation;
      // req._v1Adapter = adapter;
      // req._v1Path = path;
      
      // Transform based on operation
      if (operation === 'index') {
        // Old: GET /?q=index&adapter=local&path=local://documents
        // New: GET /storages/local/nodes/documents
        
        // Extract path without adapter prefix
        let nodePath = path;
        if (nodePath.startsWith(`${adapter}://`)) {
          nodePath = nodePath.substring(`${adapter}://`.length);
        }
        
        // Build new v2 URL with full baseUrl
        let url = `${request.baseUrl}/storages/${adapter}/nodes`;
        if (nodePath) {
          url += "/" + nodePath;
        }
        req.url = url;
        
        // Remove old query params
        delete req.params.q;
        delete req.params.adapter;
        delete req.params.path;
        
        console.log("Transformed to v2:", req);
      }
    }
    
    return req;
  },

  // Transform v2 API responses back to v1 format
  // transformResponse: (response: any, request: any) => {
  //   console.log("Original response:", response);
  //   console.log("Request context:", request);
    
  //   // Only transform if this was a v1 API call
  //   if (request._v1Operation === 'index') {
  //     // V2 response format:
  //     // {
  //     //   "nodes": [{ "path": "documents", "type": "dir", "name": "documents", "storage": "local", "children": [...] }],
  //     //   "path": "documents",
  //     //   "storage": "local",
  //     //   "storages": ["local"]
  //     // }
  //     //
  //     // V1 response format:
  //     // {
  //     //   "adapter": "local",
  //     //   "storages": ["local"],
  //     //   "dirname": "local://documents",
  //     //   "files": [{ "path": "local://documents/file.txt", "type": "file", "basename": "file.txt", "storage": "local", ... }]
  //     // }
      
  //     const adapter = request._v1Adapter || 'local';
  //     const path = request._v1Path || `${adapter}://`;
      
  //     // Extract children from the first node (which is the directory itself)
  //     const files = response.nodes?.[0]?.children || [];
      
  //     // Convert v2 nodes to v1 files format
  //     const v1Files = files.map((node: any) => ({
  //       path: `${adapter}://${node.path}`,
  //       type: node.type,
  //       basename: node.name,
  //       storage: node.storage,
  //       extension: node.extension,
  //       mime_type: node.mime_type,
  //       file_size: node.size,
  //       last_modified: node.modified_at,
  //     }));
      
  //     const v1Response = {
  //       adapter: adapter,
  //       storages: response.storages || [adapter],
  //       dirname: path || `${adapter}://`,
  //       files: v1Files,
  //     };
      
  //     console.log("Transformed response to v1:", v1Response);
  //     return v1Response;
  //   }
    
  //   return response;
  // },
}

</script>

<style scoped>
.finder {
  width: 100%;
}
</style>