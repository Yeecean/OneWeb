<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { getSyncList, saveSyncList, getSyncListTree } from '@/api/synclist'
import type { SyncListResponse, DirNode } from '@/api/synclist'
import DirTreeBrowser from '@/components/sync/DirTreeBrowser.vue'

const route = useRoute()
const id = computed(() => route.params.id as string)

const data = ref<SyncListResponse | null>(null)
const source = ref('')
const loading = ref(true)
const saving = ref(false)
const saved = ref(false)
const error = ref<string | null>(null)

const treeNodes = ref<DirNode[]>([])
const treeLoading = ref(false)

const baseSha = computed(() => data.value?.version_meta?.sha256 || '')

async function load() {
  loading.value = true
  error.value = null
  try {
    data.value = await getSyncList(id.value)
    source.value = data.value.source || ''
  } catch (e: any) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}

async function loadTree() {
  treeLoading.value = true
  try {
    treeNodes.value = await getSyncListTree(id.value, 3)
  } catch (e) {
    // 静默失败，树面板为空
    treeNodes.value = []
  } finally {
    treeLoading.value = false
  }
}

// appendRule 将新规则追加到编辑器，避免重复
function appendRule(rule: string) {
  const existing = source.value.split('\n').map((l) => l.trim())
  if (existing.includes(rule)) return
  source.value = source.value ? source.value.replace(/\s*$/, '') + '\n' + rule : rule
}

async function doSave() {
  saving.value = true
  saved.value = false
  error.value = null
  try {
    data.value = await saveSyncList(id.value, {
      source: source.value,
      base_sha256: baseSha.value,
    })
    saved.value = true
    setTimeout(() => (saved.value = false), 4000)
  } catch (e: any) {
    error.value = e.message
  } finally {
    saving.value = false
  }
}

onMounted(async () => {
  await load()
  loadTree()
})
</script>

<template>
  <div>
    <header class="page-head">
      <h1>选择性同步 (SyncList)</h1>
      <button class="btn primary" :disabled="saving" @click="doSave">
        {{ saving ? '保存中...' : '保存规则' }}
      </button>
    </header>

    <div v-if="loading" class="state">加载中...</div>
    <div v-else-if="error && !data" class="error-banner">{{ error }}</div>
    <template v-else-if="data">
      <div class="rule-legend">
        <span class="legend-item include"><code>/路径</code> — 仅同步此路径</span>
        <span class="legend-item exclude"><code>!/路径</code> — 排除此路径</span>
        <span class="legend-item comment"><code>#注释</code> — 注释行，不生效</span>
      </div>

      <div v-if="saved" class="banner ok">已保存</div>
      <div v-if="data.resync_required" class="banner warn">
        ⚠️ 修改同步规则后必须执行全量重同步 (--resync) 才能生效
      </div>

      <div class="warnings" v-if="data.warnings && data.warnings.length">
        <div v-for="(w, i) in data.warnings" :key="i" class="warning" :class="w.level">
          <span class="w-badge">{{ w.level }}</span>
          L{{ w.line }} · {{ w.rule }}
          <span class="w-msg">{{ w.message }}</span>
        </div>
      </div>

      <div class="split-layout">
        <div class="editor-pane">
          <h3 class="pane-title">📝 规则编辑器</h3>
          <textarea
            v-model="source"
            class="editor"
            spellcheck="false"
            placeholder="# 在此输入选择性同步规则，例如:&#10;/Documents/*&#10;!/Documents/temp/*"
          ></textarea>

          <div class="rules-preview" v-if="data.rules && data.rules.length">
            <h4>规则模型 ({{ data.rules.length }} 行)</h4>
            <div v-for="(r, i) in data.rules" :key="i" class="rule-row" :class="r.type">
              <span class="ln">{{ r.line_number }}</span>
              <code>{{ r.raw_text }}</code>
            </div>
          </div>
        </div>

        <div class="tree-pane">
          <div class="pane-header">
            <h3 class="pane-title">📂 本地目录浏览</h3>
            <button class="btn-icon" title="刷新目录树" @click="loadTree">🔄</button>
          </div>
          <p class="pane-desc">浏览已同步的本地目录，点击按钮快速生成规则</p>
          <DirTreeBrowser :nodes="treeNodes" :loading="treeLoading" @add-rule="appendRule" />
        </div>
      </div>
    </template>
  </div>
