<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { useProfileStore } from '@/stores/profile'

const route = useRoute()
const profileStore = useProfileStore()

const links = computed(() => {
  const base = [{ to: '/', label: '概览', icon: '◉' }]
  if (profileStore.profiles.length > 0) {
    const defaultId = (route.params.id as string) || profileStore.profiles[0].id
    base.push(
      { to: `/profiles/${defaultId}/config`, label: '配置中心', icon: '⚙' },
      { to: `/profiles/${defaultId}/runtime`, label: '服务运行时', icon: '⚡' },
      { to: `/profiles/${defaultId}/synclist`, label: '选择性同步', icon: '📁' },
    )
  }
  base.push({ to: '/profiles', label: '账号管理', icon: '👤' })
  return base
})
</script>

<template>
  <aside class="sidebar">
    <div class="brand">
      <span class="logo">1W</span>
      <span class="brand-name">OneWeb</span>
    </div>

    <nav class="nav">
      <router-link
        v-for="l in links"
        :key="l.to"
        :to="l.to"
        class="nav-link"
        :class="{ active: route.path === l.to }"
      >
        <span class="icon">{{ l.icon }}</span>{{ l.label }}
      </router-link>
    </nav>

    <div class="sidebar-section">
      <div class="section-title">Profiles</div>
      <router-link
        v-for="p in profileStore.profiles"
        :key="p.id"
        :to="`/profiles/${p.id}`"
        class="profile-link"
        :class="{ active: route.params.id === p.id }"
      >
        {{ p.display_name || p.id }}
      </router-link>
    </div>
  </aside>
</template>

<style scoped>
.sidebar {
  width: 220px;
  background: #0f172a;
  color: #e2e8f0;
  display: flex;
  flex-direction: column;
  padding: 20px 12px;
}
.brand {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 0 8px 20px;
  border-bottom: 1px solid #1e293b;
}
.logo {
  background: var(--oneweb-primary);
  border-radius: 8px;
  padding: 4px 8px;
  font-weight: 700;
  font-size: 14px;
  color: #fff;
}
.brand-name {
  font-weight: 600;
  font-size: 15px;
}
.nav {
  display: flex;
  flex-direction: column;
  gap: 4px;
  padding: 16px 0;
}
.nav-link,
.profile-link {
  color: #94a3b8;
  text-decoration: none;
  padding: 8px 10px;
  border-radius: 6px;
  font-size: 14px;
  display: flex;
  align-items: center;
  gap: 8px;
}
.nav-link:hover,
.profile-link:hover {
  background: #1e293b;
  color: #e2e8f0;
}
.nav-link.active,
.profile-link.active {
  background: #1e293b;
  color: #fff;
}
.icon {
  width: 18px;
  text-align: center;
}
.sidebar-section {
  padding-top: 8px;
  border-top: 1px solid #1e293b;
  display: flex;
  flex-direction: column;
  gap: 2px;
  overflow-y: auto;
}
.section-title {
  font-size: 12px;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: #64748b;
  padding: 8px 10px;
}
</style>
