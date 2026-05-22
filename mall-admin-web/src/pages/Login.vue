<template>
  <div class="login-page">
    <el-card class="login-card">
      <h2 class="title">二手商城管理后台</h2>
      <el-form :model="form" :rules="rules" ref="formRef" @submit.prevent="handleLogin">
        <el-form-item prop="username">
          <el-input v-model="form.username" placeholder="用户名" prefix-icon="User" />
        </el-form-item>
        <el-form-item prop="password">
          <el-input v-model="form.password" type="password" placeholder="密码"
            prefix-icon="Lock" show-password @keyup.enter="handleLogin" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" class="login-btn" :loading="loading" @click="handleLogin">
            登录
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { useAuthStore } from '@/stores/auth'
import { adminLogin } from '@/api/auth'

const router = useRouter()
const auth = useAuthStore()
const formRef = ref()
const loading = ref(false)

const form = reactive({ username: '', password: '' })
const rules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }],
}

async function handleLogin() {
  await formRef.value?.validate()
  loading.value = true
  try {
    const res = await adminLogin(form.username, form.password)
    if (res.code === 0) {
      auth.setAuth(res.data.token, res.data.username)
      router.push('/users')
    } else {
      ElMessage.error(res.msg || '登录失败')
    }
  } catch {
    ElMessage.error('网络错误，请重试')
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-page {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 100vh;
  background: #f1f5f9;
}
.login-card {
  width: 380px;
}
.title {
  text-align: center;
  margin-bottom: 24px;
  font-size: 20px;
  color: #1d2d44;
}
.login-btn {
  width: 100%;
}
</style>
