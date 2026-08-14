# mall-web 首页 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在 `mall-web/`（当前为空目录）初始化 Vue3 + JavaScript + Vite 工程，实现二手商城 H5 首页（搜索入口、分类/地区筛选、商品列表分页浏览），并搭好路由骨架供后续页面接入。

**Architecture:** Vite 脚手架 + vue-router（路由骨架）+ Pinia（token / 首页数据状态）+ Vant 4（手动按需引入组件）+ axios（统一请求封装，拦截器处理 token 注入与错误提示）。首页数据来自现有后端接口 `GET /api/product/search`，不改动 `mall-server`。地区数据从 `mall-mini/miniprogram/data/china-regions.ts` 移植为纯 JS。

**Tech Stack:** Vue 3 (`<script setup>`, JavaScript), Vite, vue-router 4, Pinia, Vant 4, axios

## Global Constraints

- 全部代码使用 JavaScript，不使用 TypeScript
- Vant 组件通过具名 ES import 手动按需引入（`import { X as VanX } from 'vant'`），不引入 `unplugin-vue-components`/`unplugin-auto-import` 等额外构建插件
- 不新增自动化测试框架（无 vitest/jest）；验证方式为：纯函数用临时 node 脚本断言（不写入仓库）、其余用 `npm run build` 编译校验，最终以手动浏览器 QA 收尾
- 不修改 `mall-server/**` 与 `mall-mini/**` 任何文件
- 接口 base url 通过 `.env.development` 的 `VITE_API_BASE_URL` 配置，默认为 `http://localhost:8080`
- 本次范围仅首页；`/detail/:id`、`/publish`、`/chat`、`/my` 均为占位页，不实现具体逻辑
- 所有新文件路径以仓库根目录 `second-hand-mall/` 为基准，工程本身位于 `mall-web/`

---

### Task 1: 项目脚手架、依赖安装、路由骨架与占位页

**Files:**
- Create: `mall-web/package.json`, `mall-web/vite.config.js`, `mall-web/index.html`（由脚手架生成）
- Create: `mall-web/.env.development`
- Create: `mall-web/src/main.js`
- Create: `mall-web/src/App.vue`
- Create: `mall-web/src/styles/main.css`
- Create: `mall-web/src/router/index.js`
- Create: `mall-web/src/views/Home.vue`（占位版本，Task 6 会替换为完整实现）
- Create: `mall-web/src/views/Detail.vue`
- Create: `mall-web/src/views/Publish.vue`
- Create: `mall-web/src/views/ChatList.vue`
- Create: `mall-web/src/views/My.vue`
- Delete: `mall-web/src/assets/`, `mall-web/src/components/`, `mall-web/src/style.css`（脚手架默认模板文件，本项目不使用）

**Interfaces:**
- Produces: `router`（默认导出，`mall-web/src/router/index.js`），供 `main.js` 使用；路由名 `home` / `detail`（props: `id`）/ `publish` / `chat-list` / `my`

- [ ] **Step 1: 用 create-vite 生成 Vue3 + JS 模板**

```bash
cd mall-web
npx --yes create-vite@latest . --template vue
```

- [ ] **Step 2: 验证脚手架生成的文件**

Run: `ls mall-web`
Expected: 看到 `package.json`, `index.html`, `vite.config.js`, `src`, `public` 等文件/目录

- [ ] **Step 3: 安装依赖**

```bash
cd mall-web
npm install
npm install vue-router@4 pinia axios vant@4
```

- [ ] **Step 4: 删除模板自带的示例文件**

```bash
cd mall-web
rm -rf src/assets src/components src/style.css
```

- [ ] **Step 5: 创建环境变量文件**

`mall-web/.env.development`:

```
VITE_API_BASE_URL=http://localhost:8080
```

- [ ] **Step 6: 创建全局基础样式**

`mall-web/src/styles/main.css`:

```css
* {
  box-sizing: border-box;
}

body {
  margin: 0;
  background: #f5f5f5;
  font-family: -apple-system, BlinkMacSystemFont, "PingFang SC", "Microsoft YaHei", sans-serif;
}
```

- [ ] **Step 7: 创建占位页面（Detail / Publish / ChatList / My）**

`mall-web/src/views/Detail.vue`:

