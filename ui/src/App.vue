<template>
  <main>
    <header>
      <img class="logo" src="/logo.png" alt="Timeship Logo" />
      <h1>Timeship</h1>
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
  width: fit-content;
  height: 64px;
  margin-bottom: -16px;
}

.browser {
  width: 100%;
  height: 600px;
}

header {
  display: flex;
  flex-direction: column;
  margin-bottom: 16px;
}

h1 {
  font-family: Kanit, sans-serif;
  letter-spacing: -0.6px;
  font-weight: 700;
  font-size: 18px;
  font-style: italic;
  margin-top: 0;
}

</style>
