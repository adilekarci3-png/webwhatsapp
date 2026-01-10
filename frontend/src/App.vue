<!-- src/App.vue -->
<template>
  <router-view />
</template>

<script setup lang="ts">
// Root App: Tüm sayfalar vue-router üzerinden render edilir.
// Burada ekstra layout kullanmak istersen ileride <MainLayout> gibi sarmalayabilirsin.
</script>

<style>
/* Global baseline (isteğe bağlı) */
html, body, #app {
  height: 100%;
}

body {
  margin: 0;
}
</style>
