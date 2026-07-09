import Foundation

/// HTTP 请求封装，接口约定与 mall-mini/utils/request.ts 保持一致：
/// - BASE_URL 相同
/// - 响应统一为 { code, msg, data }，code == 0 表示成功
/// - 已登录时自动注入 Authorization: Bearer <token>
enum API {
    /// 与 mall-mini/utils/request.ts 的 BASE_URL 保持一致
   // static let baseURL = URL(string: "https://yangzhongyu.site")!
    static let baseURL = URL(string: "http://localhost:8080")!

    /// 登录后写入（对应小程序的 wx.setStorageSync('token', ...)）
    static var token: String? {
        get { UserDefaults.standard.string(forKey: "token") }
        set { UserDefaults.standard.set(newValue, forKey: "token") }
    }

    enum Error: LocalizedError {
        case invalidURL
        case business(code: Int, message: String)
        case emptyData

        var errorDescription: String? {
            switch self {
            case .invalidURL: return "请求地址错误"
            case .business(_, let message): return message
            case .emptyData: return "响应数据缺失"
            }
        }
    }

    private struct Response<T: Decodable>: Decodable {
        let code: Int
        let msg: String
        let data: T?
    }

    static func get<T: Decodable>(_ path: String, query: [String: String] = [:]) async throws -> T {
        try await request(path, method: "GET", query: query, body: Optional<Int>.none)
    }

    static func post<T: Decodable, Body: Encodable>(_ path: String, body: Body) async throws -> T {
        try await request(path, method: "POST", query: [:], body: body)
    }

    private static func request<T: Decodable, Body: Encodable>(
        _ path: String,
        method: String,
        query: [String: String],
        body: Body?
    ) async throws -> T {
        guard var components = URLComponents(
            url: baseURL.appending(path: path),
            resolvingAgainstBaseURL: false
        ) else { throw Error.invalidURL }

        if !query.isEmpty {
            components.queryItems = query.map { URLQueryItem(name: $0.key, value: $0.value) }
        }
        guard let url = components.url else { throw Error.invalidURL }

        var urlRequest = URLRequest(url: url)
        urlRequest.httpMethod = method
        if let token, !token.isEmpty {
            urlRequest.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization")
        }
        if let body {
            urlRequest.setValue("application/json", forHTTPHeaderField: "Content-Type")
            urlRequest.httpBody = try JSONEncoder().encode(body)
        }

        let (data, _) = try await URLSession.shared.data(for: urlRequest)
        let response = try JSONDecoder().decode(Response<T>.self, from: data)

        guard response.code == 0 else {
            throw Error.business(code: response.code, message: response.msg)
        }
        guard let payload = response.data else {
            throw Error.emptyData
        }
        return payload
    }
}
