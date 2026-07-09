import SwiftUI

@MainActor
@Observable
final class ProductDetailViewModel {
    let productID: Int
    var detail: ProductDetail?
    var loading = false
    var errorMessage: String?
    var toast: String?

    init(productID: Int) {
        self.productID = productID
    }

    func load() async {
        loading = true
        defer { loading = false }
        do {
            detail = try await ProductAPI.detail(id: productID)
            errorMessage = nil
        } catch {
            errorMessage = error.localizedDescription
        }
    }

    func toggleFavorite() async {
        guard detail != nil else { return }
        guard let token = API.token, !token.isEmpty else {
            toast = "请先登录"
            return
        }
        do {
            let isFavorited = try await FavoriteAPI.toggle(productID: productID)
            detail?.isFavorited = isFavorited
            toast = isFavorited ? "收藏成功" : "已取消收藏"
        } catch {
            toast = error.localizedDescription
        }
    }
}

struct ProductDetailView: View {
    @State private var viewModel: ProductDetailViewModel

    init(productID: Int) {
        _viewModel = State(initialValue: ProductDetailViewModel(productID: productID))
    }

    var body: some View {
        Group {
            if let detail = viewModel.detail {
                content(detail)
            } else if viewModel.loading {
                ProgressView("加载中...")
                    .frame(maxWidth: .infinity, maxHeight: .infinity)
            } else if let message = viewModel.errorMessage {
                errorView(message)
            }
        }
        .background(Color(.systemGray6))
        .navigationTitle("商品详情")
        .navigationBarTitleDisplayMode(.inline)
        .toolbarBackground(Color(.systemGray6), for: .navigationBar)
        .task { await viewModel.load() }
        .overlay(alignment: .bottom) {
            if let toast = viewModel.toast {
                toastView(toast)
            }
        }
    }

    private func content(_ detail: ProductDetail) -> some View {
        VStack(spacing: 0) {
            ScrollView {
                VStack(spacing: 12) {
                    imageGallery(detail)
                    infoCard(detail)
                    descriptionCard(detail)
                    sellerCard(detail)
                    if !detail.contactValue.isEmpty {
                        contactCard(detail)
                    }
                    Text("发布于 \(detail.createTimeText)")
                        .font(.footnote)
                        .foregroundStyle(.tertiary)
                        .padding(.vertical, 8)
                }
                .padding(12)
            }

            actionBar(detail)
        }
    }

    // MARK: - 图片轮播

    private func imageGallery(_ detail: ProductDetail) -> some View {
        Group {
            if detail.images.isEmpty {
                Color(.systemGray5)
                    .overlay {
                        Image(systemName: "photo")
                            .font(.system(size: 48))
                            .foregroundStyle(.secondary)
                    }
            } else {
                TabView {
                    ForEach(detail.images, id: \.self) { urlString in
                        AsyncImage(url: URL(string: urlString)) { phase in
                            switch phase {
                            case .success(let image):
                                image
                                    .resizable()
                                    .scaledToFill()
                            default:
                                Color(.systemGray5)
                                    .overlay { ProgressView() }
                            }
                        }
                    }
                }
                .tabViewStyle(.page)
                .indexViewStyle(.page(backgroundDisplayMode: .always))
            }
        }
        .frame(height: 320)
        .frame(maxWidth: .infinity)
        .clipShape(RoundedRectangle(cornerRadius: 12))
    }

    // MARK: - 标题 / 价格 / 基本信息

    private func infoCard(_ detail: ProductDetail) -> some View {
        VStack(alignment: .leading, spacing: 12) {
            HStack(alignment: .top) {
                Text(detail.title)
                    .font(.title2.bold())
                Spacer()
                if detail.status != 0 {
                    Text("已下架")
                        .font(.caption)
                        .foregroundStyle(.secondary)
                        .padding(.horizontal, 8)
                        .padding(.vertical, 4)
                        .background(Color(.systemGray5))
                        .clipShape(Capsule())
                }
            }

            Text(detail.priceText)
                .font(.title.bold())
                .foregroundStyle(Color.mallGreen)

            Divider()

            infoRow(label: "分类", value: detail.category.isEmpty ? "其他" : detail.category)
            infoRow(label: "地点", value: detail.location)
        }
        .padding(16)
        .frame(maxWidth: .infinity, alignment: .leading)
        .background(Color(.systemBackground))
        .clipShape(RoundedRectangle(cornerRadius: 12))
    }

