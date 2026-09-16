<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useProfileStore } from '@/stores/profile'
import StatusBadge from '@/components/runtime/StatusBadge.vue'
import { getRuntimeLogs, getRuntimeStatus, runtimeAction } from '@/api/runtime'
import type { LogEntry, RuntimeStatus } from '@/api/runtime'

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
const recentLogs = ref<LogEntry[]>([])
const acting = ref(false)
const actError = ref<string | null>(null)
let pollTimer: number | undefined

async function fetchStatus() {
  if (!activeProfile.value) return
  try {
    runtime.value = await getRuntimeStatus(activeProfile.value.id)
    if (runtime.value?.state === 'running') {
      const logs = await getRuntimeLogs(activeProfile.value.id, 25)
      if (Array.isArray(logs)) {
        recentLogs.value = logs
      }
    } else {
      recentLogs.value = []
    }
  } catch {
    runtime.value = null
  }
}

interface SyncActivity {
  phase: 'idle' | 'syncing' | 'stopped' | 'starting'
  label: string
  lastSyncTime?: string
  detail?: string
}

const syncActivity = computed<SyncActivity>(() => {
  if (!runtime.value || runtime.value.state === 'stopped' || runtime.value.state === 'inactive') {
    return { phase: 'stopped', label: '服务已停止' }
  }
  if (runtime.value.state === 'starting') {
    return { phase: 'starting', label: '服务正在启动中...' }
  }
  if (runtime.value.state === 'failed') {
    return { phase: 'stopped', label: '服务运行异常' }
  }
  if (recentLogs.value.length === 0) {
    return {
      phase: 'idle',
      label: '空闲待命（实时监听中）',
      detail: '守护进程正在通过 inotify 监听本地文件变动',
    }
  }

  for (let i = recentLogs.value.length - 1; i >= 0; i--) {
    const msg = recentLogs.value[i].message || ''
    const ts = recentLogs.value[i].timestamp
      ? new Date(recentLogs.value[i].timestamp).toLocaleTimeString()
      : ''
    if (msg.includes('Sync with Microsoft OneDrive is complete')) {
      return {
        phase: 'idle',
        label: '空闲待命（已同步至最新）',
        lastSyncTime: ts,
        detail: '正在后台实时监听本地与云端变动',
      }
    }
    if (
      msg.includes('Starting a sync') ||
      msg.includes('Syncing changes') ||
      msg.includes('Fetching items') ||
      msg.includes('Scanning the local file system') ||
      msg.includes('Uploading') ||
      msg.includes('Downloading')
    ) {
      return {
        phase: 'syncing',
        label: '正在同步与比对数据...',
        detail: msg,
      }
    }
  }

  return {
    phase: 'idle',
    label: '空闲待命（实时监听中）',
    detail: '正在后台实时监听本地与云端变动',
  }
})

const syncingNow = ref(false)
const syncFeedback = ref<string | null>(null)

