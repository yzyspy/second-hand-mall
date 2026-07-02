## 1. 共享数据模型

- [ ] 1.1 新增 `Core/Models/Product.swift`：列表项与详情模型，处理 `images` 逗号分隔字符串 → `[String]`
- [ ] 1.2 新增 `Core/Models/ProductCategory.swift`：分类常量数组
- [ ] 1.3 新增 `Core/Data/ChinaRegions.swift`：转换 `mall-mini/miniprogram/data/china-regions.ts` 为 Swift 静态省市区数据

## 2. 首页浏览

- [ ] 2.1 重写 `Features/Home/ViewModel/HomeViewModel.swift`：分页状态、筛选状态（分类/省/市/区）、`loadProducts()` 对接 `/api/product/search`
- [ ] 2.2 重写 `Features/Home/View/HomeView.swift`：商品卡片列表、下拉刷新、上拉加载、搜索入口按钮
- [ ] 2.3 实现分类筛选面板 UI 与交互
- [ ] 2.4 实现省市区三级联动地区筛选面板 UI 与交互

## 3. 搜索

- [ ] 3.1 新增 `Features/Search/ViewModel/SearchViewModel.swift`：关键字/排序/状态/分类/地区筛选状态与分页
- [ ] 3.2 新增 `Features/Search/View/SearchView.swift`：搜索输入框、排序与状态选择器、结果列表
- [ ] 3.3 复用首页的分类/地区筛选面板组件

## 4. 商品详情

- [ ] 4.1 新增 `Features/ProductDetail/ViewModel/ProductDetailViewModel.swift`：加载详情、错误状态
- [ ] 4.2 新增 `Features/ProductDetail/View/ProductDetailView.swift`：图片轮播、信息展示、收藏/联系卖家占位按钮（点击展示"即将上线"）
- [ ] 4.3 首页与搜索页商品卡片接入跳转到 `ProductDetailView`

## 5. 测试

- [ ] 5.1 测试 `Product` 模型的 `images` 字符串解析（含空字符串、含逗号分隔的正常场景）
- [ ] 5.2 测试 `HomeViewModel`：分页追加逻辑、筛选变化重置分页、`hasMore` 判断
- [ ] 5.3 测试 `SearchViewModel`：关键字/排序/状态组合查询参数拼装、基于 `total` 判断 `hasMore`
- [ ] 5.4 测试 `ProductDetailViewModel`：加载成功/失败状态

## 6. 验证

- [ ] 6.1 `xcodebuild build` 或等效命令通过
- [ ] 6.2 运行单元测试全部通过
- [ ] 6.3 手动验证：首页浏览 → 分类/地区筛选 → 下拉刷新/上拉加载 → 进入搜索 → 组合筛选搜索 → 点击结果进入详情页展示完整信息
