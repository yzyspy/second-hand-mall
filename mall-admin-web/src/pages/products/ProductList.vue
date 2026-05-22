<template>
  <el-card>
    <el-form inline class="search-form">
      <el-form-item>
        <el-input v-model="params.keyword" placeholder="搜索商品标题" clearable style="width:200px" />
      </el-form-item>
      <el-form-item>
        <el-select v-model="params.category" placeholder="全部分类" clearable style="width:120px">
          <el-option v-for="c in categories" :key="c" :label="c" :value="c" />
        </el-select>
      </el-form-item>
      <el-form-item>
        <el-select v-model="params.status" placeholder="全部状态" clearable style="width:120px">
          <el-option label="在售" :value="0" />
          <el-option label="已售出" :value="1" />
          <el-option label="已下架" :value="2" />
        </el-select>
      </el-form-item>
      <el-form-item>
        <el-button type="primary" @click="handleSearch">搜索</el-button>
        <el-button @click="handleReset">重置</el-button>
      </el-form-item>
    </el-form>

    <el-table :data="products" v-loading="loading" border stripe>
      <el-table-column prop="id" label="ID" width="70" />
      <el-table-column prop="title" label="标题" show-overflow-tooltip />
      <el-table-column prop="price" label="价格" width="90">
        <template #default="{ row }">¥{{ row.price }}</template>
      </el-table-column>
      <el-table-column prop="category" label="分类" width="100" />
      <el-table-column label="状态" width="90">
        <template #default="{ row }">
          <el-tag :type="row.status === 0 ? 'success' : row.status === 1 ? 'warning' : 'info'">
            {{ row.status === 0 ? '在售' : row.status === 1 ? '已售出' : '已下架' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="location" label="地区" width="120" />
      <el-table-column prop="seller" label="发布者" width="100" />
      <el-table-column prop="create_time" label="发布时间" width="170" />
      <el-table-column label="操作" width="140">
        <template #default="{ row }">
          <el-button size="small" @click="router.push(`/products/${row.id}`)">详情</el-button>
          <el-button size="small" type="danger" :disabled="row.status === 2"
            @click="handleDelist(row)">下架</el-button>
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
      @change="loadProducts"
    />
  </el-card>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getProductList, delistProduct, type AdminProduct } from '@/api/products'

const router = useRouter()
const loading = ref(false)
const products = ref<AdminProduct[]>([])
const total = ref(0)
const categories = ['电子产品', '服装鞋帽', '图书文具', '生活用品', '数码配件', '其他']

const params = reactive<{
  keyword: string; category: string | undefined; status: number | undefined; page: number; page_size: number
}>({ keyword: '', category: undefined, status: undefined, page: 1, page_size: 10 })

async function loadProducts() {
  loading.value = true
  try {
    const res = await getProductList(params)
    if (res.code === 0) {
      products.value = res.data.list
      total.value = res.data.total
    }
  } finally {
    loading.value = false
  }
}

function handleSearch() { params.page = 1; loadProducts() }
function handleReset() {
  params.keyword = ''; params.category = undefined; params.status = undefined; params.page = 1
  loadProducts()
}

async function handleDelist(row: AdminProduct) {
  await ElMessageBox.confirm(`确认强制下架商品 "${row.title}"？`, '提示', { type: 'warning' })
  try {
    const res = await delistProduct(row.id)
    if (res.code === 0) {
      ElMessage.success('已下架')
      loadProducts()
    }
  } catch {
    ElMessage.error('操作失败')
  }
}

onMounted(loadProducts)
</script>

<style scoped>
.search-form { margin-bottom: 16px; }
.pagination { margin-top: 16px; justify-content: flex-end; display: flex; }
</style>
