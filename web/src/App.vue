<script setup lang="ts">
import { onMounted } from 'vue'
import { useProfileStore } from '@/stores/profile'
import AppSidebar from '@/components/layout/AppSidebar.vue'

const profileStore = useProfileStore()

onMounted(() => {
  profileStore.fetchAll()
})
</script>

<template>
  <div class="app-shell">
    <AppSidebar />
    <main class="app-main">
      <router-view />
    </main>
  </div>
</template>

<style scoped>
.app-shell {
  display: flex;
  min-height: 100vh;
}
.app-main {
  flex: 1;
  padding: 24px 32px;
  overflow-x: hidden;
}
</style>
