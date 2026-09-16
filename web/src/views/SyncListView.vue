<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { getSyncList, saveSyncList } from '@/api/synclist'
import type { SyncListResponse } from '@/api/synclist'

const route = useRoute()
const id = computed(() => route.params.id as string)

const data = ref<SyncListResponse | null>(null)
const source = ref('')
const loading = ref(true)
const saving = ref(false)
const saved = ref(false)
const error = ref<string | null>(null)

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

onMounted(load)
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
    <div v-else-if="error" class="error-banner">{{ error }}</div>
    <template v-else-if="data">
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

      <textarea v-model="source" class="editor" spellcheck="false" placeholder="# 在此输入选择性同步规则，例如:&#10;/Documents/*&#10;!/Documents/temp/*"></textarea>

      <div class="rules-preview" v-if="data.rules && data.rules.length">
        <h4>规则模型 ({{ data.rules.length }} 行)</h4>
        <div v-for="(r, i) in data.rules" :key="i" class="rule-row" :class="r.type">
          <span class="ln">{{ r.line_number }}</span>
          <code>{{ r.raw_text }}</code>
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
</style>
