import SwiftUI

@MainActor
@Observable
final class HomeViewModel {
    var products: [Product] = []
    var loading = false
    var initialLoaded = false
    var errorMessage: String?

    var keyword = ""
    var selectedCategory = "全部分类"
    var selectedProvince = "全部地区"

    private var page = 1
    private var hasMore = true
    private let pageSize = 10

    /// 与小程序 pages/home/home.ts 的分类保持一致
    let categories = ["全部分类", "电子产品", "服装鞋帽", "图书文具", "生活用品", "数码配件", "其他"]
    let provinces = [
        "全部地区", "北京市", "天津市", "上海市", "重庆市",
        "河北省", "山西省", "辽宁省", "吉林省", "黑龙江省",
        "江苏省", "浙江省", "安徽省", "福建省", "江西省", "山东省",
        "河南省", "湖北省", "湖南省", "广东省", "海南省",
        "四川省", "贵州省", "云南省", "陕西省", "甘肃省", "青海省",
        "内蒙古自治区", "广西壮族自治区", "西藏自治区", "宁夏回族自治区", "新疆维吾尔自治区",
    ]

    func reload() async {
        page = 1
        hasMore = true
        await load(replace: true)
    }

    func loadMoreIfNeeded(current product: Product) async {
        guard product.id == products.last?.id, hasMore, !loading else { return }
        await load(replace: false)
    }

    private func load(replace: Bool) async {
        guard !loading else { return }
        loading = true
        defer {
            loading = false
            initialLoaded = true
        }

        do {
            let result = try await ProductAPI.search(
                keyword: keyword,
                category: selectedCategory == "全部分类" ? "" : selectedCategory,
                province: selectedProvince == "全部地区" ? "" : selectedProvince,
                page: page,
                pageSize: pageSize
            )
            products = replace ? result.list : products + result.list
            hasMore = result.list.count >= pageSize
            page += 1
            errorMessage = nil
        } catch {
            errorMessage = error.localizedDescription
            if replace { products = [] }
        }
    }
}

struct HomeView: View {
    @State private var viewModel = HomeViewModel()
    @State private var path = NavigationPath()

    var body: some View {
        NavigationStack(path: $path) {
            VStack(spacing: 0) {
                searchBar
                filterBar
                productList
            }
            .background(Color(.systemGray6))
            .navigationTitle("旧物新语")
            .navigationBarTitleDisplayMode(.inline)
            .navigationDestination(for: Int.self) { productID in
                ProductDetailView(productID: productID)
            }
        }
        .task {
            if !viewModel.initialLoaded {
                await viewModel.reload()
            }
        }
        #if DEBUG
        // 供 UI 调试/截图自动化直达详情页：SIMCTL_CHILD_DEBUG_OPEN_PRODUCT_ID=<id> simctl launch ...
        .onAppear {
            if let idString = ProcessInfo.processInfo.environment["DEBUG_OPEN_PRODUCT_ID"],
               let id = Int(idString) {
                path.append(id)
            }
        }
        #endif
    }

    private var searchBar: some View {
        HStack(spacing: 8) {
            Image(systemName: "magnifyingglass")
                .foregroundStyle(.secondary)
            TextField("搜索商品", text: $viewModel.keyword)
                .submitLabel(.search)
                .onSubmit { Task { await viewModel.reload() } }
            if !viewModel.keyword.isEmpty {
                Button {
                    viewModel.keyword = ""
                    Task { await viewModel.reload() }
                } label: {
                    Image(systemName: "xmark.circle.fill")
                        .foregroundStyle(.tertiary)
                }
            }
        }
        .padding(.horizontal, 14)
        .padding(.vertical, 10)
        .background(Color(.systemBackground))
        .clipShape(RoundedRectangle(cornerRadius: 10))
        .padding(.horizontal, 12)
        .padding(.vertical, 10)
        .background(Color(.systemGray6))
    }

    private var filterBar: some View {
        HStack(spacing: 12) {
            filterMenu(title: viewModel.selectedCategory, options: viewModel.categories) {
                viewModel.selectedCategory = $0
                Task { await viewModel.reload() }
            }
            filterMenu(title: viewModel.selectedProvince, options: viewModel.provinces) {
                viewModel.selectedProvince = $0
                Task { await viewModel.reload() }
            }
            Spacer()
        }
        .padding(.horizontal, 12)
        .padding(.bottom, 10)
    }

