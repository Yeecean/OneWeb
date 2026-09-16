<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useProfileStore } from '@/stores/profile'
import StatusBadge from '@/components/runtime/StatusBadge.vue'
import { getRuntimeStatus, runtimeAction } from '@/api/runtime'
import type { RuntimeStatus } from '@/api/runtime'

const profileStore = useProfileStore()
const router = useRouter()
const systemInfo = ref<any>(null)

const activeProfile = computed(() => {
  if (profileStore.profiles.length === 1) {
    return profileStore.profiles[0]
  }
  return null
})

const runtime = ref<RuntimeStatus | null>(null)
const acting = ref(false)
const actError = ref<string | null>(null)
let pollTimer: number | undefined

async function fetchStatus() {
  if (!activeProfile.value) return
  try {
    runtime.value = await getRuntimeStatus(activeProfile.value.id)
  } catch {
    runtime.value = null
  }
}

async function handleAction(action: 'start' | 'stop' | 'restart') {
  if (!activeProfile.value) return
  acting.value = true
  actError.value = null
  try {
    runtime.value = await runtimeAction(activeProfile.value.id, action)
  } catch (e: any) {
    actError.value = e.message
  } finally {
    acting.value = false
  }
}

watch(
  () => activeProfile.value?.id,
  (newId) => {
    if (newId) fetchStatus()
  },
  { immediate: true },
)

onMounted(async () => {
  await profileStore.fetchAll()
  try {
    const { getSystemInfo } = await import('@/api/system')
    systemInfo.value = await getSystemInfo()
  } catch {
    systemInfo.value = null
  }
  pollTimer = window.setInterval(fetchStatus, 3000)
})

onUnmounted(() => {
  if (pollTimer) clearInterval(pollTimer)
})
</script>

<template>
  <div>
    <header class="page-head">
      <div>
        <h1>概览</h1>
        <span class="subtitle">OneWeb — OneDrive Web Control Plane</span>
      </div>
      <div class="head-actions">
        <button class="btn secondary" @click="router.push('/profiles')">
          + 添加其他账号
        </button>
      </div>
    </header>

    <!-- 系统环境概况 -->
    <section v-if="systemInfo" class="system-card">
      <div class="sys-item"><span class="label">版本</span>{{ systemInfo.version }}</div>
      <div v-if="systemInfo.client_version" class="sys-item">
        <span class="label">onedrive</span>{{ systemInfo.client_version }}
      </div>
      <div v-if="systemInfo.capabilities" class="sys-item">
        <span class="label">能力</span>Schema {{ systemInfo.capabilities.schema_version }}
        <span v-if="systemInfo.capabilities.device_auth" class="chip">Device Auth</span>
        <span v-if="systemInfo.capabilities.sync_list" class="chip">SyncList</span>
      </div>
      <div class="sys-item">
        <span class="label">systemd</span>
        <span v-if="systemInfo.runtimes?.systemd?.available">可用</span>
        <span v-else class="muted">不可用</span>
      </div>
    </section>

    <!-- 场景 1: 单账户沉浸式控制看板 (95% 用户的默认状态) -->
    <section v-if="activeProfile" class="account-dashboard">
      <div class="account-header-panel">
        <div class="account-info">
          <div class="account-title-row">
            <h2>{{ activeProfile.display_name }}</h2>
            <StatusBadge :state="runtime?.state || 'stopped'" />
            <span v-if="runtime?.pid" class="pid-badge">PID {{ runtime.pid }}</span>
          </div>
          <div class="account-meta-row">
            <span class="meta-item"><span class="label">配置目录:</span> <code>{{ activeProfile.confdir }}</code></span>
            <span class="meta-item"><span class="label">运行服务:</span> <code>{{ activeProfile.runtime_target }}</code></span>
          </div>
        </div>

        <!-- 启停快捷控制栏 -->
        <div class="quick-controls">
          <button
            v-if="runtime?.state !== 'running'"
            class="btn primary"
            :disabled="acting || runtime?.state === 'starting'"
            @click="handleAction('start')"
          >
            {{ acting || runtime?.state === 'starting' ? '启动中...' : '▶ 启动同步' }}
          </button>
          <button
            v-else
            class="btn danger"
            :disabled="acting"
            @click="handleAction('stop')"
          >
            {{ acting ? '停止中...' : '⏹ 停止同步' }}
          </button>
          <button
            class="btn secondary"
            :disabled="acting"
            @click="handleAction('restart')"
          >
            ↻ 重启
          </button>
        </div>
      </div>

      <div v-if="actError" class="act-error">
        {{ actError }}
      </div>

      <!-- 核心功能直通卡片 -->
      <div class="feature-grid">
        <router-link :to="`/profiles/${activeProfile.id}/config`" class="feature-card">
          <div class="card-icon">⚙️</div>
          <div class="card-body">
            <div class="card-title">配置中心 (Configuration)</div>
            <div class="card-desc">Schema 动态表单、生效值与默认值三层对比、无损编辑保存与冲突保护</div>
          </div>
          <div class="card-arrow">→</div>
        </router-link>

        <router-link :to="`/profiles/${activeProfile.id}/runtime`" class="feature-card">
          <div class="card-icon">⚡</div>
          <div class="card-body">
            <div class="card-title">运行状态与实时日志 (Runtime & Logs)</div>
            <div class="card-desc">直通 systemd 与 journalctl 日志流，实时监控同步活动与性能</div>
          </div>
          <div class="card-arrow">→</div>
        </router-link>

        <router-link :to="`/profiles/${activeProfile.id}/synclist`" class="feature-card">
          <div class="card-icon">📁</div>
          <div class="card-body">
            <div class="card-title">选择性同步 (SyncList)</div>
            <div class="card-desc">白名单筛选规则管理、前导斜杠性能预警、重同步 (--resync) 联动</div>
          </div>
          <div class="card-arrow">→</div>
        </router-link>

        <router-link :to="`/profiles/${activeProfile.id}`" class="feature-card">
          <div class="card-icon">🔑</div>
          <div class="card-body">
            <div class="card-title">账号与认证 (Auth & Settings)</div>
            <div class="card-desc">微软 OAuth 2.0 登录引导、设备代码认证流、账号信息与安全设置</div>
          </div>
          <div class="card-arrow">→</div>
        </router-link>
      </div>
    </section>

    <!-- 场景 2: 多账户卡片列表 (≥ 2 个账号) -->
    <section v-else-if="profileStore.profiles.length > 1" class="cards">
      <div
        v-for="p in profileStore.profiles"
        :key="p.id"
        class="profile-card"
        @click="router.push(`/profiles/${p.id}`)"
      >
        <div class="card-head">
          <span class="card-title">{{ p.display_name || p.id }}</span>
          <StatusBadge :state="'stopped'" />
        </div>
        <div class="card-meta">
          <span>{{ p.confdir }}</span>
          <span class="chip">{{ p.runtime_type }}</span>
        </div>
      </div>
    </section>

    <!-- 场景 3: 初始化中 -->
    <section v-else class="empty">
      <p>正在连接并初始化本地 OneDrive 账户...</p>
    </section>
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
  font-size: 22px;
  margin: 0;
}
.subtitle {
  color: #64748b;
  font-size: 13px;
}
.head-actions {
  display: flex;
  gap: 10px;
}
.system-card {
  background: #fff;
  border: 1px solid var(--oneweb-border);
  border-radius: 10px;
  padding: 14px 20px;
  margin-bottom: 20px;
  display: flex;
  flex-wrap: wrap;
  gap: 24px;
}
.sys-item {
  font-size: 13px;
  display: flex;
  align-items: center;
  gap: 8px;
}
.sys-item .label {
  color: #64748b;
  font-size: 12px;
}
.chip {
  background: #e0f2fe;
  color: #0369a1;
  border-radius: 20px;
  padding: 2px 10px;
  font-size: 12px;
}
.muted {
  color: #94a3b8;
}