```vue
<template>
  <div class="placeholder-page">
    <p>商品详情页开发中，敬请期待</p>
  </div>
</template>
```

`mall-web/src/views/Publish.vue`:

```vue
<template>
  <div class="placeholder-page">
    <p>发布页开发中，敬请期待</p>
  </div>
</template>
```

`mall-web/src/views/ChatList.vue`:

```vue
<template>
  <div class="placeholder-page">
    <p>消息页开发中，敬请期待</p>
  </div>
</template>
```

`mall-web/src/views/My.vue`:

```vue
<template>
  <div class="placeholder-page">
    <p>我的页面开发中，敬请期待</p>
  </div>
</template>
```

- [ ] **Step 8: 创建首页占位版本**

`mall-web/src/views/Home.vue`:

```vue
<template>
  <div class="placeholder-page">
    <p>首页开发中</p>
  </div>
</template>
```

- [ ] **Step 9: 创建路由**

`mall-web/src/router/index.js`:

```js
import { createRouter, createWebHistory } from 'vue-router'
import Home from '../views/Home.vue'
import Detail from '../views/Detail.vue'
import Publish from '../views/Publish.vue'
import ChatList from '../views/ChatList.vue'
import My from '../views/My.vue'

const routes = [
  { path: '/', name: 'home', component: Home },
  { path: '/detail/:id', name: 'detail', component: Detail, props: true },
  { path: '/publish', name: 'publish', component: Publish },
  { path: '/chat', name: 'chat-list', component: ChatList },
  { path: '/my', name: 'my', component: My }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

export default router
```

- [ ] **Step 10: 创建 App.vue 与 main.js**

`mall-web/src/App.vue`:

```vue
<template>
  <router-view />
</template>
```

`mall-web/src/main.js`:

```js
import { createApp } from 'vue'
import { createPinia } from 'pinia'
import 'vant/lib/index.css'
import './styles/main.css'
import App from './App.vue'
import router from './router'

const app = createApp(App)
app.use(createPinia())
app.use(router)
app.mount('#app')
```

- [ ] **Step 11: 编译校验**

```bash
cd mall-web
npm run build
```

Expected: 命令成功退出（exit code 0），输出 `dist/` 构建产物，无报错

- [ ] **Step 12: 提交**

```bash
git add mall-web/
git commit -m "feat(mall-web): 初始化 Vue3+Vite 工程，搭建路由骨架与占位页"
```

---

### Task 2: Auth Store（token 状态管理）

**Files:**
- Create: `mall-web/src/stores/auth.js`

**Interfaces:**
- Produces: `useAuthStore()` Pinia store，`state.token: string`，`actions.setToken(token: string)`，`actions.clearToken()`。供 Task 3 的 `request.js` 拦截器使用

- [ ] **Step 1: 编写 auth store**

`mall-web/src/stores/auth.js`:

```js
import { defineStore } from 'pinia'

const TOKEN_KEY = 'token'

export const useAuthStore = defineStore('auth', {
  state: () => ({
    token: localStorage.getItem(TOKEN_KEY) || ''
  }),
  actions: {
    setToken(token) {
      this.token = token
      localStorage.setItem(TOKEN_KEY, token)
    },
    clearToken() {
      this.token = ''
      localStorage.removeItem(TOKEN_KEY)
    }
  }
})
```

- [ ] **Step 2: 用临时 node 脚本验证 store 逻辑（不写入仓库，仅本地断言）**

```bash
cd mall-web
node --input-type=module -e "
global.localStorage = {
  _data: {},
  getItem(k) { return Object.prototype.hasOwnProperty.call(this._data, k) ? this._data[k] : null },
  setItem(k, v) { this._data[k] = String(v) },
  removeItem(k) { delete this._data[k] }
}
const { createPinia, setActivePinia } = await import('pinia')
const { useAuthStore } = await import('./src/stores/auth.js')
setActivePinia(createPinia())
const store = useAuthStore()
console.assert(store.token === '', 'expected initial token to be empty, got ' + store.token)
store.setToken('abc123')
console.assert(store.token === 'abc123', 'expected token to be abc123, got ' + store.token)
console.assert(localStorage.getItem('token') === 'abc123', 'expected localStorage token to be abc123')
store.clearToken()
console.assert(store.token === '', 'expected token cleared, got ' + store.token)
console.assert(localStorage.getItem('token') === null, 'expected localStorage token removed')
console.log('OK')
"
```

