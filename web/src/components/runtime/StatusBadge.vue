<script setup lang="ts">
defineProps<{ state: string }>()

const stateClass: Record<string, string> = {
  running: 'ok',
  stopped: 'stopped',
  failed: 'failed',
  inactive: 'stopped',
  activating: 'pending',
}
</script>

<template>
  <span class="badge" :class="stateClass[state] || 'stopped'">{{ state }}</span>
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
