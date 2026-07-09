import SwiftUI
import PhotosUI

struct PublishView: View {
    @State private var session = UserSession.shared

    var body: some View {
        NavigationStack {
            Group {
                if session.isLoggedIn {
                    PublishFormView()
                } else {
                    ScrollView {
                        Text("发布商品需要登录")
                            .font(.subheadline)
                            .foregroundStyle(.secondary)
                            .padding(.top, 24)
                        AuthFormView()
                    }
                }
            }
            .background(Color(.systemGray6))
            .navigationTitle("发布")
            .navigationBarTitleDisplayMode(.inline)
        }
    }
}

/// 已选图片：保留压缩后的 JPEG 数据；上传成功后记录 remoteURL，重试时跳过重复上传
private struct PickedImage: Identifiable {
    let id = UUID()
    let thumbnail: UIImage
    let jpegData: Data
    var remoteURL: String?
}

struct PublishFormView: View {
    private static let maxImages = 9
    private let categories = ["电子产品", "服装鞋帽", "图书文具", "生活用品", "数码配件", "其他"]

    @State private var images: [PickedImage] = []
    @State private var photoSelections: [PhotosPickerItem] = []
    @State private var description = ""
    @State private var price = ""
    @State private var category = "电子产品"
    @State private var provinceIndex = 0
    @State private var cityIndex = 0
    @State private var districtIndex = 0
    @State private var contactType = "phone"
    @State private var contactValue = ""
    @State private var submitting = false
    @State private var progressText = ""
    @State private var toast: String?

    private let contactTypes = [("phone", "手机号"), ("wechat", "微信"), ("qq", "QQ")]

    private var provinces: [Province] { ChinaRegions.provinces }
    private var cities: [City] { provinces.isEmpty ? [] : provinces[provinceIndex].children }
    private var districts: [String] { cities.isEmpty ? [] : cities[cityIndex].children }

    /// 与小程序一致：省市同名（直辖市）时地点省略市
    private var location: String {
        guard !provinces.isEmpty, !cities.isEmpty, !districts.isEmpty else { return "" }
        let province = provinces[provinceIndex].name
        let city = cities[cityIndex].name
        let district = districts[districtIndex]
        return province == city ? "\(province)\(district)" : "\(province)\(city)\(district)"
    }

    var body: some View {
        ScrollView {
            VStack(spacing: 12) {
                imageSection
                descriptionSection
                fieldsSection
                contactSection
                submitButton
            }
            .padding(12)
        }
        .scrollDismissesKeyboard(.interactively)
        .overlay(alignment: .bottom) {
            if let toast {
                toastView(toast)
            }
        }
        #if DEBUG
        // 供 UI 调试/自动化验证整条发布链路（生成图片→七牛直传→发布接口）：
        // SIMCTL_CHILD_DEBUG_AUTO_PUBLISH=1
        .onAppear {
            guard ProcessInfo.processInfo.environment["DEBUG_AUTO_PUBLISH"] == "1",
                  images.isEmpty, !submitting else { return }
            let size = CGSize(width: 400, height: 400)
            let image = UIGraphicsImageRenderer(size: size).image { context in
                UIColor(red: 7 / 255.0, green: 193 / 255.0, blue: 96 / 255.0, alpha: 1).setFill()
                context.fill(CGRect(origin: .zero, size: size))
            }
            guard let jpeg = image.jpegData(compressionQuality: 0.8),
                  let thumbnail = UIImage(data: jpeg) else { return }
            images = [PickedImage(thumbnail: thumbnail, jpegData: jpeg)]
            description = "iOS自动化测试商品（可删除）"
            price = "9.9"
            contactValue = "13800000000"
            Task { await submit() }
        }
        #endif
    }

    // MARK: - 图片

