# 站内私信功能设计文档

**日期：** 2026-05-27  
**场景：** 买卖双方一对一议价沟通  
**规模：** 同时在线几千人  
**方案：** 自研 REST + 轮询，不依赖第三方 IM 服务

---

## 1. 功能范围

- 买家在商品详情页发起会话，向卖家发送文字消息
- 双方可在「消息」页查看所有会话及历史记录
- 进入小程序时检测未读消息数，在 tabBar 显示红点提示
- 聊天详情页每 3 秒轮询拉取新消息
- 不使用微信订阅消息，不依赖任何第三方推送服务

---

## 2. 架构总览

```
买家/卖家（小程序）                后端（Go/Gin）              数据库（SQLite）
       │                               │                           │
       │── POST /api/chat/send ──────▶ │── INSERT message ────────▶│
       │                               │── UPSERT conversation ───▶│
       │                               │                           │
       │  进入小程序 onShow()           │                           │
       │── GET /api/chat/unread-count ▶│◀── COUNT unread ─────────▶│
       │◀── { count: 3 } ─────────────│                           │
       │  tabBar 显示红点              │                           │
       │                               │                           │
       │  打开消息页                    │                           │
       │── GET /api/chat/conversations▶│◀── SELECT conversations ──▶│
       │                               │                           │
       │  进入对话（每 3s 轮询）        │                           │
       │── GET /api/chat/messages ────▶│◀── SELECT messages ───────▶│
       │── PUT /api/chat/read/:id ────▶│── UPDATE is_read ─────────▶│
```

---

## 3. 数据模型

### conversation（会话表）

| 字段         | 类型          | 约束                    | 说明               |
|--------------|---------------|-------------------------|--------------------|
| id           | INTEGER       | PK, AUTO_INCREMENT      |                    |
| product_id   | INTEGER       | NOT NULL                | 关联商品           |
| buyer_id     | INTEGER       | NOT NULL                | 发起询价的用户     |
| seller_id    | INTEGER       | NOT NULL                | 商品卖家           |
| last_message | VARCHAR(200)  | NOT NULL, DEFAULT ''    | 列表预览用         |
| last_at      | DATETIME      | NOT NULL                | 会话排序依据       |
| created_at   | DATETIME      | NOT NULL                |                    |

**唯一约束：** `UNIQUE(product_id, buyer_id)` — 同一商品买家只有一个会话

### message（消息表）

| 字段            | 类型         | 约束       | 说明                     |
|-----------------|--------------|------------|--------------------------|
| id              | INTEGER      | PK         |                          |
| conversation_id | INTEGER      | NOT NULL   | 所属会话                 |
| sender_id       | INTEGER      | NOT NULL   | 发送方用户 ID            |
| content         | VARCHAR(500) | NOT NULL   | 消息文字内容             |
| is_read         | BOOLEAN      | DEFAULT 0  | 接收方是否已读           |
| created_at      | DATETIME     | NOT NULL   |                          |

---

## 4. API 接口

所有接口需 JWT 认证（`Authorization: Bearer <token>`）。

### 4.1 发送消息

**POST** `/api/chat/send`

请求体：
```json
{
  "product_id": 1,
  "receiver_id": 2,
  "content": "还在吗？"
}
```

响应：
```json
{
  "conversation_id": 10,
  "message_id": 100
}
```

逻辑：
1. 根据 `product_id` 查出卖家，确定 buyer/seller 角色
2. `UPSERT` conversation（已存在则复用）
3. `INSERT` message，`is_read = false`
4. 更新 conversation 的 `last_message` 和 `last_at`

### 4.2 未读总数

**GET** `/api/chat/unread-count`

响应：
```json
{ "count": 3 }
```

逻辑：统计当前用户作为接收方（`sender_id != me`）且 `is_read = false` 的消息数。

### 4.3 会话列表

**GET** `/api/chat/conversations`

响应：
```json
[
  {
    "conversation_id": 10,
    "product": { "id": 1, "title": "二手iPhone", "cover": "url" },
    "other_user": { "id": 2, "nickname": "小明", "avatar": "url" },
    "last_message": "还在吗？",
    "last_at": "2026-05-27T10:00:00Z",
    "unread_count": 2
  }
]
```

按 `last_at DESC` 排序。

### 4.4 消息列表（轮询）

**GET** `/api/chat/messages?conv_id=10&last_id=99`

- `last_id`：客户端已有的最大消息 id，只返回 `id > last_id` 的消息
- 首次进入传 `last_id=0`，返回最近 50 条

响应：
```json
[
  {
    "id": 100,
    "sender_id": 2,
    "content": "还在吗？",
    "created_at": "2026-05-27T10:00:00Z"
  }
]
```

### 4.5 标记已读

**PUT** `/api/chat/read/:conv_id`

将该会话内所有 `sender_id != me` 且 `is_read = false` 的消息置为已读。

响应：`{ "ok": true }`

---

## 5. 前端新增页面

### 5.1 消息列表页 `pages/chat-list/`

- 加入 tabBar（消息图标）
- 展示所有会话，按最新消息排序
- 每个会话显示：商品封面、对方昵称、最后消息预览、未读红点

### 5.2 聊天详情页 `pages/chat/`

- 路由参数：`conversation_id` 或 `product_id + receiver_id`（首次发起时）
- 进入时调用标记已读接口
- 每 3 秒调用 `/api/chat/messages?conv_id=&last_id=` 拉取增量消息
- 离开页面时清除轮询定时器

### 5.3 进入小程序提示

在 `app.ts` 的 `onShow()` 中：
```
GET /api/chat/unread-count
→ count > 0：wx.setTabBarBadge({ index: 消息tab索引, text: count })
→ count = 0：wx.removeTabBarBadge
```

---

## 6. 后端路由变更

在 `router.go` 的认证路由组中新增：

```
POST   /api/chat/send
GET    /api/chat/unread-count
GET    /api/chat/conversations
GET    /api/chat/messages
PUT    /api/chat/read/:conv_id
```

---

## 7. 不在范围内

- 图片/语音消息（仅文字）
- 消息撤回
- 多人群聊
- 微信订阅消息推送
- 消息加密存储
