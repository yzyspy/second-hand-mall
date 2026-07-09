import Foundation

/// 七牛云直传，流程与小程序 utils/qiniu-upload.ts 一致：
/// 1. POST /api/upload/qiniu-token 获取凭证（服务端生成唯一 uploadKey）
/// 2. multipart/form-data 直传七牛 uploadUrl，字段 key/token + 文件字段 file
/// 3. 最终图片地址 = domain/uploadKey
enum QiniuUploader {
    struct Token: Decodable {
        let uploadKey: String
        let upToken: String
        let domain: String
        let uploadUrl: String
    }

    enum Error: LocalizedError {
        case uploadFailed(statusCode: Int)

        var errorDescription: String? {
            switch self {
            case .uploadFailed(let code): return "图片上传失败（\(code)）"
            }
        }
    }

    private struct TokenRequest: Encodable {
        let key: String?
    }

    static func fetchToken() async throws -> Token {
        try await API.post("/api/upload/qiniu-token", body: TokenRequest(key: nil))
    }

    /// 上传一张 JPEG 图片，返回最终可访问的 URL
    static func upload(jpegData: Data) async throws -> String {
        let token = try await fetchToken()

        let boundary = "MallApp-\(UUID().uuidString)"
        var request = URLRequest(url: URL(string: token.uploadUrl)!)
        request.httpMethod = "POST"
        request.setValue("multipart/form-data; boundary=\(boundary)", forHTTPHeaderField: "Content-Type")

        var body = Data()
        func appendField(_ name: String, _ value: String) {
            body.append(Data("--\(boundary)\r\n".utf8))
            body.append(Data("Content-Disposition: form-data; name=\"\(name)\"\r\n\r\n".utf8))
            body.append(Data("\(value)\r\n".utf8))
        }
        appendField("key", token.uploadKey)
        appendField("token", token.upToken)

        body.append(Data("--\(boundary)\r\n".utf8))
        body.append(Data("Content-Disposition: form-data; name=\"file\"; filename=\"image.jpg\"\r\n".utf8))
        body.append(Data("Content-Type: image/jpeg\r\n\r\n".utf8))
        body.append(jpegData)
        body.append(Data("\r\n--\(boundary)--\r\n".utf8))

        let (_, response) = try await URLSession.shared.upload(for: request, from: body)
        guard let httpResponse = response as? HTTPURLResponse, httpResponse.statusCode == 200 else {
            throw Error.uploadFailed(statusCode: (response as? HTTPURLResponse)?.statusCode ?? -1)
        }

        return "\(token.domain)/\(token.uploadKey)"
    }
}
