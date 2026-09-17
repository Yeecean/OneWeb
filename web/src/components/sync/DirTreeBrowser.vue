<script setup lang="ts">
import { computed, reactive } from 'vue'
import type { DirNode } from '@/api/synclist'
import DirTreeNode from './DirTreeNode.vue'

const props = defineProps<{
  nodes: DirNode[]
  loading: boolean
}>()

const emit = defineEmits<{ (e: 'add-rule', rule: string): void }>()

const expanded = reactive(new Set<string>())

function toggle(path: string) {
  if (expanded.has(path)) expanded.delete(path)
  else expanded.add(path)
}

function addInclude(node: DirNode) {
  emit('add-rule', node.path)
}

function addExclude(node: DirNode) {
  emit('add-rule', `!${node.path}`)
}

const hasNodes = computed(() => props.nodes.length > 0)
</script>

<template>
  <div class="tree-browser">
    <div v-if="loading" class="tree-loading">扫描本地目录中...</div>
    <template v-else>
      <div v-if="hasNodes" class="tree-hint">
        点击 <kbd>+ 包含</kbd> 或 <kbd>− 排除</kbd> 将文件夹快速添加到规则
      </div>
      <DirTreeNode
        v-for="node in nodes"
        :key="node.path"
        :node="node"
        :expanded="expanded"
        @toggle="toggle"
        @add-include="addInclude"
        @add-exclude="addExclude"
      />
      <div v-if="!hasNodes" class="tree-empty">
        未发现本地 OneDrive 目录，服务运行并同步后将自动出现
      </div>
    </template>
  </div>
</template>

<style scoped>
.tree-browser {
  font-size: 13px;
}
.tree-loading,
.tree-empty {
  color: #94a3b8;
  padding: 16px 4px;
  font-size: 13px;
}
.tree-hint {
  color: #64748b;
  font-size: 12px;
  margin-bottom: 10px;
}
.tree-hint kbd {
  background: #f1f5f9;
  border: 1px solid #e2e8f0;
  border-radius: 4px;
  padding: 1px 5px;
  font-size: 11px;
  font-family: inherit;
}
</style>
