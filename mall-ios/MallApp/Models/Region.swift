import Foundation

/// 省市区三级数据，来源与小程序 data/china-regions.ts 相同（打包为 china-regions.json）
struct Province: Decodable, Identifiable, Hashable {
    let name: String
    let children: [City]
    var id: String { name }
}

struct City: Decodable, Identifiable, Hashable {
    let name: String
    let children: [String]
    var id: String { name }
}

enum ChinaRegions {
    static let provinces: [Province] = {
        guard let url = Bundle.main.url(forResource: "china-regions", withExtension: "json"),
              let data = try? Data(contentsOf: url),
              let provinces = try? JSONDecoder().decode([Province].self, from: data) else {
            assertionFailure("china-regions.json 缺失或解析失败")
            return []
        }
        return provinces
    }()
}
