<script setup lang="ts">
import { computed } from 'vue'
import type { ConfigResponse, OptionSchema, UIWidget } from '@/api/config'

const props = defineProps<{
  config: ConfigResponse
  modelValue: Record<string, string>
}>()

const emit = defineEmits<{ (e: 'update:modelValue', v: Record<string, string>): void }>()

// 按 group 分组
const groups = computed(() => {
  const ui = props.config.ui
  const grouped: { id: string; label: string; options: OptionSchema[] }[] = []
  const order = new Map((ui?.groups || []).map((g) => [g.id, g]))
  const groupMap = new Map<string, OptionSchema[]>()
  for (const opt of props.config.schema.options) {
    const gid = opt.group || 'advanced'
    if (!groupMap.has(gid)) groupMap.set(gid, [])
    groupMap.get(gid)!.push(opt)
  }
  for (const [gid, opts] of groupMap) {
    const g = order.get(gid)
    grouped.push({
      id: gid,
      label: g?.label || gid,
      options: opts,
    })
  }
  grouped.sort((a, b) => {
    const oa = order.get(a.id)?.order ?? 99
    const ob = order.get(b.id)?.order ?? 99
    return oa - ob
  })
  return grouped
})

function widgetFor(opt: OptionSchema): UIWidget {
  return props.config.ui?.widgets?.[opt.key] || { widget: 'text-input' }
}

function setValue(key: string, value: string) {
  const next = { ...props.modelValue, [key]: value }
  emit('update:modelValue', next)
}

function removeKey(key: string) {
  const next = { ...props.modelValue }
  delete next[key]
  emit('update:modelValue', next)
}

function isDefault(key: string): boolean {
  return props.config.defaults[key] === props.modelValue[key]
}

function hasFileValue(key: string): boolean {
  return key in props.config.file_config
}
</script>

<template>
  <div class="schema-form">
    <section v-for="g in groups" :key="g.id" class="group">
      <h3 class="group-title">{{ g.label }}</h3>
      <div class="option-grid">
        <div
          v-for="opt in g.options"
          :key="opt.key"
          class="option"
          :class="{ 'has-file': hasFileValue(opt.key) }"
        >
          <div class="option-head">
            <code class="option-key">{{ opt.key }}</code>
            <span v-if="opt.deprecated" class="tag warn">deprecated</span>
            <span v-if="hasFileValue(opt.key)" class="tag">文件已设置</span>
            <span v-else-if="!isDefault(opt.key)" class="tag blue">非默认</span>
          </div>
          <p class="option-desc">{{ opt.description }}</p>

          <div class="option-control">
            <template v-if="widgetFor(opt).widget === 'switch'">
              <input
                type="checkbox"
                :checked="modelValue[opt.key] === 'true'"
                @change="setValue(opt.key, ($event.target as HTMLInputElement).checked ? 'true' : 'false')"
              />
            </template>
            <template v-else-if="widgetFor(opt).widget === 'slider'">
              <div class="slider-row">
                <input
                  type="range"
                  :min="opt.constraints?.min ?? 1"
                  :max="opt.constraints?.max ?? 32"
                  :value="Number(modelValue[opt.key] || opt.default || 4)"
                  @input="setValue(opt.key, ($event.target as HTMLInputElement).value)"
                />
                <input
                  class="num"
                  type="number"
                  :value="modelValue[opt.key] || ''"
                  @input="setValue(opt.key, ($event.target as HTMLInputElement).value)"
                />
              </div>
            </template>
            <template v-else-if="widgetFor(opt).widget === 'select'">
              <select :value="modelValue[opt.key] || ''" @change="setValue(opt.key, ($event.target as HTMLSelectElement).value)">
                <option v-for="v in opt.constraints?.allowed_values" :key="v" :value="v">{{ v }}</option>
              </select>
            </template>
            <template v-else>
              <input
                :type="widgetFor(opt).widget === 'password-input' ? 'password' : 'text'"
                :value="modelValue[opt.key] || ''"
                :placeholder="String(opt.default ?? '')"
                @input="setValue(opt.key, ($event.target as HTMLInputElement).value)"
              />
            </template>
            <button v-if="hasFileValue(opt.key)" class="unset" @click="removeKey(opt.key)" title="移除（转为注释）">✕</button>
          </div>
        </div>
      </div>
    </section>
  </div>
</template>

<style scoped>
.group {
  margin-bottom: 24px;
}
.group-title {
  font-size: 15px;
  font-weight: 600;
  border-bottom: 1px solid var(--oneweb-border);
  padding-bottom: 8px;
  margin-bottom: 14px;
}
.option-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: 14px;
}
.option {
  background: #fff;
  border: 1px solid var(--oneweb-border);
  border-radius: 8px;
  padding: 14px;
}
.option.has-file {
  border-left: 3px solid var(--oneweb-primary);
}
.option-head {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}
.option-key {
  font-weight: 600;
  font-size: 13px;
}
.tag {
  background: #f1f5f9;
  color: #64748b;
  border-radius: 12px;
  padding: 1px 8px;
  font-size: 11px;
}
.tag.warn {
  background: #fef3c7;
  color: #b45309;
}
.tag.blue {
  background: #e0f2fe;
  color: #0369a1;
}
.option-desc {
  color: #64748b;
  font-size: 12px;
  margin: 6px 0 10px;
}
.option-control {
  display: flex;
  align-items: center;
  gap: 10px;
}
.option-control input[type='text'],
.option-control input[type='number'],
.option-control input[type='password'],
.option-control select {
  flex: 1;
  padding: 6px 10px;
  border: 1px solid var(--oneweb-border);
  border-radius: 6px;
  font-size: 13px;
}
.slider-row {
  display: flex;
  align-items: center;
  gap: 10px;
  flex: 1;
}
.slider-row input[type='range'] {
  flex: 1;
}
.slider-row .num {
  width: 64px;
}
.unset {
  background: none;
  border: none;
  color: #94a3b8;
  font-size: 14px;
}
.unset:hover {
  color: #dc2626;
}
</style>
