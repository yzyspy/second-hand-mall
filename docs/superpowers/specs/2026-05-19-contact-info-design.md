# 商品联系方式字段设计

**Date:** 2026-05-19
**Status:** Approved

## Problem

买家在商品详情页无法获取卖家联系方式，`contactSeller()` 仅显示"功能开发中"。二手交易平台缺少最基础的买卖双方沟通入口。

## Goal

卖家发布商品时必填联系方式（手机号/微信/QQ），买家在详情页以纯文字形式查看联系方式。

## Decisions

- 联系方式**必填**，发布和编辑时均需填写
- 详情页**仅展示文字**，不触发拨号/跳转等交互
- 每个商品支持**一种**联系方式（类型 + 值）
- 数据存储用**两个独立字段**：`contact_type` + `contact_value`

## Architecture

### Files Changed

| 文件 | 变更 |
|------|------|
| `mall-server/internal/app/dao/product.entity.go` | 新增 `ContactType`、`ContactValue` 字段 |
| `mall-server/internal/app/dao/product.repo.go` | `ProductDetail` 结构体加两字段；`GetProductByID` SELECT 语句追加两列 |
| `mall-server/internal/app/service/types.go` | `PublishProductRequest`、`UpdateProductRequest` 各加两字段（binding required） |
| `mall-server/internal/app/service/product.go` | `PublishProduct`、`UpdateProduct` handler 写入两个新字段 |
| `mall-mini/miniprogram/pages/publish/publish.ts` | 新增 `contactType`/`contactValue` state、校验、提交逻辑 |
| `mall-mini/miniprogram/pages/publish/publish.wxml` | 新增联系方式单选 + 输入区块 |
| `mall-mini/miniprogram/pages/detail/detail.ts` | `Product` interface 加两字段；`loadProductDetail` 映射；移除 contactSeller toast |
| `mall-mini/miniprogram/pages/detail/detail.wxml` | 新增联系方式展示区块 |
| `mall-mini/miniprogram/pages/productEdit/productEdit.ts` | 同 publish：新增 state、回填、校验、提交 |
| `mall-mini/miniprogram/pages/productEdit/productEdit.wxml` | 同 publish：新增联系方式单选 + 输入区块 |

### 数据模型

```go
// product.entity.go 新增字段
ContactType  string `gorm:"column:contact_type;type:varchar(10);not null;default:''"  json:"contact_type"`
ContactValue string `gorm:"column:contact_value;type:varchar(100);not null;default:''" json:"contact_value"`
```

`contact_type` 合法值：`phone` / `wechat` / `qq`

GORM AutoMigrate 在服务启动时自动为 SQLite 添加新列，无需手写迁移脚本。

### 后端请求/响应 DTO

`PublishProductRequest` 和 `UpdateProductRequest` 均新增：
```go
ContactType  string `json:"contact_type"  binding:"required,oneof=phone wechat qq"`
ContactValue string `json:"contact_value" binding:"required"`
```

`ProductDetail` 新增：
```go
ContactType  string `json:"contact_type"`
ContactValue string `json:"contact_value"`
```

`GetProductByID` 的 SELECT 追加 `product.contact_type, product.contact_value`。

### 前端发布页 & 编辑页

`PublishData` / `ProductEditData` 新增：
```typescript
contactType: 'phone' | 'wechat' | 'qq'   // 默认 'phone'
contactValue: string                        // 默认 ''
contactTypes: string[]                      // ['手机号', '微信', 'QQ']
```

UI 区块（插在价格字段下方）：
- 三个单选按钮横排：手机号 / 微信 / QQ
- 选中项高亮边框
- 下方 input，placeholder 随类型动态变化：
  - phone → "请输入手机号"
  - wechat → "请输入微信号"
  - qq → "请输入QQ号"

校验（`validateForm()`）：`contactValue` 为空时 Toast "请填写联系方式"。

提交时在 `productData` 中加入 `contact_type` + `contact_value`。

编辑页在 `onLoad` 时从详情接口回填 `contact_type` / `contact_value`。

### 前端详情页

`Product` interface 新增：
```typescript
contactType: string
contactValue: string
```

`loadProductDetail` 映射：
```typescript
contactType: data.contact_type || '',
contactValue: data.contact_value || '',
```

UI 在卖家信息区块下方新增"联系卖家"区块：
```
联系卖家
[类型标签]  具体号码
```
类型标签：`phone` → "手机号"，`wechat` → "微信"，`qq` → "QQ"。
纯文字展示，无点击交互。

WXML 中删除"联系卖家"按钮，`contactSeller()` 方法同步从 `detail.ts` 中删除。

## Error Handling

| 场景 | 行为 |
|------|------|
| 发布/编辑时 contactValue 为空 | Toast "请填写联系方式"，阻止提交 |
| contact_type 值非法（后端校验） | 返回 `code: -1, msg: "参数错误"` |
| 老商品 contact_type/value 为空字符串 | 详情页联系方式区块不渲染（`wx:if="{{product.contactValue}}"` 判断） |

## Testing

1. 发布页不填联系方式直接提交 → Toast "请填写联系方式"
2. 切换类型（手机号/微信/QQ）→ placeholder 随之变化
3. 填写联系方式后正常发布 → 进详情页验证类型标签和号码展示正确
4. 编辑商品 → 联系方式已回填，修改后保存验证更新成功
5. 老数据（contactValue 为空）→ 详情页联系方式区块不显示
