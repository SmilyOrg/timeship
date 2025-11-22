<template>
  <div class="text-viewer" ref="editorContainer"></div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch } from 'vue';
import * as monaco from 'monaco-editor';

const props = defineProps<{
  content: string;
  language?: string;
  filename?: string;
}>();

const editorContainer = ref<HTMLElement | null>(null);
let editor: monaco.editor.IStandaloneCodeEditor | null = null;

// Detect language from filename or mime type
function detectLanguage(filename?: string, explicitLanguage?: string): string {
  if (explicitLanguage) {
    return explicitLanguage;
  }
  
  if (!filename) {
    return 'plaintext';
  }
  
  const ext = filename.split('.').pop()?.toLowerCase();
  
  // Map common extensions to Monaco languages
  const languageMap: Record<string, string> = {
    'js': 'javascript',
    'ts': 'typescript',
    'jsx': 'javascript',
    'tsx': 'typescript',
    'json': 'json',
    'html': 'html',
    'css': 'css',
    'scss': 'scss',
    'less': 'less',
    'md': 'markdown',
    'py': 'python',
    'rb': 'ruby',
    'go': 'go',
    'rs': 'rust',
    'java': 'java',
    'c': 'c',
    'cpp': 'cpp',
    'h': 'c',
    'hpp': 'cpp',
    'cs': 'csharp',
    'php': 'php',
    'sh': 'shell',
    'bash': 'shell',
    'zsh': 'shell',
    'yaml': 'yaml',
    'yml': 'yaml',
    'xml': 'xml',
    'sql': 'sql',
    'dockerfile': 'dockerfile',
    'txt': 'plaintext',
  };
  
  return languageMap[ext || ''] || 'plaintext';
}

onMounted(() => {
  if (!editorContainer.value) return;
  
  const language = detectLanguage(props.filename, props.language);
  
  editor = monaco.editor.create(editorContainer.value, {
    value: props.content,
    language: language,
    theme: 'vs',
    readOnly: true,
    automaticLayout: true,
    minimap: {
      enabled: true,
    },
    scrollBeyondLastLine: false,
    fontSize: 14,
    lineNumbers: 'on',
    renderWhitespace: 'selection',
    folding: true,
    lineDecorationsWidth: 10,
    lineNumbersMinChars: 3,
  });
});

// Watch for content changes
watch(() => props.content, (newContent) => {
  if (editor && newContent !== editor.getValue()) {
    editor.setValue(newContent);
  }
});

// Watch for language changes
watch(() => [props.language, props.filename], () => {
  if (editor) {
    const language = detectLanguage(props.filename, props.language);
    const model = editor.getModel();
    if (model) {
      monaco.editor.setModelLanguage(model, language);
    }
  }
});

onUnmounted(() => {
  editor?.dispose();
});
</script>

<style scoped>
.text-viewer {
  width: 100%;
  height: 400px;
  border: 1px solid rgba(34, 36, 38, 0.15);
  border-radius: 0.28571429rem;
}
</style>
