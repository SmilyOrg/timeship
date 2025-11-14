# FileTable Component

## Overview

FileTable is a composable, reusable file browser component for the Timeship UI. It provides a focused, table-based view of files and directories with support for selection and navigation.

## Philosophy

### Pure Presentation Component

FileTable is a **controlled, presentational component** - it receives data via props and emits events for interactions. It doesn't:
- Fetch data from APIs
- Parse paths or URLs
- Manage application state
- Handle routing or navigation logic

This makes it extremely flexible and reusable in various contexts:
- Traditional file browsers
- Search results viewers
- File pickers / selection dialogs
- Snapshot comparisons
- Filtered/sorted views

### Composability Over Monolithic Design

FileTable focuses on doing one thing well: displaying a list of files/folders and managing user interactions. Navigation controls, path inputs, data fetching, and toolbar actions are intentionally left to parent components. This allows:

- **Flexible composition**: Combine FileTable with different data sources and navigation patterns
- **Easier testing**: Test display and interaction logic without mocking APIs
- **Reusability**: Use FileTable anywhere you need to show a list of nodes

### Controlled Component Pattern

FileTable doesn't manage navigation state. Instead:

- Accepts `nodes` as a prop (array of Node objects)
- Accepts optional `loading` and `error` props for state display
- Emits `navigate` events when the user wants to open a folder
- Parent component controls the data fetching and path changes

This pattern makes the component predictable and allows parent components to:
- Use any data source (API, local state, search results, etc.)
- Intercept navigation (e.g., show confirmation dialogs)
- Implement custom data transformation (filtering, sorting, etc.)
- Synchronize with other components

## Requirements

### Core Functionality

1. **File Display**
   - Table layout with semantic HTML (`<table>`)
   - Columns: icon, name, size, modified date
   - Unicode icons (📁 folders, 📄 files) - easily replaceable
   - Format sizes (B, KB, MB, GB, TB)
   - Format timestamps (locale-aware)

2. **Navigation**
   - Double-click folder to emit `navigate` event
   - Enter key on selected folder to emit `navigate` event
   - Parent controls actual navigation and data fetching

3. **Selection**
   - Single-click to select
   - Ctrl/Cmd+click to toggle individual items
   - Shift+click for range selection
   - Arrow keys to move selection
   - Shift+Arrow for range selection with keyboard
   - Ctrl+Arrow to add to selection with keyboard
   - Emit complete selection state (array of paths)

4. **States**
   - Loading: Pulsing opacity animation (via `loading` prop)
   - Error: Large warning icon with error message (via `error` prop)
   - Empty: Large folder icon with "Folder is empty" message

5. **Accessibility**
   - Focusable table (tabindex="0")
   - Keyboard interactions only work when focused
   - Proper semantic HTML structure
   - Scroll selected items into view

### Props

```typescript
interface Props {
  nodes: Node[];        // Array of files/folders to display
  loading?: boolean;    // Show loading animation
  error?: string | null; // Error message to display
}
```

### Events

```typescript
interface Events {
  'update:selection': [paths: string[]]; // Selection changed
  'navigate': [path: string];            // User wants to open a folder
}
```

### Node Interface

```typescript
interface Node {
  path: string;          // Full path with storage prefix
  type: 'file' | 'dir';  // Node type
  basename: string;      // Display name
  extension: string;     // File extension (empty for dirs)
  file_size: number;     // Size in bytes (0 for dirs)
  last_modified: number; // Unix timestamp
  mime_type?: string;    // Optional MIME type
  url?: string | null;   // Optional public URL
}
```

## Non-Requirements

Intentionally excluded to maintain composability:

- **Data fetching**: Parent handles API calls
- **Path parsing**: Parent handles URL parsing
- **Path bar / breadcrumbs**: Separate component
- **Back/forward buttons**: Separate component
- **Search box**: Separate component
- **Toolbar actions**: Separate components
- **Context menu**: Will be added later as separate concern
- **Drag and drop**: Future enhancement
- **Column sorting**: Parent handles sorting before passing nodes
- **Filtering**: Parent handles filtering before passing nodes
- **View modes (grid/list)**: FileTable is table-only; grid views are separate components

## Styling Philosophy

Minimal styling in the component itself:

- Only essential styles (selection highlight, loading animation)
- No colors beyond basic grays (themeable by parent)
- No spacing/layout beyond table structure
- Parent component controls overall appearance through CSS

This allows FileTable to adapt to different design systems without fighting built-in styles.

## Usage Example

### Basic Usage

```vue
<template>
  <div class="file-browser">
    <PathBreadcrumbs :path="currentPath" @navigate="currentPath = $event" />
    
    <FileTable 
      :nodes="nodes"
      :loading="isLoading"
      :error="error"
      @navigate="currentPath = $event"
      @update:selection="selectedFiles = $event"
    />
    
    <FileActions :selection="selectedFiles" />
  </div>
</template>

<script setup>
import { ref, computed } from 'vue';
import { useApi } from './api/api';

const currentPath = ref('local://documents');
const selectedFiles = ref([]);

// Parse path and build API endpoint
const apiEndpoint = computed(() => {
  const url = new URL(currentPath.value);
  const storage = url.protocol.replace(':', '');
  const path = url.pathname.replace(/^\//, '');
  return `/storages/${storage}/nodes${path ? '/' + path : ''}`;
});

// Fetch data
const { data, error, isLoading } = useApi(apiEndpoint);

// Extract nodes
const nodes = computed(() => data.value?.files || []);
</script>
```

### With Snapshot Support

```vue
<template>
  <FileTable 
    :nodes="nodes"
    :loading="isLoading"
    :error="error"
    @navigate="handleNavigate"
  />
</template>

<script setup>
const selectedSnapshot = ref(null);

const apiEndpoint = computed(() => {
  const { storage, path } = parsePath(currentPath.value);
  let endpoint = `/storages/${storage}/nodes${path ? '/' + path : ''}`;
  if (selectedSnapshot.value) {
    endpoint += `?snapshot=${selectedSnapshot.value}`;
  }
  return endpoint;
});

const { data, error, isLoading } = useApi(apiEndpoint);
const nodes = computed(() => data.value?.files || []);
</script>
```

### With Search Results

```vue
<template>
  <FileTable 
    :nodes="searchResults"
    :loading="isSearching"
    @navigate="openFolder"
  />
</template>

<script setup>
// Search results come from a different API endpoint
const { data, isLoading: isSearching } = useApi(searchEndpoint);
const searchResults = computed(() => data.value?.files || []);

function openFolder(path) {
  // Navigate to the folder in the main browser
  emit('open-folder', path);
}
</script>
```

### With Custom Filtering

```vue
<template>
  <input v-model="filter" placeholder="Filter files..." />
  
  <FileTable 
    :nodes="filteredNodes"
    @navigate="currentPath = $event"
  />
</template>

<script setup>
const filter = ref('');

const filteredNodes = computed(() => {
  if (!filter.value) return nodes.value;
  return nodes.value.filter(node => 
    node.basename.toLowerCase().includes(filter.value.toLowerCase())
  );
});
</script>
```

## Future Enhancements

Potential additions that maintain composability:

- Custom columns (via slots)
- Inline rename (emit `rename` event)
- Context menu integration (emit `contextmenu` with position + selection)
- Drag and drop (emit `drag-start`, `drop` events)
- Virtual scrolling for large lists
- Keyboard shortcuts (Delete, Ctrl+A, etc.)

Each enhancement should follow the controlled component pattern: accept configuration via props, emit events for actions, let parent handle state and side effects.
