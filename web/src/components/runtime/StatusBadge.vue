<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{ state: string }>()

const stateClass: Record<string, string> = {
  running: 'ok',
  starting: 'pending',
  activating: 'pending',
  stopped: 'stopped',
  failed: 'failed',
  inactive: 'stopped',
}

const stateLabel: Record<string, string> = {
  running: '运行中',
  starting: '启动中...',
  activating: '启动中...',
  stopped: '已停止',
  failed: '异常失败',
  inactive: '未运行',
}

const label = computed(() => stateLabel[props.state] || props.state)
</script>

<template>
  <span class="badge" :class="stateClass[state] || 'stopped'">{{ label }}</span>
</template>

<style scoped>
.badge {
  border-radius: 20px;
  padding: 2px 10px;
  font-size: 12px;
  text-transform: uppercase;
  letter-spacing: 0.02em;
}
.ok {
  background: #dcfce7;
  color: #15803d;
}
.stopped {
  background: #f1f5f9;
  color: #64748b;
}
.failed {
  background: #fee2e2;
  color: #b91c1c;
}
.pending {
  background: #fef3c7;
  color: #b45309;
}
</style>
