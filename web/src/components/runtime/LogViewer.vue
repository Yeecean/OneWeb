<script setup lang="ts">
import { nextTick, ref, watch } from 'vue'
import type { LogEntry } from '@/api/runtime'

const props = defineProps<{ entries: LogEntry[] }>()
const container = ref<HTMLDivElement | null>(null)

watch(
  () => props.entries.length,
  async () => {
    await nextTick()
    if (container.value) container.value.scrollTop = container.value.scrollHeight
  },
)
</script>

<template>
  <div ref="container" class="log-viewer">
    <div v-if="entries.length === 0" class="empty">暂无日志</div>
    <div v-for="(e, i) in entries" :key="i" class="log-line" :class="e.stream">
      <span class="time">{{ e.timestamp ? new Date(e.timestamp).toLocaleTimeString() : '' }}</span>
      <span class="level" v-if="e.level">{{ e.level }}</span>
      <span class="msg">{{ e.message }}</span>
    </div>
  </div>
</template>

<style scoped>
.log-viewer {
  background: #0f172a;
  color: #e2e8f0;
  border-radius: 8px;
  padding: 14px;
  height: 50vh;
  overflow-y: auto;
  font-family: 'JetBrains Mono', monospace;
  font-size: 12px;
  line-height: 1.6;
}
.empty {
  color: #64748b;
}
.log-line {
  display: flex;
  gap: 10px;
  white-space: pre-wrap;
  word-break: break-all;
}
.log-line.stderr .msg {
  color: #fca5a5;
}
.time {
  color: #64748b;
  flex-shrink: 0;
}
.level {
  color: #38bdf8;
  flex-shrink: 0;
  width: 52px;
}
.msg {
  flex: 1;
}
</style>
