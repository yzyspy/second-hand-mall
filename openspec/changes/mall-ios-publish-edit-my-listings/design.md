## Context

mall-mini 的发布（`publish.ts`）与编辑（`productEdit.ts`）几乎是同一份表单：图片选择/上传、描述、价格、地区三级联动、分类、联系方式；区别仅在于编辑页需要先加载已有数据预填，且提交走 `PUT /api/product/update` 而非 `POST /api/product/publish`。图片上传统一走七牛云：前端先调用 `POST /api/upload/qiniu-token` 拿到 `{uploadKey, upToken, domain, uploadUrl}`，再直接向七牛云 `uploadUrl` 发起 multipart 上传。`myPublish.ts` 是独立的列表页，通过 `GET /api/product/mine` 分页拉取当前用户商品，操作（标记售出/下架/删除）统一调用 `POST /api/product/change-status`，"删除"在后端语义上其实是下架（`status=2`），不是物理删除。

`change #2` 已定义 `Product` 模型（列表项字段）与省市区/分类静态数据，本 change 复用；但发布/编辑提交的商品字段（`title/description/price/location/category/province/city/district/images/contact_type/contact_value`）比列表项模型更完整，需要单独的 `ProductDraft`/`ProductFormInput` 结构，不与浏览用的 `Product` 模型混用。

## Goals / Non-Goals

**Goals:**
- 发布表单：多图选择（最多 9 张）+ 上传进度反馈、描述/价格/分类/地区/联系方式录入与校验、提交成功后清空表单。
- 编辑表单：复用发布表单 UI/校验逻辑，预填已有商品数据，提交走更新接口。
- 我的发布列表：分页展示、状态标签、编辑/标记售出/下架/删除操作，操作后刷新列表。
- 未登录用户进入发布 Tab 时提示登录（对齐 mall-mini `onShow` 的登录检查）。

**Non-Goals:**
- 不做图片裁剪、滤镜等图片编辑功能（mall-mini 也没有）。
- 不做发布内容审核提示（后端无此机制）。
- 不实现真正的物理删除接口（后端没有提供，"删除"操作按 mall-mini 的实际行为调用 `change-status(2)`）。
- 不做草稿箱/离线保存未提交表单的功能。

## Decisions

### 1. 发布与编辑共享一个表单组件
- 新增 `Features/Publish/View/ProductFormView.swift`（可复用视图）+ `Features/Publish/ViewModel/ProductFormViewModel.swift`（可复用逻辑：图片管理、上传、字段校验、地区联动），`PublishView` 直接使用空白初始状态，`ProductEditView` 用加载到的商品数据初始化同一个 ViewModel，仅在提交时分流调用 `publish` 或 `update` 接口。
- **备选方案**：编辑页整个复制一份发布页代码。放弃：两者字段和校验规则完全一致，未来任一方调整字段都要同步改两份，维护成本高；mall-mini 本身也是这种"复制+改提交接口"的模式，iOS 端用共享组件规避该重复。

### 2. 图片上传：`QiniuUploader`
```swift
struct QiniuTokenResponse: Decodable {
    let uploadKey: String
    let upToken: String
    let domain: String
    let uploadUrl: String
}

enum QiniuUploader {
    static func upload(imageData: Data, key: String? = nil) async throws -> String  // 返回可访问 URL
}
```
- 先调用 `APIClient.request` 获取 `QiniuTokenResponse`（`POST /api/upload/qiniu-token`，公开接口），再用 `URLSession` 发起 `multipart/form-data` 请求（字段 `key`/`token`/`file`）到 `uploadUrl`，成功后拼接 `domain/uploadKey` 作为图片 URL（对齐 `qiniu-upload.ts` 的实现）。
- 多图上传按顺序逐张上传（对齐 mall-mini 的 `for` 循环行为，非并发），每张独立展示上传中/失败状态，允许单张失败后重试而不影响已上传的图片。
- **备选方案**：并发上传所有图片。放弃：mall-mini 是顺序上传，改成并发在没有并发数限制的情况下可能对七牛云或用户网络造成突发压力，且当前项目没有需要优化上传速度的明确诉求，遵循现状更稳妥。

### 3. 我的发布列表操作的二次确认
- 「标记售出」「下架」「删除」均先弹出 `.confirmationDialog` 或 `.alert` 二次确认（对齐 mall-mini 的 `wx.showModal`），确认后调用 `POST /api/product/change-status`，成功后重置分页重新加载列表（而非本地乐观更新单条状态），与 mall-mini 行为一致，避免本地状态与服务端不一致。
- **备选方案**：本地乐观更新单条状态，不重新拉取列表。放弃：mall-mini 选择重新拉取整个列表以保证一致性，改动行为超出本 change 范围。

## Risks / Trade-offs

- **[风险] 逐张顺序上传图片，多图发布时耗时较长** → 缓解：与 mall-mini 行为保持一致（用户预期一致），且展示逐张上传进度，用户可感知进度而非长时间无反馈。
- **[风险] 「删除」操作实际只是下架（`status=2`），与用户预期的"永久删除"不符** → 缓解：这是后端已有限制，UI 文案沿用 mall-mini 的"确认删除"提示，不承诺物理删除，行为对齐现有小程序，不在本 change 修改后端语义。
- **[权衡] 编辑与发布共享 ViewModel 而非独立实现** → 影响：初期实现复杂度略高于两份独立代码，但避免了后续维护两份几乎相同表单逻辑的成本。

## Migration Plan

无迁移概念，全部为新增视图与网络对接，不涉及已发布数据迁移。

## Open Questions

无遗留未决问题。
