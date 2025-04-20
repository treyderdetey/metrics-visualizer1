<script setup lang="ts">
import MetricsDisplay from './components/MetricsDisplay.vue';
</script>

<template>
  <main>
    <MetricsDisplay />
  </main>
</template>

<style scoped>
/* Можно удалить или оставить стандартные стили App.vue */
header {
  line-height: 1.5;
}

.logo {
  display: block;
  margin: 0 auto 2rem;
}

@media (min-width: 1024px) {
  header {
    display: flex;
    place-items: center;
    padding-right: calc(var(--section-gap) / 2);
  }

  .logo {
    margin: 0 2rem 0 0;
  }

  header .wrapper {
    display: flex;
    place-items: flex-start;
    flex-wrap: wrap;
  }
}
</style>