    private func filterMenu(title: String, options: [String], onSelect: @escaping (String) -> Void) -> some View {
        Menu {
            ForEach(options, id: \.self) { option in
                Button(option) { onSelect(option) }
            }
        } label: {
            HStack(spacing: 4) {
                Text(title)
                Image(systemName: "arrowtriangle.down.fill")
                    .font(.system(size: 8))
            }
            .font(.subheadline)
            .padding(.horizontal, 14)
            .padding(.vertical, 7)
            .background(Color(.systemBackground))
            .clipShape(Capsule())
        }
        .tint(.primary)
    }

    private var productList: some View {
        ScrollView {
            LazyVStack(spacing: 12) {
                if let message = viewModel.errorMessage, viewModel.products.isEmpty {
                    statusPlaceholder(icon: "wifi.exclamationmark", text: message, showRetry: true)
                } else if viewModel.products.isEmpty && viewModel.initialLoaded && !viewModel.loading {
                    statusPlaceholder(icon: "tray", text: "暂无相关商品", showRetry: false)
                } else {
                    ForEach(viewModel.products) { product in
                        NavigationLink(value: product.id) {
                            ProductCardView(product: product)
                        }
                        .buttonStyle(.plain)
                        .task { await viewModel.loadMoreIfNeeded(current: product) }
                    }
                    if viewModel.loading {
                        ProgressView()
                            .padding(.vertical, 16)
                    }
                }
            }
            .padding(.horizontal, 12)
            .padding(.vertical, 12)
        }
        .refreshable { await viewModel.reload() }
    }

    private func statusPlaceholder(icon: String, text: String, showRetry: Bool) -> some View {
        VStack(spacing: 12) {
            Image(systemName: icon)
                .font(.system(size: 40))
                .foregroundStyle(.tertiary)
            Text(text)
                .font(.subheadline)
                .foregroundStyle(.secondary)
            if showRetry {
                Button("重试") {
                    Task { await viewModel.reload() }
                }
                .buttonStyle(.bordered)
                .tint(.mallGreen)
            }
        }
        .frame(maxWidth: .infinity)
        .padding(.top, 100)
    }
}

struct ProductCardView: View {
    let product: Product

    var body: some View {
        HStack(alignment: .top, spacing: 16) {
            productImage

            VStack(alignment: .leading, spacing: 10) {
                Text(product.title)
                    .font(.title3.bold())
                    .foregroundStyle(.primary)
                    .lineLimit(1)

                HStack(spacing: 4) {
                    Image(systemName: "mappin")
                        .foregroundStyle(.red)
                    Text(product.location)
                    Text("卖家：\(product.seller)")
                }
                .font(.footnote)
                .foregroundStyle(.secondary)
                .lineLimit(1)
                .minimumScaleFactor(0.8)
            }
            .padding(.top, 4)

            Spacer(minLength: 0)
        }
        .padding(12)
        .frame(maxWidth: .infinity, alignment: .leading)
        .background(Color(.systemBackground))
        .clipShape(RoundedRectangle(cornerRadius: 12))
        .shadow(color: .black.opacity(0.04), radius: 4, y: 2)
    }

    private var productImage: some View {
        ZStack(alignment: .bottomLeading) {
            AsyncImage(url: product.coverURL) { phase in
                switch phase {
                case .success(let image):
                    image
                        .resizable()
                        .scaledToFill()
                default:
                    Color(.systemGray5)
                        .overlay {
                            Image(systemName: "photo")
                                .font(.system(size: 32))
                                .foregroundStyle(.secondary)
                        }
                }
            }
            .frame(width: 110, height: 110)
            .clipShape(RoundedRectangle(cornerRadius: 8))

            Text(product.priceText)
                .font(.subheadline.bold())
                .foregroundStyle(.white)
                .padding(.horizontal, 10)
                .padding(.vertical, 4)
                .background(Color.mallGreen)
                .clipShape(Capsule())
                .padding(6)
        }
    }
}

#Preview {
    HomeView()
}
