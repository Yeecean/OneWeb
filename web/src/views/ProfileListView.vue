<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useProfileStore } from '@/stores/profile'

const profileStore = useProfileStore()
const router = useRouter()
const error = ref<string | null>(null)
const creating = ref(false)

const form = reactive({
  id: '',
  display_name: '',
  confdir: '',
  runtime_type: 'systemd' as any,
  runtime_target: '',
})

async function submit() {
  error.value = null
  if (!form.id || !form.confdir) {
    error.value = 'ID 与 confdir 为必填项'
    return
  }
  creating.value = true
  try {
    await profileStore.create({
      id: form.id,
      display_name: form.display_name || form.id,
      confdir: form.confdir,
      runtime_type: form.runtime_type,
      runtime_target: form.runtime_target || `onedrive@${form.id}.service`,
    })
    router.push('/profiles')
  } catch (e: any) {
    error.value = e.message
  } finally {
    creating.value = false
  }
}
</script>

<template>
  <div class="modal-overlay" @click.self="router.push('/profiles')">
    <div class="modal">
      <h2>创建 Profile</h2>
      <div class="form">
        <label>
          ID <span class="req">*</span>
          <input v-model="form.id" placeholder="例如: default" />
        </label>
        <label>
          显示名称
          <input v-model="form.display_name" placeholder="例如: 个人 OneDrive" />
        </label>
        <label>
          ConfDir <span class="req">*</span>
          <input v-model="form.confdir" placeholder="例如: ~/.config/onedrive" />
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
          Runtime Target
          <input v-model="form.runtime_target" placeholder="例如: onedrive@default.service" />
        </label>
      </div>
      <div v-if="error" class="error">{{ error }}</div>
      <div class="actions">
        <button class="btn ghost" @click="router.push('/profiles')">取消</button>
        <button class="btn primary" :disabled="creating" @click="submit">
          {{ creating ? '创建中...' : '创建' }}
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
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
  width: 420px;
  max-width: 90vw;
}
.modal h2 {
  margin: 0 0 20px;
  font-size: 18px;
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
  padding: 8px 10px;
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
.actions {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  margin-top: 20px;
}
.btn {
  padding: 8px 16px;
  border-radius: 6px;
  border: none;
  font-size: 14px;
}
.btn.primary {
  background: var(--oneweb-primary);
  color: #fff;
}
.btn.ghost {
  background: #f1f5f9;
  color: #475569;
}
</style>
