<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useProfileStore } from '@/stores/profile'

const profileStore = useProfileStore()
const router = useRouter()
const error = ref<string | null>(null)
const creating = ref(false)
const showCreateModal = ref(false)
const discovering = ref(false)

const form = reactive({
  id: '',
  display_name: '',
  confdir: '',
  runtime_type: 'systemd' as any,
  runtime_target: '',
})

function onIdInput() {
  if (!form.id) return
  const cleanId = form.id.trim().toLowerCase()
  if (!form.display_name || form.display_name.startsWith('OneDrive (')) {
    form.display_name = `OneDrive (${cleanId})`
  }
  if (!form.confdir || form.confdir.startsWith('~/.config/onedrive')) {
    form.confdir = cleanId === 'default' ? '~/.config/onedrive' : `~/.config/onedrive-${cleanId}`
  }
  if (!form.runtime_target || form.runtime_target.startsWith('onedrive')) {
    form.runtime_target = cleanId === 'default' ? 'onedrive.service' : `onedrive@${cleanId}.service`
  }
}

async function triggerDiscover() {
  discovering.value = true
  try {
    await profileStore.discover()
  } finally {
    discovering.value = false
  }
}

async function submit() {
  error.value = null
  if (!form.id || !form.confdir) {
    error.value = 'ID 与 confdir 为必填项'
    return
  }
  creating.value = true
  try {
    await profileStore.create({
      id: form.id.trim(),
      display_name: form.display_name.trim() || form.id.trim(),
      confdir: form.confdir.trim(),
      runtime_type: form.runtime_type,
      runtime_target: form.runtime_target.trim() || `onedrive@${form.id.trim()}.service`,
    })
    showCreateModal.value = false
    // 重置表单
    form.id = ''
    form.display_name = ''
    form.confdir = ''
    form.runtime_target = ''
  } catch (e: any) {
    error.value = e.message
  } finally {
    creating.value = false
  }
}

async function confirmDelete(id: string, name: string) {
  if (!confirm(`确认移除账号 "${name}"？此操作仅移除管理绑定，绝不删除本地文件或云端数据。`)) return
  await profileStore.remove(id)
}

onMounted(() => {
  profileStore.fetchAll()
})
</script>

