<template>
  <div class="home-page">
    <div class="search-bar" @click="onSearch">
      <span class="search-icon">🔍</span>
      <span class="search-placeholder">搜索商品</span>
    </div>

    <van-dropdown-menu>
      <van-dropdown-item ref="categoryItemRef" :title="categoryTitle">
        <div
          v-for="option in categoryOptions"
          :key="option.value"
          class="option-item"
          :class="{ selected: option.value === homeStore.category }"
          @click="onSelectCategory(option.value)"
        >
          {{ option.text }}
        </div>
      </van-dropdown-item>
      <van-dropdown-item ref="regionItemRef" :title="regionTitle">
        <van-cascader
          :options="regionOptions"
          @finish="onRegionFinish"
        />
      </van-dropdown-item>
    </van-dropdown-menu>

    <van-pull-refresh v-model="refreshing" @refresh="onRefresh">
      <van-list
        v-model:loading="homeStore.loading"
        :finished="!homeStore.hasMore"
        finished-text="已经到底了~"
        @load="onLoad"
      >
        <div
          v-for="item in homeStore.products"
          :key="item.id"
          class="product-card"
          @click="goToDetail(item.id)"
        >
          <div class="img-wrap">
            <van-image class="product-image" :src="firstImage(item.images)" fit="cover" />
            <div class="price-badge">¥{{ item.price }}</div>
          </div>
          <div class="product-info">
            <div class="product-title">{{ item.title }}</div>
            <div class="product-meta">
              <span class="product-location">📍 {{ item.location }}</span>
              <span class="product-seller">卖家：{{ item.seller }}</span>
            </div>
          </div>
        </div>
      </van-list>
    </van-pull-refresh>

    <van-empty v-if="!homeStore.loading && homeStore.products.length === 0" description="暂无商品" />
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import {
  DropdownMenu as VanDropdownMenu,
  DropdownItem as VanDropdownItem,
  Cascader as VanCascader,
  PullRefresh as VanPullRefresh,
  List as VanList,
  Empty as VanEmpty,
  Image as VanImage
} from 'vant'
import { useHomeStore } from '../stores/home'
import { buildRegionCascaderOptions } from '../utils/regionCascader'
import regionsData from '../data/chinaRegions'

const router = useRouter()
const homeStore = useHomeStore()

const categoryOptions = [
  { text: '全部', value: '' },
  { text: '电子产品', value: '电子产品' },
  { text: '服装鞋帽', value: '服装鞋帽' },
  { text: '图书文具', value: '图书文具' },
  { text: '生活用品', value: '生活用品' },
  { text: '数码配件', value: '数码配件' },
  { text: '其他', value: '其他' }
]

const regionOptions = buildRegionCascaderOptions(regionsData)

const refreshing = ref(false)
const categoryItemRef = ref(null)
const regionItemRef = ref(null)

const categoryTitle = computed(() => homeStore.category || '全部分类')
const regionTitle = computed(() => homeStore.district || homeStore.city || homeStore.province || '全部地区')

function onSelectCategory(value) {
  homeStore.setCategory(value)
  categoryItemRef.value?.toggle(false)
}

function onRegionFinish({ selectedOptions }) {
  const values = selectedOptions.map((o) => o.value)
  homeStore.setRegion({
    province: values[0] || '',
    city: values[1] || '',
    district: values[2] || ''
  })
  regionItemRef.value?.toggle(false)
}

async function onRefresh() {
  await homeStore.loadProducts(true)
  refreshing.value = false
}

async function onLoad() {
  await homeStore.loadProducts(false)
}

function goToDetail(id) {
  router.push({ name: 'detail', params: { id } })
}

function firstImage(images) {
  return images && images.length > 0 ? images[0] : ''
}

function onSearch() {
  // 搜索结果页本次范围外，暂不跳转
}
</script>

<style scoped>
.home-page {
  min-height: 100vh;
  background: #f5f5f5;
  padding-bottom: 20px;
}

.search-bar {
  background: #fff;
  padding: 10px 16px;
  display: flex;
  align-items: center;
  border-radius: 20px;
  margin: 12px;
  box-shadow: 0 1px 6px rgba(0, 0, 0, 0.05);
}

.search-icon {
  margin-right: 8px;
  font-size: 16px;
}

.search-placeholder {
  color: #999;
  font-size: 14px;
}

.product-card {
  background: #fff;
  border-radius: 8px;
  margin: 0 12px 10px;
  overflow: hidden;
  display: flex;
  padding: 8px;
  box-shadow: 0 3px 10px rgba(12, 18, 29, 0.06);
  align-items: flex-start;
}

.img-wrap {
  position: relative;
  width: 110px;
  height: 110px;
  flex-shrink: 0;
  border-radius: 6px;
  overflow: hidden;
}

.product-image {
  width: 100%;
  height: 100%;
}

.price-badge {
  position: absolute;
  left: 6px;
  bottom: 6px;
  background: rgba(7, 193, 96, 0.95);
  color: #fff;
  padding: 4px 7px;
  border-radius: 10px;
  font-size: 13px;
  font-weight: 600;
}

.product-info {
  flex: 1;
  margin-left: 10px;
  display: flex;
  flex-direction: column;
  justify-content: center;
}

.product-title {
  font-size: 15px;
  color: #111;
  display: -webkit-box;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
  overflow: hidden;
  line-height: 1.3;
  margin-bottom: 5px;
}

.product-meta {
  display: flex;
  gap: 10px;
  font-size: 12px;
  color: #888;
}

.product-location,
.product-seller {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 160px;
}

.option-item {
  padding: 12px 16px;
  font-size: 14px;
  color: #333;
}

.option-item.selected {
  color: #fa5151;
  font-weight: 500;
}
</style>