/* 单账户主看板样式 */
.account-dashboard {
  display: flex;
  flex-direction: column;
  gap: 20px;
}
.account-header-panel {
  background: #fff;
  border: 1px solid var(--oneweb-border);
  border-radius: 10px;
  padding: 24px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 16px;
}
.account-title-row {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 8px;
}
.account-title-row h2 {
  font-size: 18px;
  margin: 0;
}
.pid-badge {
  background: #f1f5f9;
  color: #475569;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
}
.account-meta-row {
  display: flex;
  flex-wrap: wrap;
  gap: 20px;
  font-size: 13px;
  color: #64748b;
}
.account-meta-row .label {
  color: #94a3b8;
  margin-right: 4px;
}
.account-meta-row code {
  background: #f8fafc;
  padding: 2px 6px;
  border-radius: 4px;
}
.quick-controls {
  display: flex;
  align-items: center;
  gap: 10px;
}
.act-error {
  background: #fee2e2;
  color: #991b1b;
  padding: 10px 16px;
  border-radius: 6px;
  font-size: 13px;
}

/* 核心功能直通卡片 */
.feature-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(360px, 1fr));
  gap: 16px;
}
.feature-card {
  background: #fff;
  border: 1px solid var(--oneweb-border);
  border-radius: 10px;
  padding: 20px;
  text-decoration: none;
  color: inherit;
  display: flex;
  align-items: center;
  gap: 16px;
  transition: all 0.15s ease;
}
.feature-card:hover {
  box-shadow: 0 6px 16px rgba(0, 0, 0, 0.06);
  border-color: #cbd5e1;
  transform: translateY(-1px);
}
.card-icon {
  font-size: 28px;
  width: 48px;
  height: 48px;
  background: #f8fafc;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}
.card-body {
  flex: 1;
}
.card-title {
  font-weight: 600;
  font-size: 15px;
  margin-bottom: 4px;
}
.card-desc {
  color: #64748b;
  font-size: 12px;
  line-height: 1.4;
}
.card-arrow {
  font-size: 18px;
  color: #94a3b8;
  transition: transform 0.15s;
}
.feature-card:hover .card-arrow {
  color: var(--oneweb-primary);
  transform: translateX(3px);
}

/* 多账户网格卡片 */
.cards {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 16px;
}
.profile-card {
  background: #fff;
  border: 1px solid var(--oneweb-border);
  border-radius: 10px;
  padding: 18px;
  cursor: pointer;
  transition: box-shadow 0.15s;
}
.profile-card:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
}
.card-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 10px;
}
.card-title {
  font-weight: 600;
  font-size: 15px;
}
.card-meta {
  display: flex;
  justify-content: space-between;
  color: #64748b;
  font-size: 13px;
}
.empty {
  text-align: center;
  padding: 48px;
  color: #64748b;
}

/* 通用按钮 */
.btn {
  padding: 8px 16px;
  border-radius: 6px;
  border: 1px solid transparent;
  font-size: 13px;
  cursor: pointer;
  font-weight: 500;
  transition: background 0.15s;
}
.btn.primary {
  background: var(--oneweb-primary);
  color: #fff;
}
.btn.primary:hover {
  filter: brightness(0.95);
}
.btn.danger {
  background: #ef4444;
  color: #fff;
}
.btn.danger:hover {
  background: #dc2626;
}
.btn.secondary {
  background: #f1f5f9;
  color: #334155;
  border-color: #cbd5e1;
}
.btn.secondary:hover {
  background: #e2e8f0;
}
</style>