</template>

<style scoped>
.page-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}
.page-head h1 {
  margin: 0;
  font-size: 22px;
}
.state {
  padding: 40px;
  text-align: center;
  color: #64748b;
}
.error-banner,
.banner {
  padding: 10px 14px;
  border-radius: 8px;
  margin-bottom: 14px;
  font-size: 13px;
}
.error-banner,
.banner.error {
  background: #fee2e2;
  color: #b91c1c;
}
.banner.ok {
  background: #dcfce7;
  color: #15803d;
}
.banner.warn {
  background: #fef3c7;
  color: #b45309;
}
.warnings {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin-bottom: 14px;
}
.warning {
  padding: 8px 12px;
  border-radius: 6px;
  font-size: 13px;
  border: 1px solid;
}
.warning.performance {
  background: #fffbeb;
  border-color: #fde68a;
  color: #92400e;
}
.warning.ordering {
  background: #fdf4ff;
  border-color: #e9d5ff;
  color: #6b21a8;
}
.w-badge {
  display: inline-block;
  background: inherit;
  font-weight: 600;
  margin-right: 8px;
  text-transform: uppercase;
  font-size: 11px;
}
.w-msg {
  color: #64748b;
  margin-left: 8px;
}
.editor {
  width: 100%;
  min-height: 40vh;
  font-family: 'JetBrains Mono', monospace;
  font-size: 13px;
  padding: 16px;
  border: 1px solid var(--oneweb-border);
  border-radius: 8px;
  resize: vertical;
  line-height: 1.6;
  background: #fff;
}
.rules-preview {
  margin-top: 20px;
  background: #fff;
  border: 1px solid var(--oneweb-border);
  border-radius: 8px;
  padding: 14px;
}
.rules-preview h4 {
  margin: 0 0 10px;
  font-size: 14px;
}
.rule-row {
  display: flex;
  gap: 12px;
  padding: 3px 0;
  font-size: 13px;
}
.rule-row.exclude code {
  color: #dc2626;
}
.rule-row.include code {
  color: #15803d;
}
.rule-row.comment code,
.rule-row.blank code {
  color: #94a3b8;
}
.ln {
  color: #94a3b8;
  width: 24px;
  text-align: right;
  font-size: 12px;
}
.btn {
  padding: 10px 20px;
  border-radius: 6px;
  border: none;
  font-size: 14px;
}
.btn.primary {
  background: var(--oneweb-primary);
  color: #fff;
}
.btn.primary:disabled {
  opacity: 0.6;
}
.rule-legend {
  display: flex;
  flex-wrap: wrap;
  gap: 16px;
  margin-bottom: 16px;
  padding: 10px 14px;
  background: #f8fafc;
  border: 1px solid var(--oneweb-border);
  border-radius: 8px;
  font-size: 12px;
  color: #64748b;
}
.legend-item {
  display: inline-flex;
  align-items: center;
  gap: 6px;
}
.legend-item code {
  border-radius: 4px;
  padding: 1px 6px;
  font-family: 'JetBrains Mono', monospace;
}
.legend-item.include code {
  background: #dcfce7;
  color: #15803d;
}
.legend-item.exclude code {
  background: #fee2e2;
  color: #b91c1c;
}
.legend-item.comment code {
  background: #f1f5f9;
  color: #94a3b8;
}
.split-layout {
  display: grid;
  grid-template-columns: 3fr 2fr;
  gap: 20px;
  align-items: start;
}
@media (max-width: 900px) {
  .split-layout {
    grid-template-columns: 1fr;
  }
}
.editor-pane,
.tree-pane {
  background: #fff;
  border: 1px solid var(--oneweb-border);
  border-radius: 10px;
  padding: 16px;
}
.pane-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.pane-title {
  font-size: 14px;
  font-weight: 600;
  margin: 0 0 12px;
}
.pane-desc {
  color: #94a3b8;
  font-size: 12px;
  margin: 0 0 12px;
}
.btn-icon {
  border: 1px solid var(--oneweb-border);
  background: #fff;
  border-radius: 6px;
  padding: 3px 8px;
  font-size: 12px;
  cursor: pointer;
}
.btn-icon:hover {
  background: #f8fafc;
}
</style>
