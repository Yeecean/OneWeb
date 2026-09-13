import { createRouter, createWebHistory } from 'vue-router'
import DashboardView from '@/views/DashboardView.vue'
import ProfileListView from '@/views/ProfileListView.vue'
import ProfileDetailView from '@/views/ProfileDetailView.vue'
import ConfigEditorView from '@/views/ConfigEditorView.vue'
import RuntimeView from '@/views/RuntimeView.vue'
import SyncListView from '@/views/SyncListView.vue'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', name: 'dashboard', component: DashboardView },
    { path: '/profiles', name: 'profiles', component: ProfileListView },
    { path: '/profiles/:id', name: 'profile-detail', component: ProfileDetailView },
    { path: '/profiles/:id/config', name: 'config-editor', component: ConfigEditorView },
    { path: '/profiles/:id/runtime', name: 'runtime', component: RuntimeView },
    { path: '/profiles/:id/synclist', name: 'synclist', component: SyncListView },
  ],
})

export default router
