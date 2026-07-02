## 1. 详情页收藏真实交互

- [ ] 1.1 `ProductDetailViewModel` 新增 `toggleFavorite()`：调用 `POST /api/favorite/toggle`，未登录时提示登录不发请求
- [ ] 1.2 `ProductDetailView` 收藏按钮改为展示真实 `is_favorited` 状态并响应点击切换

## 2. 收藏列表

- [ ] 2.1 新增 `Features/Favorite/ViewModel/FavoriteViewModel.swift`：分页加载 `/api/favorite/list`、列表内取消收藏本地移除
- [ ] 2.2 新增 `Features/Favorite/View/FavoriteView.swift`：列表展示、下拉刷新、上拉加载、取消收藏操作、空状态
- [ ] 2.3 列表项点击跳转到商品详情页

## 3. 「我的」页面入口

- [ ] 3.1 「我的」页面加入「我的收藏」菜单项，未登录点击提示登录，已登录跳转到 `FavoriteView`

## 4. 测试

- [ ] 4.1 测试 `ProductDetailViewModel.toggleFavorite()`：登录态切换成功、未登录拦截
- [ ] 4.2 测试 `FavoriteViewModel`：分页追加、取消收藏后本地移除且 total 递减、取消收藏失败保留原项

## 5. 验证

- [ ] 5.1 `xcodebuild build` 或等效命令通过
- [ ] 5.2 运行单元测试全部通过
- [ ] 5.3 手动验证：详情页收藏/取消收藏 → 「我的收藏」看到/移除对应商品 → 未登录点击收藏与「我的收藏」入口均提示登录
