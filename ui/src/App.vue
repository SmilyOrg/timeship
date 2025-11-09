<template>
  <main>
    <header>
      <img class="logo" src="/logo.png" alt="Timeship Logo" />
    </header>
    <browser class="browser" />
  </main>
</template>

<script setup lang="ts">
import Browser from './components/Browser.vue';
</script>

<style scoped>

.logo {
  display: block;
  height: 64px;
}

.browser {
  width: 100%;
  height: 600px;
}
</style>