Expected: 输出 `OK`，无 assertion 报错

- [ ] **Step 3: 编译校验**

```bash
cd mall-web
npm run build
```

Expected: 成功退出

- [ ] **Step 4: 提交**

```bash
git add mall-web/src/stores/auth.js
git commit -m "feat(mall-web): 新增 auth store 管理 token"
```

---

### Task 3: API 层（request.js + product.js）

**Files:**
- Create: `mall-web/src/api/request.js`
- Create: `mall-web/src/api/product.js`

**Interfaces:**
- Consumes: `useAuthStore()` from `mall-web/src/stores/auth.js`（Task 2）
- Produces: `request`（默认导出的 axios 实例，`request.get/post/put` 返回 resolve 时为 `{code, msg, data}`，`code!==0` 或网络错误时 reject）；`searchProducts(params): Promise<{code, msg, data: {list, total, page, page_size}}>`（`mall-web/src/api/product.js` 默认导出无，具名导出 `searchProducts`），供 Task 4 的 home store 使用

- [ ] **Step 1: 编写 axios 请求封装**

`mall-web/src/api/request.js`:

```js
import axios from 'axios'
import { showToast } from 'vant'
import { useAuthStore } from '../stores/auth'

const request = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL,
  timeout: 10000
})

request.interceptors.request.use((config) => {
  const authStore = useAuthStore()
  if (authStore.token) {
    config.headers.Authorization = `Bearer ${authStore.token}`
  }
  return config
})

request.interceptors.response.use(
  (response) => {
    const body = response.data
    if (body.code === 0) {
      return body
    }
    showToast(body.msg || '请求失败')
    return Promise.reject(body)
  },
  (error) => {
    if (error.response?.status === 401) {
      const authStore = useAuthStore()
      authStore.clearToken()
      showToast('登录已过期，请重新登录')
    } else {
      showToast('网络请求失败')
    }
    return Promise.reject(error)
  }
)

export default request
```

- [ ] **Step 2: 编写商品接口封装**

`mall-web/src/api/product.js`:

```js
import request from './request'

export function searchProducts(params) {
  return request.get('/api/product/search', { params })
}
```

- [ ] **Step 3: 编译校验**

```bash
cd mall-web
npm run build
```

Expected: 成功退出，无报错（`import.meta.env` 由 Vite 在构建时替换，不会报错）

- [ ] **Step 4: 提交**

```bash
git add mall-web/src/api/
git commit -m "feat(mall-web): 新增 axios 请求封装与商品搜索接口"
```

---

### Task 4: 查询参数工具 + Home Store

**Files:**
- Create: `mall-web/src/utils/searchParams.js`
- Create: `mall-web/src/stores/home.js`

**Interfaces:**
- Consumes: `searchProducts(params)` from `mall-web/src/api/product.js`（Task 3）
- Produces: `buildSearchParams({page, category, province, city, district}): object`；`useHomeStore()` Pinia store，`state: {products, page, hasMore, loading, category, province, city, district}`，`actions: {setCategory(category), setRegion({province, city, district}), loadProducts(reset: boolean)}`。供 Task 6 的 `Home.vue` 使用

- [ ] **Step 1: 编写查询参数纯函数**

`mall-web/src/utils/searchParams.js`:

```js
export function buildSearchParams({ page, category, province, city, district }) {
  const params = {
    page,
    page_size: 10,
    status: 0
  }
  if (category) params.category = category
  if (province) params.province = province
  if (city) params.city = city
  if (district) params.district = district
  return params
}
```

- [ ] **Step 2: 用临时 node 脚本验证纯函数（不写入仓库）**

