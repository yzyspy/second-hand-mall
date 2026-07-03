import Foundation

/// mall-server 统一响应信封：{code, msg, data}。
/// code == 0 表示成功；data 为 nil 表示该接口本次响应无业务数据（见 EmptyResponse 用法）。
struct ApiResponse<T: Decodable>: Decodable {
    let code: Int
    let msg: String
    let data: T?
}

/// 标记类型：用于只需要 code == 0 语义成功、data 恒为 null 的接口。
/// 让调用方以 `T = EmptyResponse` 显式声明"无数据"，而不必把 T 声明为 Optional 多一层解包。
struct EmptyResponse: Decodable {}