    private func infoRow(label: String, value: String) -> some View {
        HStack(spacing: 16) {
            Text(label)
                .foregroundStyle(.secondary)
                .frame(width: 44, alignment: .leading)
            Text(value)
            Spacer()
        }
        .font(.subheadline)
    }

    // MARK: - 商品描述

    private func descriptionCard(_ detail: ProductDetail) -> some View {
        VStack(alignment: .leading, spacing: 10) {
            Text("商品描述")
                .font(.headline)
            Text(detail.description.isEmpty ? "暂无描述" : detail.description)
                .font(.subheadline)
                .foregroundStyle(.secondary)
                .frame(maxWidth: .infinity, alignment: .leading)
        }
        .padding(16)
        .background(Color(.systemBackground))
        .clipShape(RoundedRectangle(cornerRadius: 12))
    }

    // MARK: - 卖家信息

    private func sellerCard(_ detail: ProductDetail) -> some View {
        VStack(alignment: .leading, spacing: 10) {
            Text("卖家信息")
                .font(.headline)
            HStack(spacing: 12) {
                AsyncImage(url: URL(string: detail.avatar)) { phase in
                    if case .success(let image) = phase {
                        image.resizable().scaledToFill()
                    } else {
                        Image(systemName: "person.crop.circle.fill")
                            .resizable()
                            .foregroundStyle(Color(.systemGray3))
                    }
                }
                .frame(width: 44, height: 44)
                .clipShape(Circle())

                Text(detail.seller)
                    .font(.subheadline.bold())
                Spacer()
            }
        }
        .padding(16)
        .frame(maxWidth: .infinity, alignment: .leading)
        .background(Color(.systemBackground))
        .clipShape(RoundedRectangle(cornerRadius: 12))
    }

    // MARK: - 联系卖家

    private func contactCard(_ detail: ProductDetail) -> some View {
        VStack(alignment: .leading, spacing: 10) {
            Text("联系卖家")
                .font(.headline)
            HStack(spacing: 10) {
                Text(detail.contactTypeText)
                    .font(.caption.bold())
                    .foregroundStyle(Color.mallGreen)
                    .padding(.horizontal, 8)
                    .padding(.vertical, 4)
                    .background(Color.mallGreen.opacity(0.12))
                    .clipShape(Capsule())

                if detail.contactType == "phone",
                   let url = URL(string: "tel:\(detail.contactValue)") {
                    Link(detail.contactValue, destination: url)
                        .font(.subheadline)
                        .tint(Color.mallGreen)
                } else {
                    Text(detail.contactValue)
                        .font(.subheadline)
                        .textSelection(.enabled)
                }
                Spacer()
            }
        }
        .padding(16)
        .frame(maxWidth: .infinity, alignment: .leading)
        .background(Color(.systemBackground))
        .clipShape(RoundedRectangle(cornerRadius: 12))
    }

    // MARK: - 底部操作栏

    private func actionBar(_ detail: ProductDetail) -> some View {
        Button {
            Task { await viewModel.toggleFavorite() }
        } label: {
            Label(
                detail.isFavorited ? "已收藏" : "收藏",
                systemImage: detail.isFavorited ? "heart.fill" : "heart"
            )
            .font(.headline)
            .foregroundStyle(.white)
            .frame(maxWidth: .infinity)
            .padding(.vertical, 14)
            .background(Color.mallGreen)
            .clipShape(RoundedRectangle(cornerRadius: 12))
        }
        .padding(.horizontal, 12)
        .padding(.top, 8)
        .background(Color(.systemBackground))
    }

    // MARK: - 错误 / Toast

    private func errorView(_ message: String) -> some View {
        VStack(spacing: 12) {
            Image(systemName: "wifi.exclamationmark")
                .font(.system(size: 40))
                .foregroundStyle(.tertiary)
            Text(message)
                .font(.subheadline)
                .foregroundStyle(.secondary)
            Button("重试") {
                Task { await viewModel.load() }
            }
            .buttonStyle(.bordered)
            .tint(.mallGreen)
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity)
    }

    private func toastView(_ message: String) -> some View {
        Text(message)
            .font(.subheadline)
            .foregroundStyle(.white)
            .padding(.horizontal, 16)
            .padding(.vertical, 10)
            .background(.black.opacity(0.75))
            .clipShape(Capsule())
            .padding(.bottom, 80)
            .task {
                try? await Task.sleep(for: .seconds(2))
                viewModel.toast = nil
            }
    }
}

#Preview {
    NavigationStack {
        ProductDetailView(productID: 6)
    }
}
