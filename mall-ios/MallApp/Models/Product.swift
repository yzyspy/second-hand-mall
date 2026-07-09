import Foundation

/// 对应 mall-server GET /api/product/search 返回的商品项
/// （dao.ProductSearchResult，与小程序 pages/home/home.ts 的映射逻辑一致）
struct Product: Identifiable, Decodable {
    let id: Int
    let title: String
    let price: Double
    let location: String
    let category: String
    let seller: String
    let avatar: String
    let images: [String]
    let buyUid: Int
    let createTime: String

    enum CodingKeys: String, CodingKey {
        case id, title, price, images, location, category, seller, avatar
        case buyUid = "buy_uid"
        case createTime = "create_time"
    }

    init(from decoder: Decoder) throws {
        let c = try decoder.container(keyedBy: CodingKeys.self)
        id = try c.decode(Int.self, forKey: .id)
        title = try c.decode(String.self, forKey: .title)
        price = try c.decode(Double.self, forKey: .price)
        location = try c.decodeIfPresent(String.self, forKey: .location) ?? ""
        category = try c.decodeIfPresent(String.self, forKey: .category) ?? ""
        avatar = try c.decodeIfPresent(String.self, forKey: .avatar) ?? ""
        buyUid = try c.decodeIfPresent(Int.self, forKey: .buyUid) ?? 0
        createTime = try c.decodeIfPresent(String.self, forKey: .createTime) ?? ""

        // 卖家可能为空（LEFT JOIN），兜底为匿名用户，与小程序一致
        let sellerName = try c.decodeIfPresent(String.self, forKey: .seller) ?? ""
        seller = sellerName.isEmpty ? "匿名用户" : sellerName

        // 服务端 images 为逗号分隔的 URL 字符串
        let imagesString = try c.decodeIfPresent(String.self, forKey: .images) ?? ""
        images = imagesString.split(separator: ",").map(String.init).filter { !$0.isEmpty }
    }

    init(id: Int, title: String, price: Double, location: String, category: String = "",
         seller: String = "匿名用户", avatar: String = "", images: [String] = [],
         buyUid: Int = 0, createTime: String = "") {
        self.id = id
        self.title = title
        self.price = price
        self.location = location
        self.category = category
        self.seller = seller
        self.avatar = avatar
        self.images = images
        self.buyUid = buyUid
        self.createTime = createTime
    }

    var coverURL: URL? {
        images.first.flatMap(URL.init(string:))
    }

    var priceText: String {
        price.truncatingRemainder(dividingBy: 1) == 0
            ? "¥\(Int(price))"
            : String(format: "¥%.2f", price)
    }
}

struct ProductPage: Decodable {
    let list: [Product]
    let total: Int
    let page: Int
    let pageSize: Int

    enum CodingKeys: String, CodingKey {
        case list, total, page
        case pageSize = "page_size"
    }
}

/// POST /api/product/publish 请求体（对应服务端 PublishProductRequest）
struct PublishProductRequest: Encodable {
    let title: String
    let description: String
    let price: Double
    let location: String
    let category: String
    let province: String
    let city: String
    let district: String
    let images: [String]
    let contactType: String
    let contactValue: String

    enum CodingKeys: String, CodingKey {
        case title, description, price, location, category, province, city, district, images
        case contactType = "contact_type"
        case contactValue = "contact_value"
    }
}

enum ProductAPI {
    private struct PublishResult: Decodable {
        let id: Int
    }

    /// POST /api/product/publish（需要登录），返回新商品 id
    @discardableResult
    static func publish(_ request: PublishProductRequest) async throws -> Int {
        let result: PublishResult = try await API.post("/api/product/publish", body: request)
        return result.id
    }

    /// GET /api/product/search（公开接口）
    static func search(
        keyword: String = "",
        category: String = "",
        province: String = "",
        page: Int,
        pageSize: Int = 10
    ) async throws -> ProductPage {
        var query = [
            "page": "\(page)",
            "page_size": "\(pageSize)",
            "status": "0",
        ]
        if !keyword.isEmpty { query["keyword"] = keyword }
        if !category.isEmpty { query["category"] = category }
        if !province.isEmpty { query["province"] = province }
        return try await API.get("/api/product/search", query: query)
    }
}

extension Product {
    /// 仅用于 SwiftUI 预览
    static let previewList: [Product] = [
        Product(id: 1, title: "游戏卡", price: 10, location: "天津市和平区"),
        Product(id: 2, title: "电视机", price: 10, location: "北京市东城区"),
        Product(id: 3, title: "ipad", price: 11, location: "北京市东城区"),
        Product(id: 4, title: "iphone17", price: 100, location: "河北省廊坊市广阳区"),
    ]
}