    private var imageSection: some View {
        VStack(alignment: .leading, spacing: 10) {
            Text("商品图片（\(images.count)/\(Self.maxImages)）")
                .font(.headline)

            LazyVGrid(columns: Array(repeating: GridItem(.flexible(), spacing: 8), count: 3), spacing: 8) {
                ForEach(images) { image in
                    imageCell(image)
                }
                if images.count < Self.maxImages {
                    PhotosPicker(
                        selection: $photoSelections,
                        maxSelectionCount: Self.maxImages - images.count,
                        matching: .images
                    ) {
                        RoundedRectangle(cornerRadius: 8)
                            .fill(Color(.systemGray6))
                            .aspectRatio(1, contentMode: .fit)
                            .overlay {
                                Image(systemName: "plus")
                                    .font(.title2)
                                    .foregroundStyle(.secondary)
                            }
                    }
                }
            }
        }
        .padding(16)
        .background(Color(.systemBackground))
        .clipShape(RoundedRectangle(cornerRadius: 12))
        .onChange(of: photoSelections) { _, newItems in
            guard !newItems.isEmpty else { return }
            Task { await loadPickedPhotos(newItems) }
        }
    }

    private func imageCell(_ image: PickedImage) -> some View {
        Image(uiImage: image.thumbnail)
            .resizable()
            .scaledToFill()
            .aspectRatio(1, contentMode: .fit)
            .clipShape(RoundedRectangle(cornerRadius: 8))
            .overlay(alignment: .topTrailing) {
                Button {
                    images.removeAll { $0.id == image.id }
                } label: {
                    Image(systemName: "xmark.circle.fill")
                        .font(.title3)
                        .foregroundStyle(.white, .black.opacity(0.5))
                }
                .padding(4)
            }
    }

    private func loadPickedPhotos(_ items: [PhotosPickerItem]) async {
        for item in items {
            guard images.count < Self.maxImages,
                  let data = try? await item.loadTransferable(type: Data.self),
                  let jpeg = compressedJPEG(from: data),
                  let thumbnail = UIImage(data: jpeg) else { continue }
            images.append(PickedImage(thumbnail: thumbnail, jpegData: jpeg))
        }
        photoSelections = []
    }

    /// 等比缩到最长边 1280 并压为 JPEG，等价小程序 chooseMedia 的 sizeType: compressed
    private func compressedJPEG(from data: Data, maxDimension: CGFloat = 1280) -> Data? {
        guard let image = UIImage(data: data) else { return nil }
        let scale = min(1, maxDimension / max(image.size.width, image.size.height))
        guard scale < 1 else { return image.jpegData(compressionQuality: 0.8) }
        let newSize = CGSize(width: image.size.width * scale, height: image.size.height * scale)
        let resized = UIGraphicsImageRenderer(size: newSize).image { _ in
            image.draw(in: CGRect(origin: .zero, size: newSize))
        }
        return resized.jpegData(compressionQuality: 0.8)
    }

    // MARK: - 描述 / 价格 / 分类 / 地区

    private var descriptionSection: some View {
        VStack(alignment: .leading, spacing: 10) {
            Text("商品描述")
                .font(.headline)
            TextField("描述一下宝贝的品牌型号、新旧程度…", text: $description, axis: .vertical)
                .lineLimit(4...8)
        }
        .padding(16)
        .frame(maxWidth: .infinity, alignment: .leading)
        .background(Color(.systemBackground))
        .clipShape(RoundedRectangle(cornerRadius: 12))
    }

    private var fieldsSection: some View {
        VStack(spacing: 14) {
            HStack {
                Text("价格")
                Spacer()
                TextField("0.00", text: $price)
                    .keyboardType(.decimalPad)
                    .multilineTextAlignment(.trailing)
                    .frame(width: 140)
                Text("元")
                    .foregroundStyle(.secondary)
            }

            Divider()

            HStack {
                Text("分类")
                Spacer()
                Picker("分类", selection: $category) {
                    ForEach(categories, id: \.self) { Text($0) }
                }
                .pickerStyle(.menu)
                .tint(.primary)
            }

            Divider()

            VStack(alignment: .leading, spacing: 8) {
                Text("地区")
                regionPickers
            }
            .frame(maxWidth: .infinity, alignment: .leading)
        }
        .padding(16)
        .background(Color(.systemBackground))
        .clipShape(RoundedRectangle(cornerRadius: 12))
    }

    private var regionPickers: some View {
        HStack(spacing: 8) {
            Picker("省", selection: $provinceIndex) {
                ForEach(provinces.indices, id: \.self) { index in
                    Text(provinces[index].name).tag(index)
                }
            }
            .onChange(of: provinceIndex) {
                cityIndex = 0
                districtIndex = 0
            }

            Picker("市", selection: $cityIndex) {
                ForEach(cities.indices, id: \.self) { index in
                    Text(cities[index].name).tag(index)
                }
            }
            .onChange(of: cityIndex) {
                districtIndex = 0
            }

            Picker("区", selection: $districtIndex) {
                ForEach(districts.indices, id: \.self) { index in
                    Text(districts[index]).tag(index)
                }
            }
        }
        .pickerStyle(.menu)
        .tint(.primary)
        .labelsHidden()
        .fixedSize()
    }