async function handleAction(action: 'start' | 'stop' | 'restart' | 'sync') {
  if (!activeProfile.value) return
  acting.value = true
  actError.value = null
  if (action === 'sync') syncingNow.value = true
  try {
    runtime.value = await runtimeAction(activeProfile.value.id, action)
    if (action === 'sync') {
      syncFeedback.value = '⚡ 已触发立刻同步，正在与云端比对最新变动！'
      setTimeout(() => { syncFeedback.value = null }, 4000)
    }
    setTimeout(fetchStatus, 500)
    setTimeout(fetchStatus, 2000)
  } catch (e: any) {
    actError.value = e.message
  } finally {
    acting.value = false
    syncingNow.value = false
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

          <!-- 实时同步活动指示行 -->
          <div v-if="runtime?.state === 'running'" class="sync-activity-row">
            <span class="activity-pill" :class="syncActivity.phase">
              <span class="activity-dot" />
              <span class="activity-text">{{ syncActivity.label }}</span>
            </span>
            <span v-if="syncActivity.lastSyncTime" class="activity-time">
              上次同步完成：<strong>{{ syncActivity.lastSyncTime }}</strong>
            </span>
            <span class="activity-hint">（守护服务后台待命中，实时监听文件变动）</span>
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
            {{ acting && !syncingNow || runtime?.state === 'starting' ? '启动中...' : '▶ 启动服务' }}
          </button>
          <button
            v-else
            class="btn danger"
            :disabled="acting"
            @click="handleAction('stop')"
          >
            {{ acting && !syncingNow ? '停止中...' : '⏹ 停止服务' }}
          </button>
          <button
            class="btn sync-btn"
            :disabled="acting || runtime?.state === 'starting'"
            @click="handleAction('sync')"
            title="立即触发全量数据比对与同步"
          >
            {{ syncingNow ? '⚡ 正在触发...' : '⚡ 立刻同步' }}
          </button>
          <button
            class="btn secondary"
            :disabled="acting"
            @click="handleAction('restart')"
          >
            ↻ 重启服务
          </button>
        </div>
      </div>

      <!-- 同步操作反馈提示 -->
      <div v-if="syncFeedback" class="sync-feedback-banner">
        {{ syncFeedback }}
      </div>

      <!-- 守护进程机制说明条 -->
      <div v-if="runtime?.state === 'running'" class="daemon-info-banner">
        💡 <strong>运行机制说明</strong>：OneDrive 客户端作为 Linux 守护进程（Daemon）在后台持续驻留。当数据比对完成后，服务不会退出，而是处于<strong>“空闲待命”</strong>状态，通过 Linux 内核 <code>inotify</code> 秒级响应本地文件变动，有改动时自动同步，无需手动停止。
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
.sync-activity-row {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-top: 12px;
  flex-wrap: wrap;
  font-size: 13px;
}
.activity-pill {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 3px 10px;
  border-radius: 20px;
  font-weight: 500;
  font-size: 12px;
}
.activity-pill.idle {
  background: #f0fdf4;
  color: #166534;
  border: 1px solid #bbf7d0;
}
.activity-pill.syncing {
  background: #eff6ff;
  color: #1d4ed8;
  border: 1px solid #bfdbfe;
}
.activity-pill.starting {
  background: #fefce8;
  color: #854d0e;
  border: 1px solid #fef08a;
}
.activity-pill.stopped {
  background: #f8fafc;
  color: #64748b;
  border: 1px solid #e2e8f0;
}
.activity-dot {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: currentColor;
}
.activity-pill.idle .activity-dot {
  box-shadow: 0 0 6px #22c55e;
}
.activity-pill.syncing .activity-dot {
  animation: pulse-dot 1.2s infinite;
}
@keyframes pulse-dot {
  0% { opacity: 0.3; transform: scale(0.9); }
  50% { opacity: 1; transform: scale(1.1); }
  100% { opacity: 0.3; transform: scale(0.9); }
}
.activity-time {
  color: #334155;
}
.activity-hint {
  color: #94a3b8;
  font-size: 12px;
}
.daemon-info-banner {
  background: #f0f9ff;
  border: 1px solid #bae6fd;
  border-radius: 8px;
  padding: 12px 18px;
  color: #0369a1;
  font-size: 13px;
  line-height: 1.5;
}
.daemon-info-banner code {
  background: #e0f2fe;
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
.btn.sync-btn {
  background: #2563eb;
  color: #fff;
  border-color: #1d4ed8;
}
.btn.sync-btn:hover {
  background: #1d4ed8;
}
.btn.sync-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
.sync-feedback-banner {
  background: #f0fdf4;
  border: 1px solid #86efac;
  color: #166534;
  padding: 10px 16px;
  border-radius: 6px;
  font-size: 13px;
  font-weight: 500;
  display: flex;
  align-items: center;
  gap: 8px;
}
</style>
