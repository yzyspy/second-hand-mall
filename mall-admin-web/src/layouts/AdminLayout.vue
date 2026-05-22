<template>
  <div class="admin-layout">
    <aside class="sidebar">
      <div class="logo">⚙ 二手商城</div>
      <el-menu
        :default-active="activeMenu"
        router
        background-color="#1d2d44"
        text-color="#94a3b8"
        active-text-color="#fff"
      >
        <el-menu-item index="/users">
          <el-icon><User /></el-icon>
          <span>用户管理</span>
        </el-menu-item>
        <el-menu-item index="/products">
          <el-icon><Box /></el-icon>
          <span>商品管理</span>
        </el-menu-item>
      </el-menu>
      <div class="logout" @click="handleLogout">
        <el-icon><SwitchButton /></el-icon> 退出登录
      </div>
    </aside>

    <div class="main">
      <header class="header">
        <span class="breadcrumb">{{ currentTitle }}</span>
        <span class="admin-name">{{ auth.username }}</span>
      </header>
      <main class="content">
        <router-view />
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()

const activeMenu = computed(() => {
  if (route.path.startsWith('/users')) return '/users'
  if (route.path.startsWith('/products')) return '/products'
  return '/users'
})

const titleMap: Record<string, string> = {
  '/users': '用户管理 / 用户列表',
  '/products': '商品管理 / 商品列表',
}

const currentTitle = computed(() => titleMap[route.path] ?? '管理后台')

function handleLogout() {
  auth.logout()
  router.push('/login')
}
</script>

<style scoped>
.admin-layout {
  display: flex;
  min-height: 100vh;
}
.sidebar {
  width: 200px;
  background: #1d2d44;
  display: flex;
  flex-direction: column;
  flex-shrink: 0;
}
.logo {
  padding: 16px;
  color: #fff;
  font-size: 15px;
  font-weight: bold;
  border-bottom: 1px solid #2d4060;
}
.logout {
  margin-top: auto;
  padding: 14px 20px;
  color: #64748b;
  cursor: pointer;
  border-top: 1px solid #2d4060;
  font-size: 13px;
  display: flex;
  align-items: center;
  gap: 6px;
}
.logout:hover { color: #94a3b8; }
.main {
  flex: 1;
  display: flex;
  flex-direction: column;
  background: #f1f5f9;
  overflow: hidden;
}
.header {
  height: 56px;
  background: #fff;
  border-bottom: 1px solid #e2e8f0;
  display: flex;
  align-items: center;
  padding: 0 24px;
  justify-content: space-between;
  font-size: 13px;
  color: #64748b;
}
.admin-name { font-weight: 500; color: #334155; }
.content {
  flex: 1;
  padding: 20px;
  overflow-y: auto;
}
</style>