<template>
  <div>
    <header class="page-head">
      <div>
        <h1>账号与配置管理 (Profiles)</h1>
        <span class="subtitle">管理本地绑定的 OneDrive 账号与服务单元</span>
      </div>
      <div class="head-actions">
        <button class="btn secondary" :disabled="discovering" @click="triggerDiscover">
          {{ discovering ? '扫描中...' : '🔍 重新扫描宿主机账号' }}
        </button>
        <button class="btn primary" @click="showCreateModal = true">
          + 添加新账号
        </button>
      </div>
    </header>

    <!-- 账号列表 -->
    <section class="profiles-container">
      <div v-if="profileStore.profiles.length === 0" class="empty-state">
        <p>暂无已绑定的 OneDrive 账号</p>
        <button class="btn primary" @click="triggerDiscover">立即自动扫描</button>
      </div>

      <div v-for="p in profileStore.profiles" :key="p.id" class="profile-item-card">
        <div class="item-main">
          <div class="item-title-row">
            <h3>{{ p.display_name || p.id }}</h3>
            <span class="item-id-badge">{{ p.id }}</span>
            <span v-if="p.id === 'default'" class="default-badge">默认</span>
          </div>
          <div class="item-meta">
            <span class="meta-field"><span class="label">配置目录:</span> <code>{{ p.confdir }}</code></span>
            <span class="meta-field"><span class="label">服务:</span> <code>{{ p.runtime_target }}</code></span>
            <span class="meta-field"><span class="label">类型:</span> {{ p.runtime_type }}</span>
          </div>
        </div>

        <div class="item-actions">
          <button class="btn secondary" @click="router.push(`/profiles/${p.id}`)">
            进入控制台
          </button>
          <button
            v-if="profileStore.profiles.length > 1"
            class="btn danger-ghost"
            @click="confirmDelete(p.id, p.display_name || p.id)"
          >
            解绑
          </button>
        </div>
      </div>
    </section>

    <!-- 添加新账号 Modal -->
    <div v-if="showCreateModal" class="modal-overlay" @click.self="showCreateModal = false">
      <div class="modal">
        <h2>添加 OneDrive 账号</h2>
        <p class="modal-desc">为新的 OneDrive 实例分配独立的配置目录与服务单元。</p>

        <div class="form">
          <label>
            账号唯一标识 (ID) <span class="req">*</span>
            <input
              v-model="form.id"
              placeholder="例如: work 或 personal"
              @input="onIdInput"
            />
          </label>
          <label>
            显示名称
            <input v-model="form.display_name" placeholder="例如: 工作账号" />
          </label>
          <label>
            配置目录 (ConfDir) <span class="req">*</span>
            <input v-model="form.confdir" placeholder="例如: ~/.config/onedrive-work" />
          </label>
          <label>
            运行时类型
            <select v-model="form.runtime_type">
              <option value="systemd">systemd (Native)</option>
              <option value="docker">Docker</option>
              <option value="podman">Podman</option>
            </select>
          </label>
          <label>
            Systemd 单元 (Runtime Target)
            <input v-model="form.runtime_target" placeholder="例如: onedrive@work.service" />
          </label>
        </div>

        <div v-if="error" class="error">{{ error }}</div>
        <div class="modal-actions">
          <button class="btn secondary" @click="showCreateModal = false">取消</button>
          <button class="btn primary" :disabled="creating" @click="submit">
            {{ creating ? '添加中...' : '确定添加' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.page-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}
.page-head h1 {
  font-size: 22px;
  margin: 0;
}
.subtitle {
  color: #64748b;
  font-size: 13px;
}
.head-actions {
  display: flex;
  gap: 12px;
}
.profiles-container {
  display: flex;
  flex-direction: column;
  gap: 14px;
}
.profile-item-card {
  background: #fff;
  border: 1px solid var(--oneweb-border);
  border-radius: 10px;
  padding: 20px 24px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 16px;
  transition: box-shadow 0.15s ease;
}
.profile-item-card:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.05);
}
.item-title-row {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 8px;
}
.item-title-row h3 {
  margin: 0;
  font-size: 16px;
}
.item-id-badge {
  background: #f1f5f9;
  color: #64748b;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
}
.default-badge {
  background: #dbeafe;
  color: #1d4ed8;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 500;
}
.item-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 20px;
  font-size: 13px;
  color: #64748b;
}
.meta-field .label {
  color: #94a3b8;
  margin-right: 4px;
}
.meta-field code {
  background: #f8fafc;
  padding: 2px 6px;
  border-radius: 4px;
}
.item-actions {
  display: flex;
  align-items: center;
  gap: 10px;
}
.empty-state {
  text-align: center;
  padding: 48px;
  background: #fff;
  border-radius: 10px;
  border: 1px dashed #cbd5e1;
  color: #64748b;
}

/* 模态弹窗 */
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(15, 23, 42, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 50;
}
.modal {
  background: #fff;
  border-radius: 12px;
  padding: 28px;
  width: 440px;
  max-width: 90vw;
}
.modal h2 {
  margin: 0 0 6px;
  font-size: 18px;
}
.modal-desc {
  color: #64748b;
  font-size: 13px;
  margin: 0 0 18px;
}
.form {
  display: flex;
  flex-direction: column;
  gap: 14px;
}
.form label {
  display: flex;
  flex-direction: column;
  gap: 6px;
  font-size: 13px;
  color: #475569;
}
.form input,
.form select {
  padding: 8px 12px;
  border: 1px solid var(--oneweb-border);
  border-radius: 6px;
  font-size: 14px;
}
.req {
  color: #dc2626;
}
.error {
  margin-top: 12px;
  color: #dc2626;
  font-size: 13px;
}
.modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  margin-top: 24px;
}

/* 按钮 */
.btn {
  padding: 8px 16px;
  border-radius: 6px;
  border: 1px solid transparent;
  font-size: 13px;
  cursor: pointer;
  font-weight: 500;
  transition: all 0.15s ease;
}
.btn.primary {
  background: var(--oneweb-primary);
  color: #fff;
}
.btn.primary:hover {
  filter: brightness(0.95);
}
.btn.secondary {
  background: #f1f5f9;
  color: #334155;
  border-color: #cbd5e1;
}
.btn.secondary:hover {
  background: #e2e8f0;
}
.btn.danger-ghost {
  background: transparent;
  color: #ef4444;
}
.btn.danger-ghost:hover {
  background: #fee2e2;
}
</style>
