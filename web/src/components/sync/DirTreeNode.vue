<script setup lang="ts">
import { computed } from 'vue'
import type { DirNode } from '@/api/synclist'

const props = defineProps<{
  node: DirNode
  expanded: Set<string>
  depth?: number
}>()

const emit = defineEmits<{
  (e: 'toggle', path: string): void
  (e: 'add-include', node: DirNode): void
  (e: 'add-exclude', node: DirNode): void
}>()

const isOpen = computed(() => props.expanded.has(props.node.path))
const hasChildren = computed(() => (props.node.children?.length ?? 0) > 0)
const indent = computed(() => (props.depth ?? 0) * 16)
</script>

<template>
  <div class="tree-node">
    <div class="node-row" :style="{ paddingLeft: `${indent}px` }">
      <button v-if="hasChildren" class="expand-btn" @click="emit('toggle', node.path)">
        {{ isOpen ? '▾' : '▸' }}
      </button>
      <span v-else class="expand-placeholder"></span>
      <span class="folder-icon">📁</span>
      <span class="node-name">{{ node.name }}</span>
      <span class="node-path">{{ node.path }}</span>
      <div class="node-actions">
        <button class="action-btn include" title="添加为同步包含规则" @click="emit('add-include', node)">
          + 包含
        </button>
        <button class="action-btn exclude" title="添加为排除规则" @click="emit('add-exclude', node)">
          − 排除
        </button>
      </div>
    </div>
    <template v-if="isOpen && hasChildren">
      <DirTreeNode
        v-for="child in node.children"
        :key="child.path"
        :node="child"
        :expanded="expanded"
        :depth="(depth ?? 0) + 1"
        @toggle="emit('toggle', $event)"
        @add-include="emit('add-include', $event)"
        @add-exclude="emit('add-exclude', $event)"
      />
    </template>
  </div>
</template>

<style scoped>
.node-row {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 3px 4px;
  border-radius: 5px;
}
.node-row:hover {
  background: #f8fafc;
}
.node-row:hover .node-actions {
  visibility: visible;
}
.expand-btn {
  border: none;
  background: none;
  width: 16px;
  color: #64748b;
  font-size: 11px;
  padding: 0;
}
.expand-placeholder {
  width: 16px;
}
.folder-icon {
  font-size: 12px;
}
.node-name {
  font-weight: 500;
  color: #334155;
}
.node-path {
  color: #94a3b8;
  font-size: 11px;
  font-family: 'JetBrains Mono', monospace;
}
.node-actions {
  margin-left: auto;
  display: flex;
  gap: 4px;
  visibility: hidden;
}
.action-btn {
  border: none;
  border-radius: 4px;
  padding: 2px 8px;
  font-size: 11px;
  cursor: pointer;
}
.action-btn.include {
  background: #dcfce7;
  color: #15803d;
}
.action-btn.include:hover {
  background: #bbf7d0;
}
.action-btn.exclude {
  background: #fee2e2;
  color: #b91c1c;
}
.action-btn.exclude:hover {
  background: #fecaca;
}
</style>