    // MARK: - 联系方式

    private var contactSection: some View {
        VStack(alignment: .leading, spacing: 12) {
            Text("联系方式")
                .font(.headline)

            Picker("联系方式类型", selection: $contactType) {
                ForEach(contactTypes, id: \.0) { type in
                    Text(type.1).tag(type.0)
                }
            }
            .pickerStyle(.segmented)

            TextField("填写\(contactTypes.first { $0.0 == contactType }?.1 ?? "联系方式")", text: $contactValue)
                .keyboardType(contactType == "phone" ? .phonePad : .default)
                .padding(12)
                .background(Color(.systemGray6))
                .clipShape(RoundedRectangle(cornerRadius: 8))
        }
        .padding(16)
        .frame(maxWidth: .infinity, alignment: .leading)
        .background(Color(.systemBackground))
        .clipShape(RoundedRectangle(cornerRadius: 12))
    }

    // MARK: - 提交

    private var submitButton: some View {
        Button {
            Task { await submit() }
        } label: {
            Group {
                if submitting {
                    HStack(spacing: 8) {
                        ProgressView().tint(.white)
                        Text(progressText)
                    }
                } else {
                    Text("发布")
                }
            }
            .font(.headline)
            .foregroundStyle(.white)
            .frame(maxWidth: .infinity)
            .padding(.vertical, 14)
            .background(submitting ? Color(.systemGray3) : Color.mallGreen)
            .clipShape(RoundedRectangle(cornerRadius: 12))
        }
        .disabled(submitting)
        .padding(.top, 4)
    }

    private func validate() -> String? {
        if images.isEmpty { return "请至少上传一张图片" }
        if description.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty { return "请填写商品描述" }
        guard let priceValue = Double(price), priceValue > 0 else { return "请填写正确的价格" }
        if location.isEmpty { return "请选择交易地点" }
        if contactValue.trimmingCharacters(in: .whitespaces).isEmpty { return "请填写联系方式" }
        return nil
    }

    private func submit() async {
        if let message = validate() {
            toast = message
            return
        }

        submitting = true
        defer {
            submitting = false
            progressText = ""
        }

        do {
            // 1. 逐张上传图片到七牛（已传过的跳过）
            var imageURLs: [String] = []
            for index in images.indices {
                if let url = images[index].remoteURL {
                    imageURLs.append(url)
                    continue
                }
                progressText = "上传图片 \(index + 1)/\(images.count)"
                let url = try await QiniuUploader.upload(jpegData: images[index].jpegData)
                images[index].remoteURL = url
                imageURLs.append(url)
            }

            // 2. 发布商品（与小程序一致，取描述前50字作为标题）
            progressText = "发布中..."
            let trimmedDescription = description.trimmingCharacters(in: .whitespacesAndNewlines)
            let request = PublishProductRequest(
                title: String(trimmedDescription.prefix(50)),
                description: trimmedDescription,
                price: Double(price) ?? 0,
                location: location,
                category: category,
                province: provinces[provinceIndex].name,
                city: cities[cityIndex].name,
                district: districts[districtIndex],
                images: imageURLs,
                contactType: contactType,
                contactValue: contactValue.trimmingCharacters(in: .whitespaces)
            )
            try await ProductAPI.publish(request)

            // 3. 成功后清空表单
            images = []
            description = ""
            price = ""
            contactValue = ""
            toast = "发布成功"
        } catch {
            toast = error.localizedDescription
        }
    }

    private func toastView(_ message: String) -> some View {
        Text(message)
            .font(.subheadline)
            .foregroundStyle(.white)
            .padding(.horizontal, 16)
            .padding(.vertical, 10)
            .background(.black.opacity(0.75))
            .clipShape(Capsule())
            .padding(.bottom, 24)
            .task {
                try? await Task.sleep(for: .seconds(2))
                toast = nil
            }
    }
}

#Preview {
    PublishView()
}
