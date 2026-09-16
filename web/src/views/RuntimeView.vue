<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import StatusBadge from '@/components/runtime/StatusBadge.vue'
import LogViewer from '@/components/runtime/LogViewer.vue'
import { getRuntimeLogs, getRuntimeStatus, runtimeAction } from '@/api/runtime'
import type { LogEntry, RuntimeStatus } from '@/api/runtime'

const route = useRoute()
const id = computed(() => route.params.id as string)

const status = ref<RuntimeStatus | null>(null)
const logs = ref<LogEntry[]>([])
const error = ref<string | null>(null)
const acting = ref(false)

let pollTimer: number | undefined

async function refresh() {
  try {
    status.value = await getRuntimeStatus(id.value)
  } catch (e: any) {
    error.value = e.message
  }
}

async function loadLogs() {
  try {
    const fetched = await getRuntimeLogs(id.value, 200)
    if (fetched && Array.isArray(fetched)) {
      logs.value = fetched
    }
  } catch (e: any) {
    console.error('loadLogs failed:', e)
  }
}

async function act(action: 'start' | 'stop' | 'restart') {
  acting.value = true
  error.value = null
  try {
    status.value = await runtimeAction(id.value, action)
    setTimeout(loadLogs, 1000)
  } catch (e: any) {
    error.value = e.message
  } finally {
    acting.value = false
  }
}

onMounted(() => {
  refresh()
  loadLogs()
  pollTimer = window.setInterval(() => {
    refresh()
    loadLogs()
  }, 3000)
})

onUnmounted(() => {
  if (pollTimer) clearInterval(pollTimer)
})
</script>

<template>
  <div>
    <header class="page-head">
      <h1>运行时</h1>
      <div v-if="status" class="status-line">
        <StatusBadge :state="status.state" />
        <span v-if="status.pid" class="muted">PID {{ status.pid }}</span>
        <span v-if="status.memory_bytes" class="muted">
          {{ (status.memory_bytes / 1024 / 1024).toFixed(1) }} MB
        </span>
      </div>
      <div class="actions">
        <button class="btn ghost" :disabled="acting || status?.state === 'running' || status?.state === 'starting'" @click="act('start')">
          {{ status?.state === 'starting' ? '启动中...' : '启动' }}
        </button>
        <button class="btn ghost" :disabled="acting || status?.state === 'stopped'" @click="act('stop')">停止</button>
        <button class="btn ghost" :disabled="acting" @click="act('restart')">重启</button>
        <button class="btn ghost" @click="loadLogs">刷新日志</button>
      </div>
    </header>

    <div v-if="error" class="error-banner">{{ error }}</div>

    <LogViewer :entries="logs" />
  </div>
</template>

<style scoped>
.page-head {
  display: flex;
  align-items: center;
  gap: 16px;
  margin-bottom: 20px;
  flex-wrap: wrap;
}
.page-head h1 {
  margin: 0;
  font-size: 22px;
}
.status-line {
  display: flex;
  align-items: center;
  gap: 10px;
}
.muted {
  color: #64748b;
  font-size: 13px;
}
.actions {
  margin-left: auto;
  display: flex;
  gap: 8px;
}
.btn {
  padding: 8px 14px;
  border-radius: 6px;
  border: 1px solid var(--oneweb-border);
  background: #fff;
  font-size: 13px;
}
.btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
.error-banner {
  background: #fee2e2;
  color: #b91c1c;
  padding: 10px 14px;
  border-radius: 8px;
  margin-bottom: 14px;
  font-size: 13px;
}
</style>
