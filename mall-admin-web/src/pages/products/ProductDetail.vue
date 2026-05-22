<template>
  <el-card v-loading="loading">
    <template #header>
      <div style="display:flex;justify-content:space-between;align-items:center">
        <el-button @click="router.back()">← 返回列表</el-button>
        <el-button
          v-if="product && product.status !== 2"
          type="danger"
          @click="handleDelist"
        >强制下架</el-button>
      </div>
    </template>

    <el-descriptions v-if="product" :column="2" border>
      <el-descriptions-item label="ID">{{ product.id }}</el-descriptions-item>
      <el-descriptions-item label="标题">{{ product.title }}</el-descriptions-item>
      <el-descriptions-item label="价格">¥{{ product.price }}</el-descriptions-item>
      <el-descriptions-item label="分类">{{ product.category || '—' }}</el-descriptions-item>
      <el-descriptions-item label="状态">
        <el-tag :type="product.status === 0 ? 'success' : product.status === 1 ? 'warning' : 'info'">
          {{ product.status === 0 ? '在售' : product.status === 1 ? '已售出' : '已下架' }}
        </el-tag>
      </el-descriptions-item>
      <el-descriptions-item label="地区">{{ product.location }}</el-descriptions-item>
      <el-descriptions-item label="发布者">{{ product.seller }}</el-descriptions-item>
      <el-descriptions-item label="发布时间">{{ product.create_time }}</el-descriptions-item>
      <el-descriptions-item label="图片" :span="2">
        <div style="display:flex;gap:8px;flex-wrap:wrap">
          <el-image
            v-for="(img, i) in (product.images || '').split(',').filter(Boolean)"
            :key="i"
            :src="img"
            style="width:100px;height:100px;object-fit:cover;border-radius:4px"
            fit="cover"
          />
        </div>
      </el-descriptions-item>
    </el-descriptions>
  </el-card>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getProductDetail, delistProduct, type AdminProduct } from '@/api/products'

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const product = ref<AdminProduct | null>(null)

async function loadDetail() {
  loading.value = true
  try {
    const res = await getProductDetail(Number(route.params.id))
    if (res.code === 0) product.value = res.data
  } finally {
    loading.value = false
  }
}

async function handleDelist() {
  if (!product.value) return
  await ElMessageBox.confirm(`确认强制下架 "${product.value.title}"？`, '提示', { type: 'warning' })
  try {
    const res = await delistProduct(product.value.id)
    if (res.code === 0) {
      ElMessage.success('已下架')
      loadDetail()
    }
  } catch {
    ElMessage.error('操作失败')
  }
}

onMounted(loadDetail)
</script>
