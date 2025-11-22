<template>
  <main>
    <header>
      <router-link to="/">
        <img class="logo" src="/logo.png" alt="Cute spaceship logo" />
        <h1>Timeship</h1>
      </router-link>
    </header>
    <router-view class="browser" />
  </main>
</template>

<script setup lang="ts">
</script>

<style scoped>

.logo {
  display: block;
  width: fit-content;
  height: 64px;
  margin-bottom: -16px;
  cursor: pointer;
}

header a {
  text-decoration: none;
  display: block;
  width: fit-content;
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
  color: #cac1ab;
  cursor: pointer;
  /* text-shadow: -1px -1px 0 #fbf2da, 1px -1px 0 #fbf2da, -1px 1px 0 #fbf2da, 1px 1px 0 #fbf2da; */
}

</style>
