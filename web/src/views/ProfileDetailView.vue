<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useProfileStore } from '@/stores/profile'
import StatusBadge from '@/components/runtime/StatusBadge.vue'
import { getRuntimeStatus } from '@/api/runtime'
import type { RuntimeStatus } from '@/api/runtime'

const route = useRoute()
const router = useRouter()
const profileStore = useProfileStore()

const id = computed(() => route.params.id as string)
const profile = computed(() => profileStore.profiles.find((p) => p.id === id.value))
const runtime = ref<RuntimeStatus | null>(null)

onMounted(async () => {
  if (!profileStore.loaded) await profileStore.fetchAll()
  try {
    runtime.value = await getRuntimeStatus(id.value)
  } catch {
    runtime.value = null
  }
})

async function confirmDelete() {
  if (!confirm(`删除 Profile "${profile.value?.display_name}"？此操作只删除元数据，不会删除任何磁盘文件。`)) return
  await profileStore.remove(id.value)
  router.push('/profiles')
}
</script>

<template>
  <div v-if="profile">
    <header class="page-head">
      <div>
        <h1>{{ profile.display_name }}</h1>
        <code class="id">{{ profile.id }}</code>
      </div>
      <div class="head-actions">
        <button class="btn danger ghost" @click="confirmDelete">删除</button>
      </div>
    </header>

    <section class="panel">
      <div class="panel-row">
        <span class="label">ConfDir</span>
        <code>{{ profile.confdir }}</code>
      </div>
      <div class="panel-row">
        <span class="label">运行时</span>
        <span>{{ profile.runtime_type }} · {{ profile.runtime_target }}</span>
      </div>
      <div class="panel-row" v-if="runtime">
        <span class="label">状态</span>
        <StatusBadge :state="runtime.state" />
        <span v-if="runtime.pid" class="muted">PID {{ runtime.pid }}</span>
      </div>
    </section>

    <section class="action-cards">
      <router-link :to="`/profiles/${profile.id}/config`" class="action-card">
        <div class="action-title">配置管理</div>
        <div class="action-desc">Schema 表单 + 原文编辑 · 无损保存 · 冲突检测</div>
      </router-link>
      <router-link :to="`/profiles/${profile.id}/runtime`" class="action-card">
        <div class="action-title">运行时</div>
        <div class="action-desc">服务启停 · 状态 · 日志</div>
      </router-link>
      <router-link :to="`/profiles/${profile.id}/synclist`" class="action-card">
        <div class="action-title">选择性同步</div>
        <div class="action-desc">SyncList 规则编辑 · 性能告警 · Resync 提醒</div>
      </router-link>
    </section>
  </div>
  <div v-else class="loading">加载中...</div>
</template>

<style scoped>
.page-head {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 20px;
}
.page-head h1 {
  margin: 0;
  font-size: 22px;
}
.id {
  color: #64748b;
  font-size: 13px;
}
.panel {
  background: #fff;
  border: 1px solid var(--oneweb-border);
  border-radius: 10px;
  padding: 18px 20px;
  margin-bottom: 20px;
}
.panel-row {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 8px 0;
  font-size: 14px;
}
.panel-row .label {
  width: 90px;
  color: #64748b;
  font-size: 13px;
}
.muted {
  color: #94a3b8;
}
.action-cards {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
  gap: 16px;
}
.action-card {
  background: #fff;
  border: 1px solid var(--oneweb-border);
  border-radius: 10px;
  padding: 20px;
  text-decoration: none;
  color: inherit;
  transition: box-shadow 0.15s;
}
.action-card:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
}
.action-title {
  font-weight: 600;
  margin-bottom: 6px;
}
.action-desc {
  color: #64748b;
  font-size: 13px;
}
.head-actions {
  display: flex;
  gap: 10px;
}
.btn {
  padding: 8px 14px;
  border-radius: 6px;
  border: none;
  font-size: 13px;
}
.btn.danger {
  background: #fee2e2;
  color: #b91c1c;
}
.loading {
  color: #64748b;
  padding: 40px;
  text-align: center;
}
</style>
