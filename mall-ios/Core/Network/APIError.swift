import Foundation

/// APIClient 统一错误分类。
enum APIError: Error {
    /// 信封解析成功但 code != 0：服务端业务错误。
    case server(code: Int, msg: String)
    /// HTTP 401，或 requiresAuth == true 但本地无已保存 token。
    case unauthorized
    /// URLSession 层失败（无网络连接等）。
    case transport(Error)
    /// JSON 解码失败，或响应体结构与期望类型不匹配。
    case decoding(Error)
}
