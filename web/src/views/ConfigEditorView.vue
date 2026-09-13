<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useConfigStore } from '@/stores/config'
import SchemaForm from '@/components/config/SchemaForm.vue'
import type { SaveConfigRequest } from '@/api/config'

const route = useRoute()
const configStore = useConfigStore()
const id = computed(() => route.params.id as string)

const mode = ref<'schema' | 'raw'>('schema')
const model = reactive<Record<string, string>>({})
const rawText = ref('')
const saving = ref(false)
const saved = ref(false)

const baseSha = computed(() => configStore.config?.version_meta?.sha256 || '')

function buildModel() {
  const c = configStore.config
  if (!c) return
  // 仅展示文件已设置 + 有默认值的项，未设置的留空以便填写
  const keys = new Set<string>([
    ...Object.keys(c.file_config),
    ...Object.keys(c.defaults),
  ])
  for (const k of keys) {
    model[k] = c.file_config[k] ?? ''
  }
}

function buildRaw() {
  const c = configStore.config
  if (!c) return
  rawText.value = Object.entries(c.file_config)
    .map(([k, v]) => `${k} = "${v}"`)
    .join('\n')
}

watch(
  () => configStore.config,
  (c) => {
    if (c) {
      buildModel()
      buildRaw()
    }
  },
)

async function doSave() {
  saving.value = true
  saved.value = false
  const req: SaveConfigRequest = {
    desired: {},
    base_sha256: baseSha.value,
  }
  if (mode.value === 'schema') {
    // 仅提交有值的项
    for (const [k, v] of Object.entries(model)) {
      if (v !== '' && v !== undefined) req.desired[k] = v
    }
  } else {
    // raw 模式：将每行解析为 desired
    for (const line of rawText.value.split('\n')) {
      const m = line.match(/^\s*([\w_]+)\s*=\s*"?([^"#]*)"?/)
      if (m) req.desired[m[1]] = m[2].trim()
    }
  }
  const ok = await configStore.save(id.value, req)
  if (ok) {
    saved.value = true
    setTimeout(() => (saved.value = false), 3000)
  }
}

onMounted(() => configStore.load(id.value))
</script>

<template>
  <div>
    <header class="page-head">
      <h1>配置管理</h1>
      <div class="tabs">
        <button :class="{ active: mode === 'schema' }" @click="mode = 'schema'">Schema 表单</button>
        <button :class="{ active: mode === 'raw' }" @click="mode = 'raw'">原文编辑</button>
      </div>
    </header>

    <div v-if="configStore.loading" class="state">加载中...</div>
    <div v-else-if="configStore.lastError && !configStore.config" class="state error">
      {{ configStore.lastError }}
    </div>

    <template v-else-if="configStore.config">
      <div v-if="configStore.conflict" class="banner error">
        检测到配置已被外部修改 (409 Conflict)，请重新加载后合并再保存。
      </div>
      <div v-if="saved" class="banner ok">已保存，文件无损写入 ✓</div>

      <div v-if="mode === 'schema'">
        <SchemaForm
          :config="configStore.config"
          v-model="model"
        />
      </div>
      <div v-else>
        <textarea v-model="rawText" class="raw-editor" spellcheck="false"></textarea>
      </div>

      <div class="save-bar">
        <span v-if="configStore.config.file_exists" class="meta">
          指纹 {{ baseSha.slice(0, 12) }}… · 最近修改 {{ new Date(configStore.config.version_meta?.mtime || '').toLocaleString() }}
        </span>
        <span v-else class="meta">配置文件尚不存在，保存将创建</span>
        <button class="btn primary" :disabled="saving" @click="doSave">
          {{ saving ? '保存中...' : '保存配置' }}
        </button>
      </div>
    </template>
  </div>
</template>

<style scoped>
.page-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}
.page-head h1 {
  margin: 0;
  font-size: 22px;
}
.tabs {
  display: flex;
  gap: 4px;
  background: #f1f5f9;
  border-radius: 8px;
  padding: 3px;
}
.tabs button {
  border: none;
  background: transparent;
  padding: 6px 14px;
  border-radius: 6px;
  font-size: 13px;
  color: #64748b;
}
.tabs button.active {
  background: #fff;
  color: #1e293b;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}
.state {
  padding: 40px;
  text-align: center;
  color: #64748b;
}
.state.error {
  color: #dc2626;
}
.banner {
  padding: 10px 14px;
  border-radius: 8px;
  margin-bottom: 16px;
  font-size: 14px;
}
.banner.ok {
  background: #dcfce7;
  color: #15803d;
}
.banner.error {
  background: #fee2e2;
  color: #b91c1c;
}
.raw-editor {
  width: 100%;
  min-height: 60vh;
  font-family: 'JetBrains Mono', monospace;
  font-size: 13px;
  padding: 16px;
  border: 1px solid var(--oneweb-border);
  border-radius: 8px;
  resize: vertical;
  line-height: 1.6;
  background: #fff;
}
.save-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 20px;
  padding: 16px 0;
  border-top: 1px solid var(--oneweb-border);
}
.meta {
  color: #64748b;
  font-size: 13px;
}
.btn {
  padding: 10px 22px;
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
