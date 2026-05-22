<template>
  <el-card v-loading="loading">
    <template #header>
      <div style="display:flex;justify-content:space-between;align-items:center">
        <el-button @click="router.back()">← 返回列表</el-button>
        <el-button
          v-if="user"
          :type="user.is_banned ? 'success' : 'danger'"
          @click="handleToggleBan"
        >
          {{ user?.is_banned ? '解封用户' : '封禁用户' }}
        </el-button>
      </div>
    </template>

    <el-descriptions v-if="user" :column="2" border>
      <el-descriptions-item label="ID">{{ user.id }}</el-descriptions-item>
      <el-descriptions-item label="用户名">{{ user.user_name }}</el-descriptions-item>
      <el-descriptions-item label="昵称">{{ user.nick_name || '—' }}</el-descriptions-item>
      <el-descriptions-item label="手机号">{{ user.phone || '—' }}</el-descriptions-item>
      <el-descriptions-item label="状态">
        <el-tag :type="user.is_banned ? 'danger' : 'success'">
          {{ user.is_banned ? '已封禁' : '正常' }}
        </el-tag>
      </el-descriptions-item>
      <el-descriptions-item label="发布商品数">{{ user.product_count }}</el-descriptions-item>
      <el-descriptions-item label="注册时间">{{ user.created_at }}</el-descriptions-item>
    </el-descriptions>
  </el-card>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getUserDetail, banUser, unbanUser, type AdminUser } from '@/api/users'

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const user = ref<AdminUser | null>(null)

async function loadDetail() {
  loading.value = true
  try {
    const res = await getUserDetail(Number(route.params.id))
    if (res.code === 0) user.value = res.data
  } finally {
    loading.value = false
  }
}

async function handleToggleBan() {
  if (!user.value) return
  const action = user.value.is_banned ? '解封' : '封禁'
  await ElMessageBox.confirm(`确认${action}用户 "${user.value.user_name}"？`, '提示', { type: 'warning' })
  try {
    const res = user.value.is_banned
      ? await unbanUser(user.value.id)
      : await banUser(user.value.id)
    if (res.code === 0) {
      ElMessage.success(`${action}成功`)
      loadDetail()
    }
  } catch {
    ElMessage.error('操作失败')
  }
}

onMounted(loadDetail)
</script>
