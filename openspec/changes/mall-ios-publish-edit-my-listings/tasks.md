## 1. 图片上传

- [ ] 1.1 新增 `Core/Network/QiniuUploader.swift`：获取上传凭证 + multipart 上传 + 拼接可访问 URL
- [ ] 1.2 支持单张图片上传失败后重试，不影响其他图片状态

## 2. 共享发布表单组件

- [ ] 2.1 新增 `Features/Publish/ViewModel/ProductFormViewModel.swift`：图片管理、字段状态、地区联动、校验逻辑
- [ ] 2.2 新增 `Features/Publish/View/ProductFormView.swift`：图片九宫格选择/预览/删除、描述/价格/分类/地区/联系方式输入 UI
- [ ] 2.3 实现表单校验（图片/描述/价格/地点/联系方式）与错误提示

## 3. 发布

- [ ] 3.1 重写 `Features/Publish/View/PublishView.swift`：未登录拦截 + 使用共享表单组件（空白初始状态）
- [ ] 3.2 实现提交逻辑：上传全部图片 → 组装商品数据 → `POST /api/product/publish` → 成功后清空表单

## 4. 编辑

- [ ] 4.1 新增 `Features/ProductEdit/ViewModel/ProductEditViewModel.swift`：加载详情预填共享表单、提交走更新接口
- [ ] 4.2 新增 `Features/ProductEdit/View/ProductEditView.swift`：复用共享表单组件展示预填数据
- [ ] 4.3 实现提交逻辑：`PUT /api/product/update` → 成功后返回上一页

## 5. 我的发布

- [ ] 5.1 新增 `Features/MyPublish/ViewModel/MyPublishViewModel.swift`：分页加载 `/api/product/mine`、状态标签映射
- [ ] 5.2 新增 `Features/MyPublish/View/MyPublishView.swift`：列表展示、下拉刷新、上拉加载
- [ ] 5.3 实现编辑/标记售出/下架/删除四个操作入口及二次确认弹窗，操作后刷新列表
- [ ] 5.4 「我的」Tab 加入「我发布的」入口，跳转到 `MyPublishView`

## 6. 测试

- [ ] 6.1 测试 `ProductFormViewModel` 校验逻辑（各必填字段缺失场景）
- [ ] 6.2 测试图片上传成功/失败状态流转
- [ ] 6.3 测试 `MyPublishViewModel` 分页追加与状态标签映射
- [ ] 6.4 测试编辑页预填逻辑（详情数据 → 表单字段映射）

## 7. 验证

- [ ] 7.1 `xcodebuild build` 或等效命令通过
- [ ] 7.2 运行单元测试全部通过
- [ ] 7.3 手动验证：登录 → 发布带图商品成功 → 「我发布的」看到该商品 → 编辑并保存成功 → 标记售出/下架/删除各操作生效并刷新列表
