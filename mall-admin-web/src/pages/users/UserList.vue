<template>
  <el-card>
    <el-form inline class="search-form">
      <el-form-item>
        <el-input v-model="params.keyword" placeholder="搜索用户名/昵称" clearable style="width:200px" />
      </el-form-item>
      <el-form-item>
        <el-select v-model="bannedFilter" placeholder="全部状态" clearable style="width:120px">
          <el-option label="正常" :value="false" />
          <el-option label="已封禁" :value="true" />
        </el-select>
      </el-form-item>
      <el-form-item>
        <el-button type="primary" @click="handleSearch">搜索</el-button>
        <el-button @click="handleReset">重置</el-button>
      </el-form-item>
    </el-form>

    <el-table :data="users" v-loading="loading" border stripe>
      <el-table-column prop="id" label="ID" width="70" />
      <el-table-column prop="user_name" label="用户名" />
      <el-table-column prop="nick_name" label="昵称" />
      <el-table-column label="状态" width="90">
        <template #default="{ row }">
          <el-tag :type="row.is_banned ? 'danger' : 'success'">
            {{ row.is_banned ? '已封禁' : '正常' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="product_count" label="商品数" width="80" />
      <el-table-column prop="created_at" label="注册时间" width="170" />
      <el-table-column label="操作" width="160">
        <template #default="{ row }">
          <el-button size="small" @click="router.push(`/users/${row.id}`)">详情</el-button>
          <el-button size="small" :type="row.is_banned ? 'success' : 'danger'"
            @click="handleToggleBan(row)">
            {{ row.is_banned ? '解封' : '封禁' }}
          </el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-pagination
      class="pagination"
      v-model:current-page="params.page"
      v-model:page-size="params.page_size"
      :total="total"
      :page-sizes="[10, 20, 50]"
      layout="total, sizes, prev, pager, next"
      @change="loadUsers"
    />
  </el-card>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getUserList, banUser, unbanUser, type AdminUser } from '@/api/users'

const router = useRouter()
const loading = ref(false)
const users = ref<AdminUser[]>([])
const total = ref(0)
const bannedFilter = ref<boolean | undefined>(undefined)

const params = reactive({ keyword: '', page: 1, page_size: 10 })

async function loadUsers() {
  loading.value = true
  try {
    const res = await getUserList({
      ...params,
      is_banned: bannedFilter.value,
    })
    if (res.code === 0) {
      users.value = res.data.list
      total.value = res.data.total
    }
  } finally {
    loading.value = false
  }
}

function handleSearch() {
  params.page = 1
  loadUsers()
}

function handleReset() {
  params.keyword = ''
  bannedFilter.value = undefined
  params.page = 1
  loadUsers()
}

async function handleToggleBan(row: AdminUser) {
  const action = row.is_banned ? '解封' : '封禁'
  await ElMessageBox.confirm(`确认${action}用户 "${row.user_name}"？`, '提示', { type: 'warning' })
  try {
    const res = row.is_banned ? await unbanUser(row.id) : await banUser(row.id)
    if (res.code === 0) {
      ElMessage.success(`${action}成功`)
      loadUsers()
    }
  } catch {
    ElMessage.error('操作失败')
  }
}

onMounted(loadUsers)
</script>

<style scoped>
.search-form { margin-bottom: 16px; }
.pagination { margin-top: 16px; justify-content: flex-end; display: flex; }
</style>
