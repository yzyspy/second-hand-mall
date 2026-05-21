# 二手商城管理后台 Design Spec

**日期：** 2026-05-21  
**状态：** 已确认

---

## 概述

为二手商城平台新增 Web 管理后台，支持管理员对用户和商品进行管理操作。

- **前端：** `mall-admin-web/`，Vue3 + Vite + TypeScript + Element Plus
- **后端：** 管理接口追加到 `mall-server/`，路由前缀 `/admin`

---

## 功能范围

### 用户管理
- 用户列表（分页、按用户名/昵称搜索、按封禁状态筛选）
- 用户详情（基本信息 + 发布商品数 + 收藏数）
- 封禁 / 解封用户

### 商品管理
- 商品列表（分页、按标题/分类/状态/地区搜索）
- 商品详情（完整信息 + 发布者信息）
- 强制下架（将 status 设为 2）

---

## 后端设计（mall-server）

### 数据库变更

**新增表 `admin_user`**
```sql
id            INTEGER PRIMARY KEY AUTOINCREMENT
username      VARCHAR(50) NOT NULL UNIQUE
password_hash VARCHAR(255) NOT NULL   -- bcrypt
created_at    DATETIME
updated_at    DATETIME
deleted_at    DATETIME
```

**`sys_user` 新增字段**
```sql
is_banned BOOLEAN NOT NULL DEFAULT false
```

> `is_banned = true` 的用户调用 `/api/user/wx-login` 或 `/user/login` 时，两个登录入口均需检查该字段，返回 403 并提示「账号已被封禁」。

### CLI 命令

```bash
# 创建第一个管理员账号
./mall-server admin create-admin --username admin --password <密码>
```

`create-admin` 命令：读取参数，bcrypt 哈希密码，写入 `admin_user` 表。若用户名已存在则报错退出。

### 认证

- **端点：** `POST /admin/login`，body: `{ username, password }`
- **响应：** `{ token: "<admin JWT>" }`
- **JWT Claims：** `{ admin_id, username, is_admin: true, exp: now+7days }`
- **密钥：** 复用现有 JWT 密钥
- **AdminAuthMiddleware：** 解析 Bearer token，验证 `is_admin == true`，否则返回 401

### Admin API 路由

**公开路由（无需 auth）**

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /admin/login | 登录，返回 admin JWT |

**受保护路由（需 `AdminAuthMiddleware`）**

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /admin/users | 用户列表（分页 + 搜索） |
| GET | /admin/users/:id | 用户详情 |
| POST | /admin/users/:id/ban | 封禁用户（设 is_banned=true） |
| POST | /admin/users/:id/unban | 解封用户（设 is_banned=false） |
| GET | /admin/products | 商品列表（分页 + 搜索） |
| GET | /admin/products/:id | 商品详情 |
| POST | /admin/products/:id/delist | 强制下架（直接设 status=2，不校验所有者） |

### 查询参数

**GET /admin/users**
- `keyword` — 模糊匹配 username / nick_name
- `is_banned` — `true` / `false` / 空（全部）
- `page`、`page_size`（默认 10，最大 50）

**GET /admin/products**
- `keyword` — 模糊匹配 title
- `category`、`province`、`city`、`district` — 精确匹配
- `status` — 0/1/2 / 空（全部）
- `page`、`page_size`（默认 10，最大 50）

### CORS

在 `CORSMiddleware` 中追加允许来源 `http://localhost:5174`（admin 开发服务器端口）。

---

## 前端设计（mall-admin-web）

### 技术栈
- Vue 3 + Vite + TypeScript
- Element Plus（UI 组件库）
- Vue Router 4（路由）
- Pinia（状态管理）
- Axios（HTTP 请求）

### 目录结构

```
mall-admin-web/
├── index.html
├── vite.config.ts           # proxy: /admin → http://localhost:10088
├── tsconfig.json
├── package.json
├── src/
│   ├── main.ts
│   ├── App.vue
│   ├── api/
│   │   ├── auth.ts          # login()
│   │   ├── users.ts         # getUsers(), getUserDetail(), banUser(), unbanUser()
│   │   └── products.ts      # getProducts(), getProductDetail(), delistProduct()
│   ├── router/
│   │   └── index.ts         # 路由守卫：未登录跳转 /login
│   ├── stores/
│   │   └── auth.ts          # Pinia store：token（localStorage）、adminInfo、logout()
│   ├── layouts/
│   │   └── AdminLayout.vue  # 左侧导航栏 + 顶部 Header + <router-view>
│   ├── pages/
│   │   ├── Login.vue
│   │   ├── users/
│   │   │   ├── UserList.vue
│   │   │   └── UserDetail.vue
│   │   └── products/
│   │       ├── ProductList.vue
│   │       └── ProductDetail.vue
│   └── utils/
│       └── request.ts       # Axios 实例，拦截器自动注入 Authorization: Bearer <token>
```

### 路由

```
/login              → Login.vue（无需登录）
/users              → AdminLayout > UserList.vue
/users/:id          → AdminLayout > UserDetail.vue
/products           → AdminLayout > ProductList.vue
/products/:id       → AdminLayout > ProductDetail.vue
/                   → 重定向到 /users
```

### 布局

**AdminLayout.vue**
- 左侧深色侧边栏（宽 200px）：Logo + 「用户管理」「商品管理」菜单项
- 顶部 Header（高 56px）：当前页面面包屑 + 管理员用户名 + 退出按钮
- 内容区：`<router-view>`，灰色背景 (#f1f5f9)，内边距 20px

**列表页通用结构**
1. 搜索栏（ElInput + ElSelect + 搜索按钮 + 重置按钮）
2. 数据表格（ElTable + 分页 ElPagination）
3. 操作列：「详情」跳转到详情页；「封禁/解封」或「下架」弹出 ElMessageBox 确认后执行

**详情页通用结构**
- ElDescriptions 展示字段
- 页面顶部「返回列表」按钮
- 操作按钮（封禁/解封 或 下架）放在右上角

### Token 管理
- 登录成功后将 token 存入 `localStorage`，Pinia store 同步持有
- Axios 拦截器从 store 读取 token 注入请求头
- 响应拦截器：收到 401 时清空 token 并跳转 `/login`
- 刷新页面时从 `localStorage` 恢复 token 到 store（main.ts 初始化）

---

## 开发端口约定

| 服务 | 端口 |
|------|------|
| mall-server | 10088 |
| mall-mini devtools | — |
| mall-admin-web (dev) | 5174 |

Vite 配置 proxy 将 `/admin/*` 转发到 `http://localhost:10088`，避免跨域问题。

---

## 不在范围内

- 管理员权限分级（超级管理员 / 普通管理员）
- 统计看板 / 数据报表
- 商品内容编辑
- 删除用户 / 删除商品
- 操作日志
