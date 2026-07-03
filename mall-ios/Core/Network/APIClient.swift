import Foundation

final class APIClient {
    static let shared = APIClient()

    private let session: URLSession
    private let baseURL: String

    init(session: URLSession = .shared, baseURL: String = "http://localhost:8080") {
        self.session = session
        self.baseURL = baseURL
    }

    /// 通用请求方法：解析统一信封 {code, msg, data}。
    /// - requiresAuth == true 时从 TokenStore 读取 token 注入 Authorization 头；
    ///   token 缺失时直接抛 .unauthorized，不发起网络请求。
    func request<T: Decodable>(
        _ path: String,
        method: HTTPMethod = .get,
        body: Encodable? = nil,
        requiresAuth: Bool = false
    ) async throws -> T {
        guard let url = URL(string: baseURL + path) else {
            throw APIError.transport(URLError(.badURL))
        }

        var urlRequest = URLRequest(url: url)
        urlRequest.httpMethod = method.rawValue
        urlRequest.setValue("application/json", forHTTPHeaderField: "Content-Type")

        if let body {
            urlRequest.httpBody = try JSONEncoder().encode(body)
        }

        if requiresAuth {
            guard let token = TokenStore.shared.getToken() else {
                throw APIError.unauthorized
            }
            urlRequest.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization")
        }

        let data: Data
        let response: URLResponse
        do {
            (data, response) = try await session.data(for: urlRequest)
        } catch {
            throw APIError.transport(error)
        }

        guard let http = response as? HTTPURLResponse else {
            throw APIError.transport(URLError(.badServerResponse))
        }

        if http.statusCode == 401 {
            throw APIError.unauthorized
        }

        let envelope: ApiResponse<T>
        do {
            envelope = try JSONDecoder().decode(ApiResponse<T>.self, from: data)
        } catch {
            throw APIError.decoding(error)
        }

        guard envelope.code == 0 else {
            throw APIError.server(code: envelope.code, msg: envelope.msg)
        }

        if let payload = envelope.data {
            return payload
        }
        if let empty = EmptyResponse() as? T {
            return empty
        }
        throw APIError.decoding(
            DecodingError.valueNotFound(T.self, DecodingError.Context(codingPath: [], debugDescription: "data 字段缺失"))
        )
    }

    /// /user/save 不遵循信封格式，直接返回 {"message": String}。
    /// 任何非 2xx 响应或缺少 message 字段均视为注册失败，不解析具体后端错误原因。
    func register(username: String, password: String) async throws -> String {
        guard let url = URL(string: baseURL + "/user/save") else {
            throw APIError.transport(URLError(.badURL))
        }

        var urlRequest = URLRequest(url: url)
        urlRequest.httpMethod = HTTPMethod.post.rawValue
        urlRequest.setValue("application/json", forHTTPHeaderField: "Content-Type")
        urlRequest.httpBody = try JSONEncoder().encode([
            "username": username,
            "password": password
        ])

        let data: Data
        let response: URLResponse
        do {
            (data, response) = try await session.data(for: urlRequest)
        } catch {
            throw APIError.transport(error)
        }

        guard let http = response as? HTTPURLResponse, (200..<300).contains(http.statusCode) else {
            let code = (response as? HTTPURLResponse)?.statusCode ?? -1
            throw APIError.server(code: code, msg: "注册失败")
        }

        struct RegisterResponse: Decodable {
            let message: String
        }
        do {
            return try JSONDecoder().decode(RegisterResponse.self, from: data).message
        } catch {
            throw APIError.decoding(error)
        }
    }
}
