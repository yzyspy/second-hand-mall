import Foundation

/// 对应 mall-server GET /api/product/detail 返回的 dao.ProductDetail
struct ProductDetail: Decodable {
    let id: Int
    let userID: Int
    let title: String
    let description: String
    let price: Double
    let images: [String]
    let location: String
    let status: Int
    let buyUid: Int
    let category: String
    let seller: String
    let avatar: String
    let createTime: String
    let contactType: String
    let contactValue: String
    var isFavorited: Bool

    enum CodingKeys: String, CodingKey {
        case id, title, description, price, images, location, status, category, seller, avatar
        case userID = "user_id"
        case buyUid = "buy_uid"
        case createTime = "create_time"
        case contactType = "contact_type"
        case contactValue = "contact_value"
        case isFavorited = "is_favorited"
    }

    init(from decoder: Decoder) throws {
        let c = try decoder.container(keyedBy: CodingKeys.self)
        id = try c.decode(Int.self, forKey: .id)
        userID = try c.decodeIfPresent(Int.self, forKey: .userID) ?? 0
        title = try c.decode(String.self, forKey: .title)
        description = try c.decodeIfPresent(String.self, forKey: .description) ?? ""
        price = try c.decode(Double.self, forKey: .price)
        location = try c.decodeIfPresent(String.self, forKey: .location) ?? ""
        status = try c.decodeIfPresent(Int.self, forKey: .status) ?? 0
        buyUid = try c.decodeIfPresent(Int.self, forKey: .buyUid) ?? 0
        category = try c.decodeIfPresent(String.self, forKey: .category) ?? ""
        avatar = try c.decodeIfPresent(String.self, forKey: .avatar) ?? ""
        createTime = try c.decodeIfPresent(String.self, forKey: .createTime) ?? ""
        contactType = try c.decodeIfPresent(String.self, forKey: .contactType) ?? ""
        contactValue = try c.decodeIfPresent(String.self, forKey: .contactValue) ?? ""
        isFavorited = try c.decodeIfPresent(Bool.self, forKey: .isFavorited) ?? false

        let sellerName = try c.decodeIfPresent(String.self, forKey: .seller) ?? ""
        seller = sellerName.isEmpty ? "匿名用户" : sellerName

        let imagesString = try c.decodeIfPresent(String.self, forKey: .images) ?? ""
        images = imagesString.split(separator: ",").map(String.init).filter { !$0.isEmpty }
    }

    var priceText: String {
        price.truncatingRemainder(dividingBy: 1) == 0
            ? "¥\(Int(price))"
            : String(format: "¥%.2f", price)
    }

    var contactTypeText: String {
        switch contactType {
        case "phone": return "手机号"
        case "wechat": return "微信"
        default: return "QQ"
        }
    }

    /// "2026-05-21T19:31:56.508557+08:00" → "2026-05-21 19:31"
    var createTimeText: String {
        let parser = ISO8601DateFormatter()
        parser.formatOptions = [.withInternetDateTime, .withFractionalSeconds]
        guard let date = parser.date(from: createTime)
            ?? ISO8601DateFormatter().date(from: createTime) else { return createTime }
        let formatter = DateFormatter()
        formatter.dateFormat = "yyyy-MM-dd HH:mm"
        return formatter.string(from: date)
    }
}

extension ProductAPI {
    /// GET /api/product/detail（公开接口，带 token 时返回 is_favorited）
    static func detail(id: Int) async throws -> ProductDetail {
        try await API.get("/api/product/detail", query: ["id": "\(id)"])
    }
}

enum FavoriteAPI {
    private struct ToggleBody: Encodable {
        let productID: Int
        enum CodingKeys: String, CodingKey { case productID = "product_id" }
    }

    private struct ToggleResult: Decodable {
        let isFavorited: Bool
        enum CodingKeys: String, CodingKey { case isFavorited = "is_favorited" }
    }

    /// POST /api/favorite/toggle（需要登录），返回操作后的收藏状态
    static func toggle(productID: Int) async throws -> Bool {
        let result: ToggleResult = try await API.post("/api/favorite/toggle", body: ToggleBody(productID: productID))
        return result.isFavorited
    }
}
