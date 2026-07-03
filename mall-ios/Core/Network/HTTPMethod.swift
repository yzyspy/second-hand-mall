/// HTTP 方法集合。本 change 只用到 .get/.post；.put/.delete 一并定义，
/// 为后续 change（如 PUT /api/product/update）预留，避免回头修改本文件。
enum HTTPMethod: String {
    case get = "GET"
    case post = "POST"
    case put = "PUT"
    case delete = "DELETE"
}