```bash
cd mall-web
node --input-type=module -e "
import { buildSearchParams } from './src/utils/searchParams.js'
const p1 = buildSearchParams({ page: 1 })
console.assert(
  JSON.stringify(p1) === JSON.stringify({ page: 1, page_size: 10, status: 0 }),
  'base case failed: ' + JSON.stringify(p1)
)
const p2 = buildSearchParams({ page: 2, category: '数码配件', province: '广东省', city: '广州市', district: '天河区' })
console.assert(p2.page === 2 && p2.category === '数码配件' && p2.province === '广东省' && p2.city === '广州市' && p2.district === '天河区', 'full case failed: ' + JSON.stringify(p2))
const p3 = buildSearchParams({ page: 1, category: '', province: '', city: '', district: '' })
console.assert(!('category' in p3) && !('province' in p3), 'empty filters should be omitted: ' + JSON.stringify(p3))
console.log('OK')
"
```

Expected: 输出 `OK`

- [ ] **Step 3: 编写 home store**

`mall-web/src/stores/home.js`:

```js
import { defineStore } from 'pinia'
import { searchProducts } from '../api/product'
import { buildSearchParams } from '../utils/searchParams'

export const useHomeStore = defineStore('home', {
  state: () => ({
    products: [],
    page: 1,
    hasMore: true,
    loading: false,
    category: '',
    province: '',
    city: '',
    district: ''
  }),
  actions: {
    setCategory(category) {
      this.category = category
      this.loadProducts(true)
    },
    setRegion({ province, city, district }) {
      this.province = province
      this.city = city
      this.district = district
      this.loadProducts(true)
    },
    async loadProducts(reset) {
      if (this.loading) return
      if (reset) {
        this.page = 1
        this.hasMore = true
        this.products = []
      }
      this.loading = true
      try {
        const params = buildSearchParams({
          page: this.page,
          category: this.category,
          province: this.province,
          city: this.city,
          district: this.district
        })
        const res = await searchProducts(params)
        const list = res.data?.list || []
        const items = list.map((item) => ({
          id: item.id,
          title: item.title,
          price: item.price,
          location: item.location,
          images: item.images ? item.images.split(',').filter(Boolean) : [],
          seller: item.seller || '匿名用户',
          avatar: item.avatar || ''
        }))
        this.products = reset ? items : [...this.products, ...items]
        this.hasMore = list.length >= 10
        this.page += 1
      } finally {
        this.loading = false
      }
    }
  }
})
```

- [ ] **Step 4: 编译校验**

```bash
cd mall-web
npm run build
```

Expected: 成功退出

- [ ] **Step 5: 提交**

```bash
git add mall-web/src/utils/searchParams.js mall-web/src/stores/home.js
git commit -m "feat(mall-web): 新增查询参数工具与 home store"
```

---

### Task 5: 移植省市区数据 + 级联选项工具

**Files:**
- Create: `mall-web/src/data/chinaRegions.js`（内容移植自 `mall-mini/miniprogram/data/china-regions.ts`）
- Create: `mall-web/src/utils/regionCascader.js`

**Interfaces:**
- Produces: `chinaRegions.js` 默认导出 `regionsData: Array<{name, children: Array<{name, children: string[]}>}>`；`buildRegionCascaderOptions(regionsData): Array<{text, value, children: Array<{text, value, children: Array<{text, value}>}>}>`（Vant Cascader 所需格式）。供 Task 6 的 `Home.vue` 使用

- [ ] **Step 1: 移植省市区数据（去掉 TS 类型标注）**

```bash
cd /Users/yangzhongyu/Desktop/code/github/second-hand-mall
sed -n '11,422p' mall-mini/miniprogram/data/china-regions.ts \
  | sed '1s/: RegionItem\[\]//' \
  > mall-web/src/data/chinaRegions.js
```

- [ ] **Step 2: 验证生成的文件头尾正确**

```bash
head -3 mall-web/src/data/chinaRegions.js
tail -3 mall-web/src/data/chinaRegions.js
```

Expected: 第一行为 `const regionsData = [`（不含 `: RegionItem[]`），最后一行为 `export default regionsData`

- [ ] **Step 3: 编写级联选项转换工具**

`mall-web/src/utils/regionCascader.js`:

```js
export function buildRegionCascaderOptions(regionsData) {
  return regionsData.map((province) => ({
    text: province.name,
    value: province.name,
    children: province.children.map((city) => ({
      text: city.name,
      value: city.name,
      children: city.children.map((district) => ({
        text: district,
        value: district
      }))
    }))
  }))
}
```

- [ ] **Step 4: 用临时 node 脚本验证转换结果（不写入仓库）**

