import Foundation

enum APIError: Error {
    case invalidResponse
    case serverError(Int)
}

final class APIClient {
    static let shared = APIClient()

    private let baseURL = "http://localhost:8080"

    private init() {}

    // 示例骨架方法——后续真实功能替换此处实现
    func get<T: Decodable>(_ path: String, as type: T.Type) async throws -> T {
        guard let url = URL(string: baseURL + path) else {
            throw APIError.invalidResponse
        }
        let (data, response) = try await URLSession.shared.data(from: url)
        guard let http = response as? HTTPURLResponse, (200..<300).contains(http.statusCode) else {
            let code = (response as? HTTPURLResponse)?.statusCode ?? -1
            print("[APIClient] error: HTTP \(code) for \(path)")
            throw APIError.serverError(code)
        }
        return try JSONDecoder().decode(type, from: data)
    }
}
