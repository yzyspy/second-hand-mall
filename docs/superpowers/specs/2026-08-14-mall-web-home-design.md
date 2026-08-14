# mall-web 首页功能设计

## 概述

初始化 `mall-web/`（Vue 3 + JavaScript + Vite）项目，参照 `mall-mini` 小程序首页（`pages/home/`）实现二手商城 H5 首页：搜索入口、分类/地区筛选、商品列表（下拉刷新 + 触底加载）。同时搭好路由骨架，其余页面（详情/发布/消息/我的）留占位页，本次不实现具体逻辑。

复用后端已有接口 `GET /api/product/search`（见 `mall-server/internal/app/service`），不新增/修改后端代码。

## 技术栈

- **构建工具**：Vite + Vue 3（`<script setup>`，JavaScript，不使用 TypeScript）
- **路由**：vue-router 4
- **状态管理**：Pinia
- **UI 组件库**：Vant 4（按需引入）
- **HTTP**：axios，封装 request 拦截器自动注入 token；响应拦截器统一处理业务 `code !== 0` 和网络错误
- **接口地址**：`.env.development` 中 `VITE_API_BASE_URL=http://localhost:8080`；后端 CORS 已放行 `http://localhost:5173`（Vite 默认端口），本地开发无需反向代理

## 项目结构

```
mall-web/
├── .env.development          # VITE_API_BASE_URL=http://localhost:8080
├── vite.config.js
├── package.json
├── index.html
└── src/
    ├── main.js                # 挂载 App，注册 router / pinia / vant 样式
    ├── App.vue
    ├── router/
    │   └── index.js           # /, /detail/:id, /publish, /chat, /my
    ├── api/
    │   ├── request.js         # axios 实例 + 拦截器（token 注入、401 清 token + toast、统一错误 toast）
    │   └── product.js         # searchProducts(params) 封装 GET /api/product/search
    ├── stores/
    │   ├── auth.js            # Pinia store：token（读写 localStorage）
    │   └── home.js            # Pinia store：products/page/hasMore/loading/筛选条件 + loadProducts action
    ├── data/
    │   └── chinaRegions.js    # 从 mall-mini/data/china-regions.ts 移植的省市区数据（去掉 TS 类型）
    ├── views/
    │   ├── Home.vue            # 首页（本次核心实现）
    │   ├── Detail.vue          # 占位页
    │   ├── Publish.vue         # 占位页
    │   ├── ChatList.vue        # 占位页
    │   └── My.vue              # 占位页
    └── styles/
        └── main.css            # 全局基础样式（重置样式等）
```

## 首页设计（Home.vue）

### 布局对照小程序 home.wxml

```
┌─────────────────────────────────┐
│  🔍 搜索商品                     │  ← 搜索栏，点击跳转 /search（暂未实现，先占位路由或跳过）
├─────────────────────────────────┤
│  [全部分类 ▾] [全部地区 ▾]       │  ← van-dropdown-menu
├─────────────────────────────────┤
│  下拉刷新区域（van-pull-refresh） │
│  ┌───────────────────────────┐  │
│  │ [图] ¥价格   标题(2行省略)  │  │  ← van-list 触底加载
│  │      📍位置  卖家：xxx      │  │
│  └───────────────────────────┘  │
│  ...                            │
│  加载中 / 没有更多 / van-empty   │
└─────────────────────────────────┘
```

### 筛选交互（对照 home.ts 状态机简化）

- **分类筛选**：`van-dropdown-item` 内渲染分类列表（全部 + 6 个分类），选中后收起下拉、重置分页、重新加载 — 对应 `onSelectCategory`/`onClearCategory`
- **地区筛选**：`van-dropdown-item` 内嵌 `van-cascader`，数据源为 `chinaRegions.js` 转换成 `{text, value, children}` 树；选择到「区」层级或点确认后收起、重置分页、重新加载 — 对应 `onSelectProvince`/`onSelectCity`/`onSelectDistrict`/`onConfirmRegion`/`onClearRegion`。用 cascader 组件替代小程序手写的三级 step 面板，交互效果等价但实现更简单

### 数据流（`stores/home.js`）

```js
state: {
  products: [],
  page: 1,
  hasMore: true,
  loading: false,
  selectedCategory: '',
  selectedProvince: '', selectedCity: '', selectedDistrict: '',
}
actions: {
  async loadProducts(reset) {
    // reset=true 时 page=1, products=[]
    // 组装 params: page, page_size:10, status:0, 按需附加 category/province/city/district
    // 调 api/product.js -> GET /api/product/search
    // 追加/替换 products，hasMore = list.length >= 10，page += 1
  }
}
```

对应关系：
- `onLoad`/`onShow` → 组件 `onMounted`
- `onPullDownRefresh` → `van-pull-refresh` 的 `refresh` 事件，调用 `loadProducts(true)`
- `onReachBottom` → `van-list` 的 `load` 事件（`finished` 绑定 `!hasMore`）

### 商品卡片

字段与小程序一致：`images[0]` 主图 + 价格徽标 + 2 行省略标题 + 位置/卖家。`images` 字段后端返回逗号分隔字符串，前端 `split(',').filter(Boolean)` 处理，与 `home.ts` 一致。点击卡片跳转 `/detail/:id`（占位页，仅接收参数展示，不做详情逻辑）。

## 错误处理

- axios 响应拦截器：`code !== 0` 或网络错误 → `van-toast` 提示（对应小程序 `wx.showToast`）
- 401：清除本地 token（`localStorage`），toast「登录已过期」。不做自动重登（Web 无微信 `wx.login` 能力），登录页本次不实现
- 图片加载失败：使用 `<van-image>` 的 `error` 插槽兜底（可选，非必须）

## 测试与验收

- `npm run dev` 启动后浏览器访问首页：
  - 商品列表正常拉取展示（需本地起 `mall-server` 且有商品数据）
  - 分类筛选、地区筛选（省/市/区级联）能正确带参数重新拉取
  - 下拉刷新、滚动到底加载更多正常工作
  - 无商品时展示空状态
- 不写自动化测试（前端小工程 + 无现有测试基础设施），以手动浏览器验证为主

## 文件清单

### 新增

- `mall-web/` 整个 Vite 工程（`package.json`, `vite.config.js`, `index.html`, `src/**`）
- `mall-web/src/data/chinaRegions.js`（移植自 `mall-mini/miniprogram/data/china-regions.ts`）

### 不涉及修改

- `mall-server/**`（复用现有 `/api/product/search` 接口，不改动后端）
- `mall-mini/**`（仅作为实现参照，不修改小程序代码）

## 范围外（后续迭代）

- 登录/注册页面与鉴权流程
- 商品详情页 `/detail/:id` 具体内容
- 发布、消息（chat）、我的（my）等页面具体功能
- 搜索结果页 `/search`