```bash
cd mall-web
node --input-type=module -e "
import regionsData from './src/data/chinaRegions.js'
import { buildRegionCascaderOptions } from './src/utils/regionCascader.js'
console.assert(Array.isArray(regionsData) && regionsData.length > 0, 'regionsData should be a non-empty array')
const options = buildRegionCascaderOptions(regionsData)
const beijing = options.find((o) => o.value === '北京市')
console.assert(beijing, 'expected to find 北京市')
console.assert(beijing.children.length > 0, 'expected 北京市 to have city children')
console.assert(beijing.children[0].children.length > 0, 'expected city to have district children')
console.assert(typeof beijing.children[0].children[0].text === 'string', 'district option should have text field')
console.log('OK')
"
```

Expected: 输出 `OK`

- [ ] **Step 5: 编译校验**

```bash
cd mall-web
npm run build
```

Expected: 成功退出

- [ ] **Step 6: 提交**

```bash
git add mall-web/src/data/chinaRegions.js mall-web/src/utils/regionCascader.js
git commit -m "feat(mall-web): 移植省市区数据并新增级联选项工具"
```

---

### Task 6: 首页完整实现（Home.vue）

**Files:**
- Modify: `mall-web/src/views/Home.vue`（替换 Task 1 的占位内容）

**Interfaces:**
- Consumes: `useHomeStore()`（Task 4，含 `products/loading/hasMore/setCategory/setRegion/loadProducts`）、`buildRegionCascaderOptions` + `chinaRegions.js` 默认导出（Task 5）、`useRouter()` from vue-router（跳转 `detail` 路由，Task 1 已注册）

- [ ] **Step 1: 编写首页完整实现**

`mall-web/src/views/Home.vue`:

```vue
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
```

- [ ] **Step 2: 编译校验**

```bash
cd mall-web
npm run build
```

Expected: 成功退出，无报错

- [ ] **Step 3: 提交**

```bash
git add mall-web/src/views/Home.vue
git commit -m "feat(mall-web): 实现首页搜索/筛选/商品列表完整功能"
```

---

### Task 7: 最终联调与手动 QA

**Files:**
- No new files；本任务为验证性任务

- [ ] **Step 1: 确认后端服务可用**

```bash
cd /Users/yangzhongyu/Desktop/code/github/second-hand-mall/mall-server
go build -o mall-server
./mall-server web -config configs/config.yaml &
sleep 1
curl -s http://localhost:8080/actuator/health/readiness
```

Expected: 返回健康检查成功响应（服务已在监听 8080）

- [ ] **Step 2: 启动前端开发服务器**

```bash
cd mall-web
npm run dev &
sleep 2
curl -s http://localhost:5173 | grep -o '<div id="app">'
```

Expected: 输出 `<div id="app">`，说明开发服务器正常返回 SPA 入口页面

- [ ] **Step 3: 浏览器手动 QA 清单**

在浏览器打开 `http://localhost:5173`，逐项确认：

- [ ] 商品列表正常拉取展示（若数据库无商品数据，确认展示 `van-empty` 空状态而非报错）
- [ ] 点击「全部分类」筛选栏，选择某个分类后列表按分类重新加载，筛选栏标题变为所选分类名
- [ ] 点击「全部地区」筛选栏，逐级选择省/市/区后列表按地区重新加载，筛选栏标题变为所选区名
- [ ] 下拉刷新触发重新加载第一页
- [ ] 滚动到列表底部触发加载下一页（若总数超过 10 条）
- [ ] 点击商品卡片跳转到 `/detail/:id` 占位页，URL 参数正确
- [ ] 浏览器控制台无报错

- [ ] **Step 4: 关闭本地服务**

```bash
pkill -f "mall-server web" || true
pkill -f "vite" || true
```

- [ ] **Step 5: 记录 QA 结果**

若清单全部通过，直接进行下一步；若有失败项，加载 `superpowers:systematic-debugging` 定位根因后修复、重新验证，不得直接跳过。

- [ ] **Step 6: 最终提交（如有修复）**

```bash
git add mall-web/
git commit -m "fix(mall-web): 首页手动 QA 修复"
```

（若 QA 全部通过且无需修复，本步骤跳过）
